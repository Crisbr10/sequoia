package codex_test

import (
	"os"
	"path/filepath"
	"testing"

	"sequoia-ai/adapters"
	"sequoia-ai/adapters/codex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_ID(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "codex", a.ID())
}

func TestAdapter_Name(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "OpenAI Codex", a.Name())
}

func TestAdapter_PromptStrategy(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, adapters.StrategyTOMLMerge, a.PromptStrategy())
}

func TestAdapter_Detect_DirExists(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	a := codex.NewAdapter(tmp)
	assert.True(t, a.Detect())
}

func TestAdapter_Detect_NoDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := codex.NewAdapter(tmp)
	assert.False(t, a.Detect())
}

func TestAdapter_IsInstalled_DirAndTablePresent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	sequoiaDir := filepath.Join(codexDir, "sequoia")
	require.NoError(t, os.MkdirAll(sequoiaDir, 0o755))

	// Write config.toml with [sequoia] table.
	configContent := `[settings]
theme = "dark"

[sequoia]
skills_path = "/path"
`
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(configContent), 0o644))

	a := codex.NewAdapter(tmp)
	assert.True(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_SequoiaDirMissing(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	// config.toml exists but no sequoia/ dir.
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte("[settings]\n"), 0o644))

	a := codex.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_ConfigMissing(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	sequoiaDir := filepath.Join(codexDir, "sequoia")
	require.NoError(t, os.MkdirAll(sequoiaDir, 0o755))

	a := codex.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_CodexDirMissing(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := codex.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_Status_NotInstalled(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := codex.NewAdapter(tmp)
	s := a.Status()
	assert.False(t, s.Installed)
	assert.Equal(t, "", s.Version)
	assert.NotEmpty(t, s.Path)
}

func TestAdapter_Status_ReadsVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	sequoiaDir := filepath.Join(codexDir, "sequoia")
	skillsDir := filepath.Join(sequoiaDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))

	// Write config.toml with [sequoia] table so IsInstalled returns true.
	configContent := `[sequoia]
skills_path = "/path"
`
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(configContent), 0o644))

	// Write the version file.
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, ".sequoia-version"), []byte("0.2.0"), 0o644))

	a := codex.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true when sequoia dir and [sequoia] table exist")
	assert.Equal(t, "0.2.0", s.Version, "Status().Version should read .sequoia-version content")
}

func TestAdapter_Status_VersionMissingLegacy(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	sequoiaDir := filepath.Join(codexDir, "sequoia")
	skillsDir := filepath.Join(sequoiaDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))

	configContent := `[sequoia]
skills_path = "/path"
`
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(configContent), 0o644))

	a := codex.NewAdapter(tmp)
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
