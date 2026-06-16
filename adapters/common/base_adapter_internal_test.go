//nolint:gosec // test file: all os.* operations use t.TempDir() test fixtures, not production paths
package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
)

// =========================================================================
// Test helpers
// =========================================================================

// centralAdapterTestHome is the temp dir the userConfigDir override returns
// for these tests. Capture once per test so overrideUserConfigDir and the
// assertions see the same path (t.TempDir() returns a fresh subdir per call).
func centralAdapterTestHome(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// makeBaseAdapterWithSentinel builds a BaseAdapter whose BackupPathBuilder
// carries a SENTINEL backupPathFn. The helper tests assert the central-home
// path is used and the sentinel does NOT leak in. The BackupPathBuilder
// itself still delegates to BackupHomeDir() on the happy path.
func makeBaseAdapterWithSentinel(t *testing.T, home string) *BaseAdapter {
	t.Helper()
	a := &BaseAdapter{}
	a.SetIDName("central-test", "Central Test")
	a.SetPaths(NewPathResolver(
		func(string) (string, error) { return home, nil },
		home,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(NewPromptManager(adapters.StrategyFileReplace, nil, nil))
	// SENTINEL_FN_PATH must not appear in any backup path on the happy path.
	a.SetBackup(NewBackupPathBuilder(
		func(string) string { return "SENTINEL_FN_PATH" },
		"central-test",
	))
	return a
}

// internalFileExists is the package-internal equivalent of common_test.fileExists.
// Kept here so the package-common tests don't depend on the common_test package.
func internalFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// =========================================================================
// Task 2.1 (RED) — BaseAdapter.Prepare writes BackupDir to the central home
// =========================================================================

// TestBaseAdapter_Prepare_BackupDirUsesCentralHome verifies that after
// BaseAdapter.Prepare(), the recorded LastBackupDir starts with the central
// home returned by BackupHomeDir(). REQ-BRP-02.
//
// The sentinel backupPathFn in the BackupPathBuilder is a fallback-only
// closure; on the happy path it must not be consulted. If the assertion
// below fails with SENTINEL_FN_PATH in the result, the test confirms a
// regression in which Prepare fell back to the per-tool path.
//
// Not parallel: the userConfigDir override is package-level and races with
// other tests that also override it.
func TestBaseAdapter_Prepare_BackupDirUsesCentralHome(t *testing.T) {
	home := centralAdapterTestHome(t)
	overrideUserConfigDir(t, func() (string, error) { return home, nil })

	a := makeBaseAdapterWithSentinel(t, home)

	require.NoError(t, a.Prepare(adapters.InstallOpts{}),
		"Prepare should succeed with a writable home and a temp override")

	centralHome, err := BackupHomeDir()
	require.NoError(t, err)

	backupDir := a.LastBackupDir()
	require.NotEmpty(t, backupDir, "LastBackupDir must be set after Prepare")

	assert.True(t, strings.HasPrefix(backupDir, centralHome+string(filepath.Separator)),
		"LastBackupDir %q must start with the central home %q (REQ-BRP-02)",
		backupDir, centralHome)
	assert.NotContains(t, backupDir, "SENTINEL_FN_PATH",
		"LastBackupDir must not invoke the per-tool backupPathFn on the happy path")
	assert.NotContains(t, backupDir, ".sequoia-backup",
		"LastBackupDir must not use the legacy per-tool .sequoia-backup marker")
}

// =========================================================================
// Task 2.8 (RED) — centralBackupDir helper exists
// =========================================================================

// TestBaseAdapter_CentralBackupDir_JoinsHomeAndSubdir verifies the
// private helper centralBackupDir(targetSubdir) returns the per-install
// session root joined with targetSubdir. The session root lives under
// the central home (REQ-BRP-02) and is shared across calls within the
// same install so the per-installer subdirectories ("skills" and
// "commands") land in the same session directory.
//
// Not parallel: the userConfigDir override is package-level.
func TestBaseAdapter_CentralBackupDir_JoinsHomeAndSubdir(t *testing.T) {
	home := centralAdapterTestHome(t)
	overrideUserConfigDir(t, func() (string, error) { return home, nil })

	a := makeBaseAdapterWithSentinel(t, home)

	centralHome, err := BackupHomeDir()
	require.NoError(t, err)

	// Trigger a session dir by calling Prepare (which now uses the helper
	// internally and caches the session dir for Stage).
	require.NoError(t, a.Prepare(adapters.InstallOpts{}))
	sessionDir := a.LastBackupDir()

	// centralBackupDir("") returns the cached session dir exactly.
	gotRoot := a.centralBackupDir("")
	assert.Equal(t, sessionDir, gotRoot,
		"centralBackupDir(\"\") must return the cached session dir")

	// centralBackupDir("skills") returns sessionDir/skills.
	gotSkills := a.centralBackupDir("skills")
	assert.Equal(t, filepath.Join(sessionDir, "skills"), gotSkills,
		"centralBackupDir(\"skills\") must return <sessionDir>/skills")

	// centralBackupDir("commands") returns sessionDir/commands.
	gotCmds := a.centralBackupDir("commands")
	assert.Equal(t, filepath.Join(sessionDir, "commands"), gotCmds,
		"centralBackupDir(\"commands\") must return <sessionDir>/commands")

	// All paths must live under the central home.
	assert.True(t, strings.HasPrefix(sessionDir, centralHome+string(filepath.Separator)),
		"session dir %q must be under the central home %q", sessionDir, centralHome)
	assert.True(t, strings.HasPrefix(gotSkills, centralHome+string(filepath.Separator)),
		"skills backup path %q must be under the central home %q", gotSkills, centralHome)
	assert.True(t, strings.HasPrefix(gotCmds, centralHome+string(filepath.Separator)),
		"commands backup path %q must be under the central home %q", gotCmds, centralHome)

	// The sentinel must not leak into any result.
	for _, p := range []string{sessionDir, gotSkills, gotCmds} {
		assert.NotContains(t, p, "SENTINEL_FN_PATH",
			"centralBackupDir result %q must not consult the per-tool backupPathFn on the happy path", p)
		assert.NotContains(t, p, ".sequoia-backup",
			"centralBackupDir result %q must not use the legacy per-tool .sequoia-backup marker", p)
	}
}

// TestBaseAdapter_CentralBackupDir_CachesSessionDir verifies that multiple
// centralBackupDir calls within a single install share the same session
// root, so the per-installer subdirectories ("skills" and "commands")
// land in the same session directory under the central home (REQ-BRP-02).
//
// Not parallel: the userConfigDir override is package-level.
func TestBaseAdapter_CentralBackupDir_CachesSessionDir(t *testing.T) {
	home := centralAdapterTestHome(t)
	overrideUserConfigDir(t, func() (string, error) { return home, nil })

	a := makeBaseAdapterWithSentinel(t, home)
	require.NoError(t, a.Prepare(adapters.InstallOpts{}))

	firstRoot := a.centralBackupDir("")
	firstSkills := a.centralBackupDir("skills")
	firstCmds := a.centralBackupDir("commands")

	// All three paths share the same parent (the session dir).
	assert.Equal(t, firstRoot, filepath.Dir(firstSkills),
		"skills backup dir must share the same parent as the session root")
	assert.Equal(t, firstRoot, filepath.Dir(firstCmds),
		"commands backup dir must share the same parent as the session root")

	// A second round of calls returns the same paths (idempotent within
	// one install).
	assert.Equal(t, firstRoot, a.centralBackupDir(""),
		"centralBackupDir must cache the session dir for repeated calls")
	assert.Equal(t, firstSkills, a.centralBackupDir("skills"),
		"centralBackupDir must return the same skills dir on repeated calls")
	assert.Equal(t, firstCmds, a.centralBackupDir("commands"),
		"centralBackupDir must return the same commands dir on repeated calls")
}
