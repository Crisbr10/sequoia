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
