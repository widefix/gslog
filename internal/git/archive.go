package git

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	ArchiveRefPrefix = "refs/squash-archive/"
)

func PreservationRefName(rootFullSHA, childFullSHA string) string {
	return ArchiveRefPrefix + rootFullSHA + "/" + childFullSHA
}

func FullHash(repoPath, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref)
	if repoPath != "" {
		cmd.Dir = repoPath
	}
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse %s: %w", ref, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func CreatePreservationRefs(repoPath, rootFullSHA string, childFullSHAs []string) error {
	for _, child := range childFullSHAs {
		refName := PreservationRefName(rootFullSHA, child)
		cmd := exec.Command("git", "update-ref", refName, child)
		if repoPath != "" {
			cmd.Dir = repoPath
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git update-ref %s %s: %w: %s", refName, child, err, string(out))
		}
	}
	return nil
}

func PreservationRefsExist(repoPath, rootFullSHA string, childFullSHAs []string) (bool, error) {
	for _, child := range childFullSHAs {
		refName := PreservationRefName(rootFullSHA, child)
		cmd := exec.Command("git", "show-ref", "--verify", "--quiet", refName)
		if repoPath != "" {
			cmd.Dir = repoPath
		}
		if err := cmd.Run(); err != nil {
			return false, nil
		}
	}
	return true, nil
}
