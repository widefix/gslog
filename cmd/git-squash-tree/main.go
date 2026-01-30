package main

import (
	"fmt"
	"os"
	"strings"

	"squash-tree/internal/git"
	"squash-tree/internal/metadata"
	"squash-tree/internal/repo"
	"squash-tree/internal/tree"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	sub := os.Args[1]
	switch sub {
	case "add-metadata":
		if err := runAddMetadata(os.Args[2:]); err != nil {
			fatal(err)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		if err := runShowTree(sub); err != nil {
			fatal(err)
		}
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: git squash-tree <commit>       Show squash tree for a commit\n")
	fmt.Fprintf(os.Stderr, "       git squash-tree add-metadata --root=<ref> --base=<ref> --children=<refs>\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  git squash-tree HEAD\n")
	fmt.Fprintf(os.Stderr, "  git squash-tree add-metadata --root=HEAD --base=main --children=a1b2c3,d4e5f6\n")
}

func runShowTree(commitRef string) error {
	repoPath, err := repo.FindGitRepo(".")
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	commitHash, err := repo.ResolveCommitHash(repoPath, commitRef)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", commitRef, err)
	}

	notesReader := git.NewNotesReader(repoPath)
	builder := tree.NewBuilder(notesReader)
	rootNode, err := builder.BuildTree(commitHash)
	if err != nil {
		return fmt.Errorf("build tree: %w", err)
	}

	fmt.Print(tree.NewVisualizer().Visualize(rootNode))
	return nil
}

func runAddMetadata(args []string) error {
	opts, err := metadata.ParseAddMetadataFlags(args)
	if err != nil {
		return err
	}

	repoPath, err := repo.FindGitRepo(".")
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	notesReader := git.NewNotesReader(repoPath)
	if notesReader.HasMetadata(opts.RootRef) {
		return nil
	}

	rootShort, err := repo.ResolveCommitHash(repoPath, opts.RootRef)
	if err != nil {
		return fmt.Errorf("invalid root: %w", err)
	}
	baseShort, err := repo.ResolveCommitHash(repoPath, opts.BaseRef)
	if err != nil {
		return fmt.Errorf("invalid base: %w", err)
	}
	childrenShort, err := repo.ResolveRefs(repoPath, splitTrim(opts.ChildrenRefs, ","))
	if err != nil {
		return fmt.Errorf("children: %w", err)
	}

	if err := git.WriteMetadata(repoPath, rootShort, baseShort, childrenShort, opts.Strategy); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}
	return nil
}

func splitTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
