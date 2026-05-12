package screens_test

import (
	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/model"
)

// dummyAdapter is a minimal ToolAdapter for screen tests.
type dummyAdapter struct {
	id   string
	name string
	inst bool
	ver  string
	path string
}

func (d *dummyAdapter) ID() string                                { return d.id }
func (d *dummyAdapter) Name() string                              { return d.name }
func (d *dummyAdapter) Detect() bool                              { return true }
func (d *dummyAdapter) IsInstalled() bool                         { return d.inst }
func (d *dummyAdapter) Install(opts adapters.InstallOpts) error   { _ = opts.Language; return nil }
func (d *dummyAdapter) Uninstall(opts adapters.InstallOpts) error { _ = opts.Language; return nil }
func (d *dummyAdapter) Status() model.ToolStatus {
	return model.ToolStatus{Installed: d.inst, Version: d.ver, Path: d.path}
}
func (d *dummyAdapter) SkillsPath() string       { return "" }
func (d *dummyAdapter) CommandsPath() string     { return "" }
func (d *dummyAdapter) SystemPromptPath() string { return "" }
func (d *dummyAdapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyMarkdownSections
}

// Compile-time check: dummyAdapter satisfies model.ToolInfo.
var _ model.ToolInfo = (*dummyAdapter)(nil)
