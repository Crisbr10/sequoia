// Package tui provides the terminal user interface for the Sequoia installer.
// It contains screen navigation messages, styles, and screen renderers.
package tui

import (
	"github.com/Crisbr10/sequoia/internal/model"
)

// NavigateMsg is a Bubbletea message instructing the root model to
// switch to the given screen.
type NavigateMsg struct {
	// Target is the screen to navigate to.
	Target model.Screen
}
