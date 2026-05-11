package gemini_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
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
	require.NoError(t, a.Install())

	// Verify the version file exists with the correct content.
	versionFile := filepath.Join(geminiDir, "sequoia", ".sequoia-version")
	data, err := os.ReadFile(versionFile)
	require.NoError(t, err)
	assert.Equal(t, "0.1.0\n", string(data))
}

func TestAdapter_Uninstall_RemovesSequoiaDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install())

	// Confirm sequoia dir exists before uninstall.
	sequoiaDir := filepath.Join(geminiDir, "sequoia")
	_, err := os.Stat(sequoiaDir)
	require.NoError(t, err, "sequoia dir must exist before uninstall")

	require.NoError(t, a.Uninstall())

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
	require.NoError(t, a.Install())

	s := a.Status()
	assert.True(t, s.Installed, "should be installed after Install()")
	assert.Equal(t, "0.1.0", s.Version, "Status().Version should match the adapter Version constant")
}

func TestAdapter_Reinstall_OverwritesVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install())

	s := a.Status()
	assert.Equal(t, "0.1.0", s.Version, "first install should write version 0.1.0")

	// Reinstall should overwrite.
	require.NoError(t, a.Install())
	s = a.Status()
	assert.Equal(t, "0.1.0", s.Version, "reinstall should still report 0.1.0")
}

func TestAdapter_Install_ValidatesSkillFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install())

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
	require.NoError(t, a.Install())

	geminiMD := filepath.Join(geminiDir, "GEMINI.md")
	data, err := os.ReadFile(geminiMD)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<!-- sequoia:start -->")
	assert.Contains(t, string(data), "Sequoia v0.1.0")
	assert.Contains(t, string(data), "<!-- sequoia:end -->")
}

func TestAdapter_Install_ValidatesCommands(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	geminiDir := filepath.Join(tmp, ".gemini")
	require.NoError(t, os.MkdirAll(geminiDir, 0o755))

	a := gemini.NewAdapter(tmp)
	require.NoError(t, a.Install())

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
	require.NoError(t, a.Install())

	require.NoError(t, a.Uninstall())

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
