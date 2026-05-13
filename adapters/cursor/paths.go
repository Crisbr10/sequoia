package cursor

import (
	"path/filepath"
)

// cursorBase returns the ~/.cursor/rules/ directory.
// BaseAdapter.base() handles home directory resolution and symlink detection.
func cursorBase(homeDir string) (string, error) {
	return filepath.Join(homeDir, ".cursor", "rules"), nil
}

// skillsPath returns the root .cursor/rules/ directory where skills are stored.
func skillsPath(base string) string { return base }

// commandsPath returns the root .cursor/rules/ directory where commands are stored.
func commandsPath(base string) string { return base }

// systemPromptPath returns ~/.cursor/rules/sequoia-ai.md
func systemPromptPath(base string) string {
	return filepath.Join(base, "sequoia-ai.md")
}

// versionFilePath returns the path to the .sequoia-version marker file
// inside the rules directory.
func versionFilePath(base string) string {
	return filepath.Join(base, ".sequoia-version")
}

// backupPath returns the temp backup dir used for rollback: ~/.cursor/rules/.sequoia-backup/
func backupPath(base string) string { return filepath.Join(base, ".sequoia-backup") }
