package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"squash-tree/internal/git"
	"squash-tree/internal/tree"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	sub := os.Args[1]
	switch sub {
	case "init":
		runInit(os.Args[2:])
	case "add-metadata":
		runAddMetadata(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		runShowTree(sub)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: git squash-tree <commit>       Show squash tree for a commit\n")
	fmt.Fprintf(os.Stderr, "       git squash-tree init [--global] Install hooks in repo (or globally)\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  git squash-tree HEAD\n")
	fmt.Fprintf(os.Stderr, "  git squash-tree init\n")
}

func runShowTree(commitRef string) {
	repoPath, err := findGitRepo(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repository or .git not found: %v\n", err)
		os.Exit(1)
	}

	notesReader := git.NewNotesReader(repoPath)
	commitHash, err := resolveCommitHash(repoPath, commitRef)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to resolve commit reference '%s': %v\n", commitRef, err)
		os.Exit(1)
	}

	builder := tree.NewBuilder(notesReader)
	rootNode, err := builder.BuildTree(commitHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building squash tree: %v\n", err)
		os.Exit(1)
	}

	visualizer := tree.NewVisualizer()
	fmt.Print(visualizer.Visualize(rootNode))
}

func runAddMetadata(args []string) {
	fs := flag.NewFlagSet("add-metadata", flag.ExitOnError)
	root := fs.String("root", "", "Squash commit (root) hash or ref")
	base := fs.String("base", "", "Base commit hash or ref")
	children := fs.String("children", "", "Comma-separated child commit hashes (order preserved)")
	strategy := fs.String("strategy", "auto", "Strategy: auto or manual")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if *root == "" || *base == "" || *children == "" {
		fmt.Fprintf(os.Stderr, "add-metadata requires --root, --base, and --children\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	repoPath, err := findGitRepo(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repository: %v\n", err)
		os.Exit(1)
	}

	notesReader := git.NewNotesReader(repoPath)
	if notesReader.HasMetadata(*root) {
		return
	}

	rootShort, err := resolveCommitHash(repoPath, *root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid root: %v\n", err)
		os.Exit(1)
	}
	baseShort, err := resolveCommitHash(repoPath, *base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid base: %v\n", err)
		os.Exit(1)
	}
	childList := strings.Split(*children, ",")
	var childrenShort []string
	for _, c := range childList {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		short, err := resolveCommitHash(repoPath, c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid child %q: %v\n", c, err)
			os.Exit(1)
		}
		childrenShort = append(childrenShort, short)
	}
	if len(childrenShort) == 0 {
		fmt.Fprintf(os.Stderr, "Error: at least one child required\n")
		os.Exit(1)
	}

	if err := git.WriteMetadata(repoPath, rootShort, baseShort, childrenShort, *strategy); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing metadata: %v\n", err)
		os.Exit(1)
	}
}

func runInit(args []string) {
	global := false
	for _, a := range args {
		if a == "--global" {
			global = true
			break
		}
	}

	if global {
		runInitGlobal()
		return
	}

	repoPath, err := findGitRepo(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not a git repository: %v\n", err)
		os.Exit(1)
	}
	hooksDir := filepath.Join(repoPath, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := writeHooks(hooksDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing hooks: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Git hooks installed. Squash metadata will be recorded automatically.")
}

func runInitGlobal() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	hooksDir := filepath.Join(home, ".config", "git", "squash-tree-hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := writeHooks(hooksDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing hooks: %v\n", err)
		os.Exit(1)
	}
	cmd := exec.Command("git", "config", "--global", "core.hooksPath", hooksDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting core.hooksPath: %v\n%s", err, string(out))
		os.Exit(1)
	}
	fmt.Printf("Global hooks installed at %s. All repos will use squash-tree hooks.\n", hooksDir)
}

func writeHooks(hooksDir string) error {
	for name, body := range hookScripts() {
		p := filepath.Join(hooksDir, name)
		if err := os.WriteFile(p, []byte(body), 0755); err != nil {
			return err
		}
	}
	return nil
}

func hookScripts() map[string]string {
	return map[string]string{
		"pre-rebase": preRebaseHook,
		"post-rewrite": postRewriteHook,
		"post-merge":  postMergeHook,
		"prepare-commit-msg": prepareCommitMsgHook,
	}
}

const preRebaseHook = `#!/bin/bash
if [ -n "$2" ] && [ "$2" != "" ]; then
    UPSTREAM="$2"
    if [ -n "$3" ]; then
        git rev-list "$UPSTREAM..$3" > .git/SQUASH_PRE_REBASE_COMMITS 2>/dev/null || true
    else
        git rev-list "$UPSTREAM..HEAD" > .git/SQUASH_PRE_REBASE_COMMITS 2>/dev/null || true
    fi
    echo "$UPSTREAM" > .git/SQUASH_PRE_REBASE_BASE 2>/dev/null || true
fi
exit 0
`

const postRewriteHook = `#!/bin/bash
if [ "$1" != "rebase" ] && [ ! -f .git/rebase-merge ] && [ ! -f .git/rebase-apply ]; then
    exit 0
fi
if [ -f .git/SQUASH_PRE_REBASE_COMMITS ] && [ -f .git/SQUASH_PRE_REBASE_BASE ]; then
    BASE=$(cat .git/SQUASH_PRE_REBASE_BASE)
    OLD_COMMITS=($(cat .git/SQUASH_PRE_REBASE_COMMITS))
    while read old_sha new_sha extra; do
        if [ "$old_sha" != "$new_sha" ] && [ -n "$new_sha" ]; then
            SQUASHED=()
            for old in "${OLD_COMMITS[@]}"; do
                if git rev-parse "$old" &>/dev/null; then
                    git merge-base --is-ancestor "$old" "$new_sha" 2>/dev/null && SQUASHED+=("$old")
                else
                    SQUASHED+=("$old")
                fi
            done
            if [ ${#SQUASHED[@]} -gt 1 ]; then
                CHILDREN=$(IFS=,; echo "${SQUASHED[*]}")
                git squash-tree add-metadata --root="$new_sha" --base="$BASE" --children="$CHILDREN" --strategy=auto 2>/dev/null || true
            fi
        fi
    done
    rm -f .git/SQUASH_PRE_REBASE_COMMITS .git/SQUASH_PRE_REBASE_BASE
else
    while read old_sha new_sha extra; do
        if [ "$old_sha" != "$new_sha" ] && [ -n "$new_sha" ]; then
            BASE=$(git merge-base "$old_sha" "$new_sha" 2>/dev/null || git rev-parse "$new_sha^" 2>/dev/null || echo "")
            if [ -n "$BASE" ]; then
                CHILDREN=$(git rev-list --reverse "$BASE..$old_sha" 2>/dev/null | tr '\n' ',')
                CHILDREN="${CHILDREN%,}"
                if [ -n "$CHILDREN" ] && [ $(echo "$CHILDREN" | tr ',' '\n' | wc -l) -gt 1 ]; then
                    git squash-tree add-metadata --root="$new_sha" --base="$BASE" --children="$CHILDREN" --strategy=auto 2>/dev/null || true
                fi
            fi
        fi
    done
fi
exit 0
`

const postMergeHook = `#!/bin/bash
if [ ! -f .git/SQUASH_HEAD ]; then
    exit 0
fi
MERGE_HEAD=$(cat .git/MERGE_HEAD 2>/dev/null)
SQUASH_HEAD=$(cat .git/SQUASH_HEAD 2>/dev/null)
CURRENT_HEAD=$(git rev-parse HEAD)
if [ -n "$MERGE_HEAD" ] && [ -n "$SQUASH_HEAD" ]; then
    BASE=$(git merge-base "$CURRENT_HEAD" "$MERGE_HEAD" 2>/dev/null || git rev-parse "$CURRENT_HEAD^" 2>/dev/null || echo "")
    if [ -n "$BASE" ]; then
        COMMITS=$(git rev-list --reverse "$BASE..$MERGE_HEAD" 2>/dev/null | tr '\n' ',')
        COMMITS="${COMMITS%,}"
        if [ -n "$COMMITS" ]; then
            git squash-tree add-metadata --root="$CURRENT_HEAD" --base="$BASE" --children="$COMMITS" --strategy=auto 2>/dev/null || true
        fi
    fi
fi
rm -f .git/SQUASH_HEAD
exit 0
`

const prepareCommitMsgHook = `#!/bin/bash
if [ "$2" = "squash" ] || [ "$2" = "merge" ]; then
    touch .git/SQUASH_IN_PROGRESS
    if [ -f .git/rebase-merge/stopped-sha ]; then
        cat .git/rebase-merge/stopped-sha >> .git/SQUASH_COMMITS_LIST 2>/dev/null || true
    fi
fi
exit 0
`

func findGitRepo(startPath string) (string, error) {
	path, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}
	for {
		gitPath := filepath.Join(path, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return path, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return "", fmt.Errorf("not a git repository")
}

func resolveCommitHash(repoPath, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", ref)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
