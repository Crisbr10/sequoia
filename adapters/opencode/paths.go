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

// backupPath returns the legacy per-tool temp backup dir used by the
// safety-net fallback in common.BackupPathBuilder.Build.
//
// PR 3 moved the main backup flow to the central backup home
// (<UserConfigDir>/sequoia/backups/<adapterID>/<session>/) so this
// function is no longer consulted on the happy path. It is retained
// for backwards compatibility with any external callers that may
// reference it, and as a documentary anchor for the legacy
// per-tool layout.
//
// The PR 3 safety-net in BackupPathBuilder.Build uses a hard-coded
// `<base>/.sequoia-backup/<adapterID>/<suffix>` shape and does not
// call this function — see backup_path_builder.go for details.
func backupPath(base string) string { return filepath.Join(base, ".sequoia-backup") }
