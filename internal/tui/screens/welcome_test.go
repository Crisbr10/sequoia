// Package screens_test provides tests for the TUI screen implementations.
package screens_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "github.com/charmbracelet/bubbletea"

	"sequoia-ai/adapters"
	"sequoia-ai/internal/model"
	"sequoia-ai/internal/tui"
	"sequoia-ai/internal/tui/screens"
)

// dummyAdapter is a minimal ToolAdapter for screen tests.
type dummyAdapter struct {
	id   string
	name string
	inst bool
}

func (d *dummyAdapter) ID() string                     { return d.id }
func (d *dummyAdapter) Name() string                   { return d.name }
func (d *dummyAdapter) Detect() bool                   { return true }
func (d *dummyAdapter) IsInstalled() bool              { return d.inst }
func (d *dummyAdapter) Install() error                 { return nil }
func (d *dummyAdapter) Uninstall() error               { return nil }
func (d *dummyAdapter) Status() adapters.AdapterStatus {
	return adapters.AdapterStatus{Installed: d.inst}
}
func (d *dummyAdapter) SkillsPath() string       { return "" }
func (d *dummyAdapter) CommandsPath() string     { return "" }
func (d *dummyAdapter) SystemPromptPath() string { return "" }
func (d *dummyAdapter) PromptStrategy() adapters.PromptStrategy {
	return adapters.StrategyMarkdownSections
}

var _ adapters.ToolAdapter = (*dummyAdapter)(nil)

func TestWelcomeView_ContainsBrandingElements(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{}
	version := "v0.1.0"
	view := screens.WelcomeView(tools, version)

	// Must contain the ASCII logo.
	assert.Contains(t, view, "Sequoia", "Welcome screen should contain Sequoia branding")

	// Must contain the version string.
	assert.Contains(t, view, version, "Welcome screen should display the version")

	// Must contain the tagline.
	assert.Contains(t, view, "Audit quality for AI coding tools",
		"Welcome screen should contain the tagline")

	// Must contain navigation hint.
	assert.Contains(t, view, "Enter",
		"Welcome screen should show Enter key hint")
	assert.Contains(t, view, "quit",
		"Welcome screen should show quit hint")

	// Must be non-empty and multi-line.
	assert.NotEmpty(t, view, "Welcome view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 5, "Welcome view should span at least 5 lines of content")
}

func TestWelcomeView_ListsToolsByName(t *testing.T) {
	t.Parallel()

	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code"}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode"}, Selected: false},
		{Adapter: &dummyAdapter{id: "gemini", name: "Gemini CLI"}, Selected: false},
	}
	view := screens.WelcomeView(tools, "v0.1.0")

	// Each tool name must appear in the view.
	for _, ts := range tools {
		assert.Contains(t, view, ts.Adapter.Name(),
			"Welcome view should list tool %q by name", ts.Adapter.Name())
	}
}

func TestWelcomeView_ShowsInstallStatus(t *testing.T) {
	t.Parallel()

	// One tool installed, one not.
	tools := []model.ToolState{
		{
			Adapter:  &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true},
			Selected: false,
		},
		{
			Adapter:  &dummyAdapter{id: "opencode", name: "OpenCode", inst: false},
			Selected: false,
		},
	}
	view := screens.WelcomeView(tools, "v0.1.0")

	// The view should contain tool names.
	assert.Contains(t, view, "Claude Code",
		"Welcome view should show tool name")
	assert.Contains(t, view, "OpenCode",
		"Welcome view should show tool name")

	// The view should use distinct visual indicators for installed vs not-installed.
	hasInstalled := strings.Contains(view, "✓") || strings.Contains(view, "installed")
	hasNotInstalled := strings.Contains(view, "✗") || strings.Contains(view, "not installed")
	assert.True(t, hasInstalled || hasNotInstalled,
		"Welcome view should show install status indicators")
}

func TestWelcomeUpdate_EnterReturnsNavigateToToolSelection(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := screens.WelcomeUpdate(msg)

	require.NotNil(t, cmd, "Enter key should produce a command on Welcome screen")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Enter should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"Enter on Welcome should navigate to ToolSelection")
}

func TestWelcomeUpdate_RightArrowReturnsNavigateToToolSelection(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRight}
	cmd := screens.WelcomeUpdate(msg)

	require.NotNil(t, cmd, "Right arrow should produce a command on Welcome screen")
	result := cmd()
	nav, ok := result.(tui.NavigateMsg)
	require.True(t, ok, "Right arrow should produce NavigateMsg, got %T", result)
	assert.Equal(t, model.ScreenToolSelection, nav.Target,
		"Right arrow on Welcome should navigate to ToolSelection")
}

func TestWelcomeUpdate_QReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	cmd := screens.WelcomeUpdate(msg)

	require.NotNil(t, cmd, "q key should produce a command on Welcome screen")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "q key should produce tea.QuitMsg, got %T", result)
}

func TestWelcomeUpdate_CtrlCReturnsQuit(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	cmd := screens.WelcomeUpdate(msg)

	require.NotNil(t, cmd, "ctrl+c should produce a command on Welcome screen")
	result := cmd()
	_, ok := result.(tea.QuitMsg)
	assert.True(t, ok, "ctrl+c should produce tea.QuitMsg, got %T", result)
}

func TestWelcomeUpdate_UnknownKeyReturnsNil(t *testing.T) {
	t.Parallel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	cmd := screens.WelcomeUpdate(msg)

	assert.Nil(t, cmd, "Unknown key should produce no command on Welcome screen")
}
