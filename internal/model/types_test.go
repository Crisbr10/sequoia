package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/model"
)

// mockAdapter is a minimal ToolAdapter test double for model tests.
type mockAdapter struct {
	id   string
	name string
}

func (m *mockAdapter) ID() string                     { return m.id }
func (m *mockAdapter) Name() string                   { return m.name }
func (m *mockAdapter) Detect() bool                   { return false }
func (m *mockAdapter) IsInstalled() bool              { return false }
func (m *mockAdapter) Install(opts adapters.InstallOpts) error   { _ = opts.Language; return nil }
func (m *mockAdapter) Uninstall(opts adapters.InstallOpts) error { _ = opts.Language; return nil }
func (m *mockAdapter) Status() adapters.AdapterStatus { return adapters.AdapterStatus{} }
func (m *mockAdapter) SkillsPath() string             { return "" }
func (m *mockAdapter) CommandsPath() string           { return "" }
func (m *mockAdapter) SystemPromptPath() string       { return "" }
func (m *mockAdapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyMarkdownSections
}

var _ adapters.ToolAdapter = (*mockAdapter)(nil)

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
		adapter  adapters.ToolAdapter
		selected bool
		result   *model.InstallResult
	}{
		{
			name:     "unselected tool with no result",
			adapter:  &mockAdapter{id: "claude-code", name: "Claude Code"},
			selected: false,
			result:   nil,
		},
		{
			name:     "selected tool with install result",
			adapter:  &mockAdapter{id: "opencode", name: "OpenCode"},
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

			assert.Equal(t, tc.adapter, ts.Adapter)
			assert.Equal(t, tc.selected, ts.Selected)
			assert.Equal(t, tc.result, ts.Result)
		})
	}
}

func TestInstallResult_Construction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		result  model.InstallResult
		wantOK  bool
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
		name   string
		step   model.StepResult
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
