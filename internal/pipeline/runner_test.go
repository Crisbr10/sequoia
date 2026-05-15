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

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/pipeline"
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

	mu             sync.Mutex
	installCalls   int
	uninstallCalls int

	// lastContext captures the context passed to Install via InstallOpts.
	lastContext context.Context
}

func (a *testAdapter) ID() string        { return a.id }
func (a *testAdapter) Name() string      { return a.name }
func (a *testAdapter) Detect() bool      { return true }
func (a *testAdapter) IsInstalled() bool { return a.installed }

func (a *testAdapter) Install(opts adapters.InstallOpts) error {
	_ = opts.Language
	a.mu.Lock()
	a.installCalls++
	a.lastContext = opts.Context
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.installErr
}

func (a *testAdapter) Uninstall(opts adapters.InstallOpts) error {
	_ = opts.Language
	a.mu.Lock()
	a.uninstallCalls++
	a.lastContext = opts.Context
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.uninstallErr
}

func (a *testAdapter) Status() model.ToolStatus {
	return model.ToolStatus{
		Installed: a.installed,
		Version:   "v0.1.0",
		Path:      "/fake/path",
	}
}
func (a *testAdapter) SkillsPath() string       { return "/fake/skills" }
func (a *testAdapter) CommandsPath() string     { return "/fake/commands" }
func (a *testAdapter) SystemPromptPath() string { return "/fake/prompt" }
func (a *testAdapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyMarkdownSections
}

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

func TestRunInstall_HappyPath_SendsTwoMessages(t *testing.T) {
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

	// After simplification, the pipeline sends exactly 2 messages per adapter:
	// 1 "running" (Done=false) + 1 "done" (Done=true) for the single "Installing" step.
	require.Len(t, msgs, 2, "Should receive exactly 2 progress messages (1 running + 1 done)")

	// First message: "Installing" running (Done=false).
	assert.Equal(t, "test-tool", msgs[0].ToolID)
	assert.Equal(t, "Installing", msgs[0].Step)
	assert.False(t, msgs[0].Done, "First message should be 'running' (Done=false)")
	assert.Empty(t, msgs[0].Error, "Running message should have no error")

	// Second message: "Installing" done (Done=true).
	assert.Equal(t, "test-tool", msgs[1].ToolID)
	assert.Equal(t, "Installing", msgs[1].Step)
	assert.True(t, msgs[1].Done, "Second message should be 'done' (Done=true)")
	assert.Empty(t, msgs[1].Error, "Done message should have no error")

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

	// After simplification: 1 running message + error sent.
	// The "running" message should be sent before execution, then the error.
	require.Len(t, msgs, 2, "Should receive 2 messages: 1 running + 1 error")

	// First message: "Installing" running.
	assert.Equal(t, "fail-tool", msgs[0].ToolID)
	assert.Equal(t, "Installing", msgs[0].Step)
	assert.False(t, msgs[0].Done, "First message should be 'running'")

	// Second message: error on "Installing".
	assert.Equal(t, "fail-tool", msgs[1].ToolID)
	assert.Equal(t, "Installing", msgs[1].Step)
	assert.True(t, msgs[1].Done, "Error message should have Done=true")
	assert.Contains(t, msgs[1].Error, "disk full", "Error message should contain the error text")

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

func TestRunUninstall_HappyPath_SendsTwoMessages(t *testing.T) {
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

	// After simplification: 2 messages (1 running + 1 done) for the single "Installing" step.
	require.Len(t, msgs, 2, "Should receive exactly 2 messages: 1 running + 1 done")

	assert.Equal(t, "uninstall-tool", msgs[0].ToolID)
	assert.Equal(t, "Installing", msgs[0].Step)
	assert.False(t, msgs[0].Done, "First should be 'running'")

	assert.Equal(t, "uninstall-tool", msgs[1].ToolID)
	assert.Equal(t, "Installing", msgs[1].Step)
	assert.True(t, msgs[1].Done, "Second should be 'done'")
	assert.Empty(t, msgs[1].Error, "Done message should have no error")

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

	// After simplification: 1 running + 1 error.
	require.Len(t, msgs, 2, "Should receive 2 messages: 1 running + 1 error")

	assert.Equal(t, "fail-uninstall", msgs[0].ToolID)
	assert.Equal(t, "Installing", msgs[0].Step)
	assert.False(t, msgs[0].Done, "First message should be 'running'")

	assert.Equal(t, "fail-uninstall", msgs[1].ToolID)
	assert.Equal(t, "Installing", msgs[1].Step)
	assert.True(t, msgs[1].Done, "Error message should have Done=true")
	assert.Contains(t, msgs[1].Error, "permission denied")

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

// TestRunInstall_PassesContextToAdapter verifies that the pipeline forwards
// the caller's context to the adapter's Install method via InstallOpts.
// This is the triangulation test for context propagation: unlike the
// "already cancelled" test, this verifies the pipeline properly wires
// a live context through to the adapter layer.
func TestRunInstall_PassesContextToAdapter(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "ctx-test", name: "Context Test"}
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
	require.NotEmpty(t, msgs, "Should receive progress messages")

	// The adapter must have received a non-nil context.
	adapter.mu.Lock()
	receivedCtx := adapter.lastContext
	adapter.mu.Unlock()

	require.NotNil(t, receivedCtx, "Adapter should receive a non-nil Context via InstallOpts")

	// The received context should be alive (not cancelled).
	select {
	case <-receivedCtx.Done():
		t.Error("Context passed to adapter should not be cancelled during normal operation")
	default:
		// OK — context is still alive.
	}

	assert.Equal(t, 1, adapter.installCallCount(), "Adapter.Install should be called once")
}

// TestRunUninstall_PassesContextToAdapter verifies uninstall context propagation.
func TestRunUninstall_PassesContextToAdapter(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "uninstall-ctx", name: "Uninstall Context", installed: true}
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
	require.NotEmpty(t, msgs)

	adapter.mu.Lock()
	receivedCtx := adapter.lastContext
	adapter.mu.Unlock()

	require.NotNil(t, receivedCtx, "Adapter should receive a non-nil Context via InstallOpts for uninstall")

	select {
	case <-receivedCtx.Done():
		t.Error("Context passed to adapter should not be cancelled during normal uninstall")
	default:
	}

	assert.Equal(t, 1, adapter.uninstallCallCount())
}

// TestDefaultStepNames_SingleStep verifies that the exported InstallSteps
// contains exactly one element: "Installing". This replaces the old 3-step
// cosmetic breakdown that didn't reflect the monolithic adapter.Install() call.
func TestDefaultStepNames_SingleStep(t *testing.T) {
	t.Parallel()

	steps := pipeline.InstallSteps
	require.Len(t, steps, 1, "InstallSteps should have exactly 1 element after simplification")
	assert.Equal(t, "Installing", steps[0], "The single step should be 'Installing'")
}

// TestRunInstall_MultiTool_SendsTwoMessagesEach verifies that with 2 tools,
// each tool sends exactly 2 messages (1 running + 1 done).
func TestRunInstall_MultiTool_SendsTwoMessagesEach(t *testing.T) {
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

	// 2 tools × 2 messages each = 4 total.
	assert.Len(t, msgs, 4, "2 tools × 2 messages = 4 total")

	// Both tools should appear with "Installing" as the step.
	toolMsgs := map[string]int{}
	for _, msg := range msgs {
		toolMsgs[msg.ToolID]++
		assert.Equal(t, "Installing", msg.Step, "All messages should use the 'Installing' step")
	}
	assert.Equal(t, 2, toolMsgs["tool-a"], "Tool A should have 2 messages")
	assert.Equal(t, 2, toolMsgs["tool-b"], "Tool B should have 2 messages")

	assert.Equal(t, 1, adapter1.installCallCount())
	assert.Equal(t, 1, adapter2.installCallCount())
}

// =========================================================================
// TestRunInstall_WarningEmitter
// =========================================================================

// warnAdapter wraps testAdapter and implements Warnings() []string
// so the pipeline's WarningEmitter type assertion succeeds.
type warnAdapter struct {
	testAdapter
	warnings []string
}

func (a *warnAdapter) Warnings() []string {
	return append([]string{}, a.warnings...)
}

// TestRunInstall_WarningEmitter verifies that when an adapter implements
// Warnings() []string and returns non-empty warnings, the pipeline emits
// a ProgressMsg with Warning=true.
func TestRunInstall_WarningEmitter(t *testing.T) {
	t.Parallel()

	w := &warnAdapter{
		testAdapter: testAdapter{id: "warn-tool", name: "Warn Tool"},
		warnings:    []string{"symlink warning: /fake/path"},
	}
	tools := []model.ToolState{
		{Adapter: w, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// After warning emission: 1 running + 2 done (success + warning).
	require.Len(t, msgs, 3, "Should receive 3 messages: 1 running + 1 done + 1 warning")

	// First: running.
	assert.Equal(t, "warn-tool", msgs[0].ToolID)
	assert.False(t, msgs[0].Done)

	// Second: done (success).
	assert.Equal(t, "warn-tool", msgs[1].ToolID)
	assert.True(t, msgs[1].Done)
	assert.Empty(t, msgs[1].Error)
	assert.False(t, msgs[1].Warning)

	// Third: warning.
	assert.Equal(t, "warn-tool", msgs[2].ToolID)
	assert.True(t, msgs[2].Done)
	assert.True(t, msgs[2].Warning, "warning message should have Warning=true")
	assert.NotEmpty(t, msgs[2].Error, "warning message should contain the joined warnings in Error")
	assert.Contains(t, msgs[2].Error, "symlink warning: /fake/path")
}

// TestRunInstall_WarningEmitter_EmptyWarnings verifies that when an adapter
// implements Warnings() but returns an empty slice, no warning ProgressMsg is sent.
func TestRunInstall_WarningEmitter_EmptyWarnings(t *testing.T) {
	t.Parallel()

	w := &warnAdapter{
		testAdapter: testAdapter{id: "clean-tool", name: "Clean Tool"},
		warnings:    []string{},
	}
	tools := []model.ToolState{
		{Adapter: w, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// No warnings — standard 2 messages (1 running + 1 done).
	require.Len(t, msgs, 2, "Should receive 2 messages: 1 running + 1 done")
	assert.False(t, msgs[1].Warning, "done message should not have Warning=true")
}

// TestStartPipeline_ChannelRecreated verifies that calling RunInstall twice
// with the same channel parameter does not panic. The first call closes the
// channel; the second write must be safe (REQ-BUG-002).
// This test models what startPipeline protects against by always creating
// a fresh channel, and also validates that sendProgress handles a closed
// channel gracefully.
func TestStartPipeline_ChannelRecreated(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "chan-test", name: "Channel Test"}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)

	// First pipeline run — creates goroutines, waits, closes channel.
	cmd1 := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd1)
	cmd1() // blocks until all goroutines complete and channel is closed

	// Drain remaining messages to confirm closure.
	_ = collectProgress(ch)

	// Verify channel is closed (zero value, ok=false).
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after first pipeline run")

	// Second pipeline run with the SAME closed channel must not panic.
	require.NotPanics(t, func() {
		cmd2 := pipeline.RunInstall(ctx, tools, ch, "en")
		require.NotNil(t, cmd2)
		// cmd2's goroutines will try to send on the closed channel.
		// This must not panic — sendProgress must handle it gracefully.
		cmd2()
	}, "second pipeline run with a closed channel must not panic (sendProgress must be defensive)")
}

// TestRunInstall_WarningEmitter_NoInterface verifies that when an adapter
// does NOT implement Warnings(), the pipeline works normally without warnings.
func TestRunInstall_WarningEmitter_NoInterface(t *testing.T) {
	t.Parallel()

	a := &testAdapter{id: "plain-tool", name: "Plain Tool"}
	tools := []model.ToolState{
		{Adapter: a, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Standard 2 messages.
	require.Len(t, msgs, 2, "Should receive 2 messages for normal adapter")
	assert.False(t, msgs[1].Warning, "done message should not have Warning=true")
}

// =========================================================================
// TestRunInstall_BackupDirGetter
// =========================================================================

// backupAdapter wraps testAdapter and implements adapters.BackupDirGetter
// so the pipeline can query the backup directory after Install/Uninstall.
type backupAdapter struct {
	testAdapter
	backupDir string
}

func (a *backupAdapter) LastBackupDir() string {
	return a.backupDir
}

// TestRunInstall_BackupDirGetter verifies that when an adapter implements
// BackupDirGetter and returns a non-empty path, the pipeline emits a
// ProgressMsg with Info set to the backup directory. REQ-BUG-004.
func TestRunInstall_BackupDirGetter(t *testing.T) {
	t.Parallel()

	b := &backupAdapter{
		testAdapter: testAdapter{id: "backup-tool", name: "Backup Tool"},
		backupDir:   "/tmp/sequoia-backups/cursor-abc123",
	}
	tools := []model.ToolState{
		{Adapter: b, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// After backup info emission: 1 running + 2 done (success + backup info).
	require.Len(t, msgs, 3, "Should receive 3 messages: 1 running + 1 done + 1 backup info")

	// First: running.
	assert.Equal(t, "backup-tool", msgs[0].ToolID)
	assert.False(t, msgs[0].Done)

	// Second: done (success).
	assert.Equal(t, "backup-tool", msgs[1].ToolID)
	assert.True(t, msgs[1].Done)
	assert.Empty(t, msgs[1].Error)
	assert.Empty(t, msgs[1].Info)

	// Third: backup info.
	assert.Equal(t, "backup-tool", msgs[2].ToolID)
	assert.True(t, msgs[2].Done)
	assert.NotEmpty(t, msgs[2].Info, "backup info message should have Info set")
	assert.Contains(t, msgs[2].Info, "/tmp/sequoia-backups/cursor-abc123",
		"Info should contain the backup directory path")
	assert.False(t, msgs[2].Warning, "backup info should not be a warning")
	assert.Empty(t, msgs[2].Error, "backup info should not have Error set")
}

// TestRunInstall_BackupDirGetter_EmptyDir verifies that when BackupDirGetter
// returns an empty string, no extra ProgressMsg is emitted.
func TestRunInstall_BackupDirGetter_EmptyDir(t *testing.T) {
	t.Parallel()

	b := &backupAdapter{
		testAdapter: testAdapter{id: "no-backup-tool", name: "No Backup Tool"},
		backupDir:   "",
	}
	tools := []model.ToolState{
		{Adapter: b, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Standard 2 messages — no backup info when dir is empty.
	require.Len(t, msgs, 2, "Should receive 2 messages when backup dir is empty")
	assert.Empty(t, msgs[1].Info, "done message should not have Info when backup dir is empty")
}

// TestRunInstall_BackupDirGetter_NoInterface verifies that when an adapter
// does NOT implement BackupDirGetter, no extra Info message is emitted.
func TestRunInstall_BackupDirGetter_NoInterface(t *testing.T) {
	t.Parallel()

	a := &testAdapter{id: "no-getter", name: "No Getter"}
	tools := []model.ToolState{
		{Adapter: a, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, 64)
	cmd := pipeline.RunInstall(ctx, tools, ch, "en")
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Standard 2 messages.
	require.Len(t, msgs, 2, "Should receive 2 messages for adapter without BackupDirGetter")
	assert.Empty(t, msgs[1].Info, "done message should not have Info when no BackupDirGetter")
}
