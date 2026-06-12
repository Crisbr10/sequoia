// Package tui provides the terminal user interface for the Sequoia installer.
// It contains screen navigation messages, styles, and screen renderers.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crisbr10/sequoia/internal/model"
)

// NavigateMsg is a Bubbletea message instructing the root model to
// switch to the given screen.
type NavigateMsg struct {
	// Target is the screen to navigate to.
	Target model.Screen
}

// KeyHandler is the contract between the Router and the root *app.Model.
// The Router operates exclusively through this interface so that it does
// not need to import internal/app (which would create an import cycle:
// internal/tui is imported by internal/app). *app.Model satisfies this
// interface via the wrapper methods added in internal/app/model.go.
//
// The interface embeds tea.Model so that the Router's DispatchKey can
// return h as a tea.Model (the Bubbletea framework expects a tea.Model
// return value from Update). *app.Model satisfies tea.Model via its
// existing Init/Update/View methods.
//
// The getter methods use Get* prefix to avoid Go's "field and method
// with the same name" restriction (the Model struct has fields named
// Screen, Tools, Cursor, etc.). The setter methods use Set* prefix as
// is conventional.
//
// UpdateScreenKey is a temporary bridge method that delegates to the
// legacy updateScreenKey function in internal/app/update.go. It exists
// for the duration of the PR7 migration: commit 2 uses it to preserve
// behavior; commits 3-9 migrate each screen to a per-screen handleX
// method; commit 10 removes UpdateScreenKey and the legacy function.
//
// NOTE: ProgressTool and ProgressStep live in internal/tui/screens, but
// screens imports tui (for NavigateMsg), so the KeyHandler interface
// cannot directly reference them without creating an import cycle.
// The Router only needs the count of progress tools (for the
// InstallProgress screen's auto-transition logic), exposed as
// InstallProgressCount(). Full ProgressTools mutation stays in
// updateScreenMsg, which is not part of the Router refactor.
//
// REQ-TUI-07: a new Router type SHALL encapsulate per-screen key
// dispatch. The current updateScreenKey switch in internal/app/update.go
// SHALL be replaced by a method dispatch on the Router.
type KeyHandler interface {
	tea.Model

	// Screen accessors.
	GetScreen() model.Screen
	SetScreen(s model.Screen)
	GetPreviousScreen() model.Screen
	SetPreviousScreen(s model.Screen)

	// Cursor accessors.
	GetCursor() int
	SetCursor(c int)

	// Tools accessors.
	GetTools() []model.ToolState
	SetTools(t []model.ToolState)

	// ErrorMsg accessors.
	GetErrorMsg() string
	SetErrorMsg(s string)

	// OperationMode accessors.
	GetOperationMode() string
	SetOperationMode(mode string)

	// UninstallConfirming accessors.
	GetUninstallConfirming() bool
	SetUninstallConfirming(b bool)

	// Quitting accessors.
	GetQuitting() bool
	SetQuitting(q bool)

	// Install counters accessors.
	GetInstallCompleted() int
	SetInstallCompleted(n int)
	GetInstallFailed() int
	SetInstallFailed(n int)
	GetInstallWarned() int
	SetInstallWarned(n int)

	// InstallProgressCount returns the number of tools currently
	// tracked in the per-tool install progress list. The Router uses
	// this to compute the auto-transition threshold for the
	// InstallProgress screen without needing direct access to the
	// []screens.ProgressTool slice.
	InstallProgressCount() int

	// Per-screen dispatch helpers. These wrap the screen Update
	// functions in internal/tui/screens so the Router does not need
	// to import internal/tui/screens (which would create an import
	// cycle: internal/tui -> internal/tui/screens -> internal/tui
	// for NavigateMsg). *app.Model implements each helper by
	// calling the corresponding screens function with the current
	// Model state as arguments.
	WelcomeUpdate(msg tea.KeyMsg) (int, string)
	ToolSelectionUpdate(msg tea.KeyMsg) (int, bool, string)
	InstallProgressUpdate(msg tea.KeyMsg) string
	StatusUpdate(msg tea.KeyMsg) (int, string)
	UninstallUpdate(msg tea.KeyMsg) (int, bool, string)
	CountSelectedTools() int
	HasSelectedInstalled() bool

	// UpdateScreenKey is the temporary bridge to the legacy dispatch
	// surface. It delegates to the unexported updateScreenKey function
	// in internal/app/update.go. The Router's DispatchKey stub calls
	// this for behavior preservation during the PR7 migration. After
	// all screens are migrated to per-screen handleX methods (commits
	// 3-9), this method is removed (commit 10).
	UpdateScreenKey(msg tea.KeyMsg) (tea.Model, tea.Cmd)

	// Action methods.
	LoadTools(toolID string)
	Cancel()
	StartPipeline(mode string) tea.Cmd
}

// Router encapsulates per-screen key dispatch. It is the single source
// of truth for which screen handles which key and what command each
// key produces. The dispatch logic previously lived in the 200-line
// updateScreenKey switch in internal/app/update.go; that switch has
// been broken into per-screen handle methods on the Router.
//
// REQ-TUI-07: a new Router type SHALL encapsulate per-screen key
// dispatch. The current updateScreenKey switch SHALL be replaced by
// a method dispatch on the Router.
type Router struct{}

// NewRouter returns a stateless Router. The type is empty because
// dispatch is purely functional — all per-call state lives on the
// KeyHandler argument.
func NewRouter() *Router {
	return &Router{}
}

// DispatchKey routes msg to the handler for the screen reported by
// h.GetScreen(). The model state on h is mutated in place; the returned
// (tea.Model, tea.Cmd) tuple carries the updated model (for framework
// compatibility) and the command to execute.
//
// PR7 commit 2 (GREEN shell): this method is a stub that delegates to
// h.UpdateScreenKey(msg) — the legacy dispatch surface. The Router
// exists, the KeyHandler interface is satisfied by *app.Model, and
// the full test suite still passes through the public m.Update API.
//
// PR7 commits 3-9 (per-screen migration): each commit replaces the
// corresponding case in updateScreenKey with a call to r.handleX(h, msg)
// here. The per-screen handleX methods are populated incrementally.
//
// PR7 commit 10 (final cleanup): the per-screen switch below is the
// single dispatch surface; the legacy updateScreenKey and h.UpdateScreenKey
// are removed.
func (r *Router) DispatchKey(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if h.GetScreen() == model.ScreenWelcome {
		return r.handleWelcome(h, msg)
	}
	return h.UpdateScreenKey(msg)
}

// handleWelcome is the per-screen dispatch for ScreenWelcome. It is
// extracted from the ScreenWelcome case of the legacy updateScreenKey
// function in internal/app/update.go. The behavior is byte-equivalent
// to the pre-refactor switch case.
//
// Per-screen logic:
//   - WelcomeUpdate produces a (newCursor, action) tuple.
//   - The "install" / "status" / "uninstall" actions navigate to
//     the corresponding screen after loading tools.
//   - The "quit" action invokes the quit pipeline (sets Quitting,
//     calls Cancel, returns tea.Quit).
//   - All other actions return nil (no-op).
func (r *Router) handleWelcome(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	newCursor, action := h.WelcomeUpdate(msg)
	h.SetCursor(newCursor)
	switch action {
	case "install":
		h.LoadTools("")
		return h, func() tea.Msg {
			return NavigateMsg{Target: model.ScreenToolSelection}
		}
	case "status":
		h.LoadTools("")
		return h, func() tea.Msg {
			return NavigateMsg{Target: model.ScreenStatus}
		}
	case "uninstall":
		h.LoadTools("")
		return h, func() tea.Msg {
			return NavigateMsg{Target: model.ScreenUninstall}
		}
	case "quit":
		h.SetQuitting(true)
		h.Cancel()
		return h, tea.Quit
	}
	return h, nil
}

// handleToolSelection is the per-screen dispatch for ScreenToolSelection.
// Populated in commit 4.
func (r *Router) handleToolSelection(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return h, nil
}

// handleInstallProgress is the per-screen dispatch for
// ScreenInstallProgress. Populated in commit 5.
func (r *Router) handleInstallProgress(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return h, nil
}

// handleComplete is the per-screen dispatch for ScreenComplete.
// Populated in commit 6.
func (r *Router) handleComplete(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return h, nil
}

// handleError is the per-screen dispatch for ScreenError. Populated in
// commit 7.
func (r *Router) handleError(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return h, nil
}

// handleStatus is the per-screen dispatch for ScreenStatus. Populated
// in commit 8.
func (r *Router) handleStatus(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return h, nil
}

// handleUninstall is the per-screen dispatch for ScreenUninstall.
// Populated in commit 9.
func (r *Router) handleUninstall(h KeyHandler, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return h, nil
}
