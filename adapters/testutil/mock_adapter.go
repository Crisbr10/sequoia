// Package testutil provides shared test doubles for adapter tests.
package testutil

import "github.com/Crisbr10/sequoia/adapters"

// MockAdapter is a configurable ToolAdapter test double.
// Set function fields to customize behavior; nil fields use sensible defaults.
type MockAdapter struct {
	IDFunc              func() string
	NameFunc            func() string
	DetectFunc          func() bool
	IsInstalledFunc     func() bool
	InstallFunc         func(adapters.InstallOpts) error
	UninstallFunc       func(adapters.InstallOpts) error
	StatusFunc          func() adapters.AdapterStatus
	SkillsPathFunc      func() string
	CommandsPathFunc    func() string
	SystemPromptPathFunc func() string
	PromptStrategyFunc  func() adapters.PromptStrategy
	IDVal    string
	NameVal  string
}

func (m *MockAdapter) ID() string {
	if m.IDFunc != nil { return m.IDFunc() }
	return m.IDVal
}
func (m *MockAdapter) Name() string {
	if m.NameFunc != nil { return m.NameFunc() }
	return m.NameVal
}
func (m *MockAdapter) Detect() bool {
	if m.DetectFunc != nil { return m.DetectFunc() }
	return false
}
func (m *MockAdapter) IsInstalled() bool {
	if m.IsInstalledFunc != nil { return m.IsInstalledFunc() }
	return false
}
func (m *MockAdapter) Install(opts adapters.InstallOpts) error {
	if m.InstallFunc != nil { return m.InstallFunc(opts) }
	return nil
}
func (m *MockAdapter) Uninstall(opts adapters.InstallOpts) error {
	if m.UninstallFunc != nil { return m.UninstallFunc(opts) }
	return nil
}
func (m *MockAdapter) Status() adapters.AdapterStatus {
	if m.StatusFunc != nil { return m.StatusFunc() }
	return adapters.AdapterStatus{}
}
func (m *MockAdapter) SkillsPath() string {
	if m.SkillsPathFunc != nil { return m.SkillsPathFunc() }
	return ""
}
func (m *MockAdapter) CommandsPath() string {
	if m.CommandsPathFunc != nil { return m.CommandsPathFunc() }
	return ""
}
func (m *MockAdapter) SystemPromptPath() string {
	if m.SystemPromptPathFunc != nil { return m.SystemPromptPathFunc() }
	return ""
}
func (m *MockAdapter) PromptStrategy() adapters.PromptStrategy {
	if m.PromptStrategyFunc != nil { return m.PromptStrategyFunc() }
	return adapters.StrategyMarkdownSections
}

var _ adapters.ToolAdapter = (*MockAdapter)(nil)
