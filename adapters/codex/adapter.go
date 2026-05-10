// Package codex implements the ToolAdapter for OpenAI Codex.
package codex

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"sequoia-ai/adapters"
	"sequoia-ai/adapters/common"
)

// Adapter implements adapters.ToolAdapter for OpenAI Codex.
// homeDir overrides os.UserHomeDir() for testing. Leave empty for production use.
type Adapter struct {
	homeDir string
}

func init() {
	adapters.DefaultRegistry.Register(&Adapter{})
}

// NewAdapter creates an Adapter with an overridden home directory.
// Pass an empty string to use the real home directory (production use).
// Pass a temp directory in tests to avoid touching ~/.codex/.
func NewAdapter(homeDir string) *Adapter {
	return &Adapter{homeDir: homeDir}
}

func (a *Adapter) base() (string, error) {
	return codexBase(a.homeDir)
}

// ID returns the unique machine-readable identifier for this adapter.
func (a *Adapter) ID() string { return "codex" }

// Name returns the human-readable display name.
func (a *Adapter) Name() string { return "OpenAI Codex" }

// Detect reports whether OpenAI Codex appears to be present on this machine.
// It returns true if ~/.codex/ directory exists.
func (a *Adapter) Detect() bool {
	base, err := a.base()
	if err != nil {
		return false
	}
	_, err = os.Stat(base)
	return err == nil
}

// IsInstalled reports whether Sequoia has already been installed for Codex.
// It checks if ~/.codex/sequoia/ directory exists AND [sequoia] table is in config.toml.
func (a *Adapter) IsInstalled() bool {
	base, err := a.base()
	if err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(base, "sequoia")); os.IsNotExist(err) {
		return false
	}
	// Check for [sequoia] in config.toml.
	data, err := os.ReadFile(configPath(base))
	if err != nil {
		return false
	}
	return containsSequoiaSection(string(data))
}

// PromptStrategy returns the injection strategy used by this adapter.
func (a *Adapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyTOMLMerge
}

// SkillsPath returns the absolute path to the skills directory for this adapter.
func (a *Adapter) SkillsPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return skillsPath(base)
}

// CommandsPath returns the absolute path to the commands directory for this adapter.
func (a *Adapter) CommandsPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return commandsPath(base)
}

// SystemPromptPath returns the absolute path to the config.toml file.
func (a *Adapter) SystemPromptPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return systemPromptPath(base)
}

// Status returns the current installation status of this adapter.
// It populates Version from the .sequoia-version file when present.
func (a *Adapter) Status() adapters.AdapterStatus {
	installed := a.IsInstalled()
	version := ""
	if installed {
		base, err := a.base()
		if err == nil {
			data, err := os.ReadFile(versionFilePath(base))
			if err == nil {
				version = string(data)
				version = trimSpace(version)
			}
		}
	}
	return adapters.AdapterStatus{
		Installed: installed,
		Version:   version,
		Path:      a.SkillsPath(),
	}
}

// Install installs Sequoia files for OpenAI Codex.
func (a *Adapter) Install() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("install: resolve home: %w", err)
	}

	data := templateData{
		Version:      Version,
		SkillsPath:   skillsPath(base),
		CommandsPath: commandsPath(base),
	}

	// Render templates into a temp staging dir for common.Installer.
	staging, err := os.MkdirTemp("", "sequoia-codex-*")
	if err != nil {
		return fmt.Errorf("install: create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)

	// Render and stage the skill file.
	skillContent, err := renderTemplate("templates/skill.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := stageFile(staging, "SKILL.md", []byte(skillContent)); err != nil {
		return fmt.Errorf("install: stage skill: %w", err)
	}

	// Stage command files (static — no rendering needed).
	for _, cmd := range commandFiles {
		content, err := templateFS.ReadFile("templates/commands/" + cmd)
		if err != nil {
			return fmt.Errorf("install: read command %q: %w", cmd, err)
		}
		if err := stageFile(staging, cmd, content); err != nil {
			return fmt.Errorf("install: stage command %q: %w", cmd, err)
		}
	}

	// Create target directories before Prepare (Prepare probes for write access).
	for _, dir := range []string{skillsPath(base), commandsPath(base)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("install: create dir %q: %w", dir, err)
		}
	}

	// Install skill file via the common framework.
	skillInstaller := common.NewInstaller(common.InstallerConfig{
		SourceDir: staging,
		TargetDir: skillsPath(base),
		BackupDir: backupPath(base),
		Files:     []string{"SKILL.md"},
	})
	if err := runInstaller(skillInstaller); err != nil {
		return fmt.Errorf("install: skill: %w", err)
	}

	// Install command files via the common framework.
	cmdInstaller := common.NewInstaller(common.InstallerConfig{
		SourceDir: staging,
		TargetDir: commandsPath(base),
		BackupDir: backupPath(base),
		Files:     commandFiles,
	})
	if err := runInstaller(cmdInstaller); err != nil {
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: commands: %w", err)
	}

	// Merge the [sequoia] table into config.toml.
	sequoiaTable := map[string]interface{}{
		"skills_path":  skillsPath(base),
		"commands_path": commandsPath(base),
	}
	if err := MergeConfig(configPath(base), sequoiaTable); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: merge config: %w", err)
	}

	// Write the version marker file.
	if err := os.WriteFile(versionFilePath(base), []byte(Version+"\n"), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files for OpenAI Codex.
func (a *Adapter) Uninstall() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("uninstall: resolve home: %w", err)
	}

	// Remove skill file, version marker, and command files (best-effort — missing files are not errors).
	_ = os.Remove(filepath.Join(skillsPath(base), "SKILL.md"))
	_ = os.Remove(versionFilePath(base))
	for _, cmd := range commandFiles {
		_ = os.Remove(filepath.Join(commandsPath(base), cmd))
	}

	// Remove the [sequoia] table from config.toml.
	if err := RemoveConfig(configPath(base)); err != nil {
		return fmt.Errorf("uninstall: remove config: %w", err)
	}

	// Remove the sequoia directory tree.
	if err := os.RemoveAll(filepath.Join(base, "sequoia")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("uninstall: remove sequoia dir: %w", err)
	}
	return nil
}

// containsSequoiaSection checks whether the TOML content contains a [sequoia] table header.
func containsSequoiaSection(tomlContent string) bool {
	var data map[string]interface{}
	if _, err := toml.Decode(tomlContent, &data); err != nil {
		return false
	}
	_, ok := data["sequoia"]
	return ok
}

// trimSpace removes leading and trailing whitespace from s.
func trimSpace(s string) string {
	// Simple implementation equivalent to strings.TrimSpace without importing strings.
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	j := len(s) - 1
	for j >= i && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
		j--
	}
	return s[i : j+1]
}
