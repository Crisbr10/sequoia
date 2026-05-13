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
