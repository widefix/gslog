package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// requireGit skips the test if git is not available.
func requireGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available:", err)
	}
}

func TestWriteMetadata_ReadMetadata_RoundTrip(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	root := "abc1234"
	base := "def5678"
	children := []string{"c1", "c2"}
	strategy := "rebase"

	// We need real commit hashes. Create a commit and use its hash.
	hash := makeCommit(t, repoPath, "initial")
	root = hash
	base = hash
	children = []string{hash}

	err := WriteMetadata(repoPath, root, base, children, strategy)
	if err != nil {
		t.Fatalf("WriteMetadata: %v", err)
	}

	nr := NewNotesReader(repoPath)
	if !nr.HasMetadata(root) {
		t.Fatal("HasMetadata: expected true after WriteMetadata")
	}
	meta, err := nr.ReadMetadata(root)
	if err != nil {
		t.Fatalf("ReadMetadata: %v", err)
	}
	if meta.Root != root || meta.Base != base {
		t.Errorf("Root=%q Base=%q, want Root=%q Base=%q", meta.Root, meta.Base, root, base)
	}
	if len(meta.Children) != 1 || meta.Children[0].Hash != hash {
		t.Errorf("Children: %+v", meta.Children)
	}
}

func TestNotesReader_CommitExists(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	hash := makeCommit(t, repoPath, "only commit")
	nr := NewNotesReader(repoPath)

	if !nr.CommitExists(hash) {
		t.Error("CommitExists(existing): got false")
	}
	if nr.CommitExists("nonexistent00000000000000000000") {
		t.Error("CommitExists(nonexistent): got true")
	}
}

func TestNotesReader_GetShortHash(t *testing.T) {
	requireGit(t)
	repoPath, cleanup := initTempRepo(t)
	defer cleanup()

	fullHash := makeCommit(t, repoPath, "commit")
	nr := NewNotesReader(repoPath)

	short, err := nr.getShortHash(fullHash)
	if err != nil {
		t.Fatalf("getShortHash: %v", err)
	}
	if len(short) == 0 {
		t.Error("getShortHash: empty result")
	}
	// Short hash should be prefix of full or at least valid hex
	if len(short) > len(fullHash) {
		t.Errorf("getShortHash: %q longer than full %q", short, fullHash)
	}
}

func initTempRepo(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "squash-tree-notes-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	cleanup := func() { os.RemoveAll(dir) }

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		t.Fatalf("git init: %v %s", err, out)
	}
	return dir, cleanup
}

func makeCommit(t *testing.T, repoPath, msg string) string {
	t.Helper()
	// Set user so commit succeeds
	for _, args := range [][]string{{"git", "config", "user.email", "test@test"}, {"git", "config", "user.name", "Test"}} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoPath
		cmd.Run()
	}

	f := filepath.Join(repoPath, "f.txt")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	cmd := exec.Command("git", "add", "f.txt")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v %s", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", msg)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@test", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@test")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return string(out)[:len(out)-1] // trim newline
}
