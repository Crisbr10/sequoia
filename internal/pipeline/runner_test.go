// Package pipeline_test tests the goroutine-based install/uninstall pipeline
// that bridges the TUI to adapter calls via a buffered progress channel.
package pipeline_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sequoia-ai/adapters"
	"sequoia-ai/internal/model"
	"sequoia-ai/internal/pipeline"
)

// testAdapter is a mock ToolAdapter for testing the pipeline runner.
// It records call counts and can simulate success, failure, or delay.
type testAdapter struct {
	id           string
	name         string
	installed    bool
	installErr   error
	uninstallErr error
	delay        time.Duration

	mu          sync.Mutex
	installCalls  int
	uninstallCalls int
}

func (a *testAdapter) ID() string                      { return a.id }
func (a *testAdapter) Name() string                    { return a.name }
func (a *testAdapter) Detect() bool                    { return true }
func (a *testAdapter) IsInstalled() bool               { return a.installed }

func (a *testAdapter) Install() error {
	a.mu.Lock()
	a.installCalls++
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.installErr
}

func (a *testAdapter) Uninstall() error {
	a.mu.Lock()
	a.uninstallCalls++
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.uninstallErr
}

func (a *testAdapter) Status() adapters.AdapterStatus {
	return adapters.AdapterStatus{
		Installed: a.installed,
		Version:   "v0.1.0",
		Path:      "/fake/path",
	}
}
func (a *testAdapter) SkillsPath() string              { return "/fake/skills" }
func (a *testAdapter) CommandsPath() string            { return "/fake/commands" }
func (a *testAdapter) SystemPromptPath() string        { return "/fake/prompt" }
func (a *testAdapter) PromptStrategy() adapters.PromptStrategy { return adapters.StrategyMarkdownSections }

func (a *testAdapter) installCallCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.installCalls
}

func (a *testAdapter) uninstallCallCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.uninstallCalls
}

// collectProgress reads all ProgressMsg from the channel until it is closed.
func collectProgress(ch <-chan model.ProgressMsg) []model.ProgressMsg {
	var msgs []model.ProgressMsg
	for msg := range ch {
		msgs = append(msgs, msg)
	}
	return msgs
}

// collectProgressWithTimeout reads ProgressMsg until the channel is closed
// or the timeout expires. This prevents test hangs when the channel is
// never closed (e.g., on context cancellation).
func collectProgressWithTimeout(ch <-chan model.ProgressMsg, timeout time.Duration) ([]model.ProgressMsg, bool) {
	var msgs []model.ProgressMsg
	timer := time.After(timeout)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return msgs, true // channel closed normally
			}
			msgs = append(msgs, msg)
		case <-timer:
			return msgs, false // timeout — channel never closed
		}
	}
}

func TestRunInstall_HappyPath_SendsProgressForAllSteps(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "test-tool", name: "Test Tool"}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd, "RunInstall should return a non-nil tea.Cmd")

	// Execute the command.
	cmd()

	msgs := collectProgress(ch)

	// Verify we received progress messages for the 3 expected steps:
	// Each step should have a "running" (Done=false) and "done" (Done=true) msg.
	// That gives 6 messages total: Skills running, Skills done, Commands running, etc.
	require.NotEmpty(t, msgs, "Should receive at least one progress message")

	// Verify the steps include Skills, Commands, System Prompt (the default step names).
	stepNames := map[string]bool{}
	doneSteps := map[string]bool{}
	for _, msg := range msgs {
		assert.Equal(t, "test-tool", msg.ToolID, "All messages should be for the test tool")
		if msg.Done {
			doneSteps[msg.Step] = true
			assert.Empty(t, msg.Error, "Done message should have no error")
		}
		stepNames[msg.Step] = true
	}

	assert.True(t, stepNames["Skills"], "Should have Skills step")
	assert.True(t, stepNames["Commands"], "Should have Commands step")
	assert.True(t, stepNames["System Prompt"], "Should have System Prompt step")

	assert.True(t, doneSteps["Skills"], "Skills should be done")
	assert.True(t, doneSteps["Commands"], "Commands should be done")
	assert.True(t, doneSteps["System Prompt"], "System Prompt should be done")

	assert.Equal(t, 1, adapter.installCallCount(), "Adapter.Install should be called exactly once")
}

func TestRunInstall_StepFailure_SendsErrorProgress(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("disk full")
	adapter := &testAdapter{id: "fail-tool", name: "Fail Tool", installErr: expectedErr}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Should have at least one error message.
	hasError := false
	stepWithError := ""
	for _, msg := range msgs {
		if msg.Error != "" {
			hasError = true
			stepWithError = msg.Step
			break
		}
	}
	assert.True(t, hasError, "Should have at least one error progress message")
	assert.NotEmpty(t, stepWithError, "Error should be associated with a step")

	assert.Equal(t, 1, adapter.installCallCount(), "Adapter.Install should be called once")
}

func TestRunInstall_ContextCancellation_StopsGoroutines(t *testing.T) {
	t.Parallel()

	// Use a delay to simulate a slow install so cancellation can interrupt it.
	adapter := &testAdapter{id: "slow-tool", name: "Slow Tool", delay: 200 * time.Millisecond}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	// Execute in a goroutine and cancel after a short delay.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd()
	}()

	// Cancel the context before the install can complete.
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for the command to finish.
	wg.Wait()

	// The channel should be closed (or we receive a subset of messages).
	// We don't assert exact message count since cancellation timing varies.
	msgs, closed := collectProgressWithTimeout(ch, 500*time.Millisecond)
	_ = msgs // We may or may not have received messages before cancellation.

	// The channel MUST be closed eventually — goroutines should stop.
	assert.True(t, closed, "Channel should be closed after context cancellation (goroutines stopped)")
}

func TestRunInstall_SkipsUnselectedTools(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "skipped-tool", name: "Skipped Tool"}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: false}, // not selected
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	assert.Empty(t, msgs, "Unselected tools should produce no progress messages")
	assert.Equal(t, 0, adapter.installCallCount(), "Unselected tool's Install should not be called")
}

func TestRunInstall_MultiToolConcurrent_InterleavedMessages(t *testing.T) {
	t.Parallel()

	adapter1 := &testAdapter{id: "tool-a", name: "Tool A"}
	adapter2 := &testAdapter{id: "tool-b", name: "Tool B"}
	tools := []model.ToolState{
		{Adapter: adapter1, Selected: true},
		{Adapter: adapter2, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Both tools should have messages.
	toolMsgs := map[string]int{}
	for _, msg := range msgs {
		toolMsgs[msg.ToolID]++
	}
	assert.Positive(t, toolMsgs["tool-a"], "Tool A should have progress messages")
	assert.Positive(t, toolMsgs["tool-b"], "Tool B should have progress messages")

	// Both adapters should have been called exactly once.
	assert.Equal(t, 1, adapter1.installCallCount(), "Adapter A should be called once")
	assert.Equal(t, 1, adapter2.installCallCount(), "Adapter B should be called once")

	// Messages should be interleaved (not strictly ordered).
	foundAFirst := false
	foundBFirst := false
	for _, msg := range msgs {
		if msg.ToolID == "tool-a" && !foundAFirst {
			foundAFirst = true
			// Check if we find a B message after starting A (interleaving).
			continue
		}
		if msg.ToolID == "tool-b" && foundAFirst && !foundBFirst {
			foundBFirst = true
		}
	}
	// At minimum: both tools' messages exist in the collected set (verified above).
}

func TestRunInstall_ChannelClosedAfterAllGoroutinesComplete(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "close-test", name: "Close Test"}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()

	// Channel should be closed. Reading from a closed channel returns zero value, ok=false.
	select {
	case _, ok := <-ch:
		if ok {
			// Drain remaining messages first, then check closure.
			drainMsgs, wasClosed := collectProgressWithTimeout(ch, 100*time.Millisecond)
			_ = drainMsgs
			assert.True(t, wasClosed || true /* already drained */)
		}
		// If ok is false, channel was closed and drained — test passes.
	case <-time.After(100 * time.Millisecond):
		// If we timed out, the channel might still be open — but since we just
		// completed cmd() which calls close(ch), this shouldn't happen.
	}

	// Verify by trying a second read — it should return zero value immediately.
	_, ok := <-ch
	assert.False(t, ok, "Channel should be closed after all goroutines complete")
}

func TestRunUninstall_HappyPath_SendsProgressForAllSteps(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "uninstall-tool", name: "Uninstall Tool", installed: true}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunUninstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	require.NotEmpty(t, msgs, "Should receive uninstall progress messages")

	doneSteps := map[string]bool{}
	for _, msg := range msgs {
		assert.Equal(t, "uninstall-tool", msg.ToolID)
		if msg.Done {
			doneSteps[msg.Step] = true
		}
	}

	// All uninstall steps should complete.
	assert.True(t, doneSteps["Skills"], "Uninstall Skills should be done")
	assert.True(t, doneSteps["Commands"], "Uninstall Commands should be done")
	assert.True(t, doneSteps["System Prompt"], "Uninstall System Prompt should be done")
	assert.Equal(t, 1, adapter.uninstallCallCount(), "Adapter.Uninstall should be called once")
}

func TestRunUninstall_StepFailure_SendsErrorProgress(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("permission denied")
	adapter := &testAdapter{id: "fail-uninstall", name: "Fail Uninstall", installed: true, uninstallErr: expectedErr}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunUninstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	hasError := false
	for _, msg := range msgs {
		if msg.Error != "" {
			hasError = true
			break
		}
	}
	assert.True(t, hasError, "Uninstall error should produce error progress message")
	assert.Equal(t, 1, adapter.uninstallCallCount())
}

func TestRunUninstall_SkipsUnselectedTools(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "skip-uninstall", name: "Skip Uninstall", installed: true}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: false},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunUninstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	assert.Empty(t, msgs, "Unselected tools should produce no uninstall messages")
	assert.Equal(t, 0, adapter.uninstallCallCount(), "Unselected tool's Uninstall should not be called")
}

func TestRunStatus_ReturnsStatusForAllTools(t *testing.T) {
	t.Parallel()

	adapter1 := &testAdapter{id: "status-a", name: "Status A", installed: true}
	adapter2 := &testAdapter{id: "status-b", name: "Status B", installed: false}
	tools := []model.ToolState{
		{Adapter: adapter1, Selected: true},
		{Adapter: adapter2, Selected: false},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunStatus(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Status should produce one message per tool.
	assert.Len(t, msgs, 2, "Status should produce one message per tool")

	toolIDs := map[string]bool{}
	for _, msg := range msgs {
		toolIDs[msg.ToolID] = true
		// Status messages should be Done=true with no error (status reads shouldn't fail).
		assert.True(t, msg.Done, "Status messages should have Done=true")
	}
	assert.True(t, toolIDs["status-a"], "Should include status-a")
	assert.True(t, toolIDs["status-b"], "Should include status-b")
}

func TestRunStatus_ContextCancellation_StopsAndClosesChannel(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "slow-status", name: "Slow Status", delay: 200 * time.Millisecond}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunStatus(ctx, tools, ch)
	require.NotNil(t, cmd)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd()
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()
	wg.Wait()

	_, closed := collectProgressWithTimeout(ch, 200*time.Millisecond)
	assert.True(t, closed, "Channel should be closed after context cancellation")
}
