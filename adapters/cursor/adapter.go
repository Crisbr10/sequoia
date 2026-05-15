// Package cursor implements the ToolAdapter for Cursor IDE.
package cursor

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

// Adapter implements adapters.ToolAdapter for Cursor IDE.
// It embeds common.BaseAdapter for shared Install/Uninstall/Status logic.
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
	a.SetIDName("cursor", "Cursor IDE")
	a.SetHomeDir(homeDir)
	a.ResolveBase(cursorBase)
	a.SetPathFns(skillsPath, commandsPath, systemPromptPath, versionFilePath, backupPath)
	a.SetStrategy(adapters.StrategyFileReplace,
		func(base, content string) error { return common.ReplaceFile(systemPromptPath(base), content) },
		func(base string) error { return common.RestoreOrRemoveFile(systemPromptPath(base)) })
	a.SetInstallTemplates(templateFS, "sequoia-cursor-*",
		"templates/rules.md.tmpl",
		func() interface{} { return templateData{Version: common.Version} })
	a.SetRollbackOnSystemPromptError(true)
	a.SetIsInstalledFn(func(base string) bool {
		_, err := os.Stat(versionFilePath(base))
		return err == nil
	})
	a.SetDetectFn(func() bool {
		if base, err := cursorBase(homeDir); err == nil {
			if _, err := os.Stat(filepath.Join(base, "..")); err == nil {
				return true
			}
		}
		_, err := exec.LookPath("cursor")
		return err == nil
	})
	return a
}
