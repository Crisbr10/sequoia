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

// RegisterIn registers this adapter in the given registry.
// Use this for constructor DI; init() delegates to it for backward compatibility.
func RegisterIn(reg *adapters.Registry) {
	reg.Register(newAdapter(""))
}

func init() {
	RegisterIn(adapters.DefaultRegistry)
}

// NewAdapter creates an Adapter with an overridden home directory.
func NewAdapter(homeDir string) *Adapter {
	return newAdapter(homeDir)
}

func newAdapter(homeDir string) *Adapter {
	a := &Adapter{}
	a.SetIDName("opencode", "OpenCode")
	a.SetPaths(common.NewPathResolver(opencodeBase, homeDir,
		skillsPath, commandsPath, systemPromptPath, versionFilePath, backupPath,
		a.AddWarning))
	pm := common.NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return common.ReplaceFile(systemPromptPath(base), content) },
		func(base string) error { return common.RestoreOrRemoveFile(systemPromptPath(base)) })
	pm.SetRollbackOnError(true)
	a.SetPromptManager(pm)
	a.SetBackup(common.NewBackupPathBuilder(backupPath, "opencode"))
	a.SetInstallTemplates(templateFS, "sequoia-opencode-*",
		"templates/agents-md-section.md.tmpl",
		func() interface{} { return templateData{Version: common.Version} })
	a.SetDetector(common.NewDetector(
		a.Base,
		func(base string) bool {
			data, err := os.ReadFile(systemPromptPath(base))
			if err != nil {
				return false
			}
			return strings.Contains(string(data), sequoiaMarker)
		},
		func() bool {
			if base, err := opencodeBase(homeDir); err == nil {
				if _, err := os.Stat(base); err == nil {
					return true
				}
			}
			_, err := exec.LookPath("opencode")
			return err == nil
		},
	))
	return a
}
