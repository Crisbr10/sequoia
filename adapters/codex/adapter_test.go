package codex_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/codex"

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

// setupCodexInstalled creates a codex adapter with a full Sequoia
// installation for uninstall testing. Returns the adapter and the
// paths to the skills dir and version file.
func setupCodexInstalled(t *testing.T) (*codex.Adapter, string, string) {
	t.Helper()

	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	// Write config.toml with [sequoia] so IsInstalled returns true.
	configContent := `[sequoia]
skills_path = "/path"
`
	require.NoError(t, os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte(configContent), 0o644))

	// Create sequoia dir and subdirectories for skills and commands.
	sequoiaDir := filepath.Join(codexDir, "sequoia")
	skillsDir := filepath.Join(sequoiaDir, "skills")
	cmdsDir := filepath.Join(sequoiaDir, "commands")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))
	require.NoError(t, os.MkdirAll(cmdsDir, 0o755))

	// Create SKILL.md, version file, and command files.
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("skill"), 0o644))
	versionFile := filepath.Join(skillsDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte("0.1.0\n"), 0o644))
	for _, cmd := range []string{"sequoia-init.md", "sequoia-audit.md"} {
		require.NoError(t, os.WriteFile(filepath.Join(cmdsDir, cmd), []byte("cmd"), 0o644))
	}

	a := codex.NewAdapter(tmp)
	return a, skillsDir, versionFile
}

// TestAdapter_Uninstall_CollectsErrors verifies that the Codex adapter's
// Uninstall collects errors from failed file removals instead of silently
// discarding them.
func TestAdapter_Uninstall_CollectsErrors(t *testing.T) {
	t.Parallel()

	a, _, versionFile := setupCodexInstalled(t)

	// Make the version file path a non-empty directory so os.Remove fails.
	require.NoError(t, os.Remove(versionFile))
	require.NoError(t, os.MkdirAll(versionFile, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(versionFile, "blocker.txt"), []byte("x"), 0o644))

	err := a.Uninstall(adapters.InstallOpts{})
	require.Error(t, err, "Uninstall should return an error when file removal fails")
	t.Logf("Uninstall error: %v", err)
	assert.Contains(t, err.Error(), "version file", "error should mention the version file")
}

// TestAdapter_Uninstall_ReturnsSentinelError verifies that Codex's
// Uninstall wraps adapters.ErrUninstallFailed.
func TestAdapter_Uninstall_ReturnsSentinelError(t *testing.T) {
	t.Parallel()

	a, _, versionFile := setupCodexInstalled(t)

	// Block version file removal.
	require.NoError(t, os.Remove(versionFile))
	require.NoError(t, os.MkdirAll(versionFile, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(versionFile, "blocker.txt"), []byte("x"), 0o644))

	err := a.Uninstall(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrUninstallFailed),
		"error should wrap ErrUninstallFailed, got: %v", err)
}

// TestAdapter_Install_ReturnsSentinelError verifies that Codex's
// Install wraps adapters.ErrInstallFailed on failure.
func TestAdapter_Install_ReturnsSentinelError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	// Make skills dir a file instead of a directory so MkdirAll fails.
	skillsPath := filepath.Join(codexDir, "sequoia", "skills")
	require.NoError(t, os.MkdirAll(filepath.Dir(skillsPath), 0o755))
	require.NoError(t, os.WriteFile(skillsPath, []byte("not a dir"), 0o644))

	a := codex.NewAdapter(tmp)
	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err, "Install should fail when skills dir is a file")
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed),
		"error should wrap ErrInstallFailed, got: %v", err)
}
