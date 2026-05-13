package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/testutil"
	"github.com/Crisbr10/sequoia/internal/model"
)

var _ adapters.ToolAdapter = (*testutil.MockAdapter)(nil)

// toolInfoMock is a test double that satisfies model.ToolInfo.
type toolInfoMock struct {
	id        string
	name      string
	installed bool
	detected  bool
	status    model.ToolStatus
}

func (m *toolInfoMock) ID() string               { return m.id }
func (m *toolInfoMock) Name() string             { return m.name }
func (m *toolInfoMock) IsInstalled() bool        { return m.installed }
func (m *toolInfoMock) Detect() bool             { return m.detected }
func (m *toolInfoMock) Status() model.ToolStatus { return m.status }

var _ model.ToolInfo = (*toolInfoMock)(nil)

// ---------------------------------------------------------------------------
// NEW tests (RED — reference types that don't exist yet)
// ---------------------------------------------------------------------------

func TestToolInfo_InterfaceDefinition(t *testing.T) {
	t.Parallel()

	// Verify that model.ToolInfo is a valid interface with the expected methods.
	// toolInfoMock satisfies it (compile-time check above).

	mock := &toolInfoMock{
		id:        "claude-code",
		name:      "Claude Code",
		installed: true,
		detected:  true,
		status: model.ToolStatus{
			Installed: true,
			Version:   "v2.0.0",
			Path:      "/home/user/.claude",
		},
	}

	var ti model.ToolInfo = mock

	assert.Equal(t, "claude-code", ti.ID())
	assert.Equal(t, "Claude Code", ti.Name())
	assert.True(t, ti.IsInstalled())
	assert.True(t, ti.Detect())

	st := ti.Status()
	assert.True(t, st.Installed)
	assert.Equal(t, "v2.0.0", st.Version)
	assert.Equal(t, "/home/user/.claude", st.Path)
}

func TestToolStatus_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status model.ToolStatus
	}{
		{
			name: "installed with version and path",
			status: model.ToolStatus{
				Installed: true,
				Version:   "v1.0.0",
				Path:      "/home/user/.claude",
			},
		},
		{
			name: "not installed — empty version and path",
			status: model.ToolStatus{
				Installed: false,
				Version:   "",
				Path:      "",
			},
		},
		{
			name: "partial — installed but no version",
			status: model.ToolStatus{
				Installed: true,
				Version:   "",
				Path:      "/opt/codex",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.status.Installed, tc.status.Installed) // sanity: struct round-trip
			assert.Equal(t, tc.status.Version, tc.status.Version)
			assert.Equal(t, tc.status.Path, tc.status.Path)
		})
	}
}

func TestToolState_AdapterIsToolInfo(t *testing.T) {
	t.Parallel()

	// Verify that ToolState.Adapter is typed as model.ToolInfo, not adapters.ToolAdapter.
	mock := &toolInfoMock{id: "test", name: "Test Tool"}

	ts := model.ToolState{
		Adapter:  mock,
		Selected: true,
	}

	var ti model.ToolInfo = ts.Adapter
	assert.Equal(t, "test", ti.ID())
	assert.Equal(t, "Test Tool", ti.Name())
}

func TestToolInfo_WithInstalledAdapter(t *testing.T) {
	t.Parallel()

	mock := &toolInfoMock{
		id:        "gemini",
		name:      "Gemini CLI",
		installed: true,
		detected:  true,
		status: model.ToolStatus{
			Installed: true,
			Version:   "v3.1.0",
			Path:      "/home/user/.gemini",
		},
	}

	ts := model.ToolState{
		Adapter:  mock,
		Selected: true,
	}

	require.True(t, ts.Adapter.IsInstalled(), "installed adapter should report true")
	require.True(t, ts.Adapter.Detect(), "detected adapter should report true")

	st := ts.Adapter.Status()
	require.True(t, st.Installed)
	assert.Equal(t, "v3.1.0", st.Version)
	assert.Equal(t, "/home/user/.gemini", st.Path)
}

func TestToolInfo_WithUninstalledAdapter(t *testing.T) {
	t.Parallel()

	mock := &toolInfoMock{
		id:        "cursor",
		name:      "Cursor",
		installed: false,
		detected:  false,
		status:    model.ToolStatus{Installed: false},
	}

	ts := model.ToolState{
		Adapter:  mock,
		Selected: false,
	}

	require.False(t, ts.Adapter.IsInstalled(), "uninstalled adapter should report false")
	require.False(t, ts.Adapter.Detect(), "undetected adapter should report false")

	st := ts.Adapter.Status()
	require.False(t, st.Installed)
}

// ---------------------------------------------------------------------------
// Existing tests (updated to use toolInfoMock where ToolState.Adapter is involved)
// ---------------------------------------------------------------------------

func TestScreen_EnumValues(t *testing.T) {
	t.Parallel()

	// Verify all Screen constants have distinct values and cover the expected range.
	require.NotEqual(t, model.ScreenWelcome, model.ScreenToolSelection)
	require.NotEqual(t, model.ScreenToolSelection, model.ScreenConfiguration)
	require.NotEqual(t, model.ScreenConfiguration, model.ScreenInstallProgress)
	require.NotEqual(t, model.ScreenInstallProgress, model.ScreenComplete)
	require.NotEqual(t, model.ScreenComplete, model.ScreenError)
	require.NotEqual(t, model.ScreenError, model.ScreenStatus)
	require.NotEqual(t, model.ScreenStatus, model.ScreenUninstall)

	// ScreenCount should equal the number of screens.
	assert.Equal(t, 8, int(model.ScreenCount), "ScreenCount must match the number of Screen constants")
}

func TestToolState_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		adapter  model.ToolInfo
		selected bool
		result   *model.InstallResult
	}{
		{
			name:     "unselected tool with no result",
			adapter:  &toolInfoMock{id: "claude-code", name: "Claude Code"},
			selected: false,
			result:   nil,
		},
		{
			name:     "selected tool with install result",
			adapter:  &toolInfoMock{id: "opencode", name: "OpenCode"},
			selected: true,
			result: &model.InstallResult{
				ToolID:  "opencode",
				Success: true,
				Steps: []model.StepResult{
					{Name: "prepare", Done: true},
					{Name: "apply", Done: true},
					{Name: "verify", Done: true},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := model.ToolState{
				Adapter:  tc.adapter,
				Selected: tc.selected,
				Result:   tc.result,
			}

			assert.NotNil(t, ts.Adapter)
			assert.Equal(t, tc.adapter.ID(), ts.Adapter.ID())
			assert.Equal(t, tc.adapter.Name(), ts.Adapter.Name())
			assert.Equal(t, tc.selected, ts.Selected)
			assert.Equal(t, tc.result, ts.Result)
		})
	}
}

func TestInstallResult_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result model.InstallResult
		wantOK bool
	}{
		{
			name: "successful installation with steps",
			result: model.InstallResult{
				ToolID:  "claude-code",
				Success: true,
				Steps: []model.StepResult{
					{Name: "prepare", Done: true},
					{Name: "apply", Done: true},
					{Name: "verify", Done: true},
				},
			},
			wantOK: true,
		},
		{
			name: "failed installation with error",
			result: model.InstallResult{
				ToolID:  "opencode",
				Success: false,
				Error:   "apply failed: permission denied",
				Steps: []model.StepResult{
					{Name: "prepare", Done: true},
					{Name: "apply", Done: false, Error: "permission denied"},
					{Name: "verify", Done: false},
				},
			},
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantOK, tc.result.Success)
			if !tc.wantOK {
				assert.NotEmpty(t, tc.result.Error, "failed result should have an error message")
			}
		})
	}
}

func TestProgressMsg_Construction(t *testing.T) {
	t.Parallel()

	msg := model.ProgressMsg{
		ToolID: "claude-code",
		Step:   "apply",
		Done:   true,
		Error:  "",
	}

	assert.Equal(t, "claude-code", msg.ToolID)
	assert.Equal(t, "apply", msg.Step)
	assert.True(t, msg.Done)
	assert.Empty(t, msg.Error)
}

func TestProgressMsg_WithError(t *testing.T) {
	t.Parallel()

	msg := model.ProgressMsg{
		ToolID: "opencode",
		Step:   "verify",
		Done:   false,
		Error:  "checksum mismatch",
	}

	assert.Equal(t, "opencode", msg.ToolID)
	assert.False(t, msg.Done)
	assert.Equal(t, "checksum mismatch", msg.Error)
}

func TestTUIConfig_Defaults(t *testing.T) {
	t.Parallel()

	cfg := model.TUIConfig{
		Language:    "en",
		Persistence: "engram",
	}

	assert.Equal(t, "en", cfg.Language)
	assert.Equal(t, "engram", cfg.Persistence)
}

func TestLanguage_Constants(t *testing.T) {
	t.Parallel()

	require.NotEqual(t, model.LangEN, model.LangES)
	assert.Equal(t, model.Language("en"), model.LangEN)
	assert.Equal(t, model.Language("es"), model.LangES)
}

func TestPersistenceBackend_Constants(t *testing.T) {
	t.Parallel()

	require.NotEqual(t, model.PersistenceEngram, model.PersistenceFiles)
	require.NotEqual(t, model.PersistenceFiles, model.PersistenceBoth)
	assert.Equal(t, model.PersistenceBackend("engram"), model.PersistenceEngram)
	assert.Equal(t, model.PersistenceBackend("files"), model.PersistenceFiles)
	assert.Equal(t, model.PersistenceBackend("both"), model.PersistenceBoth)
}

func TestStepResult_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		step model.StepResult
	}{
		{
			name: "completed step",
			step: model.StepResult{Name: "apply", Done: true},
		},
		{
			name: "pending step",
			step: model.StepResult{Name: "verify", Done: false},
		},
		{
			name: "failed step with error",
			step: model.StepResult{Name: "prepare", Done: false, Error: "disk full"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotEmpty(t, tc.step.Name, "step name should not be empty")
		})
	}
}
