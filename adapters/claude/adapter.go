// Package claude implements the ToolAdapter for Claude Code.
package claude

import (
	"os"
	"os/exec"
	"strings"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

const sequoiaMarker = "<!-- sequoia:start -->"

// Adapter implements adapters.ToolAdapter for Claude Code.
// It embeds common.BaseAdapter for shared Install/Uninstall/Status logic.
type Adapter struct {
	common.BaseAdapter
}

// RegisterIn registers this adapter in the given registry.
// Use this for constructor DI; init() delegates to it for backward compatibility.
func RegisterIn(reg *adapters.Registry) {
	reg.RegisterFactory("claude-code", func() adapters.ToolAdapter { return newAdapter("") })
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
	a.SetIDName("claude-code", "Claude Code")
	a.SetPaths(common.NewPathResolver(claudeBase, homeDir,
		skillsPath, commandsPath, systemPromptPath, versionFilePath, backupPath,
		a.AddWarning))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyMarkdownSections,
		func(base, content string) error { return common.InjectMarkdownSection(systemPromptPath(base), content) },
		func(base string) error { return common.RemoveMarkdownSection(systemPromptPath(base)) }))
	a.SetBackup(common.NewBackupPathBuilder(backupPath, "claude-code"))
	a.SetInstallTemplates(templateFS, "sequoia-claude-*",
		"templates/claude-md-section.md.tmpl",
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
			if base, err := claudeBase(homeDir); err == nil {
				if _, err := os.Stat(base); err == nil {
					return true
				}
			}
			_, err := exec.LookPath("claude")
			return err == nil
		},
	))
	return a
}
