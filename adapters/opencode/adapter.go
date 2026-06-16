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

// Compile-time check: Adapter satisfies common.Strategy (P3-003 ISP narrowing).
var _ common.Strategy = (*Adapter)(nil)

// RegisterIn registers this adapter in the given registry.
// Use this for constructor DI; init() delegates to it for backward compatibility.
func RegisterIn(reg *adapters.Registry) {
	reg.RegisterFactory("opencode", func() adapters.ToolAdapter { return newAdapter("") })
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
		func(base, content string) error {
			return common.ReplaceFile("opencode", systemPromptPath(base), content)
		},
		func(base string) error { return common.RestoreOrRemoveFile("opencode", systemPromptPath(base)) })
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
			if base, err := a.Base(); err == nil {
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
