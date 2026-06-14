// Package app contains internal (white-box) tests for screen-level quit
// handlers in updateScreenKey. These tests bypass the global quit handler
// in Update (which already calls m.cancel at update.go:33) by calling
// m.updateScreenKey directly. They verify the REQ-TUI-01 defense-in-depth
// requirement: every screen-level quit branch that returns tea.Quit must
// also invoke m.cancel() so the pipeline context is released even if the
// global handler is refactored away in the future.
//
// Note: the ToolSelection "quit" case at update.go:97-99 is unreachable
// from any real input today (ToolSelectionUpdate never returns "quit" —
// see TestToolSelectionUpdate_QNoLongerReturnsQuit). The fix at line 99
// is verified by code review only, not by a behavioral test.
package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

// assertCancelledWithin asserts the provided context is cancelled within
// the given duration. Used to confirm m.cancel() fired in the screen-level
// quit branch.
func assertCancelledWithin(t *testing.T, ctx context.Context, dur time.Duration, msg string) {
	t.Helper()
	select {
	case <-ctx.Done():
		// pass
	case <-time.After(dur):
		t.Fatalf("context not cancelled within %s: %s", dur, msg)
	}
}

// withCancelableModel returns a Model wired to a real context.Context so
// the test can detect m.cancel() firing. The model uses NewModel to allocate
// the Progress channel and registry; the ctx/cancel pair replaces the
// internal ones so the test owns the cancellation observable.
func withCancelableModel(t *testing.T, screen model.Screen) (Model, context.Context) {
	t.Helper()
	m := NewModel("", "test", adapters.NewRegistry())
	ctx, cancel := context.WithCancel(context.Background())
	m.ctx = ctx
	m.cancel = cancel
	m.Screen = screen
	return m, ctx
}

// TestInstallProgress_Q_CancelsContext — REQ-TUI-01
//
// When q is pressed on the InstallProgress screen, the screen-level quit
// branch (update.go:106-108) must invoke m.cancel() before returning
// tea.Quit, so the pipeline goroutines observe ctx.Done().
//
// RED on current code: line 108 returns tea.Quit without calling m.cancel().
func TestInstallProgress_Q_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenInstallProgress)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, _ = m.updateScreenKey(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"q on InstallProgress did not invoke m.cancel() before returning tea.Quit")
}

// TestErrorScreen_Q_CancelsContext — REQ-TUI-01
//
// When q is pressed on the Error screen, the inlined Error handler
// (update.go:138-140) must invoke m.cancel() before returning tea.Quit.
//
// RED on current code: line 139 returns tea.Quit without calling m.cancel().
func TestErrorScreen_Q_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenError)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, _ = m.updateScreenKey(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"q on Error screen did not invoke m.cancel() before returning tea.Quit")
}

// TestErrorScreen_CtrlC_CancelsContext — REQ-TUI-01
//
// When ctrl+c is pressed on the Error screen, the inlined Error handler
// (update.go:130-131) must invoke m.cancel() before returning tea.Quit.
//
// RED on current code: line 131 returns tea.Quit without calling m.cancel().
func TestErrorScreen_CtrlC_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenError)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, _ = m.updateScreenKey(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"ctrl+c on Error screen did not invoke m.cancel() before returning tea.Quit")
}

// TestComplete_Q_CancelsContext — REQ-TUI-01
//
// When q is pressed on the Complete screen, the inlined Complete handler
// (added by this PR — replaces the delegation to screens.CompleteUpdate)
// must invoke m.cancel() before returning tea.Quit.
//
// RED on current code: update.go:120-121 delegates to screens.CompleteUpdate
// which returns tea.Quit for q but has no access to m.cancel, so the
// pipeline ctx is never cancelled.
func TestComplete_Q_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenComplete)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, _ = m.updateScreenKey(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"q on Complete screen did not invoke m.cancel() before returning tea.Quit")
}

// TestComplete_CtrlC_CancelsContext — REQ-TUI-01
//
// When ctrl+c is pressed on the Complete screen, the inlined Complete
// handler (added by this PR) must invoke m.cancel() before returning
// tea.Quit.
//
// RED on current code: update.go:120-121 delegates to screens.CompleteUpdate
// which returns tea.Quit for ctrl+c but has no access to m.cancel.
func TestComplete_CtrlC_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenComplete)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, _ = m.updateScreenKey(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"ctrl+c on Complete screen did not invoke m.cancel() before returning tea.Quit")
}

// Sanity check: the model in this file satisfies tea.Model.
var _ tea.Model = Model{}

// Sanity check: withCancelableModel and assertCancelledWithin are reachable
// from this package's test entry points.
var _ = require.New
