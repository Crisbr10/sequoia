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
