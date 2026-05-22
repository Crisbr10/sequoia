package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// StageFile writes content to filepath.Join(dir, name), creating dir and any
// missing parent directories (mode 0o750). The file itself is written with
// mode 0o644.
func StageFile(dir, name string, content []byte) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("stage %s: %w", name, err)
	}
	return os.WriteFile(filepath.Join(dir, name), content, 0o644)
}

// StageFileIfNotExist writes content to filepath.Join(dir, name) only if the
// target file does not already exist. Creates dir and any missing parent
// directories (mode 0o755). Returns nil if the file already exists (not an error).
func StageFileIfNotExist(dir, name string, content []byte) error {
	target := filepath.Join(dir, name)
	if _, err := os.Stat(target); err == nil {
		return nil // file exists, skip
	}
	return StageFile(dir, name, content)
}
