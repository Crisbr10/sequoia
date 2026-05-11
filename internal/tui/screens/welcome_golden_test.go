package screens_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/tui/screens"
)

// goldenPath returns the path to a golden file for the given name.
func goldenPath(name string) string {
	return filepath.Join("testdata", "golden", name)
}

// updateGolden is true when UPDATE_GOLDEN=1 env var is set.
var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

func TestWelcomeView_Golden_StandardTerminal(t *testing.T) {
	tools := []model.ToolState{
		{Adapter: &dummyAdapter{id: "claude-code", name: "Claude Code", inst: true}, Selected: false},
		{Adapter: &dummyAdapter{id: "opencode", name: "OpenCode", inst: false}, Selected: false},
	}
	version := "v0.1.0"
	view := screens.WelcomeView(tools, version)

	golden := goldenPath("welcome_standard.txt")
	if updateGolden {
		dir := filepath.Dir(golden)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(golden, []byte(view), 0644))
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing — run with UPDATE_GOLDEN=1 go test ... to generate")
	assert.Equal(t, string(expected), view, "golden file mismatch — run with UPDATE_GOLDEN=1 to regenerate")
}

func TestWelcomeView_Golden_EmptyTools(t *testing.T) {
	tools := []model.ToolState{}
	version := "v0.1.0"
	view := screens.WelcomeView(tools, version)

	golden := goldenPath("welcome_empty_tools.txt")
	if updateGolden {
		dir := filepath.Dir(golden)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(golden, []byte(view), 0644))
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing — run with UPDATE_GOLDEN=1 go test ... to generate")
	assert.Equal(t, string(expected), view, "golden file mismatch — run with UPDATE_GOLDEN=1 to regenerate")
}
