package claude

import (
	"path/filepath"
)

// claudeBase returns the ~/.claude/ directory.
// BaseAdapter.base() handles home directory resolution and symlink detection.
func claudeBase(homeDir string) (string, error) {
	return filepath.Join(homeDir, ".claude"), nil
}

// skillsPath returns ~/.claude/skills/sequoia/
func skillsPath(base string) string { return filepath.Join(base, "skills", "sequoia") }

// skillFilePath returns ~/.claude/skills/sequoia/SKILL.md
func skillFilePath(base string) string { return filepath.Join(skillsPath(base), "SKILL.md") }

// commandsPath returns ~/.claude/commands/
func commandsPath(base string) string { return filepath.Join(base, "commands") }

// systemPromptPath returns ~/.claude/CLAUDE.md
func systemPromptPath(base string) string { return filepath.Join(base, "CLAUDE.md") }

// versionFilePath returns the path to the .sequoia-version marker file
// inside the skills directory.
func versionFilePath(base string) string {
	return filepath.Join(skillsPath(base), ".sequoia-version")
}

// backupPath returns the temp backup dir used for rollback: ~/.claude/.sequoia-backup/
func backupPath(base string) string { return filepath.Join(base, ".sequoia-backup") }
