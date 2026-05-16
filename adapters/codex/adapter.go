// Package codex implements the ToolAdapter for OpenAI Codex.
package codex

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

// Adapter implements adapters.ToolAdapter for OpenAI Codex.
// It embeds common.BaseAdapter for Status and path methods.
// Install and Uninstall are custom because Codex uses TOML merging
// instead of the standard markdown/file-replace strategies.
type Adapter struct {
	common.BaseAdapter
}

func init() {
	adapters.DefaultRegistry.Register(newAdapter(""))
}

// NewAdapter creates an Adapter with an overridden home directory.
func NewAdapter(homeDir string) *Adapter {
	return newAdapter(homeDir)
}

func newAdapter(homeDir string) *Adapter {
	a := &Adapter{}
	a.SetIDName("codex", "OpenAI Codex")
	a.SetHomeDir(homeDir)
	a.ResolveBase(codexBase)
	a.SetPathFns(skillsPath, commandsPath, systemPromptPath, versionFilePath, backupPath)
	a.SetStrategy(adapters.StrategyTOMLMerge, nil, nil) // TOML strategy — custom Install/Uninstall
	a.SetDetectFn(func() bool {
		base, err := codexBase(homeDir)
		if err != nil {
			return false
		}
		_, err = os.Stat(base)
		return err == nil
	})
	a.SetIsInstalledFn(func(base string) bool {
		if _, err := os.Stat(filepath.Join(base, "sequoia")); os.IsNotExist(err) {
			return false
		}
		data, err := os.ReadFile(configPath(base))
		if err != nil {
			return false
		}
		return containsSequoiaSection(string(data))
	})
	return a
}

// Install installs Sequoia files for OpenAI Codex using TOML merging.
// Overrides BaseAdapter.Install because Codex uses a custom TOML merge strategy
// and its template data includes runtime paths.
//
// On failure, the returned error wraps adapters.ErrInstallFailed.
func (a *Adapter) Install(opts adapters.InstallOpts) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", adapters.ErrInstallFailed, err)
		}
	}()
	base, err := a.Base()
	if err != nil {
		return fmt.Errorf("install: resolve home: %w", err)
	}

	data := templateData{
		Version:      common.Version,
		SkillsPath:   skillsPath(base),
		CommandsPath: commandsPath(base),
	}

	staging, err := os.MkdirTemp("", "sequoia-codex-*")
	if err != nil {
		return fmt.Errorf("install: create staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	skillContent, err := common.RenderTemplate(templateFS, "templates/skill.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := common.StageFile(staging, "SKILL.md", []byte(skillContent)); err != nil {
		return fmt.Errorf("install: stage skill: %w", err)
	}

	for _, cmd := range common.CommandFiles {
		content, err := common.CommandFS.ReadFile("templates/commands/" + cmd)
		if err != nil {
			return fmt.Errorf("install: read command %q: %w", cmd, err)
		}
		if err := common.StageFile(staging, cmd, content); err != nil {
			return fmt.Errorf("install: stage command %q: %w", cmd, err)
		}
	}

	for _, dir := range []string{skillsPath(base), commandsPath(base)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("install: create dir %q: %w", dir, err)
		}
	}

	// Generate a unique session suffix for backup dirs to avoid collisions.
	sessionSuffix := strconv.FormatInt(time.Now().UnixMilli(), 36)

	skillInstaller := common.NewInstaller(common.InstallerConfig{
		SourceDir: staging,
		TargetDir: skillsPath(base),
		BackupDir: backupPath(base) + "-" + sessionSuffix,
		Files:     []string{"SKILL.md"},
	})
	if err := skillInstaller.Run(); err != nil {
		return fmt.Errorf("install: skill: %w", err)
	}

	cmdInstaller := common.NewInstaller(common.InstallerConfig{
		SourceDir: staging,
		TargetDir: commandsPath(base),
		BackupDir: backupPath(base) + "-" + sessionSuffix,
		Files:     common.CommandFiles,
	})
	if err := cmdInstaller.Run(); err != nil {
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: commands: %w", err)
	}

	sequoiaTable := map[string]interface{}{
		"skills_path":   skillsPath(base),
		"commands_path": commandsPath(base),
	}
	if err := MergeConfig(configPath(base), sequoiaTable); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: merge config: %w", err)
	}

	if err := os.WriteFile(versionFilePath(base), []byte(common.Version+"\n"), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files for OpenAI Codex.
// Overrides BaseAdapter.Uninstall because Codex uses TOML config merging
// and removes a sequoia/ subdirectory tree.
//
// Errors from individual file removals are collected via errors.Join.
// On failure, the returned error wraps adapters.ErrUninstallFailed.
func (a *Adapter) Uninstall(opts adapters.InstallOpts) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", adapters.ErrUninstallFailed, err)
		}
	}()

	base, err := a.Base()
	if err != nil {
		return fmt.Errorf("uninstall: resolve home: %w", err)
	}

	var errs []error

	if err := os.Remove(filepath.Join(skillsPath(base), "SKILL.md")); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove skill file: %w", err))
	}
	if err := os.Remove(versionFilePath(base)); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove version file: %w", err))
	}
	for _, cmd := range common.CommandFiles {
		if err := os.Remove(filepath.Join(commandsPath(base), cmd)); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove command %s: %w", cmd, err))
		}
	}

	if err := RemoveConfig(configPath(base)); err != nil {
		errs = append(errs, fmt.Errorf("remove config: %w", err))
	}

	if err := os.RemoveAll(filepath.Join(base, "sequoia")); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove sequoia dir: %w", err))
	}

	return errors.Join(errs...)
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
