package cursor

import (
	"os"
	"path/filepath"
)

// cursorBase returns the ~/.cursor/rules/ directory.
// If homeDir is non-empty it is used directly; otherwise os.UserHomeDir() is called.
// Symlinks in homeDir are resolved via filepath.EvalSymlinks before joining.
// On resolution failure, the unresolved path is used as a fallback.
func cursorBase(homeDir string) (string, error) {
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	resolved, err := filepath.EvalSymlinks(homeDir)
	if err != nil {
		// Fall back to unresolved path on any error.
		resolved = homeDir
	}
	return filepath.Join(resolved, ".cursor", "rules"), nil
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
