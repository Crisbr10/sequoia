package tui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sequoia-ai/internal/model"
	"sequoia-ai/internal/tui"
)

func TestTransitionMap_CoversAllScreens(t *testing.T) {
	t.Parallel()

	tm := tui.TransitionMap

	// Every screen constant (except ScreenCount) should be a key in the transition map.
	expectedScreens := []model.Screen{
		model.ScreenWelcome,
		model.ScreenToolSelection,
		model.ScreenConfiguration,
		model.ScreenInstallProgress,
		model.ScreenComplete,
		model.ScreenError,
		model.ScreenStatus,
		model.ScreenUninstall,
	}

	for _, s := range expectedScreens {
		_, ok := tm[s]
		assert.True(t, ok, "TransitionMap should have an entry for screen %v", s)
	}

	// ScreenCount should NOT be in the map — it's not a real screen.
	_, ok := tm[model.ScreenCount]
	assert.False(t, ok, "TransitionMap should NOT contain ScreenCount")
}

func TestTransitionMap_ValidPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from model.Screen
		to   model.Screen
		want bool
	}{
		// Happy path — forward navigation.
		{"welcome to tool selection", model.ScreenWelcome, model.ScreenToolSelection, true},
		{"tool selection to configuration", model.ScreenToolSelection, model.ScreenConfiguration, true},
		{"configuration to install progress", model.ScreenConfiguration, model.ScreenInstallProgress, true},
		// InstallProgress branches.
		{"install progress to complete", model.ScreenInstallProgress, model.ScreenComplete, true},
		{"install progress to error", model.ScreenInstallProgress, model.ScreenError, true},
		// Error retry and back.
		{"error to install progress (retry)", model.ScreenError, model.ScreenInstallProgress, true},
		{"error to tool selection (back)", model.ScreenError, model.ScreenToolSelection, true},
		// Complete to status.
		{"complete to status", model.ScreenComplete, model.ScreenStatus, true},
		// Status to uninstall.
		{"status to uninstall", model.ScreenStatus, model.ScreenUninstall, true},
		// Uninstall reuses InstallProgress.
		{"uninstall to install progress", model.ScreenUninstall, model.ScreenInstallProgress, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, tui.IsValidTransition(tc.from, tc.to),
				"transition from %v to %v should be valid", tc.from, tc.to)
		})
	}
}

func TestTransitionMap_InvalidPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from model.Screen
		to   model.Screen
	}{
		// Backward navigation not allowed.
		{"tool selection to welcome", model.ScreenToolSelection, model.ScreenWelcome},
		{"configuration to tool selection", model.ScreenConfiguration, model.ScreenToolSelection},
		// Complete only goes to Status.
		{"complete to welcome", model.ScreenComplete, model.ScreenWelcome},
		{"complete to tool selection", model.ScreenComplete, model.ScreenToolSelection},
		{"complete to install progress", model.ScreenComplete, model.ScreenInstallProgress},
		// Welcome only goes to ToolSelection.
		{"welcome to configuration", model.ScreenWelcome, model.ScreenConfiguration},
		{"welcome to install progress", model.ScreenWelcome, model.ScreenInstallProgress},
		// Self-transitions not valid.
		{"welcome to self", model.ScreenWelcome, model.ScreenWelcome},
		{"install progress to self", model.ScreenInstallProgress, model.ScreenInstallProgress},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, tui.IsValidTransition(tc.from, tc.to),
				"transition from %v to %v should be invalid", tc.from, tc.to)
		})
	}
}

func TestNextScreen_ValidTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current model.Screen
		action  string
		want    model.Screen
	}{
		{"welcome enter → tool selection", model.ScreenWelcome, "enter", model.ScreenToolSelection},
		{"tool selection enter → configuration", model.ScreenToolSelection, "enter", model.ScreenConfiguration},
		{"configuration enter → install progress", model.ScreenConfiguration, "enter", model.ScreenInstallProgress},
		{"install progress success → complete", model.ScreenInstallProgress, "success", model.ScreenComplete},
		{"install progress fail → error", model.ScreenInstallProgress, "fail", model.ScreenError},
		{"complete status → status", model.ScreenComplete, "status", model.ScreenStatus},
		{"error retry → install progress", model.ScreenError, "retry", model.ScreenInstallProgress},
		{"error back → tool selection", model.ScreenError, "back", model.ScreenToolSelection},
		{"status uninstall → uninstall", model.ScreenStatus, "uninstall", model.ScreenUninstall},
		{"uninstall enter → install progress", model.ScreenUninstall, "enter", model.ScreenInstallProgress},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tui.NextScreen(tc.current, tc.action)
			assert.Equal(t, tc.want, got, "NextScreen(%v, %q) should return %v", tc.current, tc.action, tc.want)
		})
	}
}

func TestNextScreen_InvalidAction_ReturnsSameScreen(t *testing.T) {
	t.Parallel()

	// An unknown action should return the current screen unchanged.
	got := tui.NextScreen(model.ScreenWelcome, "nonexistent")
	assert.Equal(t, model.ScreenWelcome, got, "unknown action should return current screen unchanged")
}

func TestNextScreen_InvalidTransition_ReturnsSameScreen(t *testing.T) {
	t.Parallel()

	// Enter on Complete (terminal) should stay on Complete.
	got := tui.NextScreen(model.ScreenComplete, "enter")
	assert.Equal(t, model.ScreenComplete, got, "terminal screen should stay unchanged")
}

func TestIsValidTransition_EdgeCases(t *testing.T) {
	t.Parallel()

	// ScreenCount is not a valid screen — all transitions involving it are invalid.
	assert.False(t, tui.IsValidTransition(model.ScreenWelcome, model.ScreenCount))
	assert.False(t, tui.IsValidTransition(model.ScreenCount, model.ScreenWelcome))
	assert.False(t, tui.IsValidTransition(model.ScreenCount, model.ScreenCount))
}

func TestTransitionMap_NoDuplicateTargets(t *testing.T) {
	t.Parallel()

	tm := tui.TransitionMap
	for from, targets := range tm {
		seen := make(map[model.Screen]bool)
		for _, to := range targets {
			require.False(t, seen[to],
				"duplicate target %v in transitions from %v", to, from)
			seen[to] = true
		}
	}
}
