package codex

import (
	"os"
	"path/filepath"
)

// codexBase returns the ~/.codex/ directory.
// If homeDir is non-empty it is used directly; otherwise os.UserHomeDir() is called.
// Symlinks in homeDir are resolved via filepath.EvalSymlinks before joining.
// On resolution failure, the unresolved path is used as a fallback.
func codexBase(homeDir string) (string, error) {
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
	return filepath.Join(resolved, ".codex"), nil
}

// skillsPath returns ~/.codex/sequoia/skills/
func skillsPath(base string) string { return filepath.Join(base, "sequoia", "skills") }

// commandsPath returns ~/.codex/sequoia/commands/
func commandsPath(base string) string { return filepath.Join(base, "sequoia", "commands") }

// systemPromptPath returns ~/.codex/config.toml where the [sequoia] TOML table is merged.
func systemPromptPath(base string) string { return filepath.Join(base, "config.toml") }

// configPath is an alias for systemPromptPath returning ~/.codex/config.toml.
func configPath(base string) string { return systemPromptPath(base) }

// versionFilePath returns the path to the .sequoia-version marker file
// inside the skills directory.
func versionFilePath(base string) string {
	return filepath.Join(skillsPath(base), ".sequoia-version")
}

// backupPath returns the temp backup dir used for rollback: ~/.codex/.sequoia-backup/
func backupPath(base string) string { return filepath.Join(base, ".sequoia-backup") }
