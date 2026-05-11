package claude_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/claude"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAdapter(t *testing.T) *claude.Adapter {
	t.Helper()
	return claude.NewAdapter(t.TempDir())
}

func TestAdapter_ID(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "claude-code", a.ID())
}

func TestAdapter_Name(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, "Claude Code", a.Name())
}

func TestAdapter_PromptStrategy(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	assert.Equal(t, adapters.StrategyMarkdownSections, a.PromptStrategy())
}

func TestAdapter_Detect_DirExists(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	// Create the .claude directory inside the temp home.
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	a := claude.NewAdapter(tmp)
	assert.True(t, a.Detect())
}

func TestAdapter_Detect_NoDirNoBinary(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := claude.NewAdapter(tmp)

	_, binErr := exec.LookPath("claude")
	if binErr != nil {
		assert.False(t, a.Detect())
	} else {
		t.Skip("claude binary found in PATH — Detect() will return true regardless of dir")
	}
}

func TestAdapter_IsInstalled_MarkerPresent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")
	content := "# My config\n\n<!-- sequoia:start -->\nsome content\n<!-- sequoia:end -->\n"
	require.NoError(t, os.WriteFile(claudeMD, []byte(content), 0o644))

	a := claude.NewAdapter(tmp)
	assert.True(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_MarkerAbsent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")
	require.NoError(t, os.WriteFile(claudeMD, []byte("# My config\n"), 0o644))

	a := claude.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_IsInstalled_FileMissing(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := claude.NewAdapter(tmp)
	assert.False(t, a.IsInstalled())
}

func TestAdapter_SkillsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SkillsPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".claude/skills/sequoia"),
		"expected path to end with .claude/skills/sequoia, got %s", p)
}

func TestAdapter_CommandsPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.CommandsPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".claude/commands"),
		"expected path to end with .claude/commands, got %s", p)
}

func TestAdapter_SystemPromptPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	p := a.SystemPromptPath()
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), "CLAUDE.md"),
		"expected path to end with CLAUDE.md, got %s", p)
}

func TestAdapter_Status_NotInstalled(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := claude.NewAdapter(tmp)
	s := a.Status()
	assert.False(t, s.Installed)
}

// T-020-02: Status() reads .sequoia-version and populates Version.
func TestAdapter_Status_ReadsVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	skillsDir := filepath.Join(claudeDir, "skills", "sequoia")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))

	// Create CLAUDE.md with sequoia marker so IsInstalled() returns true.
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")
	require.NoError(t, os.WriteFile(claudeMD, []byte("<!-- sequoia:start -->\n"), 0o644))

	// Write the version file.
	versionFile := filepath.Join(skillsDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte("0.2.0\n"), 0o644))

	a := claude.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true when CLAUDE.md has marker")
	assert.Equal(t, "0.2.0", s.Version, "Status().Version should read .sequoia-version content")
}

// T-020-02: Status() returns empty Version when .sequoia-version is missing (legacy install).
func TestAdapter_Status_VersionMissingLegacy(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	// Installed (CLAUDE.md has marker) but no .sequoia-version file.
	claudeMD := filepath.Join(claudeDir, "CLAUDE.md")
	require.NoError(t, os.WriteFile(claudeMD, []byte("<!-- sequoia:start -->\n"), 0o644))

	a := claude.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true even without version file")
	assert.Equal(t, "", s.Version, "Status().Version should be empty for legacy installs without .sequoia-version")
}

// T-020-02: Status().Path is SkillsPath() — the Sequoia root installation directory.
func TestAdapter_Status_HasPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	s := a.Status()
	assert.Equal(t, a.SkillsPath(), s.Path, "Status().Path should equal SkillsPath()")
	assert.True(t, strings.HasSuffix(filepath.ToSlash(s.Path), ".claude/skills/sequoia"),
		"Status().Path should end with skills/sequoia, got %s", s.Path)
}

// T-020-03: Install writes .sequoia-version; round-trip Install → Status → Version.
func TestAdapter_Install_WritesVersionFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	// Create the .claude directory so Install can write to it.
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	a := claude.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	// Verify the version file exists with the correct content.
	versionFile := filepath.Join(claudeDir, "skills", "sequoia", ".sequoia-version")
	data, err := os.ReadFile(versionFile)
	require.NoError(t, err)
	assert.Equal(t, "0.1.0", strings.TrimSpace(string(data)),
		"version file should contain the adapter Version constant")
}

// T-020-03: Uninstall removes .sequoia-version.
func TestAdapter_Uninstall_RemovesVersionFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	a := claude.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	// Confirm version file exists before uninstall.
	versionFile := filepath.Join(claudeDir, "skills", "sequoia", ".sequoia-version")
	_, err := os.Stat(versionFile)
	require.NoError(t, err, "version file must exist before uninstall")

	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	// After uninstall, version file should not exist.
	_, err = os.Stat(versionFile)
	assert.True(t, os.IsNotExist(err), "version file should be removed by Uninstall")
}

// T-020-03: Round-trip: Install → Status().Version matches Version constant.
func TestAdapter_VersionRoundTrip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	a := claude.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	s := a.Status()
	assert.True(t, s.Installed, "should be installed after Install()")
	assert.NotEmpty(t, s.Version, "Version should not be empty after install")
	// The version constant is 0.1.0 (from install.go).
	assert.Equal(t, "0.1.0", s.Version,
		"Status().Version should match the adapter Version constant")
}

// T-020-03: Reinstall overwrites version file.
func TestAdapter_Reinstall_OverwritesVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	a := claude.NewAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}))

	s := a.Status()
	assert.Equal(t, "0.1.0", s.Version, "first install should write version 0.1.0")

	// Reinstall should overwrite (still 0.1.0 since the const hasn't changed).
	require.NoError(t, a.Install(adapters.InstallOpts{}))
	s = a.Status()
	assert.Equal(t, "0.1.0", s.Version, "reinstall should still report 0.1.0")
}

// T-020-06: EvalSymlinks error fallback — base() must not propagate error.
// When homeDir is a nonexistent path, filepath.EvalSymlinks will error.
// The adapter must fall back to the unresolved path.
func TestAdapter_EvalSymlinks_Fallback(t *testing.T) {
	t.Parallel()
	// Use a nonexistent path to trigger EvalSymlinks error.
	nonexistent := filepath.Join(t.TempDir(), "does-not-exist")
	a := claude.NewAdapter(nonexistent)

	// SkillsPath must return a path (not panic, not empty) even when
	// EvalSymlinks fails — fallback to unresolved path.
	p := a.SkillsPath()
	assert.NotEmpty(t, p, "SkillsPath should not be empty when EvalSymlinks fails")
	assert.Contains(t, filepath.ToSlash(p), ".claude/skills/sequoia",
		"SkillsPath should contain expected suffix even with fallback path")
}

// T-020-06: Status of a never-installed adapter returns Installed=false, Version="".
func TestAdapter_Status_NeverInstalled(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	s := a.Status()
	assert.False(t, s.Installed, "never-installed adapter should report Installed=false")
	assert.Equal(t, "", s.Version, "never-installed adapter should have empty Version")
	assert.NotEmpty(t, s.Path, "even never-installed adapter should report a Path")
}

func TestAdapter_VersionFilePath_Suffix(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	// versionFilePath returns filepath.Join(SkillsPath(), ".sequoia-version")
	// SkillsPath ends with skills/sequoia, so the version file path should
	// end with skills/sequoia/.sequoia-version
	p := filepath.Join(a.SkillsPath(), ".sequoia-version")
	assert.True(t, strings.HasSuffix(filepath.ToSlash(p), ".claude/skills/sequoia/.sequoia-version"),
		"expected path to end with .claude/skills/sequoia/.sequoia-version, got %s", p)
}

func TestAdapter_Base_SymlinkResolved(t *testing.T) {
	// Symlinks may require admin privileges on Windows.
	// If os.Symlink fails, skip the test rather than failing.
	realHome := t.TempDir()
	linkHome := filepath.Join(t.TempDir(), "link-home")

	// Create the .claude directory inside the real home so Detect() / SkillsPath work.
	claudeDir := filepath.Join(realHome, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))

	if err := os.Symlink(realHome, linkHome); err != nil {
		t.Skipf("os.Symlink not available (may require admin on Windows): %v", err)
	}

	a := claude.NewAdapter(linkHome)
	p := a.SkillsPath()

	// On Windows, filepath.EvalSymlinks may resolve to the long (full) path name
	// while realHome may contain the 8.3 short name (e.g., RUNNER~1 vs runneradmin).
	// Instead of checking exact path equality, verify resolution occurred:
	// the resolved path must differ from the symlink path and must be absolute.
	// Normalize separators to slash for reliable substring comparison across platforms.
	assert.NotContains(t, filepath.ToSlash(p), filepath.ToSlash(linkHome),
		"SkillsPath should NOT use the unresolved symlink path, got %s", p)
	assert.True(t, filepath.IsAbs(p),
		"SkillsPath should be an absolute (resolved) path, got %s", p)
}
