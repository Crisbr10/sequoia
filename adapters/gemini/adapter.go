// Package gemini implements the ToolAdapter for Gemini CLI.
package gemini

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

const sequoiaMarker = "<!-- sequoia:start -->"

// Adapter implements adapters.ToolAdapter for Gemini CLI.
// It embeds common.BaseAdapter for shared Install/Status logic.
// Uninstall is overridden because Gemini uses a sequoia/ subdirectory
// that is removed as a whole tree.
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
	a.SetIDName("gemini-cli", "Gemini CLI")
	a.SetHomeDir(homeDir)
	a.ResolveBase(geminiBase)
	a.SetPathFns(skillsPath, commandsPath, systemPromptPath, versionFilePath, backupPath)
	a.SetStrategy(adapters.StrategyConfigMerge,
		func(base, content string) error { return common.InjectMarkdownSection(systemPromptPath(base), content) },
		func(base string) error { return common.RemoveMarkdownSection(systemPromptPath(base)) })
	a.SetInstallTemplates(templateFS, "sequoia-gemini-*",
		"templates/gemini-md-section.md.tmpl",
		func() interface{} { return templateData{Version: common.Version} })
	a.SetIsInstalledFn(func(base string) bool {
		data, err := os.ReadFile(systemPromptPath(base))
		if err != nil {
			return false
		}
		return strings.Contains(string(data), sequoiaMarker)
	})
	a.SetDetectFn(func() bool {
		base, err := geminiBase(homeDir)
		if err != nil {
			return false
		}
		_, err = os.Stat(base)
		return err == nil
	})
	return a
}

// Uninstall removes Sequoia files for Gemini CLI.
// Overrides BaseAdapter.Uninstall because Gemini stores skills and commands
// under a sequoia/ subdirectory that is removed as a whole tree.
//
// Errors from removal operations are collected via errors.Join.
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

	// Remove the sequoia subdirectory.
	sequoiaDir := filepath.Join(base, "sequoia")
	if err := os.RemoveAll(sequoiaDir); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove sequoia dir: %w", err))
	}

	// Remove the Sequoia section from GEMINI.md.
	if err := common.RemoveMarkdownSection(systemPromptPath(base)); err != nil {
		errs = append(errs, fmt.Errorf("restore system prompt: %w", err))
	}

	return errors.Join(errs...)
}
