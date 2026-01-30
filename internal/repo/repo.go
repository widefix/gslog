package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func FindGitRepo(startPath string) (string, error) {
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

func ResolveCommitHash(repoPath, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", ref)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func ResolveRefs(repoPath string, refs []string) ([]string, error) {
	var hashes []string
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		short, err := ResolveCommitHash(repoPath, ref)
		if err != nil {
			return nil, fmt.Errorf("invalid ref %q: %w", ref, err)
		}
		hashes = append(hashes, short)
	}
	if len(hashes) == 0 {
		return nil, fmt.Errorf("at least one ref required")
	}
	return hashes, nil
}
