// Package app contains internal (white-box) tests for unexported functions
// in the TUI model. These tests exercise utility functions that are not
// directly testable from the external app_test package.
package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

// stubAdapter is a minimal ToolAdapter for internal unit tests.
type stubAdapter struct {
	id        string
	name      string
	installed bool
}

func (s *stubAdapter) ID() string                     { return s.id }
func (s *stubAdapter) Name() string                   { return s.name }
func (s *stubAdapter) Detect() bool                   { return s.installed }
func (s *stubAdapter) IsInstalled() bool              { return s.installed }
func (s *stubAdapter) Install(opts adapters.InstallOpts) error   { _ = opts.Language; return nil }
func (s *stubAdapter) Uninstall(opts adapters.InstallOpts) error { _ = opts.Language; return nil }
func (s *stubAdapter) Status() adapters.AdapterStatus { return adapters.AdapterStatus{} }
func (s *stubAdapter) SkillsPath() string             { return "" }
func (s *stubAdapter) CommandsPath() string           { return "" }
func (s *stubAdapter) SystemPromptPath() string       { return "" }
func (s *stubAdapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyMarkdownSections
}

var _ adapters.ToolAdapter = (*stubAdapter)(nil)

func TestHasSelectedInstalled_Empty(t *testing.T) {
	assert.False(t, hasSelectedInstalled(nil), "empty tools should return false")
	assert.False(t, hasSelectedInstalled([]model.ToolState{}), "empty slice should return false")
}

func TestHasSelectedInstalled_SelectedNotInstalled(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A", installed: false}, Selected: true},
	}
	assert.False(t, hasSelectedInstalled(tools), "selected but not installed should return false")
}

func TestHasSelectedInstalled_NotSelectedButInstalled(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A", installed: true}, Selected: false},
	}
	assert.False(t, hasSelectedInstalled(tools), "installed but not selected should return false")
}

func TestHasSelectedInstalled_SelectedAndInstalled(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A", installed: true}, Selected: true},
	}
	assert.True(t, hasSelectedInstalled(tools), "selected and installed should return true")
}

func TestHasSelectedInstalled_Mixed(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A", installed: false}, Selected: true},
		{Adapter: &stubAdapter{id: "b", name: "B", installed: true}, Selected: false},
		{Adapter: &stubAdapter{id: "c", name: "C", installed: true}, Selected: true},
	}
	assert.True(t, hasSelectedInstalled(tools), "at least one selected+installed tool should return true")
}

func TestBuildProgressTools_SingleSelected(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "tool-1", name: "Tool 1"}, Selected: true},
		{Adapter: &stubAdapter{id: "tool-2", name: "Tool 2"}, Selected: false},
	}

	result := buildProgressTools(tools)
	require.Len(t, result, 1, "only selected tools should be included")
	assert.Equal(t, "tool-1", result[0].ToolID)
	assert.Equal(t, "Tool 1", result[0].ToolName)
	assert.Len(t, result[0].Steps, 3, "should have 3 steps: Skills, Commands, System Prompt")

	// All steps should be in pending state.
	for _, step := range result[0].Steps {
		assert.Equal(t, screens.StepPending, step.Status, "all steps should start pending")
	}

	expectedSteps := []string{"Skills", "Commands", "System Prompt"}
	for i, step := range result[0].Steps {
		assert.Equal(t, expectedSteps[i], step.Name, "step names must match expected order")
	}
}

func TestBuildProgressTools_NoneSelected(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "tool-1", name: "Tool 1"}, Selected: false},
	}

	result := buildProgressTools(tools)
	assert.Empty(t, result, "no selected tools should produce empty result")
}

func TestBuildProgressTools_EmptyInput(t *testing.T) {
	result := buildProgressTools(nil)
	assert.Empty(t, result, "nil input should produce empty result")
}

func TestBuildUninstallProgressTools_OnlySelectedAndInstalled(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A", installed: true}, Selected: true},
		{Adapter: &stubAdapter{id: "b", name: "B", installed: true}, Selected: false},
		{Adapter: &stubAdapter{id: "c", name: "C", installed: false}, Selected: true},
	}

	result := buildUninstallProgressTools(tools)
	require.Len(t, result, 1, "only selected AND installed tools should be included")
	assert.Equal(t, "a", result[0].ToolID)
	assert.Equal(t, "A", result[0].ToolName)

	for _, step := range result[0].Steps {
		assert.Equal(t, screens.StepPending, step.Status, "all steps should start pending")
	}
}

func TestBuildUninstallProgressTools_NoneInstalled(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A", installed: false}, Selected: true},
	}

	result := buildUninstallProgressTools(tools)
	assert.Empty(t, result, "no installed tools should produce empty result")
}

func TestBuildUninstallProgressTools_EmptyInput(t *testing.T) {
	result := buildUninstallProgressTools(nil)
	assert.Empty(t, result, "nil input should produce empty result")
}

func TestRenderUninstallConfirm_ContainsPrompt(t *testing.T) {
	result := renderUninstallConfirm()
	assert.Contains(t, result, "Remove Sequoia", "confirmation should mention Remove Sequoia")
	assert.Contains(t, result, "y/N", "confirmation should show y/N prompt")
	assert.NotEmpty(t, result, "confirmation should not be empty")
}

func TestWaitForProgress_ReceivesMessage(t *testing.T) {
	ch := make(chan model.ProgressMsg, 1)
	expected := model.ProgressMsg{ToolID: "test", Step: "Skills", Done: true}
	ch <- expected

	cmd := waitForProgress(ch)
	require.NotNil(t, cmd, "waitForProgress should return a non-nil command")

	result := cmd()
	require.NotNil(t, result, "command should return a result when message is in channel")

	msg, ok := result.(model.ProgressMsg)
	require.True(t, ok, "result should be a ProgressMsg, got %T", result)
	assert.Equal(t, expected.ToolID, msg.ToolID)
	assert.Equal(t, expected.Step, msg.Step)
	assert.True(t, msg.Done)
}

func TestWaitForProgress_ClosedChannel_ReturnsNil(t *testing.T) {
	ch := make(chan model.ProgressMsg, 1)
	close(ch)

	cmd := waitForProgress(ch)
	require.NotNil(t, cmd, "waitForProgress should return a non-nil command even for closed channel")

	result := cmd()
	assert.Nil(t, result, "closed channel should return nil from command execution")
}

func TestCountSelected_AllSelected(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A"}, Selected: true},
		{Adapter: &stubAdapter{id: "b", name: "B"}, Selected: true},
	}
	assert.Equal(t, 2, countSelected(tools))
}

func TestCountSelected_NoneSelected(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "a", name: "A"}, Selected: false},
	}
	assert.Equal(t, 0, countSelected(tools))
}

func TestCountSelected_Empty(t *testing.T) {
	assert.Equal(t, 0, countSelected(nil))
	assert.Equal(t, 0, countSelected([]model.ToolState{}))
}

func TestBuildProgressTools_StepNamesMatchDesign(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &stubAdapter{id: "tool-1", name: "Tool 1"}, Selected: true},
	}

	result := buildProgressTools(tools)
	require.Len(t, result, 1)
	require.Len(t, result[0].Steps, 3)

	// Verify the exact step names from the design.
	assert.Equal(t, "Skills", result[0].Steps[0].Name)
	assert.Equal(t, "Commands", result[0].Steps[1].Name)
	assert.Equal(t, "System Prompt", result[0].Steps[2].Name)
}

func TestWaitForProgress_EmptyChannelThenClose(t *testing.T) {
	// Create a channel, launch a goroutine that closes it after a delay,
	// and verify waitForProgress returns nil when the channel closes.
	ch := make(chan model.ProgressMsg, 1)
	cmd := waitForProgress(ch)

	// Close the channel in a goroutine.
	go func() {
		close(ch)
	}()

	result := cmd()
	// Since we close immediately and no message is sent, result should be nil.
	assert.Nil(t, result, "closed channel with no pending messages should return nil")
}

func TestWaitForProgress_ContextCancellationIgnored(t *testing.T) {
	// waitForProgress doesn't use context directly — it blocks on channel read.
	// This test verifies that when a message is available, it's returned
	// regardless of external state.
	ch := make(chan model.ProgressMsg, 1)
	expected := model.ProgressMsg{ToolID: "ctx-test", Step: "Apply", Done: false}
	ch <- expected

	// Cancel a dummy context — waitForProgress doesn't use it.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = ctx

	cmd := waitForProgress(ch)
	result := cmd()
	msg, ok := result.(model.ProgressMsg)
	require.True(t, ok)
	assert.Equal(t, "ctx-test", msg.ToolID)
}
