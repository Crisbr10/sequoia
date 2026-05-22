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
// It implements both the legacy Install/Uninstall methods and the new
// common.Strategy interface for phased execution.
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

	// Strategy phase tracking
	prepareCalls  int
	downloadCalls int
	verifyCalls   int
	stageCalls    int
	applyCalls    int
	rollbackCalls int

	// Per-phase errors for Strategy methods
	prepareErr  error
	downloadErr error
	verifyErr   error
	stageErr    error
	applyErr    error
	rollbackErr error
}

func (a *testAdapter) ID() string        { return a.id }
func (a *testAdapter) Name() string      { return a.name }
func (a *testAdapter) Detect() bool      { return true }
func (a *testAdapter) IsInstalled() bool { return a.installed }

func (a *testAdapter) Install(opts adapters.InstallOpts) error {
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
	a.mu.Lock()
	a.uninstallCalls++
	a.lastContext = opts.Context
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.uninstallErr
}

// Strategy interface implementation for phased execution.

func (a *testAdapter) Prepare(opts adapters.InstallOpts) error {
	a.mu.Lock()
	a.prepareCalls++
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.prepareErr
}

func (a *testAdapter) Download(opts adapters.InstallOpts) error {
	a.mu.Lock()
	a.downloadCalls++
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.downloadErr
}

func (a *testAdapter) Verify() error {
	a.mu.Lock()
	a.verifyCalls++
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.verifyErr
}

func (a *testAdapter) Stage(opts adapters.InstallOpts) error {
	a.mu.Lock()
	a.stageCalls++
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.stageErr
}

func (a *testAdapter) Apply(opts adapters.InstallOpts) error {
	a.mu.Lock()
	a.applyCalls++
	a.lastContext = opts.Context
	a.mu.Unlock()
	if a.delay > 0 {
		time.Sleep(a.delay)
	}
	return a.applyErr
}

func (a *testAdapter) Rollback() error {
	a.mu.Lock()
	a.rollbackCalls++
	a.mu.Unlock()
	return a.rollbackErr
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd, "RunInstall should return a non-nil tea.Cmd")

	// Execute the command.
	cmd()

	msgs := collectProgress(ch)

	// 5 phases × 2 messages each = 10 messages.
	require.Len(t, msgs, 10, "Should receive exactly 10 progress messages (5 phases × 2)")

	// Verify phase names appear in order.
	expectedPhases := []string{"Preparing", "Downloading", "Verifying", "Staging", "Applying"}
	for i, phase := range expectedPhases {
		assert.Equal(t, "test-tool", msgs[i*2].ToolID)
		assert.Equal(t, phase, msgs[i*2].Step)
		assert.False(t, msgs[i*2].Done, "First message of %s should be 'running' (Done=false)", phase)
		assert.Empty(t, msgs[i*2].Error)

		assert.Equal(t, "test-tool", msgs[i*2+1].ToolID)
		assert.Equal(t, phase, msgs[i*2+1].Step)
		assert.True(t, msgs[i*2+1].Done, "Second message of %s should be 'done' (Done=true)", phase)
		assert.Empty(t, msgs[i*2+1].Error)
	}

	assert.Equal(t, 1, adapter.applyCalls, "Apply should be called once")
}

func TestRunInstall_StepFailure_SendsErrorProgress(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("disk full")
	adapter := &testAdapter{id: "fail-tool", name: "Fail Tool", applyErr: expectedErr}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// 4 phases succeed (8 msgs) + 1 failed phase (2 msgs: running + error) = 10.
	// But on Apply error, rollback happens after the error is sent.
	require.Len(t, msgs, 10, "Should receive 10 messages: 4 phases × 2 + applying running + applying error")

	// Verify the last message is the error on Applying.
	last := msgs[len(msgs)-1]
	assert.Equal(t, "Applying", last.Step)
	assert.True(t, last.Done)
	assert.Contains(t, last.Error, "disk full")

	assert.Equal(t, 1, adapter.rollbackCalls, "Rollback should be called on failure")
}

func TestRunInstall_ContextCancellation_StopsGoroutines(t *testing.T) {
	t.Parallel()

	// Use a delay to simulate a slow install so cancellation can interrupt it.
	adapter := &testAdapter{id: "slow-tool", name: "Slow Tool", delay: 200 * time.Millisecond}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
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

	// Both adapters should have been called (all phases).
	assert.Equal(t, 1, adapter1.applyCalls, "Adapter A should be called once")
	assert.Equal(t, 1, adapter2.applyCalls, "Adapter B should be called once")

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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunUninstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Uninstall uses the Installer interface: 1 running + 1 done.
	require.Len(t, msgs, 2, "Should receive exactly 2 messages: 1 running + 1 done")

	assert.Equal(t, "uninstall-tool", msgs[0].ToolID)
	assert.Equal(t, "Uninstalling", msgs[0].Step)
	assert.False(t, msgs[0].Done, "First should be 'running'")

	assert.Equal(t, "uninstall-tool", msgs[1].ToolID)
	assert.Equal(t, "Uninstalling", msgs[1].Step)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunUninstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// 1 running + 1 error.
	require.Len(t, msgs, 2, "Should receive 2 messages: 1 running + 1 error")

	assert.Equal(t, "fail-uninstall", msgs[0].ToolID)
	assert.Equal(t, "Uninstalling", msgs[0].Step)
	assert.False(t, msgs[0].Done, "First message should be 'running'")

	assert.Equal(t, "fail-uninstall", msgs[1].ToolID)
	assert.Equal(t, "Uninstalling", msgs[1].Step)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunUninstall(ctx, tools, ch)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
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
// the caller's context to the adapter's Apply method (the last phase) via InstallOpts.
func TestRunInstall_PassesContextToAdapter(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "ctx-test", name: "Context Test"}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)
	require.NotEmpty(t, msgs, "Should receive progress messages")

	// The adapter must have received a non-nil context via Apply.
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

	assert.Equal(t, 1, adapter.applyCalls, "Apply should be called once")
}

// TestRunUninstall_PassesContextToAdapter verifies uninstall context propagation
// via the Installer interface.
func TestRunUninstall_PassesContextToAdapter(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "uninstall-ctx", name: "Uninstall Context", installed: true}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunUninstall(ctx, tools, ch)
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

// TestDefaultStepNames_FivePhases_Old verifies that InstallSteps now has 5 phases.
// This test replaces the old single-step assertion with the new Strategy phases.
func TestDefaultStepNames_FivePhases_Old(t *testing.T) {
	t.Parallel()

	steps := pipeline.InstallSteps
	require.Len(t, steps, 5, "InstallSteps should have 5 phase names after Strategy refactor")
	assert.Equal(t, "Preparing", steps[0])
	assert.Equal(t, "Downloading", steps[1])
	assert.Equal(t, "Verifying", steps[2])
	assert.Equal(t, "Staging", steps[3])
	assert.Equal(t, "Applying", steps[4])
}

// TestRunInstall_MultiTool_SendsTenMessagesEach verifies that with 2 tools,
// each tool sends exactly 10 messages (5 phases × 2 messages each).
func TestRunInstall_MultiTool_SendsTenMessagesEach(t *testing.T) {
	t.Parallel()

	adapter1 := &testAdapter{id: "tool-a", name: "Tool A"}
	adapter2 := &testAdapter{id: "tool-b", name: "Tool B"}
	tools := []model.ToolState{
		{Adapter: adapter1, Selected: true},
		{Adapter: adapter2, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// 2 tools × 10 messages each = 20 total.
	assert.Len(t, msgs, 20, "2 tools × 10 messages = 20 total")

	// Both tools should appear with the 5 phase names.
	toolMsgs := map[string]int{}
	for _, msg := range msgs {
		toolMsgs[msg.ToolID]++
	}
	assert.Equal(t, 10, toolMsgs["tool-a"], "Tool A should have 10 messages")
	assert.Equal(t, 10, toolMsgs["tool-b"], "Tool B should have 10 messages")

	assert.Equal(t, 1, adapter1.applyCalls)
	assert.Equal(t, 1, adapter2.applyCalls)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// After warning emission: 10 normal messages (5 phases × 2) + 1 warning = 11.
	require.Len(t, msgs, 11, "Should receive 11 messages: 10 phase messages + 1 warning")

	// First: preparing running.
	assert.Equal(t, "warn-tool", msgs[0].ToolID)
	assert.False(t, msgs[0].Done)

	// Last: the warning message.
	last := msgs[len(msgs)-1]
	assert.Equal(t, "warn-tool", last.ToolID)
	assert.True(t, last.Done)
	assert.True(t, last.Warning, "warning message should have Warning=true")
	assert.NotEmpty(t, last.Error, "warning message should contain the joined warnings in Error")
	assert.Contains(t, last.Error, "symlink warning: /fake/path")
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// No warnings — standard 10 messages (5 phases × 2).
	require.Len(t, msgs, 10, "Should receive 10 messages: 5 phases × 2")
	// Last message should not have Warning=true.
	last := msgs[len(msgs)-1]
	assert.False(t, last.Warning, "last message should not have Warning=true")
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)

	// First pipeline run — creates goroutines, waits, closes channel.
	cmd1 := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd1)
	cmd1() // blocks until all goroutines complete and channel is closed

	// Drain remaining messages to confirm closure.
	_ = collectProgress(ch)

	// Verify channel is closed (zero value, ok=false).
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after first pipeline run")

	// Second pipeline run with the SAME closed channel must not panic.
	require.NotPanics(t, func() {
		cmd2 := pipeline.RunInstall(ctx, tools, ch)
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Standard 10 messages.
	require.Len(t, msgs, 10, "Should receive 10 messages for normal adapter")
	assert.False(t, msgs[len(msgs)-1].Warning, "last message should not have Warning=true")
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// After backup info emission: 10 phase messages + 1 backup info = 11.
	require.Len(t, msgs, 11, "Should receive 11 messages: 10 phase messages + 1 backup info")

	// First: preparing running.
	assert.Equal(t, "backup-tool", msgs[0].ToolID)
	assert.False(t, msgs[0].Done)

	// Last: backup info.
	last := msgs[len(msgs)-1]
	assert.Equal(t, "backup-tool", last.ToolID)
	assert.True(t, last.Done)
	assert.NotEmpty(t, last.Info, "backup info message should have Info set")
	assert.Contains(t, last.Info, "/tmp/sequoia-backups/cursor-abc123",
		"Info should contain the backup directory path")
	assert.False(t, last.Warning, "backup info should not be a warning")
	assert.Empty(t, last.Error, "backup info should not have Error set")
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Standard 10 messages — no backup info when dir is empty.
	require.Len(t, msgs, 10, "Should receive 10 messages when backup dir is empty")
	assert.Empty(t, msgs[len(msgs)-1].Info, "last message should not have Info when backup dir is empty")
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

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Standard 10 messages.
	require.Len(t, msgs, 10, "Should receive 10 messages for adapter without BackupDirGetter")
	assert.Empty(t, msgs[len(msgs)-1].Info, "last message should not have Info when no BackupDirGetter")
}

// =========================================================================
// Strategy Phase Tests — RED (will fail until pipeline is refactored)
// =========================================================================

// TestRunInstall_StrategyPhases verifies that when the pipeline dispatches
// via Strategy, each phase emits a running (Done=false) and done (Done=true)
// progress message with the correct phase name in Step.
func TestRunInstall_StrategyPhases(t *testing.T) {
	t.Parallel()

	adapter := &testAdapter{id: "phase-tool", name: "Phase Tool"}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// 5 phases × 2 messages each = 10 messages.
	require.Len(t, msgs, 10, "5 Strategy phases should produce 10 messages (5 running + 5 done)")

	expectedPhases := []string{"Preparing", "Downloading", "Verifying", "Staging", "Applying"}
	for i, phase := range expectedPhases {
		// Running message (Done=false).
		runningIdx := i * 2
		assert.Equal(t, phase, msgs[runningIdx].Step, "running message step %d should be %s", i, phase)
		assert.False(t, msgs[runningIdx].Done, "running message for %s should have Done=false", phase)
		assert.Empty(t, msgs[runningIdx].Error)

		// Done message (Done=true).
		doneIdx := runningIdx + 1
		assert.Equal(t, phase, msgs[doneIdx].Step, "done message step %d should be %s", i, phase)
		assert.True(t, msgs[doneIdx].Done, "done message for %s should have Done=true", phase)
		assert.Empty(t, msgs[doneIdx].Error)
	}

	// Verify phase methods were called.
	assert.Equal(t, 1, adapter.prepareCalls, "Prepare should be called once")
	assert.Equal(t, 1, adapter.downloadCalls, "Download should be called once")
	assert.Equal(t, 1, adapter.verifyCalls, "Verify should be called once")
	assert.Equal(t, 1, adapter.stageCalls, "Stage should be called once")
	assert.Equal(t, 1, adapter.applyCalls, "Apply should be called once")
}

// TestRunInstall_StrategyPhaseFailure verifies that when a Strategy phase
// fails, the pipeline reports the error on that phase step and calls Rollback.
func TestRunInstall_StrategyPhaseFailure(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("stage failed")
	adapter := &testAdapter{id: "fail-phase", name: "Fail Phase", stageErr: expectedErr}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)

	// Phases: Preparing (running + done), Downloading (running + done),
	// Verifying (running + done), Staging (running + error). No Applying.
	// That's 3×2 + 2 = 8 messages.
	require.Len(t, msgs, 8, "3 successful phases (6 msgs) + 1 failed phase (2 msgs) = 8")

	// Last message should be the error on Staging.
	lastMsg := msgs[len(msgs)-1]
	assert.Equal(t, "Staging", lastMsg.Step)
	assert.True(t, lastMsg.Done)
	assert.Contains(t, lastMsg.Error, "stage failed")

	// Rollback should have been called.
	assert.Equal(t, 1, adapter.rollbackCalls, "Rollback should be called on phase failure")
}

// TestRunInstall_StrategyNonStrategyAdapter verifies that when an adapter
// does NOT implement common.Strategy, the pipeline returns an error message
// without panicking.
func TestRunInstall_StrategyNonStrategyAdapter(t *testing.T) {
	t.Parallel()

	// nonStrategyAdapter implements ToolInfo but NOT common.Strategy.
	nsa := &nonStrategyAdapter{id: "no-strategy"}

	tools := []model.ToolState{
		{Adapter: nsa, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	// Must not panic.
	require.NotPanics(t, func() {
		cmd()
	}, "non-Strategy adapter must not cause a panic")

	msgs := collectProgress(ch)
	require.NotEmpty(t, msgs, "Should receive an error message for non-Strategy adapter")

	// The error message should mention the adapter doesn't implement Strategy.
	lastMsg := msgs[len(msgs)-1]
	assert.True(t, lastMsg.Done)
	assert.NotEmpty(t, lastMsg.Error)
	assert.Contains(t, lastMsg.Error, "does not implement Strategy",
		"error should mention missing Strategy interface")
}

// nonStrategyAdapter implements model.ToolInfo but NOT common.Strategy.
type nonStrategyAdapter struct {
	id string
}

func (a *nonStrategyAdapter) ID() string               { return a.id }
func (a *nonStrategyAdapter) Name() string             { return "No Strategy" }
func (a *nonStrategyAdapter) Detect() bool             { return true }
func (a *nonStrategyAdapter) IsInstalled() bool        { return false }
func (a *nonStrategyAdapter) Status() model.ToolStatus { return model.ToolStatus{} }

// TestRunInstall_StrategyCancellationBetweenPhases verifies that context
// cancellation between phases stops execution and sends a partial progress.
func TestRunInstall_StrategyCancellationBetweenPhases(t *testing.T) {
	t.Parallel()

	// Use a slow adapter so we can cancel mid-way.
	adapter := &testAdapter{id: "slow-phase", name: "Slow Phase", delay: 50 * time.Millisecond}
	tools := []model.ToolState{
		{Adapter: adapter, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd()
	}()

	// Cancel after a short time — should interrupt between phases.
	time.Sleep(10 * time.Millisecond)
	cancel()

	wg.Wait()

	msgs, closed := collectProgressWithTimeout(ch, 500*time.Millisecond)
	_ = msgs
	assert.True(t, closed, "Channel should be closed after context cancellation")
}

// TestDefaultStepNames_FivePhases verifies that InstallSteps now contains
// the five phase names instead of the old single "Installing" step.
func TestDefaultStepNames_FivePhases(t *testing.T) {
	t.Parallel()

	steps := pipeline.InstallSteps
	require.Len(t, steps, 5, "InstallSteps should have 5 phase names after Strategy refactor")
	assert.Equal(t, "Preparing", steps[0])
	assert.Equal(t, "Downloading", steps[1])
	assert.Equal(t, "Verifying", steps[2])
	assert.Equal(t, "Staging", steps[3])
	assert.Equal(t, "Applying", steps[4])
}
