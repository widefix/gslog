package githooks

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed scripts/*
var scriptsFS embed.FS

// Scripts returns the embedded hook script contents, keyed by hook name (e.g. "pre-rebase").
func Scripts() (map[string]string, error) {
	out := make(map[string]string)
	err := fs.WalkDir(scriptsFS, "scripts", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(scriptsFS, path)
		if err != nil {
			return err
		}
		out[filepath.Base(path)] = string(data)
		return nil
	})
	return out, err
}

// WriteToDir writes all embedded hook scripts into dir, with executable bit set.
func WriteToDir(dir string) error {
	scripts, err := Scripts()
	if err != nil {
		return fmt.Errorf("read embedded hooks: %w", err)
	}
	for name, body := range scripts {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0755); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}
