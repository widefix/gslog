package git

import (
	"os/exec"
	"strings"
	"testing"
)

func TestFullHash(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	shortHash := makeCommit(t, repoPath, "test commit")

	fullHash, err := FullHash(repoPath, shortHash)
	if err != nil {
		t.Fatalf("FullHash: %v", err)
	}

	if len(fullHash) != 40 {
		t.Errorf("FullHash: expected 40 chars, got %d (%q)", len(fullHash), fullHash)
	}

	if !strings.HasPrefix(fullHash, shortHash) {
		t.Errorf("FullHash: %q is not a prefix of full hash %q", shortHash, fullHash)
	}

	headFull, err := FullHash(repoPath, "HEAD")
	if err != nil {
		t.Fatalf("FullHash(HEAD): %v", err)
	}
	if headFull != fullHash {
		t.Errorf("FullHash(HEAD)=%q, want %q", headFull, fullHash)
	}
}

func TestFullHash_InvalidRef(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	makeCommit(t, repoPath, "initial")

	_, err := FullHash(repoPath, "nonexistent-ref")
	if err == nil {
		t.Error("FullHash(nonexistent): expected error, got nil")
	}
}

func TestPreservationRefName(t *testing.T) {
	root := "a1b2c3d4e5f6789012345678901234567890abcd"
	child := "b2c3d4e5f6789012345678901234567890abcde"

	refName := PreservationRefName(root, child)
	expected := "refs/squash-archive/" + root + "/" + child

	if refName != expected {
		t.Errorf("PreservationRefName=%q, want %q", refName, expected)
	}
}

func TestCreatePreservationRefs(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	hash1 := makeCommit(t, repoPath, "commit 1")
	hash2 := makeCommitUnique(t, repoPath, "commit 2", "2")
	hash3 := makeCommitUnique(t, repoPath, "commit 3", "3")

	fullRoot, _ := FullHash(repoPath, hash1)
	fullChild1, _ := FullHash(repoPath, hash2)
	fullChild2, _ := FullHash(repoPath, hash3)

	err := CreatePreservationRefs(repoPath, fullRoot, []string{fullChild1, fullChild2})
	if err != nil {
		t.Fatalf("CreatePreservationRefs: %v", err)
	}

	ref1 := PreservationRefName(fullRoot, fullChild1)
	ref2 := PreservationRefName(fullRoot, fullChild2)

	for _, ref := range []string{ref1, ref2} {
		cmd := exec.Command("git", "show-ref", "--verify", ref)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Errorf("ref %s does not exist", ref)
		}
	}

	for _, tc := range []struct {
		ref      string
		expected string
	}{
		{ref1, fullChild1},
		{ref2, fullChild2},
	} {
		cmd := exec.Command("git", "rev-parse", tc.ref)
		cmd.Dir = repoPath
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("rev-parse %s: %v", tc.ref, err)
		}
		got := strings.TrimSpace(string(out))
		if got != tc.expected {
			t.Errorf("ref %s points to %q, want %q", tc.ref, got, tc.expected)
		}
	}
}

func TestCreatePreservationRefs_Idempotent(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	hash1 := makeCommit(t, repoPath, "commit 1")
	hash2 := makeCommitUnique(t, repoPath, "commit 2", "2")

	fullRoot, _ := FullHash(repoPath, hash1)
	fullChild, _ := FullHash(repoPath, hash2)

	if err := CreatePreservationRefs(repoPath, fullRoot, []string{fullChild}); err != nil {
		t.Fatalf("CreatePreservationRefs (1st): %v", err)
	}
	if err := CreatePreservationRefs(repoPath, fullRoot, []string{fullChild}); err != nil {
		t.Fatalf("CreatePreservationRefs (2nd): %v", err)
	}
}

func TestPreservationRefsExist(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	hash1 := makeCommit(t, repoPath, "commit 1")
	hash2 := makeCommitUnique(t, repoPath, "commit 2", "2")

	fullRoot, _ := FullHash(repoPath, hash1)
	fullChild, _ := FullHash(repoPath, hash2)

	exists, err := PreservationRefsExist(repoPath, fullRoot, []string{fullChild})
	if err != nil {
		t.Fatalf("PreservationRefsExist: %v", err)
	}
	if exists {
		t.Error("PreservationRefsExist: expected false before creation")
	}

	if err := CreatePreservationRefs(repoPath, fullRoot, []string{fullChild}); err != nil {
		t.Fatalf("CreatePreservationRefs: %v", err)
	}

	exists, err = PreservationRefsExist(repoPath, fullRoot, []string{fullChild})
	if err != nil {
		t.Fatalf("PreservationRefsExist: %v", err)
	}
	if !exists {
		t.Error("PreservationRefsExist: expected true after creation")
	}
}

func TestWriteMetadata_CreatesPreservationRefs(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	makeCommit(t, repoPath, "base")
	child1Short := makeCommitUnique(t, repoPath, "child1", "c1")
	child2Short := makeCommitUnique(t, repoPath, "child2", "c2")
	rootShort := makeCommitUnique(t, repoPath, "squash", "sq")

	baseShort := child1Short

	children := []string{child1Short, child2Short}
	err := WriteMetadata(repoPath, rootShort, baseShort, children, "test")
	if err != nil {
		t.Fatalf("WriteMetadata: %v", err)
	}

	rootFull, _ := FullHash(repoPath, rootShort)
	child1Full, _ := FullHash(repoPath, child1Short)
	child2Full, _ := FullHash(repoPath, child2Short)

	exists, err := PreservationRefsExist(repoPath, rootFull, []string{child1Full, child2Full})
	if err != nil {
		t.Fatalf("PreservationRefsExist: %v", err)
	}
	if !exists {
		t.Error("WriteMetadata did not create preservation refs")
	}
}

func makeCommitUnique(t *testing.T, repoPath, msg, uniqueContent string) string {
	t.Helper()

	cmd := exec.Command("git", "config", "commit.gpgsign", "false")
	cmd.Dir = repoPath
	cmd.Run()

	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoPath
	out, _ := cmd.Output()
	currentHead := strings.TrimSpace(string(out))

	f := repoPath + "/f.txt"
	cmd = exec.Command("bash", "-c", "echo '"+uniqueContent+"' >> "+f)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("append to file: %v", err)
	}

	cmd = exec.Command("git", "add", "f.txt")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v %s", err, out)
	}

	cmd = exec.Command("git", "commit", "-m", msg)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(out), "nothing to commit") {
			return currentHead
		}
		t.Fatalf("git commit: %v %s", err, out)
	}

	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}
