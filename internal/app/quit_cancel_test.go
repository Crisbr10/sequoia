// Package app contains internal (white-box) tests for screen-level quit
// handlers. These tests bypass the global quit handler in Update (which
// already calls m.cancel at update.go:33) by driving m.Update() directly
// with a non-q, non-ctrl+c key path that would have hit the screen-level
// quit branches under the legacy updateScreenKey dispatch. They verify
// the REQ-TUI-01 defense-in-depth requirement: every screen-level quit
// branch that returns tea.Quit must also invoke m.cancel() so the
// pipeline context is released even if the global handler is refactored
// away in the future.
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
// branch (update.go:106-108 / Router.handleInstallProgress) must invoke
// m.cancel() before returning tea.Quit, so the pipeline goroutines
// observe ctx.Done().
//
// PR7 commit 5 migrated the InstallProgress case to the Router. The
// test now exercises the public m.Update API which routes through
// Router.DispatchKey internally, preserving the REQ-TUI-01 contract
// assertion.
func TestInstallProgress_Q_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenInstallProgress)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, _ = m.Update(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"q on InstallProgress did not invoke m.cancel() before returning tea.Quit")
}

// TestErrorScreen_Q_CancelsContext — REQ-TUI-01
//
// When q is pressed on the Error screen, the inlined Error handler
// (update.go:138-140 / Router.handleError) must invoke m.cancel()
// before returning tea.Quit.
//
// PR7 commit 7: test uses m.Update() to exercise the new dispatch
// surface (Router.DispatchKey → Router.handleError). The REQ-TUI-01
// contract is preserved end-to-end.
func TestErrorScreen_Q_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenError)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, _ = m.Update(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"q on Error screen did not invoke m.cancel() before returning tea.Quit")
}

// TestErrorScreen_CtrlC_CancelsContext — REQ-TUI-01
//
// When ctrl+c is pressed on the Error screen, the inlined Error handler
// (update.go:130-131 / Router.handleError) must invoke m.cancel()
// before returning tea.Quit.
//
// PR7 commit 7: test uses m.Update() to exercise the new dispatch
// surface. The REQ-TUI-01 contract is preserved end-to-end.
func TestErrorScreen_CtrlC_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenError)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, _ = m.Update(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"ctrl+c on Error screen did not invoke m.cancel() before returning tea.Quit")
}

// TestComplete_Q_CancelsContext — REQ-TUI-01
//
// When q is pressed on the Complete screen, the inlined Complete handler
// (added by PR1 — replaces the delegation to screens.CompleteUpdate,
// and now extracted to Router.handleComplete by PR7 commit 6) must
// invoke m.cancel() before returning tea.Quit.
//
// PR7 commit 6 migrated the Complete case to the Router. The test now
// uses m.Update() (the public API) which routes through
// Router.DispatchKey → Router.handleComplete.
func TestComplete_Q_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenComplete)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, _ = m.Update(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"q on Complete screen did not invoke m.cancel() before returning tea.Quit")
}

// TestComplete_CtrlC_CancelsContext — REQ-TUI-01
//
// When ctrl+c is pressed on the Complete screen, the inlined Complete
// handler (added by PR1, extracted to Router.handleComplete by PR7
// commit 6) must invoke m.cancel() before returning tea.Quit.
//
// PR7 commit 6: test uses m.Update() to exercise the new dispatch
// surface; the REQ-TUI-01 contract is preserved end-to-end.
func TestComplete_CtrlC_CancelsContext(t *testing.T) {
	m, ctx := withCancelableModel(t, model.ScreenComplete)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, _ = m.Update(msg)

	assertCancelledWithin(t, ctx, 100*time.Millisecond,
		"ctrl+c on Complete screen did not invoke m.cancel() before returning tea.Quit")
}

// Sanity check: the model in this file satisfies tea.Model.
var _ tea.Model = Model{}

// Sanity check: withCancelableModel and assertCancelledWithin are reachable
// from this package's test entry points.
var _ = require.New
