package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFindGitRepo_FromRepoRoot(t *testing.T) {
	dir, cleanup := createTempDirWithGit(t)
	defer cleanup()

	got, err := FindGitRepo(dir)
	if err != nil {
		t.Fatalf("FindGitRepo(%q): %v", dir, err)
	}
	if got != dir {
		t.Errorf("FindGitRepo(%q) = %q, want %q", dir, got, dir)
	}
}

func TestFindGitRepo_FromSubdir(t *testing.T) {
	dir, cleanup := createTempDirWithGit(t)
	defer cleanup()
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, err := FindGitRepo(sub)
	if err != nil {
		t.Fatalf("FindGitRepo(subdir): %v", err)
	}
	if got != dir {
		t.Errorf("FindGitRepo(subdir) = %q, want %q", got, dir)
	}
}

func TestFindGitRepo_NotFound(t *testing.T) {
	dir, err := os.MkdirTemp("", "squash-tree-repo-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)
	// No .git

	_, err = FindGitRepo(dir)
	if err == nil {
		t.Fatal("FindGitRepo: expected error when no .git")
	}
}

func TestResolveCommitHash(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available:", err)
	}
	dir, cleanup := initTempRepoWithCommit(t)
	defer cleanup()

	hash, err := ResolveCommitHash(dir, "HEAD")
	if err != nil {
		t.Fatalf("ResolveCommitHash: %v", err)
	}
	if len(hash) == 0 {
		t.Error("ResolveCommitHash: empty hash")
	}
}

func TestResolveCommitHash_InvalidRef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available:", err)
	}
	dir, cleanup := initTempRepoWithCommit(t)
	defer cleanup()

	_, err := ResolveCommitHash(dir, "nonexistent-ref-xyz")
	if err == nil {
		t.Fatal("ResolveCommitHash: expected error for invalid ref")
	}
}

func createTempDirWithGit(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "squash-tree-repo-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	cleanup := func() { os.RemoveAll(dir) }
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		cleanup()
		t.Fatalf("Mkdir .git: %v", err)
	}
	return dir, cleanup
}

func initTempRepoWithCommit(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "squash-tree-repo-test-*")
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

	// Configure git for tests (disable GPG signing)
	for _, args := range [][]string{
		{"git", "config", "user.email", "test@test"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Run()
	}

	// One commit so HEAD exists
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0644); err != nil {
		cleanup()
		t.Fatalf("WriteFile: %v", err)
	}
	cmd = exec.Command("git", "add", "f.txt")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@test", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@test")
	if out, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		t.Fatalf("git commit: %v %s", err, out)
	}
	return dir, cleanup
}
