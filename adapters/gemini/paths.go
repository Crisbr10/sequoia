package gemini

import (
	"path/filepath"
)

// geminiBase returns the ~/.gemini/ directory.
// BaseAdapter.base() handles home directory resolution and symlink detection.
func geminiBase(homeDir string) (string, error) {
	return filepath.Join(homeDir, ".gemini"), nil
}

// skillsPath returns ~/.gemini/sequoia/skills/
func skillsPath(base string) string { return filepath.Join(base, "sequoia", "skills") }

// commandsPath returns ~/.gemini/sequoia/commands/
func commandsPath(base string) string { return filepath.Join(base, "sequoia", "commands") }

// systemPromptPath returns ~/.gemini/GEMINI.md
func systemPromptPath(base string) string { return filepath.Join(base, "GEMINI.md") }

// versionFilePath returns the path to the .sequoia-version marker file
// inside the Sequoia install directory.
func versionFilePath(base string) string {
	return filepath.Join(base, "sequoia", ".sequoia-version")
}

// backupPath returns the temp backup dir used for rollback: ~/.gemini/.sequoia-backup/
func backupPath(base string) string { return filepath.Join(base, ".sequoia-backup") }
