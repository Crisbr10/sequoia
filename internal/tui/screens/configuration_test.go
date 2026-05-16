package screens_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

func TestConfigurationView_ShowsPersistenceOptions(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}
	view := screens.ConfigurationView(config, 0, true)

	// All three persistence options must be visible.
	assert.Contains(t, view, "Engram", "Should show Engram option")
	assert.Contains(t, view, "Files", "Should show Files option")
	assert.Contains(t, view, "Both", "Should show Both option")

	// Current persistence (engram) should be highlighted.
	assert.Contains(t, view, "► Engram", "Active persistence should have cursor indicator")
}

func TestConfigurationView_EngramGreyedOutWhenUnavailable(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}
	view := screens.ConfigurationView(config, 0, false)

	// When engram is unavailable, a note should appear.
	assert.Contains(t, view, "not detected", "Should show not-detected note when Engram unavailable")
}

func TestConfigurationView_ShowsNavigationHints(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}
	view := screens.ConfigurationView(config, 0, true)

	// Should show navigation hints.
	assert.Contains(t, view, "↑/↓/←/→", "Should show arrow keys for cycling options")
	assert.Contains(t, view, "change option", "Should show change option hint")
	assert.Contains(t, view, "Enter", "Should show Enter hint")
	assert.Contains(t, view, "Esc", "Should show Esc hint")
}

func TestConfigurationUpdate_LeftRightChangesPersistence(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}

	// Active field = 0 (persistence). Press right → Files.
	msg := tea.KeyMsg{Type: tea.KeyRight}
	_, newConfig, action := screens.ConfigurationUpdate(msg, 0, config, true)
	assert.Equal(t, "files", string(newConfig.Persistence), "Right arrow should cycle Engram → Files")
	assert.Empty(t, action)

	// Press right → Both.
	_, newConfig, action = screens.ConfigurationUpdate(msg, 0, newConfig, true)
	assert.Equal(t, "both", string(newConfig.Persistence), "Right arrow should cycle Files → Both")
	assert.Empty(t, action)

	// Press right → wrap to Engram.
	_, newConfig, action = screens.ConfigurationUpdate(msg, 0, newConfig, true)
	assert.Equal(t, "engram", string(newConfig.Persistence), "Right arrow should cycle Both → Engram")
	assert.Empty(t, action)

	// Press left → wrap to Both.
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	_, newConfig, action = screens.ConfigurationUpdate(msg, 0, config, true)
	assert.Equal(t, "both", string(newConfig.Persistence), "Left arrow should cycle Engram → Both")
	assert.Empty(t, action)
}

func TestConfigurationUpdate_EnterNavigatesToInstallProgress(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}

	// Enter should confirm and proceed.
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _, action := screens.ConfigurationUpdate(msg, 0, config, true)
	assert.Equal(t, "confirm", action, "Enter should confirm configuration")
}

func TestConfigurationUpdate_EscNavigatesBack(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, _, action := screens.ConfigurationUpdate(msg, 0, config, true)
	assert.Equal(t, "back", action, "Esc should go back to ToolSelection")
}

func TestConfigurationUpdate_UpDownCyclesPersistence(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}

	// Down: engram → files → both → engram
	msgDown := tea.KeyMsg{Type: tea.KeyDown}
	newField, nc, _ := screens.ConfigurationUpdate(msgDown, 0, config, true)
	assert.Equal(t, 0, newField, "active field should stay on persistence")
	assert.Equal(t, "files", string(nc.Persistence))

	_, nc, _ = screens.ConfigurationUpdate(msgDown, 0, nc, true)
	assert.Equal(t, "both", string(nc.Persistence))

	_, nc, _ = screens.ConfigurationUpdate(msgDown, 0, nc, true)
	assert.Equal(t, "engram", string(nc.Persistence), "should wrap back to engram")

	// Up: engram → both → files → engram
	msgUp := tea.KeyMsg{Type: tea.KeyUp}
	newField, nc, _ = screens.ConfigurationUpdate(msgUp, 0, config, true)
	assert.Equal(t, 0, newField)
	assert.Equal(t, "both", string(nc.Persistence))

	_, nc, _ = screens.ConfigurationUpdate(msgUp, 0, nc, true)
	assert.Equal(t, "files", string(nc.Persistence))

	_, nc, _ = screens.ConfigurationUpdate(msgUp, 0, nc, true)
	assert.Equal(t, "engram", string(nc.Persistence))
}

func TestConfigurationUpdate_QNoLongerReturnsQuit(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}

	// 'q' is handled globally in update.go before screen delegation.
	// ConfigurationUpdate should NOT return "quit" for 'q' — it's dead code.
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, _, action := screens.ConfigurationUpdate(msg, 0, config, true)

	assert.Empty(t, action, "q should not return quit from ConfigurationUpdate (handled globally)")
}

func TestConfigurationView_Golden_Standard(t *testing.T) {
	config := model.TUIConfig{Persistence: "engram"}
	view := screens.ConfigurationView(config, 0, true)

	golden := goldenPath("configuration_standard.txt")
	if updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0755))
		require.NoError(t, os.WriteFile(golden, []byte(view), 0644))
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing — run with UPDATE_GOLDEN=1 to generate")
	assert.Equal(t, string(expected), view, "golden file mismatch — run with UPDATE_GOLDEN=1 to regenerate")
}

func TestConfigurationView_Golden_EngramUnavailable(t *testing.T) {
	config := model.TUIConfig{Persistence: "files"}
	view := screens.ConfigurationView(config, 0, false)

	golden := goldenPath("configuration_engram_unavailable.txt")
	if updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0755))
		require.NoError(t, os.WriteFile(golden, []byte(view), 0644))
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing — run with UPDATE_GOLDEN=1 to generate")
	assert.Equal(t, string(expected), view, "golden file mismatch — run with UPDATE_GOLDEN=1 to regenerate")
}

func TestConfigurationView_NonEmptyView(t *testing.T) {
	t.Parallel()

	config := model.TUIConfig{Persistence: "engram"}
	view := screens.ConfigurationView(config, 0, true)

	assert.NotEmpty(t, view, "Configuration view should not be empty")
	lines := strings.Split(strings.TrimSpace(view), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Configuration view should span at least 3 lines")
}
