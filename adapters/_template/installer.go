//go:build ignore

package template

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = "<!-- sequoia:start -->"
	markerEnd   = "<!-- sequoia:end -->"
)

// GenerateRulesMD writes Sequoia content to the system prompt file at path.
// If the file does not exist it is created. If it already contains Sequoia
// markers the file is replaced in place. If it contains other content the
// original file is backed up to path+".sequoia-backup" before replacement.
//
// TODO: This function is for StrategyFileReplace. If your tool uses a different
// strategy, replace this with the appropriate injection logic:
//   - StrategyMarkdownSections: InjectSection() / RemoveSection()
//     (see adapters/claude/installer.go)
//   - StrategyConfigMerge: similar to MarkdownSections
//     (see adapters/gemini/installer.go)
//   - StrategyTOMLMerge: TOML table merge
//     (see adapters/codex/installer.go)
func GenerateRulesMD(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	managed, err := isSequoiaManaged(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		return os.WriteFile(path, []byte(content), 0o644)
	}

	if managed {
		return os.WriteFile(path, []byte(content), 0o644)
	}

	backup := path + ".sequoia-backup"
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backup, raw, 0o644); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// RemoveRulesMD removes Sequoia from the system prompt file at path.
// If a backup exists the original content is restored. If no backup exists
// and the file is Sequoia-managed it is deleted. If the file is missing or
// contains no markers and no backup exists the function returns nil.
//
// TODO: Match the approach used in GenerateRulesMD above.
func RemoveRulesMD(path string) error {
	backup := path + ".sequoia-backup"

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if _, berr := os.Stat(backup); berr == nil {
		raw, err := os.ReadFile(backup)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			return err
		}
		return os.Remove(backup)
	}

	managed, err := isSequoiaManaged(path)
	if err != nil {
		return err
	}
	if managed {
		return os.Remove(path)
	}

	return nil
}

func isSequoiaManaged(path string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(raw), markerStart), nil
}
