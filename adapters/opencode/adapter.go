// Package opencode implements the ToolAdapter for OpenCode.
package opencode

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sequoia-ai/adapters"
	"sequoia-ai/adapters/common"
)

const sequoiaMarker = "<!-- sequoia:start -->"

// Adapter implements adapters.ToolAdapter for OpenCode.
// homeDir overrides os.UserHomeDir() for testing. Leave empty for production use.
type Adapter struct {
	homeDir string
}

func init() {
	adapters.DefaultRegistry.Register(&Adapter{})
}

// NewAdapter creates an Adapter with an overridden home directory.
// Pass an empty string to use the real home directory (production use).
// Pass a temp directory in tests to avoid touching ~/.config/opencode/.
func NewAdapter(homeDir string) *Adapter {
	return &Adapter{homeDir: homeDir}
}

func (a *Adapter) base() (string, error) {
	return opencodeBase(a.homeDir)
}

// ID returns the unique machine-readable identifier for this adapter.
func (a *Adapter) ID() string { return "opencode" }

// Name returns the human-readable display name.
func (a *Adapter) Name() string { return "OpenCode" }

// Detect reports whether OpenCode appears to be present on this machine.
// It returns true if ~/.config/opencode/ exists OR if the opencode binary is in PATH.
func (a *Adapter) Detect() bool {
	base, err := a.base()
	if err == nil {
		if _, err := os.Stat(base); err == nil {
			return true
		}
	}
	_, err = exec.LookPath("opencode")
	return err == nil
}

// IsInstalled reports whether Sequoia has already been installed for OpenCode.
// It reads ~/.config/opencode/AGENTS.md and looks for the sequoia marker comment.
func (a *Adapter) IsInstalled() bool {
	base, err := a.base()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(systemPromptPath(base))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), sequoiaMarker)
}

// PromptStrategy returns the injection strategy used by this adapter.
func (a *Adapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyFileReplace
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

// SystemPromptPath returns the absolute path to the AGENTS.md system prompt file.
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

// Install installs Sequoia files for OpenCode.
func (a *Adapter) Install() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("install: resolve home: %w", err)
	}

	data := templateData{Version: common.Version}

	// Stage rendered templates to a temp dir for common.Installer.
	staging, err := os.MkdirTemp("", "sequoia-opencode-*")
	if err != nil {
		return fmt.Errorf("install: create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)

	// Render and stage the skill file.
	skillContent, err := common.RenderTemplate(templateFS, "templates/skill.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := common.StageFile(staging, "SKILL.md", []byte(skillContent)); err != nil {
		return fmt.Errorf("install: stage skill: %w", err)
	}

	// Stage command files (static — no rendering needed).
	for _, cmd := range common.CommandFiles {
		content, err := templateFS.ReadFile("templates/commands/" + cmd)
		if err != nil {
			return fmt.Errorf("install: read command %q: %w", cmd, err)
		}
		if err := common.StageFile(staging, cmd, content); err != nil {
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
	if err := skillInstaller.Run(); err != nil {
		return fmt.Errorf("install: skill: %w", err)
	}

	// Install command files via the common framework.
	cmdInstaller := common.NewInstaller(common.InstallerConfig{
		SourceDir: staging,
		TargetDir: commandsPath(base),
		BackupDir: backupPath(base),
		Files:     common.CommandFiles,
	})
	if err := cmdInstaller.Run(); err != nil {
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: commands: %w", err)
	}

	// Generate AGENTS.md (full file replace — uses GenerateAgentsMD, not InjectSection).
	agentsContent, err := common.RenderTemplate(templateFS, "templates/agents-md-section.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := GenerateAgentsMD(systemPromptPath(base), agentsContent); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: generate agents.md: %w", err)
	}

	// Write the version marker file.
	if err := os.WriteFile(versionFilePath(base), []byte(common.Version+"\n"), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files for OpenCode.
func (a *Adapter) Uninstall() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("uninstall: resolve home: %w", err)
	}

	// Remove skill file, version marker, and command files (best-effort — missing files are not errors).
	_ = os.Remove(skillFilePath(base))
	_ = os.Remove(versionFilePath(base))
	for _, cmd := range common.CommandFiles {
		_ = os.Remove(filepath.Join(commandsPath(base), cmd))
	}

	// Remove AGENTS.md (restore backup or delete if Sequoia-managed).
	return RemoveAgentsMD(systemPromptPath(base))
}
