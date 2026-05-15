// Package adapters defines the ToolAdapter contract and the infrastructure
// (Registry, Factory) for registering and retrieving tool integrations.
package adapters

import "context"

// PromptStrategy defines how Sequoia injects content into a tool's config.
type PromptStrategy int

const (
	// StrategyMarkdownSections injects a delimited section using start/end markers.
	StrategyMarkdownSections PromptStrategy = iota
	// StrategyFileReplace replaces the entire file, creating a backup first.
	StrategyFileReplace
	// StrategyConfigMerge injects a delimited section using start/end markers
	// for tools whose config format does not match Markdown sections but
	// still uses markers to delimit Sequoia content (e.g. Gemini CLI GEMINI.md).
	StrategyConfigMerge
	// StrategyTOMLMerge merges a [sequoia] TOML table into an existing TOML
	// config file, preserving all pre-existing keys and sections (e.g. Codex config.toml).
	StrategyTOMLMerge
)

// InstallOpts carries optional configuration for Install and Uninstall calls.
// It is designed to be extended with additional fields as needed without
// breaking the adapter interface (pass-by-value).
type InstallOpts struct {
	// Language is the ISO 639-1 or BCP 47 code (e.g. "en", "es", "pt-BR")
	// for the language to use when rendering templates and agent docs.
	Language string

	// Context is an optional context for cancellation propagation.
	// When set and cancelled, Install and Uninstall should abort early
	// and roll back any partial work. A nil Context means no cancellation
	// support (backwards-compatible with existing callers).
	Context context.Context
}

// AdapterStatus reports the current installation state of a tool adapter.
type AdapterStatus struct {
	// Installed reports whether Sequoia content is present for this tool.
	Installed bool
	// Version is the Sequoia version string present in the installation, or ""
	// if Sequoia has not been installed.
	Version string
	// Path is the absolute, OS-correct root installation path.
	Path string
}

// ToolAdapter is the contract every tool integration must satisfy.
// Each concrete adapter lives in its own sub-package (e.g. adapters/claude)
// and self-registers via its init() function.
type ToolAdapter interface {
	// ID returns the unique machine-readable identifier (e.g. "claude-code").
	ID() string
	// Name returns the human-readable display name.
	Name() string
	// Detect reports whether the tool is installed on this machine.
	Detect() bool
	// IsInstalled reports whether Sequoia has already been installed for this tool.
	IsInstalled() bool
	// Install installs Sequoia files for this tool.
	// opts carries optional configuration such as the target language.
	Install(opts InstallOpts) error
	// Uninstall removes Sequoia files for this tool.
	// opts carries optional configuration such as the target language.
	Uninstall(opts InstallOpts) error
	// Status returns the current installation status.
	Status() AdapterStatus
	// SkillsPath returns the absolute path to the skills directory for this tool.
	SkillsPath() string
	// CommandsPath returns the absolute path to the commands directory for this tool.
	CommandsPath() string
	// SystemPromptPath returns the absolute path to the system prompt file for this tool.
	SystemPromptPath() string
	// PromptStrategy returns the injection strategy used by this adapter.
	PromptStrategy() PromptStrategy
}

// BackupDirGetter is an optional interface that adapters may implement
// to expose the last backup directory path after Install/Uninstall.
// REQ-BUG-004: the pipeline queries this to surface backup feedback.
type BackupDirGetter interface {
	LastBackupDir() string
}
