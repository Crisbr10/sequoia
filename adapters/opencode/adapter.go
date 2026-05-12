// Package opencode implements the ToolAdapter for OpenCode.
package opencode

import (
	"os"
	"os/exec"
	"strings"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

const sequoiaMarker = "<!-- sequoia:start -->"

// Adapter implements adapters.ToolAdapter for OpenCode.
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
	a.SetIDName("opencode", "OpenCode")
	a.SetHomeDir(homeDir)
	a.ResolveBase(opencodeBase)
	a.SetPathFns(skillsPath, commandsPath, systemPromptPath, versionFilePath, backupPath)
	a.SetStrategy(adapters.StrategyFileReplace,
		func(base, content string) error { return common.ReplaceFile(systemPromptPath(base), content) },
		func(base string) error { return common.RestoreOrRemoveFile(systemPromptPath(base)) })
	a.SetInstallTemplates(templateFS, "sequoia-opencode-*",
		"templates/agents-md-section.md.tmpl",
		func() interface{} { return templateData{Version: common.Version} })
	a.SetRollbackOnSystemPromptError(true)
	a.SetIsInstalledFn(func(base string) bool {
		data, err := os.ReadFile(systemPromptPath(base))
		if err != nil {
			return false
		}
		return strings.Contains(string(data), sequoiaMarker)
	})
	a.SetDetectFn(func() bool {
		if base, err := opencodeBase(homeDir); err == nil {
			if _, err := os.Stat(base); err == nil {
				return true
			}
		}
		_, err := exec.LookPath("opencode")
		return err == nil
	})
	return a
}
