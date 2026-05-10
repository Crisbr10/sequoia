package cursor_test

import (
	"os"
	"path/filepath"
	"testing"

	"sequoia-ai/adapters"
	"sequoia-ai/adapters/cursor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_ID(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "cursor", a.ID())
}

func TestAdapter_Name(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "Cursor IDE", a.Name())
}

func TestAdapter_PromptStrategy(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, adapters.StrategyFileReplace, a.PromptStrategy())
}

func TestAdapter_Detect_DirExists(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	cursorDir := filepath.Join(tmp, ".cursor")
	require.NoError(t, os.MkdirAll(cursorDir, 0o755))

	a := cursor.NewAdapter(tmp)
	assert.True(t, a.Detect())
}

func TestAdapter_Detect_NoDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)
	assert.False(t, a.Detect())
}

func TestAdapter_IsInstalled_FileExists(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	rulesDir := filepath.Join(tmp, ".cursor", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	sequoiaAI := filepath.Join(rulesDir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(sequoiaAI, []byte("# Sequoia rules\n"), 0o644))

	a := cursor.NewAdapter(tmp)
	assert.True(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_FileMissing(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_DirMissing(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_Status_NotInstalled(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := cursor.NewAdapter(tmp)
	s := a.Status()
	assert.False(t, s.Installed)
	assert.Equal(t, "", s.Version)
	assert.NotEmpty(t, s.Path)
}

func TestAdapter_Status_ReadsVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	rulesDir := filepath.Join(tmp, ".cursor", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	// Create sequoia-ai.md so IsInstalled() returns true.
	sequoiaAI := filepath.Join(rulesDir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(sequoiaAI, []byte("# Sequoia\n"), 0o644))

	// Write the version file.
	versionFile := filepath.Join(rulesDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte("0.2.0\n"), 0o644))

	a := cursor.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true when sequoia-ai.md exists")
	assert.Equal(t, "0.2.0", s.Version, "Status().Version should read .sequoia-version content")
}

func TestAdapter_Status_VersionMissingLegacy(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	rulesDir := filepath.Join(tmp, ".cursor", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	sequoiaAI := filepath.Join(rulesDir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(sequoiaAI, []byte("# Sequoia\n"), 0o644))

	a := cursor.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true even without version file")
	assert.Equal(t, "", s.Version, "Status().Version should be empty for legacy installs without .sequoia-version")
}

func TestAdapter_Status_HasPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	s := a.Status()
	assert.Equal(t, a.SkillsPath(), s.Path, "Status().Path should equal SkillsPath()")
	assert.True(t, filepath.ToSlash(s.Path) != "",
		"Status().Path should not be empty")
}

func TestAdapter_SkillsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SkillsPath()
	assert.NotEmpty(t, p)
}

func TestAdapter_CommandsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.CommandsPath()
	assert.NotEmpty(t, p)
}

func TestAdapter_SystemPromptPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SystemPromptPath()
	assert.NotEmpty(t, p)
}
