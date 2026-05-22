package testutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/testutil"
)

func TestMockAdapter_Defaults(t *testing.T) {
	m := &testutil.MockAdapter{
		IDVal:   "default",
		NameVal: "Default Name",
	}
	assert.Equal(t, "default", m.ID())
	assert.Equal(t, "Default Name", m.Name())
	assert.False(t, m.Detect())
	assert.False(t, m.IsInstalled())
	assert.NoError(t, m.Install(adapters.InstallOpts{}))
	assert.NoError(t, m.Uninstall(adapters.InstallOpts{}))
	assert.Equal(t, adapters.AdapterStatus{}, m.Status())
	assert.Equal(t, "", m.SkillsPath())
	assert.Equal(t, "", m.CommandsPath())
	assert.Equal(t, "", m.SystemPromptPath())
	assert.Equal(t, adapters.StrategyMarkdownSections, m.PromptStrategy())
}

func TestMockAdapter_FunctionOverrides(t *testing.T) {
	m := &testutil.MockAdapter{
		DetectFunc:      func() bool { return true },
		IsInstalledFunc: func() bool { return true },
		InstallFunc:     func(_ adapters.InstallOpts) error { return nil },
		StatusFunc: func() adapters.AdapterStatus {
			return adapters.AdapterStatus{Installed: true, Version: "v1.0", Path: "/test"}
		},
	}
	assert.True(t, m.Detect())
	assert.True(t, m.IsInstalled())
	st := m.Status()
	assert.True(t, st.Installed)
	assert.Equal(t, "v1.0", st.Version)
	assert.Equal(t, "/test", st.Path)
}

func TestMockAdapter_SatisfiesToolAdapter(t *testing.T) {
	var _ adapters.ToolAdapter = (*testutil.MockAdapter)(nil)
}

// P3-003 ISP narrowing: verify each focused mock satisfies its role interface.

func TestMockIdentifier_SatisfiesIdentifier(t *testing.T) {
	var _ adapters.Identifier = (*testutil.MockIdentifier)(nil)
}

func TestMockDetector_SatisfiesDetector(t *testing.T) {
	var _ adapters.Detector = (*testutil.MockDetector)(nil)
}

func TestMockInstaller_SatisfiesInstaller(t *testing.T) {
	var _ adapters.Installer = (*testutil.MockInstaller)(nil)
}

func TestMockInstallStatus_SatisfiesInstallStatus(t *testing.T) {
	var _ adapters.InstallStatus = (*testutil.MockInstallStatus)(nil)
}

func TestMockAdapterPaths_SatisfiesAdapterPaths(t *testing.T) {
	var _ adapters.AdapterPaths = (*testutil.MockAdapterPaths)(nil)
}

func TestMockIdentifier_Defaults(t *testing.T) {
	m := &testutil.MockIdentifier{IDVal: "test-id", NameVal: "Test Name"}
	assert.Equal(t, "test-id", m.ID())
	assert.Equal(t, "Test Name", m.Name())
}

func TestMockIdentifier_Functions(t *testing.T) {
	m := &testutil.MockIdentifier{
		IDFunc:   func() string { return "func-id" },
		NameFunc: func() string { return "Func Name" },
	}
	assert.Equal(t, "func-id", m.ID())
	assert.Equal(t, "Func Name", m.Name())
}

func TestMockDetector_Defaults(t *testing.T) {
	m := &testutil.MockDetector{}
	assert.False(t, m.Detect())
	assert.False(t, m.IsInstalled())
}

func TestMockDetector_Functions(t *testing.T) {
	m := &testutil.MockDetector{
		DetectFunc:      func() bool { return true },
		IsInstalledFunc: func() bool { return true },
	}
	assert.True(t, m.Detect())
	assert.True(t, m.IsInstalled())
}

func TestMockInstaller_Defaults(t *testing.T) {
	m := &testutil.MockInstaller{}
	assert.NoError(t, m.Install(adapters.InstallOpts{}))
	assert.NoError(t, m.Uninstall(adapters.InstallOpts{}))
}

func TestMockInstaller_Functions(t *testing.T) {
	errInstall := assert.AnError
	m := &testutil.MockInstaller{
		InstallFunc:   func(_ adapters.InstallOpts) error { return errInstall },
		UninstallFunc: func(_ adapters.InstallOpts) error { return nil },
	}
	assert.Error(t, m.Install(adapters.InstallOpts{}))
	assert.NoError(t, m.Uninstall(adapters.InstallOpts{}))
}

func TestMockInstallStatus_Defaults(t *testing.T) {
	m := &testutil.MockInstallStatus{}
	assert.Equal(t, adapters.AdapterStatus{}, m.Status())
}

func TestMockInstallStatus_Functions(t *testing.T) {
	m := &testutil.MockInstallStatus{
		StatusFunc: func() adapters.AdapterStatus {
			return adapters.AdapterStatus{Installed: true, Version: "v2.0", Path: "/p"}
		},
	}
	st := m.Status()
	assert.True(t, st.Installed)
	assert.Equal(t, "v2.0", st.Version)
	assert.Equal(t, "/p", st.Path)
}

func TestMockAdapterPaths_Defaults(t *testing.T) {
	m := &testutil.MockAdapterPaths{}
	assert.Equal(t, "", m.SkillsPath())
	assert.Equal(t, "", m.CommandsPath())
	assert.Equal(t, "", m.SystemPromptPath())
	assert.Equal(t, adapters.StrategyMarkdownSections, m.PromptStrategy())
}

func TestMockAdapterPaths_Functions(t *testing.T) {
	m := &testutil.MockAdapterPaths{
		SkillsPathFunc:       func() string { return "/s" },
		CommandsPathFunc:     func() string { return "/c" },
		SystemPromptPathFunc: func() string { return "/p" },
		PromptStrategyFunc:   func() adapters.PromptStrategy { return adapters.StrategyFileReplace },
	}
	assert.Equal(t, "/s", m.SkillsPath())
	assert.Equal(t, "/c", m.CommandsPath())
	assert.Equal(t, "/p", m.SystemPromptPath())
	assert.Equal(t, adapters.StrategyFileReplace, m.PromptStrategy())
}
