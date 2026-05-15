package cursor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/cursor"

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

	// Both .sequoia-version AND sequoia-ai.md must exist for IsInstalled to return true.
	versionFile := filepath.Join(rulesDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte("0.2.0\n"), 0o644))

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

	// Create .sequoia-version so IsInstalled returns true (D3 fix: IsInstalled checks version file).
	versionFile := filepath.Join(rulesDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte(""), 0o644))

	sequoiaAI := filepath.Join(rulesDir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(sequoiaAI, []byte("# Sequoia\n"), 0o644))

	a := cursor.NewAdapter(tmp)
	s := a.Status()
	assert.True(t, s.Installed, "expected installed=true when .sequoia-version exists")
	assert.Equal(t, "", s.Version, "Status().Version should be empty when version file has empty content")
}

func TestAdapter_Status_HasPath(t *testing.T) {
	t.Parallel()
	a := newAdapter(t)
	s := a.Status()
	assert.Equal(t, a.SkillsPath(), s.Path, "Status().Path should equal SkillsPath()")
	assert.True(t, filepath.ToSlash(s.Path) != "",
		"Status().Path should not be empty")
}

// TestAdapter_IsInstalled_PreExistingSystemPrompt_ReturnsFalse verifies that
// a pre-existing sequoia-ai.md file does NOT cause a false positive when
// the .sequoia-version marker file is absent. REQ-BUG-003.
func TestAdapter_IsInstalled_PreExistingSystemPrompt_ReturnsFalse(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	rulesDir := filepath.Join(tmp, ".cursor", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	// Create sequoia-ai.md (pre-existing system prompt) but NOT .sequoia-version.
	sequoiaAI := filepath.Join(rulesDir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(sequoiaAI, []byte("# Sequoia rules\n"), 0o644))

	a := cursor.NewAdapter(tmp)
	assert.False(t, a.IsInstalled(),
		"IsInstalled must return false when .sequoia-version is absent, even if sequoia-ai.md exists")
}

// TestAdapter_IsInstalled_ReturnsFalseAfterUninstall verifies that after
// uninstall removes the .sequoia-version file, IsInstalled returns false
// even when sequoia-ai.md still exists (e.g., restored from backup).
// REQ-BUG-003 scenario: After clean uninstall.
func TestAdapter_IsInstalled_ReturnsFalseAfterUninstall(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	rulesDir := filepath.Join(tmp, ".cursor", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	// Simulate a full installation: both files present.
	sequoiaAI := filepath.Join(rulesDir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(sequoiaAI, []byte("# Sequoia\n"), 0o644))

	versionFile := filepath.Join(rulesDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte("0.2.0\n"), 0o644))

	// Verify IsInstalled returns true when both files exist.
	a := cursor.NewAdapter(tmp)
	assert.True(t, a.IsInstalled(), "IsInstalled must return true when both files exist")

	// Simulate uninstall: remove .sequoia-version but sequoia-ai.md may persist
	// (e.g., restored from backup by ReplaceFile).
	require.NoError(t, os.Remove(versionFile))

	assert.False(t, a.IsInstalled(),
		"IsInstalled must return false after uninstall removes .sequoia-version, even if sequoia-ai.md persists")
}

// TestAdapter_IsInstalled_VersionFileOnly_ReturnsTrue verifies that
// when both .sequoia-version and sequoia-ai.md exist, IsInstalled returns true.
// This is the triangulation: the happy path after the fix.
func TestAdapter_IsInstalled_VersionFileOnly_ReturnsTrue(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	rulesDir := filepath.Join(tmp, ".cursor", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	// Both files present = fully installed.
	sequoiaAI := filepath.Join(rulesDir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(sequoiaAI, []byte("# Sequoia rules\n"), 0o644))

	versionFile := filepath.Join(rulesDir, ".sequoia-version")
	require.NoError(t, os.WriteFile(versionFile, []byte("0.2.0\n"), 0o644))

	a := cursor.NewAdapter(tmp)
	assert.True(t, a.IsInstalled(),
		"IsInstalled must return true when both .sequoia-version and sequoia-ai.md exist")
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
