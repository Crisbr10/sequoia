// Package app provides the root Bubbletea Model for the Sequoia TUI installer.
// It manages screen state, tool registry, user configuration, and the progress
// channel used for async pipeline communication.
package app

import (
	"context"
	"os/exec"

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
	// Config holds user choices from the Configuration screen (persistence backend).
	Config model.TUIConfig
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
	// EngramAvailable indicates whether the Engram MCP backend was detected at startup.
	// When false, the Engram option on the Configuration screen is greyed out.
	EngramAvailable bool
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

// EngramDetectedMsg is a Bubbletea message sent when the asynchronous engram
// detection completes. The boolean value indicates whether engram was found on PATH.
type EngramDetectedMsg bool

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
		Version:         version,
		Screen:          model.ScreenWelcome,
		Tools:           nil, // loaded lazily by loadTools()
		Config:          model.TUIConfig{Persistence: "engram"},
		Progress:        make(chan model.ProgressMsg, model.ProgressChannelBufferSize),
		EngramAvailable: false,
		reg:             reg,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// LoadTools populates m.Tools from the registry if not already loaded.
// It is safe to call multiple times; subsequent calls are no-ops.
// toolID is the optional adapter ID to pre-select; empty means select all.
// Exported for test use; production code calls this internally in updateScreenKey.
func (m *Model) LoadTools(toolID string) {
	if m.Tools != nil {
		return // already loaded
	}
	all := m.reg.All()
	m.Tools = make([]model.ToolState, 0, len(all))
	for _, a := range all {
		ts := model.ToolState{
			Adapter: toolInfoAdapter{
				Identifier:    a,
				Detector:      a,
				InstallStatus: a,
				Installer:     a,
				Strategy:      a.(common.Strategy),
			},
			Selected: toolID == "" || a.ID() == toolID,
		}
		m.Tools = append(m.Tools, ts)
	}
}

// detectEngram is an async Bubbletea command that checks whether the engram
// binary is available on the system PATH. It returns EngramDetectedMsg so
// the Model can update EngramAvailable without blocking the initial render.
func detectEngram() tea.Msg {
	_, err := exec.LookPath("engram")
	return EngramDetectedMsg(err == nil)
}

// Init is the Bubbletea init command. It returns the batched initial
// commands: detectEngram runs asynchronously to avoid blocking the first
// TUI render.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		detectEngram,
	)
}
