// Package cursor implements the ToolAdapter for Cursor IDE.
package cursor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

// Adapter implements adapters.ToolAdapter for Cursor IDE.
// homeDir overrides os.UserHomeDir() for testing. Leave empty for production use.
type Adapter struct {
	homeDir string
}

func init() {
	adapters.DefaultRegistry.Register(&Adapter{})
}

// NewAdapter creates an Adapter with an overridden home directory.
// Pass an empty string to use the real home directory (production use).
// Pass a temp directory in tests to avoid touching ~/.cursor/.
func NewAdapter(homeDir string) *Adapter {
	return &Adapter{homeDir: homeDir}
}

func (a *Adapter) base() (string, error) {
	return cursorBase(a.homeDir)
}

// ID returns the unique machine-readable identifier for this adapter.
func (a *Adapter) ID() string { return "cursor" }

// Name returns the human-readable display name.
func (a *Adapter) Name() string { return "Cursor IDE" }

// Detect reports whether Cursor IDE appears to be present on this machine.
// It returns true if ~/.cursor/ exists OR if the cursor binary is in PATH.
func (a *Adapter) Detect() bool {
	base, err := a.base()
	if err == nil {
		if _, err := os.Stat(filepath.Join(base, "..")); err == nil {
			return true
		}
	}
	_, err = exec.LookPath("cursor")
	return err == nil
}

// IsInstalled reports whether Sequoia has already been installed for Cursor IDE.
// It checks if ~/.cursor/rules/sequoia-ai.md exists.
func (a *Adapter) IsInstalled() bool {
	base, err := a.base()
	if err != nil {
		return false
	}
	_, err = os.Stat(systemPromptPath(base))
	return err == nil
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

// SystemPromptPath returns the absolute path to the sequoia-ai.md rules file.
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

// Install installs Sequoia files for Cursor IDE.
func (a *Adapter) Install() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("install: resolve home: %w", err)
	}

	data := templateData{Version: common.Version}

	// Stage rendered templates to a temp dir for common.Installer.
	staging, err := os.MkdirTemp("", "sequoia-cursor-*")
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

	// Generate sequoia-ai.md (full file replace).
	rulesContent, err := common.RenderTemplate(templateFS, "templates/rules.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := GenerateRulesMD(systemPromptPath(base), rulesContent); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: generate rules.md: %w", err)
	}

	// Write the version marker file.
	if err := os.WriteFile(versionFilePath(base), []byte(common.Version), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files for Cursor IDE.
func (a *Adapter) Uninstall() error {
	base, err := a.base()
	if err != nil {
		return fmt.Errorf("uninstall: resolve home: %w", err)
	}

	// Remove skill file, version marker, and command files (best-effort — missing files are not errors).
	_ = os.Remove(filepath.Join(skillsPath(base), "SKILL.md"))
	_ = os.Remove(versionFilePath(base))
	for _, cmd := range common.CommandFiles {
		_ = os.Remove(filepath.Join(commandsPath(base), cmd))
	}

	// Remove sequoia-ai.md (restore backup or delete if Sequoia-managed).
	return RemoveRulesMD(systemPromptPath(base))
}
