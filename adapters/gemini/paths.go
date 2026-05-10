package gemini

import (
	"os"
	"path/filepath"
)

// geminiBase returns the ~/.gemini/ directory.
// If homeDir is non-empty it is used directly; otherwise os.UserHomeDir() is called.
// Symlinks in homeDir are resolved via filepath.EvalSymlinks before joining.
// On resolution failure, the unresolved path is used as a fallback.
func geminiBase(homeDir string) (string, error) {
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	resolved, err := filepath.EvalSymlinks(homeDir)
	if err != nil {
		resolved = homeDir
	}
	return filepath.Join(resolved, ".gemini"), nil
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
