package template

import (
	"os"
	"path/filepath"
)

// toolBase returns the root configuration directory for this tool.
// If homeDir is non-empty it is used directly; otherwise os.UserHomeDir() is called.
// Symlinks in homeDir are resolved via filepath.EvalSymlinks before joining.
// On resolution failure, the unresolved path is used as a fallback.
//
// TODO: Replace ".my-tool/config" with the actual config directory path.
//
//	Examples:
//	  Claude Code: ".claude"
//	  OpenCode:    ".config/opencode"
//	  Cursor IDE:  ".cursor/rules"
//	  Gemini CLI:  ".gemini"
//	  Codex:       ".codex"
func toolBase(homeDir string) (string, error) {
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
	// TODO: Replace ".my-tool/config" with the tool's config directory name.
	return filepath.Join(resolved, ".my-tool", "config"), nil
}

// skillsPath returns the directory where skill files are stored.
//
// TODO: Adjust if skills and commands have different directories.
// For most tools they share the same root.
func skillsPath(base string) string { return base }

// commandsPath returns the directory where command files are stored.
//
// TODO: Adjust if commands have a different directory than skills.
func commandsPath(base string) string { return base }

// systemPromptPath returns the path to the system prompt / rules file.
//
// TODO: Replace "sequoia-ai.md" with the correct filename for this tool.
//
//	Examples:
//	  Claude Code: filepath.Join(base, "CLAUDE.md")
//	  Cursor IDE:  filepath.Join(base, "sequoia-ai.md")
//	  OpenCode:    filepath.Join(base, "..", "AGENTS.md")
func systemPromptPath(base string) string {
	return filepath.Join(base, "sequoia-ai.md")
}

// versionFilePath returns the path to the .sequoia-version marker file.
func versionFilePath(base string) string {
	return filepath.Join(base, ".sequoia-version")
}

// backupPath returns the temp backup directory used for rollback.
//
// TODO: Adjust the backup directory name if needed.
// The default ".sequoia-backup" works for most tools.
func backupPath(base string) string {
	return filepath.Join(base, ".sequoia-backup")
}
