package common

import (
	"os"
	"path/filepath"
	"strings"
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

// ReplaceFile writes content to the file at path, creating a backup at
// path+".sequoia-backup" if the file already exists and is not Sequoia-managed
// (does not contain markers). If the file is already Sequoia-managed, it is
// replaced in place without backup. Creates parent directories if needed.
func ReplaceFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	managed, err := isSequoiaManaged(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return os.WriteFile(path, []byte(content), 0o644)
	}

	if managed {
		return os.WriteFile(path, []byte(content), 0o644)
	}

	backup := path + ".sequoia-backup"
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backup, raw, 0o644); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// RestoreOrRemoveFile restores the original content from the backup at
// path+".sequoia-backup" if it exists, or deletes the file if it is
// Sequoia-managed. If the file doesn't exist or is not managed and has
// no backup, returns nil.
func RestoreOrRemoveFile(path string) error {
	backup := path + ".sequoia-backup"

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if _, berr := os.Stat(backup); berr == nil {
		raw, err := os.ReadFile(backup)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			return err
		}
		return os.Remove(backup)
	}

	managed, err := isSequoiaManaged(path)
	if err != nil {
		return err
	}
	if managed {
		return os.Remove(path)
	}

	return nil
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
