package cursor_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Crisbr10/sequoia/adapters/cursor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAdapter(t *testing.T) *cursor.Adapter {
	t.Helper()
	return cursor.NewAdapter(t.TempDir())
}

func TestPaths_SkillsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SkillsPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".cursor/rules"),
		"expected path to end with .cursor/rules, got %s", p)
}

func TestPaths_CommandsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.CommandsPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".cursor/rules"),
		"expected path to end with .cursor/rules, got %s", p)
}

func TestPaths_SystemPromptPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SystemPromptPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".cursor/rules/sequoia-ai.md"),
		"expected path to end with .cursor/rules/sequoia-ai.md, got %s", p)
}

func TestPaths_VersionFilePath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := filepath.Join(a.SkillsPath(), ".sequoia-version")
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".cursor/rules/.sequoia-version"),
		"expected path to end with .cursor/rules/.sequoia-version, got %s", p)
}

func TestPaths_HomeResolution_Error(t *testing.T) {
	t.Parallel()
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist")
	a := cursor.NewAdapter(nonexistent)

	p := a.SkillsPath()
	assert.NotEmpty(t, p, "SkillsPath should not be empty when EvalSymlinks fails")
	assert.Contains(t, filepath.ToSlash(p), ".cursor/rules",
		"SkillsPath should contain expected suffix even with fallback path")
}

func TestPaths_SymlinkResolved(t *testing.T) {
	realHome := t.TempDir()
	linkHome := filepath.Join(t.TempDir(), "link-home")

	cursorDir := filepath.Join(realHome, ".cursor", "rules")
	require.NoError(t, os.MkdirAll(cursorDir, 0o755))

	if err := os.Symlink(realHome, linkHome); err != nil {
		t.Skipf("os.Symlink not available (may require admin on Windows): %v", err)
	}

	a := cursor.NewAdapter(linkHome)
	p := a.SkillsPath()

	// Resolve realHome via EvalSymlinks so the comparison works on
	// platforms where TempDir returns a path needing canonicalisation
	// (macOS /var → /private/var, Windows short names).
	absReal, err := filepath.EvalSymlinks(realHome)
	require.NoError(t, err)
	assert.Contains(t, p, absReal,
		"SkillsPath should use resolved (real) path, got %s", p)
	assert.NotContains(t, p, linkHome,
		"SkillsPath should NOT use symlink path, got %s", p)
}
