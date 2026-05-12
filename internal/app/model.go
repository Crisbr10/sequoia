// Package app provides the root Bubbletea Model for the Sequoia TUI installer.
// It manages screen state, tool registry, user configuration, and the progress
// channel used for async pipeline communication.
package app

import (
	"context"
	"os/exec"

	"github.com/Crisbr10/sequoia/adapters"
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
	// Config holds user choices from the Configuration screen (language, persistence).
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
}

// EngramDetectedMsg is a Bubbletea message sent when the asynchronous engram
// detection completes. The boolean value indicates whether engram was found on PATH.
type EngramDetectedMsg bool

// toolInfoAdapter wraps adapters.ToolAdapter to satisfy model.ToolInfo.
// It adapts the Status() method to return model.ToolStatus instead of
// adapters.AdapterStatus, breaking the internal/model → adapters dependency.
type toolInfoAdapter struct {
	adapters.ToolAdapter
}

// Status returns the adapter's installation status as a model.ToolStatus.
func (a toolInfoAdapter) Status() model.ToolStatus {
	s := a.ToolAdapter.Status()
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
// Engram detection is deferred to Init() via detectEngram() to avoid
// blocking the first TUI render on exec.LookPath.
func NewModel(toolID string, version string) Model {
	all := adapters.DefaultRegistry.All()
	tools := make([]model.ToolState, 0, len(all))
	for _, a := range all {
		ts := model.ToolState{
			Adapter:  toolInfoAdapter{a},
			Selected: toolID == "" || a.ID() == toolID,
		}
		tools = append(tools, ts)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		Version:         version,
		Screen:          model.ScreenWelcome,
		Tools:           tools,
		Config:          model.TUIConfig{Language: "en", Persistence: "engram"},
		Progress:        make(chan model.ProgressMsg, 64),
		EngramAvailable: false,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// detectEngram is an async Bubbletea command that checks whether the engram
// binary is available on the system PATH. It returns EngramDetectedMsg so
// the Model can update EngramAvailable without blocking the initial render.
func detectEngram() tea.Msg {
	_, err := exec.LookPath("engram")
	return EngramDetectedMsg(err == nil)
}

// Init is the Bubbletea init command. It returns the batched initial commands:
// detectEngram runs asynchronously to avoid blocking the first TUI render.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		detectEngram,
	)
}
