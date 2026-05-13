package common

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	markerStart = "<!-- sequoia:start -->"
	markerEnd   = "<!-- sequoia:end -->"
)

// InjectMarkdownSection writes content into the Markdown file at path
// between <!-- sequoia:start --> and <!-- sequoia:end --> markers.
// If the file does not exist it is created with the section. If markers
// are already present the content between them is replaced. Otherwise
// the section is appended at the end of the file.
func InjectMarkdownSection(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	section := markerStart + "\n" + strings.TrimRight(content, "\n") + "\n" + markerEnd + "\n"

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(path, []byte(section), 0o644)
		}
		return err
	}

	s := string(raw)
	start := strings.Index(s, markerStart)
	end := strings.Index(s, markerEnd)

	if start != -1 && end != -1 {
		replaced := s[:start] + section + s[end+len(markerEnd):]
		// Trim a single trailing newline that WriteFile will re-add via section.
		replaced = strings.TrimRight(replaced, "\n") + "\n"
		return os.WriteFile(path, []byte(replaced), 0o644)
	}

	// Append: ensure exactly one blank line separator when existing content is non-empty.
	body := strings.TrimRight(s, "\n")
	var out string
	if body == "" {
		out = section
	} else {
		out = body + "\n\n" + section
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

// RemoveMarkdownSection deletes the content between <!-- sequoia:start -->
// and <!-- sequoia:end --> markers from the file at path.
// Returns nil when the file does not exist or contains no markers.
func RemoveMarkdownSection(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	s := string(raw)
	start := strings.Index(s, markerStart)
	end := strings.Index(s, markerEnd)
	if start == -1 || end == -1 {
		return nil
	}

	before := strings.TrimRight(s[:start], "\n")
	after := strings.TrimLeft(s[end+len(markerEnd):], "\n")

	var out string
	switch {
	case before == "" && after == "":
		out = ""
	case before == "":
		out = after
	case after == "":
		out = before + "\n"
	default:
		out = before + "\n\n" + after
	}

	return os.WriteFile(path, []byte(out), 0o644)
}

// ReplaceFile writes content to the file at path, creating a backup with a
// timestamped name at path+".sequoia-backup-<suffix>" if the file already
// exists and is not Sequoia-managed. A session-tracking file at
// path+".sequoia-session" records the backup suffix so that
// RestoreOrRemoveFile can locate the correct backup during uninstall.
// Creates parent directories if needed.
func ReplaceFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	managed, err := isSequoiaManaged(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return AtomicWriteFile(path, []byte(content), 0o644)
	}

	if managed {
		return AtomicWriteFile(path, []byte(content), 0o644)
	}

	// Generate a unique timestamp suffix to avoid name collisions.
	suffix := strconv.FormatInt(time.Now().UnixMilli(), 36)
	backup := path + ".sequoia-backup-" + suffix

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := AtomicWriteFile(backup, raw, 0o600); err != nil {
		return err
	}

	// Write a session file so RestoreOrRemoveFile can find the correct backup.
	if err := AtomicWriteFile(path+".sequoia-session", []byte(suffix), 0o644); err != nil {
		// Best-effort: if session file write fails, the backup exists but
		// RestoreOrRemoveFile will fall back to scanning for backups.
	}

	return AtomicWriteFile(path, []byte(content), 0o644)
}

// RestoreOrRemoveFile restores the original content from the session-tracked
// backup (path+".sequoia-backup-<suffix>") if a .sequoia-session file exists.
// If no session file is found, it falls back to the legacy predictable backup
// name (path+".sequoia-backup") for backwards compatibility.
// If the file is Sequoia-managed and has no backup, it deletes the file.
// If the file doesn't exist or is not managed and has no backup, returns nil.
func RestoreOrRemoveFile(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// Determine the backup to restore from.
	backup := findBackupPath(path)

	if backup != "" {
		raw, err := os.ReadFile(backup)
		if err != nil {
			return err
		}
		if err := AtomicWriteFile(path, raw, 0o644); err != nil {
			return err
		}
		// Clean up the backup file.
		_ = os.Remove(backup)
		// Clean up the session file if it exists.
		_ = os.Remove(path + ".sequoia-session")
		return nil
	}

	managed, err := isSequoiaManaged(path)
	if err != nil {
		return err
	}
	if managed {
		_ = os.Remove(path + ".sequoia-session")
		return os.Remove(path)
	}

	return nil
}

// findBackupPath returns the path of the backup to restore for the given file.
// It first checks for a .sequoia-session file with the backup suffix.
// If not found, it falls back to the legacy predictable backup name.
func findBackupPath(path string) string {
	// Try session-tracked backup first.
	sessionFile := path + ".sequoia-session"
	if data, err := os.ReadFile(sessionFile); err == nil {
		suffix := strings.TrimSpace(string(data))
		if suffix != "" {
			backup := path + ".sequoia-backup-" + suffix
			if _, err := os.Stat(backup); err == nil {
				return backup
			}
		}
	}

	// Fall back to legacy predictable backup name.
	legacyBackup := path + ".sequoia-backup"
	if _, err := os.Stat(legacyBackup); err == nil {
		return legacyBackup
	}

	return ""
}

// isSequoiaManaged reports whether the file at path contains the
// sequoia marker, indicating it was previously written by Sequoia.
func isSequoiaManaged(path string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(raw), markerStart), nil
}

// AtomicWriteFile writes data to path atomically using a temporary file and
// rename. On Windows this prevents truncated files on crash, where os.WriteFile
// truncates in place. The temporary file is cleaned up if the rename fails.
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
