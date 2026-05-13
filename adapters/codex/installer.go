package codex

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MergeConfig reads the config.toml at path, merges the [sequoia] table
// with the provided table data, and writes the result back. A backup of
// the original file is saved at path+".sequoia-backup-<suffix>" with a
// unique timestamp suffix to avoid name collisions. A session-tracking
// file at path+".sequoia-session" records the suffix so RemoveConfig can
// locate the correct backup during uninstall.
func MergeConfig(path string, table map[string]interface{}) error {
	var existing string
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("merge config: read: %w", err)
	}

	// Create a timestamped backup if the file exists and has content.
	if existing != "" {
		suffix := strconv.FormatInt(time.Now().UnixMilli(), 36)
		backupPath := path + ".sequoia-backup-" + suffix
		if err := os.WriteFile(backupPath, []byte(existing), 0o600); err != nil {
			return fmt.Errorf("merge config: backup: %w", err)
		}
		// Write a session file so RemoveConfig can find the correct backup.
		if err := os.WriteFile(path+".sequoia-session", []byte(suffix), 0o644); err != nil {
			// Best-effort: if session file write fails, the backup exists
			// but RemoveConfig will fall back to scanning for backups.
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
// If a session-tracked backup exists (path+".sequoia-session"), the original
// content is restored from that backup and both the backup and session files
// are cleaned up. Falls back to the legacy backup name for backwards
// compatibility. If the file does not exist, RemoveConfig returns nil.
func RemoveConfig(path string) error {
	// Try session-tracked backup first.
	sessionFile := path + ".sequoia-session"
	if sessionData, err := os.ReadFile(sessionFile); err == nil {
		suffix := strings.TrimSpace(string(sessionData))
		backupPath := path + ".sequoia-backup-" + suffix
		if backupData, err := os.ReadFile(backupPath); err == nil {
			if err := os.WriteFile(path, backupData, 0o644); err != nil {
				return fmt.Errorf("remove config: restore backup: %w", err)
			}
			_ = os.Remove(backupPath)
			_ = os.Remove(sessionFile)
			return nil
		}
		// Session file exists but backup doesn't — clean up the stale session file.
		_ = os.Remove(sessionFile)
	}

	// Fall back to legacy predictable backup name.
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
