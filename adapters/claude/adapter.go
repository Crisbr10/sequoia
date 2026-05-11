// Package claude implements the ToolAdapter for Claude Code.
package claude

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

const sequoiaMarker = "<!-- sequoia:start -->"

// Adapter implements adapters.ToolAdapter for Claude Code.
// homeDir overrides os.UserHomeDir() for testing. Leave empty for production use.
type Adapter struct {
	homeDir string
}

func init() {
	adapters.DefaultRegistry.Register(&Adapter{})
}

// NewAdapter creates an Adapter with an overridden home directory.
// Pass an empty string to use the real home directory (production use).
// Pass a temp directory in tests to avoid touching ~/.claude/.
func NewAdapter(homeDir string) *Adapter {
	return &Adapter{homeDir: homeDir}
}

func (a *Adapter) base() (string, error) {
	return claudeBase(a.homeDir)
}

// ID returns the unique machine-readable identifier for this adapter.
func (a *Adapter) ID() string { return "claude-code" }

// Name returns the human-readable display name.
func (a *Adapter) Name() string { return "Claude Code" }

// Detect reports whether Claude Code appears to be present on this machine.
// It returns true if ~/.claude/ exists OR if the claude binary is in PATH.
func (a *Adapter) Detect() bool {
	base, err := a.base()
	if err == nil {
		if _, err := os.Stat(base); err == nil {
			return true
		}
	}
	_, err = exec.LookPath("claude")
	return err == nil
}

// IsInstalled reports whether Sequoia has already been installed for Claude Code.
// It reads ~/.claude/CLAUDE.md and looks for the sequoia marker comment.
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
	return adapters.StrategyMarkdownSections
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

// SystemPromptPath returns the absolute path to the CLAUDE.md system prompt file.
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

// Install installs Sequoia files for Claude Code.
func (a *Adapter) Install() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("install: resolve home: %w", err)
	}

	data := templateData{Version: common.Version}

	// Render templates into a temp staging dir for common.Installer.
	staging, err := os.MkdirTemp("", "sequoia-claude-*")
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

	// Inject the Sequoia section into CLAUDE.md (marker-based, not file copy).
	sectionContent, err := common.RenderTemplate(templateFS, "templates/claude-md-section.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := InjectSection(systemPromptPath(base), sectionContent); err != nil {
		return fmt.Errorf("install: inject claude.md section: %w", err)
	}

	// Write the version marker file.
	if err := os.WriteFile(versionFilePath(base), []byte(common.Version+"\n"), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files for Claude Code.
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

	// Remove the Sequoia section from CLAUDE.md.
	return RemoveSection(systemPromptPath(base))
}
