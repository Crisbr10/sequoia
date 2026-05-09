// Package tui provides the terminal user interface for the Sequoia installer.
// It contains the screen router, styles, and screen renderers.
package tui

import (
	"sequoia-ai/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

// TransitionMap defines all valid forward transitions between screens.
// Each key lists the screens that can be reached from the source screen.
// Terminal screens (Complete) have empty slices.
var TransitionMap = map[model.Screen][]model.Screen{
	model.ScreenWelcome:         {model.ScreenToolSelection},
	model.ScreenToolSelection:   {model.ScreenConfiguration},
	model.ScreenConfiguration:   {model.ScreenInstallProgress},
	model.ScreenInstallProgress: {model.ScreenComplete, model.ScreenError},
	model.ScreenComplete:        {model.ScreenStatus},
	model.ScreenError:           {model.ScreenInstallProgress, model.ScreenToolSelection},
	model.ScreenStatus:          {model.ScreenUninstall},
	model.ScreenUninstall:       {model.ScreenInstallProgress},
}

// IsValidTransition reports whether the transition from → to is allowed
// by the screen state machine defined in TransitionMap.
func IsValidTransition(from, to model.Screen) bool {
	targets, ok := TransitionMap[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// NextScreen resolves the next screen for the given action from the
// current screen. The action string encodes the semantic trigger
// (e.g., "enter", "success", "fail", "retry", "uninstall").
// If the action is unrecognized or the transition is invalid, the
// current screen is returned unchanged.
func NextScreen(current model.Screen, action string) model.Screen {
	switch current {
	case model.ScreenWelcome:
		if action == "enter" {
			return model.ScreenToolSelection
		}
	case model.ScreenToolSelection:
		if action == "enter" {
			return model.ScreenConfiguration
		}
	case model.ScreenConfiguration:
		if action == "enter" {
			return model.ScreenInstallProgress
		}
	case model.ScreenInstallProgress:
		switch action {
		case "success":
			return model.ScreenComplete
		case "fail":
			return model.ScreenError
		}
	case model.ScreenComplete:
		if action == "status" {
			return model.ScreenStatus
		}
	case model.ScreenError:
		switch action {
		case "retry":
			return model.ScreenInstallProgress
		case "back":
			return model.ScreenToolSelection
		}
	case model.ScreenStatus:
		if action == "uninstall" {
			return model.ScreenUninstall
		}
	case model.ScreenUninstall:
		if action == "enter" {
			return model.ScreenInstallProgress
		}
	}
	return current
}

// NavigateMsg is a Bubbletea message instructing the root model to
// switch to the given screen.
type NavigateMsg struct {
	// Target is the screen to navigate to.
	Target model.Screen
}

// ScreenRouter manages screen navigation and enforces transition rules.
type ScreenRouter interface {
	// NavigateTo attempts to transition to the given screen. If the
	// transition is valid, it returns a tea.Cmd that emits a NavigateMsg.
	// If the transition is invalid, it returns nil (no-op).
	NavigateTo(screen model.Screen) tea.Cmd

	// CurrentScreen returns the currently active screen.
	CurrentScreen() model.Screen
}

// router is the default ScreenRouter implementation backed by TransitionMap.
type router struct {
	current model.Screen
}

// NewRouter creates a ScreenRouter starting at ScreenWelcome.
func NewRouter() ScreenRouter {
	return &router{current: model.ScreenWelcome}
}

// CurrentScreen returns the router's current screen.
func (r *router) CurrentScreen() model.Screen {
	return r.current
}

// NavigateTo checks the transition and, if valid, updates the internal
// current screen and returns a tea.Cmd that emits NavigateMsg.
func (r *router) NavigateTo(screen model.Screen) tea.Cmd {
	if !IsValidTransition(r.current, screen) {
		return nil
	}
	r.current = screen
	return func() tea.Msg {
		return NavigateMsg{Target: screen}
	}
}
