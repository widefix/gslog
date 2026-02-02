package githooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteToDir_CreatesExpectedFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "githooks-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	if err := WriteToDir(dir); err != nil {
		t.Fatalf("WriteToDir: %v", err)
	}

	expected := []string{"pre-rebase", "post-rewrite", "post-merge", "prepare-commit-msg"}
	for _, name := range expected {
		p := filepath.Join(dir, name)
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("hook %q: %v", name, err)
			continue
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("hook %q: not executable (mode %o)", name, info.Mode())
		}
	}
}

func TestScripts_ContentContainsAddMetadata(t *testing.T) {
	scripts, err := Scripts()
	if err != nil {
		t.Fatalf("Scripts: %v", err)
	}
	if !strings.Contains(scripts["post-rewrite"], "add-metadata") {
		t.Error("post-rewrite should contain add-metadata")
	}
	if !strings.Contains(scripts["post-merge"], "add-metadata") {
		t.Error("post-merge should contain add-metadata")
	}
}
