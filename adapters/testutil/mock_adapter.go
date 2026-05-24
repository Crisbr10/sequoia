// Package testutil provides shared test doubles for adapter tests.
package testutil

import "github.com/Crisbr10/sequoia/adapters"

// -- Focused mocks (P3-003 ISP narrowing) -------------------------------------
//
// Each mock satisfies exactly one role interface. Use them when a test only
// needs a subset of the ToolAdapter contract.

// MockIdentifier satisfies adapters.Identifier (ID, Name).
type MockIdentifier struct {
	IDFunc   func() string
	NameFunc func() string
	IDVal    string
	NameVal  string
}

func (m *MockIdentifier) ID() string {
	if m.IDFunc != nil {
		return m.IDFunc()
	}
	return m.IDVal
}
func (m *MockIdentifier) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return m.NameVal
}

var _ adapters.Identifier = (*MockIdentifier)(nil)

// MockDetector satisfies adapters.Detector (Detect, IsInstalled).
type MockDetector struct {
	DetectFunc      func() bool
	IsInstalledFunc func() bool
}

func (m *MockDetector) Detect() bool {
	if m.DetectFunc != nil {
		return m.DetectFunc()
	}
	return false
}
func (m *MockDetector) IsInstalled() bool {
	if m.IsInstalledFunc != nil {
		return m.IsInstalledFunc()
	}
	return false
}

var _ adapters.Detector = (*MockDetector)(nil)

// MockInstaller satisfies adapters.Installer (Install, Uninstall).
type MockInstaller struct {
	InstallFunc   func(adapters.InstallOpts) error
	UninstallFunc func(adapters.InstallOpts) error
}

func (m *MockInstaller) Install(opts adapters.InstallOpts) error {
	if m.InstallFunc != nil {
		return m.InstallFunc(opts)
	}
	return nil
}
func (m *MockInstaller) Uninstall(opts adapters.InstallOpts) error {
	if m.UninstallFunc != nil {
		return m.UninstallFunc(opts)
	}
	return nil
}

var _ adapters.Installer = (*MockInstaller)(nil)

// MockInstallStatus satisfies adapters.InstallStatus (Status).
type MockInstallStatus struct {
	StatusFunc func() adapters.AdapterStatus
}

func (m *MockInstallStatus) Status() adapters.AdapterStatus {
	if m.StatusFunc != nil {
		return m.StatusFunc()
	}
	return adapters.AdapterStatus{}
}

var _ adapters.InstallStatus = (*MockInstallStatus)(nil)

// MockAdapterPaths satisfies adapters.AdapterPaths (SkillsPath, CommandsPath,
// SystemPromptPath, PromptStrategy).
type MockAdapterPaths struct {
	SkillsPathFunc       func() string
	CommandsPathFunc     func() string
	SystemPromptPathFunc func() string
	PromptStrategyFunc   func() adapters.PromptStrategy
}

func (m *MockAdapterPaths) SkillsPath() string {
	if m.SkillsPathFunc != nil {
		return m.SkillsPathFunc()
	}
	return ""
}
func (m *MockAdapterPaths) CommandsPath() string {
	if m.CommandsPathFunc != nil {
		return m.CommandsPathFunc()
	}
	return ""
}
func (m *MockAdapterPaths) SystemPromptPath() string {
	if m.SystemPromptPathFunc != nil {
		return m.SystemPromptPathFunc()
	}
	return ""
}
func (m *MockAdapterPaths) PromptStrategy() adapters.PromptStrategy {
	if m.PromptStrategyFunc != nil {
		return m.PromptStrategyFunc()
	}
	return adapters.StrategyMarkdownSections
}

var _ adapters.AdapterPaths = (*MockAdapterPaths)(nil)

// -- Composite mock -----------------------------------------------------------
//
// MockAdapter satisfies adapters.ToolAdapter with flat fields for backward
// compatibility. For tests that only need a subset of the contract, prefer
// the focused mocks (MockIdentifier, MockDetector, etc.).

// MockAdapter is a configurable ToolAdapter test double.
// Set function fields to customize behavior; nil fields use sensible defaults.
// It also satisfies common.Strategy so that the pipeline's type assertions
// succeed when the adapter passes through production wrappers (e.g. toolInfoAdapter).
type MockAdapter struct {
	IDFunc          func() string
	NameFunc        func() string
	DetectFunc      func() bool
	IsInstalledFunc func() bool
	InstallFunc     func(adapters.InstallOpts) error
	UninstallFunc   func(adapters.InstallOpts) error
	StatusFunc      func() adapters.AdapterStatus

	// Strategy method overrides for common.Strategy interface.
	PrepareFunc  func(adapters.InstallOpts) error
	DownloadFunc func(adapters.InstallOpts) error
	VerifyFunc   func() error
	StageFunc    func(adapters.InstallOpts) error
	ApplyFunc    func(adapters.InstallOpts) error
	RollbackFunc func() error

	SkillsPathFunc       func() string
	CommandsPathFunc     func() string
	SystemPromptPathFunc func() string
	PromptStrategyFunc   func() adapters.PromptStrategy
	IDVal                string
	NameVal              string
}

func (m *MockAdapter) ID() string {
	if m.IDFunc != nil {
		return m.IDFunc()
	}
	return m.IDVal
}
func (m *MockAdapter) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return m.NameVal
}
func (m *MockAdapter) Detect() bool {
	if m.DetectFunc != nil {
		return m.DetectFunc()
	}
	return false
}
func (m *MockAdapter) IsInstalled() bool {
	if m.IsInstalledFunc != nil {
		return m.IsInstalledFunc()
	}
	return false
}
func (m *MockAdapter) Install(opts adapters.InstallOpts) error {
	if m.InstallFunc != nil {
		return m.InstallFunc(opts)
	}
	return nil
}
func (m *MockAdapter) Uninstall(opts adapters.InstallOpts) error {
	if m.UninstallFunc != nil {
		return m.UninstallFunc(opts)
	}
	return nil
}
func (m *MockAdapter) Status() adapters.AdapterStatus {
	if m.StatusFunc != nil {
		return m.StatusFunc()
	}
	return adapters.AdapterStatus{}
}

// -- common.Strategy methods ---------------------------------------------------

func (m *MockAdapter) Prepare(opts adapters.InstallOpts) error {
	if m.PrepareFunc != nil {
		return m.PrepareFunc(opts)
	}
	return nil
}
func (m *MockAdapter) Download(opts adapters.InstallOpts) error {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(opts)
	}
	return nil
}
func (m *MockAdapter) Verify() error {
	if m.VerifyFunc != nil {
		return m.VerifyFunc()
	}
	return nil
}
func (m *MockAdapter) Stage(opts adapters.InstallOpts) error {
	if m.StageFunc != nil {
		return m.StageFunc(opts)
	}
	return nil
}
func (m *MockAdapter) Apply(opts adapters.InstallOpts) error {
	if m.ApplyFunc != nil {
		return m.ApplyFunc(opts)
	}
	return nil
}
func (m *MockAdapter) Rollback() error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc()
	}
	return nil
}

func (m *MockAdapter) SkillsPath() string {
	if m.SkillsPathFunc != nil {
		return m.SkillsPathFunc()
	}
	return ""
}
func (m *MockAdapter) CommandsPath() string {
	if m.CommandsPathFunc != nil {
		return m.CommandsPathFunc()
	}
	return ""
}
func (m *MockAdapter) SystemPromptPath() string {
	if m.SystemPromptPathFunc != nil {
		return m.SystemPromptPathFunc()
	}
	return ""
}
func (m *MockAdapter) PromptStrategy() adapters.PromptStrategy {
	if m.PromptStrategyFunc != nil {
		return m.PromptStrategyFunc()
	}
	return adapters.StrategyMarkdownSections
}

var _ adapters.ToolAdapter = (*MockAdapter)(nil)
