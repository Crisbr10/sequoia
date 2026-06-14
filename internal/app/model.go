// Package app provides the root Bubbletea Model for the Sequoia TUI installer.
// It manages screen state, tool registry, user configuration, and the progress
// channel used for async pipeline communication.
package app

import (
	"context"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the root Bubbletea model for the Sequoia TUI installer.
// It owns the screen state machine, tool registry snapshot, user preferences,
// and the buffered channel for progress messages from the install pipeline.
type Model struct {
	// Version is the Sequoia release string displayed on the Welcome screen.
	// Set at build time via ldflags, passed from cmd/sequoia/main.go.
	Version string
	// Screen tracks which screen is currently displayed.
	Screen model.Screen
	// Tools is a snapshot of registered adapters with their UI state.
	Tools []model.ToolState

	// Width is the terminal width in characters.
	Width int
	// Height is the terminal height in characters.
	Height int
	// Cursor tracks the highlighted row index in list-based screens
	// (ToolSelection, Configuration, Status, Uninstall).
	// In Configuration, it represents the active field (0=language, 1=persistence).
	Cursor int
	// ErrorMsg holds a transient validation or error message for the current screen.
	ErrorMsg string
	// OperationMode tracks the current operation: "install" or "uninstall".
	// Empty string defaults to install-variant labels in views.
	OperationMode string
	// PreviousScreen records the screen that was active before the most recent
	// navigation. It is set by the NavigateMsg handler so that screens can
	// implement source-aware back navigation.
	PreviousScreen model.Screen
	// Progress is a buffered channel receiving ProgressMsg from install goroutines.
	// Capacity is 64 to prevent pipeline blocking during bursty progress updates.
	Progress chan model.ProgressMsg
	// ProgressTools tracks per-tool install progress for the Install Progress screen.
	ProgressTools []screens.ProgressTool
	// InstallCompleted counts tools whose install has fully completed.
	InstallCompleted int
	// InstallFailed counts tools that encountered a critical installation failure.
	InstallFailed int
	// InstallWarned counts tools that completed with non-fatal warnings
	// (e.g., partial uninstall where some files could not be removed).
	InstallWarned int

	// UninstallConfirming is true when the Uninstall screen is in
	// confirmation mode (user pressed Enter and is being asked y/N).
	UninstallConfirming bool
	// Quitting is set to true when the user initiates exit.
	Quitting bool

	// ctx is the pipeline context, cancelled on quit to stop goroutines.
	ctx context.Context
	// cancel cancels the pipeline context.
	cancel context.CancelFunc

	// reg is the adapter registry this Model reads from.
	// Stored for potential lookup operations after construction.
	reg *adapters.Registry
}

// toolInfoAdapter adapts a ToolAdapter to satisfy model.ToolInfo by embedding
// narrow role interfaces (Identifier, Detector, InstallStatus, Installer,
// Strategy) instead of the full ToolAdapter, per ISP narrowing P3-003.
// The Status() override converts adapters.AdapterStatus to model.ToolStatus.
type toolInfoAdapter struct {
	adapters.Identifier
	adapters.Detector
	adapters.InstallStatus
	adapters.Installer
	common.Strategy
}

// Status returns the adapter's installation status as a model.ToolStatus.
func (a toolInfoAdapter) Status() model.ToolStatus {
	s := a.InstallStatus.Status()
	return model.ToolStatus{
		Installed: s.Installed,
		Version:   s.Version,
		Path:      s.Path,
	}
}

// NewModel creates the root Model populated with all registered adapters
// and default configuration. If toolID is non-empty, only that adapter is
// selected by default. version is the Sequoia release string (set via ldflags)
// displayed on the Welcome screen.
//
// reg provides the adapter registry for DI. In production, the caller creates
// a NewRegistry() and registers adapters via RegisterIn(). In tests, mocks
// are registered locally on an explicit *Registry.
//
// Engram detection is deferred to Init() via detectEngram() to avoid
// blocking the first TUI render on exec.LookPath.
//
// Adapter construction is deferred past the first frame render. The Welcome
// screen does not need the adapter list, so reg.All() is NOT called here.
// Tools are loaded lazily via loadTools() when the user navigates to a screen
// that requires adapter data (ToolSelection, Status, Uninstall).
func NewModel(toolID string, version string, reg *adapters.Registry) Model {
	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		Version:  version,
		Screen:   model.ScreenWelcome,
		Tools:    nil, // loaded lazily by loadTools()
		Progress: make(chan model.ProgressMsg, model.ProgressChannelBufferSize),
		reg:      reg,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// LoadTools populates m.Tools from the registry if not already loaded.
// It is safe to call multiple times; subsequent calls are no-ops.
// toolID is the optional adapter ID to force-select; empty means auto-select
// only adapters that are detected (Detect() returns true) on this system.
// Exported for test use; production code calls this internally in updateScreenKey.
func (m *Model) LoadTools(toolID string) {
	if m.Tools != nil {
		return // already loaded
	}
	all := m.reg.All()
	m.Tools = make([]model.ToolState, 0, len(all))
	for _, a := range all {
		// Pre-select only if user targeted this tool specifically OR
		// (no specific tool requested AND the tool is detected on this system).
		detected := a.Detect()
		selected := a.ID() == toolID || (toolID == "" && detected)

		ts := model.ToolState{
			Adapter: toolInfoAdapter{
				Identifier:    a,
				Detector:      a,
				InstallStatus: a,
				Installer:     a,
				Strategy:      a.(common.Strategy),
			},
			Selected: selected,
		}
		m.Tools = append(m.Tools, ts)
	}
}

// Init is the Bubbletea init command.
func (m Model) Init() tea.Cmd {
	return nil
}

// KeyHandler interface methods on *Model.
//
// These wrapper methods satisfy the tui.KeyHandler interface introduced
// in PR7 (REQ-TUI-07). The Router in internal/tui/router.go operates
// exclusively through this interface so that it does not need to
// import internal/app (which would create an import cycle).
//
// Each method is a 1-2 line accessor/mutator; the interface contract
// is documented on tui.KeyHandler. The methods are added here rather
// than on Model (value) so that the Router can mutate state in place
// when it receives a *Model via the interface.
//
// Getter methods use Get* prefix to avoid Go's "field and method with
// the same name" restriction — the Model struct has fields named
// Screen, Tools, Cursor, etc. (the same identifiers the interface
// would have used for the getters). The Set* setters do not conflict
// and are named per Go convention.

// GetScreen returns the current screen.
func (m *Model) GetScreen() model.Screen { return m.Screen }

// SetScreen sets the current screen.
func (m *Model) SetScreen(s model.Screen) { m.Screen = s }

// GetPreviousScreen returns the previous screen for source-aware back nav.
func (m *Model) GetPreviousScreen() model.Screen { return m.PreviousScreen }

// SetPreviousScreen sets the previous screen.
func (m *Model) SetPreviousScreen(s model.Screen) { m.PreviousScreen = s }

// GetCursor returns the current cursor position.
func (m *Model) GetCursor() int { return m.Cursor }

// SetCursor sets the cursor position.
func (m *Model) SetCursor(c int) { m.Cursor = c }

// GetTools returns the current tools slice.
func (m *Model) GetTools() []model.ToolState { return m.Tools }

// SetTools replaces the tools slice.
func (m *Model) SetTools(t []model.ToolState) { m.Tools = t }

// GetErrorMsg returns the current error message.
func (m *Model) GetErrorMsg() string { return m.ErrorMsg }

// SetErrorMsg sets the current error message.
func (m *Model) SetErrorMsg(s string) { m.ErrorMsg = s }

// GetOperationMode returns the current operation mode.
func (m *Model) GetOperationMode() string { return m.OperationMode }

// SetOperationMode sets the current operation mode.
func (m *Model) SetOperationMode(mode string) { m.OperationMode = mode }

// GetUninstallConfirming returns whether the uninstall confirmation prompt is active.
func (m *Model) GetUninstallConfirming() bool { return m.UninstallConfirming }

// SetUninstallConfirming sets whether the uninstall confirmation prompt is active.
func (m *Model) SetUninstallConfirming(b bool) { m.UninstallConfirming = b }

// GetQuitting returns whether the model has initiated exit.
func (m *Model) GetQuitting() bool { return m.Quitting }

// SetQuitting sets whether the model has initiated exit.
func (m *Model) SetQuitting(q bool) { m.Quitting = q }

// GetInstallCompleted returns the count of completed tool installs.
func (m *Model) GetInstallCompleted() int { return m.InstallCompleted }

// SetInstallCompleted sets the count of completed tool installs.
func (m *Model) SetInstallCompleted(n int) { m.InstallCompleted = n }

// GetInstallFailed returns the count of failed tool installs.
func (m *Model) GetInstallFailed() int { return m.InstallFailed }

// SetInstallFailed sets the count of failed tool installs.
func (m *Model) SetInstallFailed(n int) { m.InstallFailed = n }

// GetInstallWarned returns the count of warned tool installs.
func (m *Model) GetInstallWarned() int { return m.InstallWarned }

// SetInstallWarned sets the count of warned tool installs.
func (m *Model) SetInstallWarned(n int) { m.InstallWarned = n }

// InstallProgressCount returns the number of tools currently tracked
// in the per-tool install progress list. The Router uses this to
// compute the auto-transition threshold for the InstallProgress screen
// without needing direct access to the []screens.ProgressTool slice
// (which would create an import cycle: internal/tui -> internal/tui/screens
// -> internal/tui for NavigateMsg).
func (m *Model) InstallProgressCount() int { return len(m.ProgressTools) }

// Cancel invokes the pipeline context cancel function. No-op if the
// cancel function has not been initialized (e.g., in unit tests that
// build a Model manually).
func (m *Model) Cancel() {
	if m.cancel != nil {
		m.cancel()
	}
}

// StartPipeline is the exported wrapper that satisfies tui.KeyHandler.
// It delegates to the unexported startPipeline method which is the
// existing implementation.
func (m *Model) StartPipeline(mode string) tea.Cmd {
	return m.startPipeline(mode)
}

// UpdateScreenKey is the temporary bridge that satisfies
// tui.KeyHandler. It delegates to the unexported updateScreenKey
// function in internal/app/update.go. The Router's DispatchKey stub
// (PR7 commit 2) calls this for behavior preservation during the
// PR7 migration. After all screens are migrated to per-screen
// handleX methods (commits 3-9), this method is removed (commit 10)
// along with the legacy updateScreenKey function.
//
// The method takes a value-typed *Model receiver so it can satisfy
// the KeyHandler interface (which is defined on *Model), and the
// returned tea.Model is the modified Model value (preserving the
// existing type contract used by Bubbletea).
func (m *Model) UpdateScreenKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.updateScreenKey(msg)
}
