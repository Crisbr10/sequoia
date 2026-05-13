package opencode

import (
	"path/filepath"
)

// opencodeBase returns the ~/.config/opencode/ directory.
// BaseAdapter.base() handles home directory resolution and symlink detection.
func opencodeBase(homeDir string) (string, error) {
	return filepath.Join(homeDir, ".config", "opencode"), nil
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
