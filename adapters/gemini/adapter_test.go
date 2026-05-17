package gemini_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/adapters/gemini"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAdapter(t *testing.T) *gemini.Adapter {
	t.Helper()
	return gemini.NewAdapter(t.TempDir())
}

func TestAdapter_ID(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "gemini-cli", a.ID())
}

func TestAdapter_Name(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "Gemini CLI", a.Name())
}

func TestAdapter_PromptStrategy(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, adapters.StrategyConfigMerge, a.PromptStrategy())
}

func TestAdapter_Detect_DirExists(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	assert.True(t, a.Detect())
}

func TestAdapter_Detect_NoDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := gemini.NewAdapter(tmp)
	assert.False(t, a.Detect())
}

func TestAdapter_IsInstalled_MarkerPresent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	content := "# My config\n\n<!-- sequoia:start -->\nsome content\n<!-- sequoia:end -->\n"
	require.NoError(t, os.WriteFile(geminiMD, []byte(content), 0o644))

	a := gemini.NewAdapter(tmp)
	assert.True(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_MarkerAbsent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	require.NoError(t, os.WriteFile(geminiMD, []byte("# My config\n"), 0o644))

	a := gemini.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_FileMissing(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := gemini.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_Status_NotInstalled(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := gemini.NewAdapter(tmp)
	s := a.Status()
	assert.False(t, s.Installed)
	assert.Equal(t, "", s.Version)
	assert.NotEmpty(t, s.Path)
}

func TestAdapter_Status_ReadsVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	sequoiaDir := filepath.Join(geminiDir, "sequoia")
	require.NoError(t, os.MkdirAll(sequoiaDir, 0o755))

	// Create GEMINI.md with sequoia marker so IsInstalled() returns true.
	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	require.NoError(t, os.WriteFile(geminiMD, []byte("<!-- sequoia:start -->\n"), 0o644))

	// Write the version file.
	versionFile := filepath.Join(sequoiaDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte("0.2.0\n"), 0o644))

	a := gemini.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true when GEMINI.md has marker")
	assert.Equal(t, "0.2.0", s.Version, "Status().Version should read .sequoia-version content")
}

func TestAdapter_Status_VersionMissingLegacy(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	// Installed (GEMINI.md has marker) but no .sequoia-version file.
	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	require.NoError(t, os.WriteFile(geminiMD, []byte("<!-- sequoia:start -->\n"), 0o644))

	a := gemini.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true even without version file")
	assert.Equal(t, "", s.Version, "Status().Version should be empty for legacy installs without .sequoia-version")
}

func TestAdapter_Status_HasPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	s := a.Status()
	assert.Equal(t, a.SkillsPath(), s.Path, "Status().Path should equal SkillsPath()")
	assert.NotEmpty(t, s.Path)
}

func TestAdapter_Install_WritesVersionFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	// Verify the version file exists with the correct content.
	versionFile := filepath.Join(geminiDir, "sequoia", ".sequoia-version")
	data, err := os.ReadFile(versionFile)
	require.NoError(t, err)
	assert.Equal(t, common.Version+"\n", string(data))
}

func TestAdapter_Uninstall_RemovesSequoiaDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	// Confirm sequoia dir exists before uninstall.
	sequoiaDir := filepath.Join(geminiDir, "sequoia")
	_, err := os.Stat(sequoiaDir)
	require.NoError(t, err, "sequoia dir must exist before uninstall")

	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	// After uninstall, sequoia dir should not exist.
	_, err = os.Stat(sequoiaDir)
	assert.True(t, os.IsNotExist(err), "sequoia dir should be removed by Uninstall")
}

func TestAdapter_VersionRoundTrip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	s := a.Status()
	assert.True(t, s.Installed, "should be installed after Install()")
	assert.Equal(t, common.Version, s.Version, "Status().Version should match the adapter Version constant")
}

func TestAdapter_Reinstall_OverwritesVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	s := a.Status()
	assert.Equal(t, common.Version, s.Version, "first install should write the correct version")

	// Reinstall should overwrite with the same version.
	require.NoError(t, a.Install(adapters.InstallOpts{}))
	s = a.Status()
	assert.Equal(t, common.Version, s.Version, "reinstall should still report the correct version")
}

func TestAdapter_Install_ValidatesSkillFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	skillFile := filepath.Join(geminiDir, "sequoia", "skills", "SKILL.md")
	data, err := os.ReadFile(skillFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "name: sequoia")
}

func TestAdapter_Install_ValidatesGeminiMD(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	data, err := os.ReadFile(geminiMD)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<!-- sequoia:start -->")
	assert.Contains(t, string(data), "Sequoia v"+common.Version)
	assert.Contains(t, string(data), "<!-- sequoia:end -->")
}

func TestAdapter_Install_ValidatesCommands(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	commands := []string{"sequoia-init.md", "sequoia-audit.md", "sequoia-review.md", "sequoia-fix.md", "sequoia-diff.md"}
	for _, cmd := range commands {
		cmdPath := filepath.Join(geminiDir, "sequoia", "commands", cmd)
		_, err := os.Stat(cmdPath)
		assert.NoError(t, err, "command file %s should exist after install", cmd)
	}
}

func TestAdapter_Uninstall_RemovesMarkers(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	data, err := os.ReadFile(geminiMD)
	// GEMINI.md may or may not exist after uninstall depending on original state.
	// If it was created by install and had only sequoia section, it might be empty or deleted.
	if err == nil {
		assert.NotContains(t, string(data), "<!-- sequoia:start -->")
		assert.NotContains(t, string(data), "<!-- sequoia:end -->")
	}
}

func TestAdapter_EvalSymlinks_Fallback(t *testing.T) {
	t.Parallel()
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist")
	a := gemini.NewAdapter(nonexistent)

	p := a.SkillsPath()
	assert.NotEmpty(t, p, "SkillsPath should not be empty when EvalSymlinks fails")
	assert.Contains(t, filepath.ToSlash(p), ".gemini/sequoia/skills",
		"SkillsPath should contain expected suffix even with fallback path")
}

// setupGeminiInstalled creates a Gemini adapter with a full Sequoia
// installation for uninstall testing. Returns the adapter and the
// sequoia directory path.
func setupGeminiInstalled(t *testing.T) (*gemini.Adapter, string) {
	t.Helper()

	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	// Create GEMINI.md with full sequoia marker pair so RemoveMarkdownSection
	// finds and attempts to remove the section (triggering a write).
	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	require.NoError(t, os.WriteFile(geminiMD, []byte("<!-- sequoia:start -->\ncontent\n<!-- sequoia:end -->\n"), 0o644))

	// Create sequoia dir with subdirectories.
	sequoiaDir := filepath.Join(geminiDir, "sequoia")
	skillsDir := filepath.Join(sequoiaDir, "skills")
	cmdsDir := filepath.Join(sequoiaDir, "commands")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))
	require.NoError(t, os.MkdirAll(cmdsDir, 0o755))

	// Create files.
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("skill"), 0o644))
	for _, cmd := range []string{"sequoia-init.md", "sequoia-audit.md"} {
		require.NoError(t, os.WriteFile(filepath.Join(cmdsDir, cmd), []byte("cmd"), 0o644))
	}

	a := gemini.NewAdapter(tmp)
	return a, sequoiaDir
}

// TestAdapter_Uninstall_ReturnsSentinelError verifies that Gemini's
// Uninstall wraps adapters.ErrUninstallFailed when removal fails.
func TestAdapter_Uninstall_ReturnsSentinelError(t *testing.T) {
	t.Parallel()

	a, sequoiaDir := setupGeminiInstalled(t)

	// Make GEMINI.md read-only so that RemoveMarkdownSection's
	// os.WriteFile fails with a permission error.
	geminiDir := filepath.Dir(sequoiaDir)
	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	require.NoError(t, os.Chmod(geminiMD, 0o444))

	err := a.Uninstall(adapters.InstallOpts{})

	// Restore permissions for cleanup.
	_ = os.Chmod(geminiMD, 0o644)

	require.Error(t, err, "Uninstall should return an error when system prompt restore fails")
	t.Logf("Uninstall error: %v", err)
	assert.True(t, errors.Is(err, adapters.ErrUninstallFailed),
		"error should wrap ErrUninstallFailed, got: %v", err)
}

// TestAdapter_Install_ReturnsSentinelError verifies that Gemini's
// Install wraps adapters.ErrInstallFailed on failure.
func TestAdapter_Install_ReturnsSentinelError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	// Make skills dir a file instead of directory so MkdirAll fails.
	skillsPath := filepath.Join(geminiDir, "sequoia", "skills")
	require.NoError(t, os.MkdirAll(filepath.Dir(skillsPath), 0o755))
	require.NoError(t, os.WriteFile(skillsPath, []byte("not a dir"), 0o644))

	a := gemini.NewAdapter(tmp)
	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err, "Install should fail when skills dir is a file")
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed),
		"error should wrap ErrInstallFailed, got: %v", err)
}

// TestAdapter_Uninstall_ProductionPath verifies that Gemini's Uninstall,
// when homeDir="" (production), resolves the correct home directory via
// a.base() instead of operating on a relative path.
// NOT parallel-safe: uses t.Setenv which is incompatible with t.Parallel().
func TestAdapter_Uninstall_ProductionPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("USERPROFILE", tmpHome) // Windows
	t.Setenv("HOME", tmpHome)        // Unix

	// Create .gemini dir with sequoia content inside the controlled home.
	geminiDir := filepath.Join(tmpHome, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	sequoiaDir := filepath.Join(geminiDir, "sequoia")
	require.NoError(t, os.MkdirAll(sequoiaDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(sequoiaDir, "test-file.txt"),
		[]byte("content"), 0o644))

	// Write GEMINI.md with sequoia section so RemoveMarkdownSection works.
	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	require.NoError(t, os.WriteFile(geminiMD,
		[]byte("<!-- sequoia:start -->\ncontent\n<!-- sequoia:end -->\n"), 0o644))

	// Create adapter with empty homeDir (production path).
	a := gemini.NewAdapter("")
	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	// After uninstall, sequoia dir under the controlled home must be gone.
	_, err := os.Stat(sequoiaDir)
	assert.True(t, os.IsNotExist(err),
		"sequoia dir must be removed from production home after uninstall, but it still exists")
}
