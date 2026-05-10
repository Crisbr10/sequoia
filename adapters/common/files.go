package common

import (
	"os"
	"path/filepath"
)

// StageFile writes content to filepath.Join(dir, name), creating dir and any
// missing parent directories (mode 0o755). The file itself is written with
// mode 0o644.
func StageFile(dir, name string, content []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), content, 0o644)
}
