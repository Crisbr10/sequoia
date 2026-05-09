package opencode

import (
	"os"
	"path/filepath"
)

// opencodeBase returns the ~/.config/opencode/ directory.
// If homeDir is non-empty it is used directly; otherwise os.UserHomeDir() is called.
// Symlinks in homeDir are resolved via filepath.EvalSymlinks before joining.
// On resolution failure, the unresolved path is used as a fallback.
func opencodeBase(homeDir string) (string, error) {
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
	return filepath.Join(resolved, ".config", "opencode"), nil
}

// skillsPath returns ~/.config/opencode/skills/sequoia/
func skillsPath(base string) string { return filepath.Join(base, "skills", "sequoia") }

// skillFilePath returns ~/.config/opencode/skills/sequoia/SKILL.md
func skillFilePath(base string) string { return filepath.Join(skillsPath(base), "SKILL.md") }

// commandsPath returns ~/.config/opencode/commands/
func commandsPath(base string) string { return filepath.Join(base, "commands") }

// systemPromptPath returns ~/.config/opencode/AGENTS.md
func systemPromptPath(base string) string { return filepath.Join(base, "AGENTS.md") }

// versionFilePath returns the path to the .sequoia-version marker file
// inside the skills directory.
func versionFilePath(base string) string {
	return filepath.Join(skillsPath(base), ".sequoia-version")
}

// backupPath returns the temp backup dir used for rollback: ~/.config/opencode/.sequoia-backup/
func backupPath(base string) string { return filepath.Join(base, ".sequoia-backup") }
