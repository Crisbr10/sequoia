package testutil_test

import (
	"errors"
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

// -- close-coverage-gaps REQ-COV-02 -------------------------------------------
//
// TestMockAdapter_StrategyMethods closes the 1pp coverage gap in
// adapters/testutil by exercising both branches of each common.Strategy
// method on MockAdapter: the nil-Func default (returns nil) and the
// Func-override path (propagates the function's error).
//
// 6 subtests cover: Prepare, Download, Verify, Stage, Apply, Rollback.
// 4 take adapters.InstallOpts: Prepare, Download, Stage, Apply.
// 2 take no args: Verify, Rollback.

// TestMockAdapter_StrategyMethods verifies the dual-branch (nil-Func /
// Func-override) pattern of each common.Strategy method on MockAdapter.
// Closes REQ-COV-02: brings adapters/testutil from 69.0% to >= 70.0%.
func TestMockAdapter_StrategyMethods(t *testing.T) {
	t.Run("Prepare", func(t *testing.T) {
		// nil Func: returns nil (mock_adapter.go:206-211).
		m := &testutil.MockAdapter{}
		assert.NoError(t, m.Prepare(adapters.InstallOpts{}),
			"nil PrepareFunc should return nil")

		// Func override: propagates the function's error.
		expected := errors.New("prepare boom")
		m = &testutil.MockAdapter{
			PrepareFunc: func(_ adapters.InstallOpts) error { return expected },
		}
		assert.ErrorIs(t, m.Prepare(adapters.InstallOpts{}), expected,
			"PrepareFunc override should propagate its error")
	})

	t.Run("Download", func(t *testing.T) {
		// nil Func: returns nil (mock_adapter.go:212-217).
		m := &testutil.MockAdapter{}
		assert.NoError(t, m.Download(adapters.InstallOpts{}),
			"nil DownloadFunc should return nil")

		// Func override: propagates the function's error.
		expected := errors.New("download boom")
		m = &testutil.MockAdapter{
			DownloadFunc: func(_ adapters.InstallOpts) error { return expected },
		}
		assert.ErrorIs(t, m.Download(adapters.InstallOpts{}), expected,
			"DownloadFunc override should propagate its error")
	})

	t.Run("Verify", func(t *testing.T) {
		// nil Func: returns nil (mock_adapter.go:218-223).
		m := &testutil.MockAdapter{}
		assert.NoError(t, m.Verify(),
			"nil VerifyFunc should return nil")

		// Func override: propagates the function's error.
		expected := errors.New("verify boom")
		m = &testutil.MockAdapter{
			VerifyFunc: func() error { return expected },
		}
		assert.ErrorIs(t, m.Verify(), expected,
			"VerifyFunc override should propagate its error")
	})

	t.Run("Stage", func(t *testing.T) {
		// nil Func: returns nil (mock_adapter.go:224-229).
		m := &testutil.MockAdapter{}
		assert.NoError(t, m.Stage(adapters.InstallOpts{}),
			"nil StageFunc should return nil")

		// Func override: propagates the function's error.
		expected := errors.New("stage boom")
		m = &testutil.MockAdapter{
			StageFunc: func(_ adapters.InstallOpts) error { return expected },
		}
		assert.ErrorIs(t, m.Stage(adapters.InstallOpts{}), expected,
			"StageFunc override should propagate its error")
	})

	t.Run("Apply", func(t *testing.T) {
		// nil Func: returns nil (mock_adapter.go:230-235).
		m := &testutil.MockAdapter{}
		assert.NoError(t, m.Apply(adapters.InstallOpts{}),
			"nil ApplyFunc should return nil")

		// Func override: propagates the function's error.
		expected := errors.New("apply boom")
		m = &testutil.MockAdapter{
			ApplyFunc: func(_ adapters.InstallOpts) error { return expected },
		}
		assert.ErrorIs(t, m.Apply(adapters.InstallOpts{}), expected,
			"ApplyFunc override should propagate its error")
	})

	t.Run("Rollback", func(t *testing.T) {
		// nil Func: returns nil (mock_adapter.go:236-241).
		m := &testutil.MockAdapter{}
		assert.NoError(t, m.Rollback(),
			"nil RollbackFunc should return nil")

		// Func override: propagates the function's error.
		expected := errors.New("rollback boom")
		m = &testutil.MockAdapter{
			RollbackFunc: func() error { return expected },
		}
		assert.ErrorIs(t, m.Rollback(), expected,
			"RollbackFunc override should propagate its error")
	})
}
