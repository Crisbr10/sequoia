package codex_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Crisbr10/sequoia/adapters/codex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAdapter(t *testing.T) *codex.Adapter {
	t.Helper()
	return codex.NewAdapter(t.TempDir())
}

func TestPaths_SkillsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SkillsPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".codex/sequoia/skills"),
		"expected path to end with .codex/sequoia/skills, got %s", p)
}

func TestPaths_CommandsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.CommandsPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".codex/sequoia/commands"),
		"expected path to end with .codex/sequoia/commands, got %s", p)
}

func TestPaths_SystemPromptPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SystemPromptPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".codex/config.toml"),
		"expected path to end with .codex/config.toml, got %s", p)
}

func TestPaths_VersionFilePath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := filepath.Join(a.SkillsPath(), ".sequoia-version")
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".codex/sequoia/skills/.sequoia-version"),
		"expected path to end with .codex/sequoia/skills/.sequoia-version, got %s", p)
}

func TestPaths_HomeResolution_Error(t *testing.T) {
	t.Parallel()
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist")
	a := codex.NewAdapter(nonexistent)

	p := a.SkillsPath()
	assert.NotEmpty(t, p, "SkillsPath should not be empty when EvalSymlinks fails")
	assert.Contains(t, filepath.ToSlash(p), ".codex/sequoia",
		"SkillsPath should contain expected suffix even with fallback path")
}

func TestPaths_SymlinkResolved(t *testing.T) {
	realHome := t.TempDir()
	linkHome := filepath.Join(t.TempDir(), "link-home")

	codexDir := filepath.Join(realHome, ".codex", "sequoia", "skills")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	if err := os.Symlink(realHome, linkHome); err != nil {
		t.Skipf("os.Symlink not available (may require admin on Windows): %v", err)
	}

	a := codex.NewAdapter(linkHome)
	p := a.SkillsPath()

	assert.Contains(t, p, realHome,
		"SkillsPath should use resolved (real) path, got %s", p)
	assert.NotContains(t, p, linkHome,
		"SkillsPath should NOT use symlink path, got %s", p)
}
