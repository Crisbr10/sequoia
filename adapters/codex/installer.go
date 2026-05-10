package codex

import (
	"fmt"
	"os"
	"path/filepath"
)

// MergeConfig reads the config.toml at path, merges the [sequoia] table
// with the provided table data, and writes the result back. A backup of
// the original file is saved at path + ".sequoia-backup" if the file exists.
func MergeConfig(path string, table map[string]interface{}) error {
	var existing string
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("merge config: read: %w", err)
	}

	// Create backup if the file exists and has content.
	if existing != "" {
		backupPath := path + ".sequoia-backup"
		if err := os.WriteFile(backupPath, []byte(existing), 0o644); err != nil {
			return fmt.Errorf("merge config: backup: %w", err)
		}
	}

	result, err := MergeSection(existing, table)
	if err != nil {
		return fmt.Errorf("merge config: %w", err)
	}

	// Ensure the parent directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("merge config: mkdir: %w", err)
	}

	return os.WriteFile(path, []byte(result), 0o644)
}

// RemoveConfig removes the [sequoia] table from the config.toml at path.
// If a backup exists at path + ".sequoia-backup", the original content
// is restored and the backup is deleted. If the file does not exist,
// RemoveConfig returns nil.
func RemoveConfig(path string) error {
	// If a backup exists, restore it and remove the backup.
	backupPath := path + ".sequoia-backup"
	if backupData, err := os.ReadFile(backupPath); err == nil {
		if err := os.WriteFile(path, backupData, 0o644); err != nil {
			return fmt.Errorf("remove config: restore backup: %w", err)
		}
		return os.Remove(backupPath)
	}

	// No backup — parse and remove [sequoia] section.
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("remove config: read: %w", err)
	}

	result, err := RemoveSection(string(existing))
	if err != nil {
		return fmt.Errorf("remove config: %w", err)
	}

	return os.WriteFile(path, []byte(result), 0o644)
}
