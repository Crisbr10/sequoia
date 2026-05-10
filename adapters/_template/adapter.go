// Package template is a copy-paste boilerplate for creating new tool adapters.
// Replace every "TODO" comment with your tool-specific logic.
//
// Steps to use this template:
//   1. Copy this entire directory to adapters/{tool}/
//   2. Replace "template" with your tool ID in all files
//   3. Follow the TODO comments in order
//   4. Register in cmd/sequoia/main.go with a blank import
//
// See CONTRIBUTING.md for the full step-by-step guide.
package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sequoia-ai/adapters"
	"sequoia-ai/adapters/common"
)

// Adapter implements adapters.ToolAdapter.
// homeDir overrides os.UserHomeDir() for testing. Leave empty for production use.
type Adapter struct {
	homeDir string
}

func init() {
	adapters.DefaultRegistry.Register(&Adapter{})
}

// NewAdapter creates an Adapter with an overridden home directory.
// Pass an empty string to use the real home directory (production use).
// Pass a temp directory in tests to avoid touching the real config directory.
func NewAdapter(homeDir string) *Adapter {
	return &Adapter{homeDir: homeDir}
}

func (a *Adapter) base() (string, error) {
	return toolBase(a.homeDir)
}

// ID returns the unique machine-readable identifier.
//
// TODO: Replace "template" with your adapter ID (e.g. "my-tool").
// Must be lowercase, kebab-case, and unique across all adapters.
func (a *Adapter) ID() string { return "template" }

// Name returns the human-readable display name.
//
// TODO: Replace "Template Tool" with the tool's brand name (e.g. "My Tool").
func (a *Adapter) Name() string { return "Template Tool" }

// Detect reports whether the tool appears to be installed on this machine.
//
// TODO: Replace with tool-specific detection logic:
//   - Check if the config directory exists (os.Stat)
//   - Check if the binary is in PATH (exec.LookPath)
//   - Return true if either check succeeds
func (a *Adapter) Detect() bool {
	base, err := a.base()
	if err == nil {
		if _, err := os.Stat(filepath.Join(base, "..")); err == nil {
			return true
		}
	}
	// TODO: Replace "my-tool-binary" with the actual binary name.
	_, err = exec.LookPath("my-tool-binary")
	return err == nil
}

// IsInstalled reports whether Sequoia has already been installed for this tool.
//
// TODO: Choose the right check for your tool:
//   - If the system prompt is a dedicated file, check os.Stat(systemPromptPath(base))
//   - If it's a section in a config file, check for marker presence
func (a *Adapter) IsInstalled() bool {
	base, err := a.base()
	if err != nil {
		return false
	}
	_, err = os.Stat(systemPromptPath(base))
	return err == nil
}

// PromptStrategy returns the injection strategy used by this adapter.
//
// TODO: Choose the right strategy for your tool:
//   - StrategyMarkdownSections: inject markers into a Markdown file
//   - StrategyFileReplace: dedicated file, full replace with backup
//   - StrategyConfigMerge: config file with markers (non-Markdown format)
//   - StrategyTOMLMerge: TOML config merge
// See CONTRIBUTING.md for guidance on each strategy.
func (a *Adapter) PromptStrategy() adapters.PromptStrategy {
	// TODO: Replace with the correct strategy.
	return adapters.StrategyFileReplace
}

// SkillsPath returns the absolute path to the skills directory.
func (a *Adapter) SkillsPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return skillsPath(base)
}

// CommandsPath returns the absolute path to the commands directory.
func (a *Adapter) CommandsPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return commandsPath(base)
}

// SystemPromptPath returns the absolute path to the system prompt file.
func (a *Adapter) SystemPromptPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return systemPromptPath(base)
}

// Status returns the current installation status.
// It populates Version from the .sequoia-version file when present.
func (a *Adapter) Status() adapters.AdapterStatus {
	installed := a.IsInstalled()
	version := ""
	if installed {
		base, err := a.base()
		if err == nil {
			data, err := os.ReadFile(versionFilePath(base))
			if err == nil {
				version = strings.TrimSpace(string(data))
			}
		}
	}
	return adapters.AdapterStatus{
		Installed: installed,
		Version:   version,
		Path:      a.SkillsPath(),
	}
}

// Install installs Sequoia files for this tool.
func (a *Adapter) Install() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("install: resolve home: %w", err)
	}

	data := templateData{Version: Version}

	// Stage rendered templates to a temp dir for common.Installer.
	staging, err := os.MkdirTemp("", "sequoia-template-*")
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

	// Generate the system prompt file.
	//
	// TODO: Choose the right system prompt generation approach:
	//   - StrategyFileReplace: use GenerateRulesMD (full file replace)
	//   - StrategyMarkdownSections: use InjectSection/RemoveSection (marker-based)
	//   - StrategyConfigMerge: use InjectSection/RemoveSection
	//   - StrategyTOMLMerge: implement TOML merge logic
	// See existing adapters for examples of each approach.
	rulesContent, err := renderTemplate("templates/rules.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := GenerateRulesMD(systemPromptPath(base), rulesContent); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: generate rules.md: %w", err)
	}

	// Write the version marker file.
	if err := os.WriteFile(versionFilePath(base), []byte(Version), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files for this tool.
func (a *Adapter) Uninstall() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("uninstall: resolve home: %w", err)
	}

	// Remove skill file, version marker, and command files.
	// Best-effort — missing files are not errors.
	_ = os.Remove(filepath.Join(skillsPath(base), "SKILL.md"))
	_ = os.Remove(versionFilePath(base))
	for _, cmd := range commandFiles {
		_ = os.Remove(filepath.Join(commandsPath(base), cmd))
	}

	// Remove the system prompt file.
	//
	// TODO: Match the approach used in Install():
	//   - StrategyFileReplace: use RemoveRulesMD
	//   - StrategyMarkdownSections: use RemoveSection
	//   - StrategyConfigMerge: use RemoveSection
	return RemoveRulesMD(systemPromptPath(base))
}
