package gemini

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = "<!-- sequoia:start -->"
	markerEnd   = "<!-- sequoia:end -->"
)

// InjectSection writes the Sequoia section into the file at path.
// If the file does not exist it is created. If markers are already present
// the content between them is replaced. Otherwise the section is appended.
func InjectSection(path, content string) error {
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

// RemoveSection deletes the Sequoia section from the file at path.
// Returns nil when the file does not exist or contains no markers.
func RemoveSection(path string) error {
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
