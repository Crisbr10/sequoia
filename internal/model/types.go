// Package model defines the shared domain types used across the TUI installer.
// Types in this package have no external dependencies and serve as the
// single source of truth for screen state, tool state, and configuration.
package model

// Screen represents a distinct screen in the TUI state machine.
type Screen int

const (
	// ScreenWelcome is the initial landing screen showing branding and tool detection.
	ScreenWelcome Screen = iota
	// ScreenToolSelection shows a checkbox list for picking tools to install.
	ScreenToolSelection
	// ScreenConfiguration offers language and persistence backend choices.
	ScreenConfiguration
	// ScreenInstallProgress displays per-tool step-by-step progress during installation.
	ScreenInstallProgress
	// ScreenComplete shows the success summary after all tools are installed.
	ScreenComplete
	// ScreenError displays failed tools with retry/quit options.
	ScreenError
	// ScreenStatus shows the current installation state for all tools.
	ScreenStatus
	// ScreenUninstall provides checkbox selection and confirmation for removal.
	ScreenUninstall

	// ScreenCount tracks the total number of Screen constants.
	ScreenCount
)

// ToolStatus mirrors adapters.AdapterStatus without importing the adapters package.
// It reports the current installation state of a tool.
type ToolStatus struct {
	// Installed reports whether Sequoia content is present for this tool.
	Installed bool
	// Version is the Sequoia version string present in the installation, or "".
	Version string
	// Path is the absolute, OS-correct root installation path.
	Path string
}

// ToolInfo is a local interface that ToolState.Adapter must satisfy.
// It includes only the methods needed by TUI consumers, breaking the
// internal/model → adapters dependency. adapters.ToolAdapter can be
// adapted to satisfy this interface via a thin wrapper.
type ToolInfo interface {
	// ID returns the unique machine-readable identifier.
	ID() string
	// Name returns the human-readable display name.
	Name() string
	// IsInstalled reports whether Sequoia has already been installed for this tool.
	IsInstalled() bool
	// Status returns the current installation status.
	Status() ToolStatus
	// Detect reports whether the tool is installed on this machine.
	Detect() bool
}

// ToolState groups a ToolInfo implementation with its TUI selection state and
// installation result.
type ToolState struct {
	// Adapter is the registered tool adapter providing TUI-visible behavior.
	Adapter ToolInfo
	// Selected indicates whether the user has toggled this tool for installation.
	Selected bool
	// Result holds the outcome of installation, populated after the pipeline completes.
	Result *InstallResult
}

// InstallResult reports the final outcome of an installation pipeline run.
type InstallResult struct {
	// ToolID identifies the tool this result is for.
	ToolID string
	// Success is true when all steps completed without error.
	Success bool
	// Error contains the first error encountered, or "" on success.
	Error string
	// Steps records per-step status in execution order.
	Steps []StepResult
}

// StepResult records the outcome of a single pipeline step.
type StepResult struct {
	// Name identifies the step ("prepare", "apply", or "verify").
	Name string
	// Done is true when the step has completed (successfully or not).
	Done bool
	// Error contains the error message if the step failed.
	Error string
}

// ProgressMsg is sent from install goroutines through a buffered channel
// to the TUI event loop so the UI updates without blocking the pipeline.
type ProgressMsg struct {
	// ToolID identifies which tool's progress is being reported.
	ToolID string
	// Step names the current pipeline step ("prepare", "apply", "verify").
	Step string
	// Done reports whether this step has completed.
	Done bool
	// Error contains the error message if the step failed.
	Error string
	// Warning is true when the step completed but with non-fatal warnings
	// (e.g., partial uninstall where some files could not be removed).
	Warning bool
	// FailedCount reports how many individual sub-operations failed during
	// a partial failure (used for warning messages like "3 files could not be removed").
	FailedCount int
}

// Language represents a supported UI language.
type Language string

const (
	// LangEN is English.
	LangEN Language = "en"
	// LangES is Spanish.
	LangES Language = "es"
)

// PersistenceBackend represents a supported artifact persistence backend.
type PersistenceBackend string

const (
	// PersistenceEngram uses the Engram MCP server for artifact storage.
	PersistenceEngram PersistenceBackend = "engram"
	// PersistenceFiles uses the local filesystem (openspec) for artifact storage.
	PersistenceFiles PersistenceBackend = "files"
	// PersistenceBoth writes artifacts to both Engram and the local filesystem.
	PersistenceBoth PersistenceBackend = "both"
)

// TUIConfig holds the user's choices from the Configuration screen.
type TUIConfig struct {
	// Language sets the UI language ("en" or "es").
	Language string
	// Persistence selects the artifact storage backend ("engram", "files", or "both").
	Persistence string
}
