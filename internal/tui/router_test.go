package tui_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui"
)

func TestNavigateMsg_StillExists(t *testing.T) {
	t.Parallel()

	// Verify NavigateMsg type is defined and can be created.
	msg := tui.NavigateMsg{Target: model.ScreenWelcome}

	require.Equal(t, model.ScreenWelcome, msg.Target,
		"NavigateMsg.Target should be ScreenWelcome")

	// Verify it can target other screens too.
	msg2 := tui.NavigateMsg{Target: model.ScreenToolSelection}
	require.Equal(t, model.ScreenToolSelection, msg2.Target,
		"NavigateMsg should support other screen targets")
}
