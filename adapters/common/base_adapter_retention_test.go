//nolint:gosec // test file: uses t.TempDir() and overrideUserConfigDir hook, no real user paths
package common

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
)

// =========================================================================
// Test helpers
// =========================================================================

// retentionTestHome is the temp dir the userConfigDir override returns
// for the applyRetention tests. Capture once per test (via a local) so
// the override and the assertions see the same path — t.TempDir()
// returns a fresh subdir on every call.
func retentionTestHome(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// makeRetentionAdapter builds a BaseAdapter wired to a temp
// userConfigDir so the test never touches the real central home. The
// adapter has the minimum surface Install needs (prompt manager with
// a no-op write, install templates, detector).
func makeRetentionAdapter(t *testing.T, home string) *BaseAdapter {
	t.Helper()
	a := &BaseAdapter{}
	a.SetIDName("retention-test", "Retention Test")
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
	a.SetPromptManager(NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return nil },
		nil,
	))
	a.SetBackup(NewBackupPathBuilder(
		func(string) string { return filepath.Join(home, "backup") },
		"retention-test",
	))
	a.SetDetector(NewDetector(
		a.Base,
		func(base string) bool {
			_, err := os.Stat(filepath.Join(base, "version"))
			return err == nil
		},
		func() bool { return true },
	))
	return a
}

// seedRetentionSessions pre-seeds `count` session directories under
// the central home for the "retention-test" adapter, each with a
// strictly increasing ISO-8601 timestamp. The return value is the
// adapter's session directory (the parent of every session dir) so
// callers can read the dir count without recomputing the path.
func seedRetentionSessions(t *testing.T, count int) string {
	t.Helper()
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	centralHome, err := BackupHomeDir()
	require.NoError(t, err)
	adapterDir := filepath.Join(centralHome, "retention-test")
	require.NoError(t, os.MkdirAll(adapterDir, 0o700))

	for i := 0; i < count; i++ {
		ts := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC).
			Add(time.Duration(i) * time.Hour)
		iso := ts.Format("2006-01-02T15-04-05.000Z")
		suffix := strconv.FormatInt(ts.UnixNano(), 36)
		dirName := iso + "-" + suffix
		require.NoError(t, os.MkdirAll(filepath.Join(adapterDir, dirName), 0o700))
	}
	return adapterDir
}

// =========================================================================
// Task 3.7 (RED) — applyRetention keeps at most DefaultMaxBackupsPerAdapter
// =========================================================================

// TestApplyRetention_PrunesExcessSessions verifies REQ-BRP-04 Scenario
// 1: pre-seeding 7 session dirs in the central home, then calling
// applyRetention, leaves exactly DefaultMaxBackupsPerAdapter (= 5)
// session dirs. The 2 oldest (by ISO-8601 lex order) are removed.
//
// Not parallel: the userConfigDir override is package-level.
func TestApplyRetention_PrunesExcessSessions(t *testing.T) {
	home := retentionTestHome(t)
	adapterDir := seedRetentionSessions(t, 7)
	_ = home // adapterDir is the per-adapter subdir; the override is set inside the seeder

	a := makeRetentionAdapter(t, home)
	a.applyRetention()

	// After applyRetention, exactly DefaultMaxBackupsPerAdapter session
	// dirs must remain (the 2 oldest are removed).
	entries, err := os.ReadDir(adapterDir)
	require.NoError(t, err)
	assert.Equal(t, DefaultMaxBackupsPerAdapter, len(entries),
		"applyRetention must leave exactly DefaultMaxBackupsPerAdapter session dirs (got %d, want %d)",
		len(entries), DefaultMaxBackupsPerAdapter)
}

// TestApplyRetention_NoOpAtOrBelowMax verifies REQ-BRP-04 Scenario 2:
// when the count of session dirs is at or below DefaultMaxBackupsPerAdapter,
// applyRetention is a no-op.
//
// Not parallel: the userConfigDir override is package-level.
func TestApplyRetention_NoOpAtOrBelowMax(t *testing.T) {
	home := retentionTestHome(t)
	adapterDir := seedRetentionSessions(t, 3)

	a := makeRetentionAdapter(t, home)
	a.applyRetention()

	entries, err := os.ReadDir(adapterDir)
	require.NoError(t, err)
	assert.Equal(t, 3, len(entries),
		"applyRetention must be a no-op below the cap (got %d)", len(entries))
}

// TestApplyRetention_RunsAfterApply_NotAfterPrepare verifies the
// retention hook is wired to BaseAdapter.Apply, NOT to Prepare or any
// earlier phase. Concretely, calling Prepare alone (which clears
// strategyState) does NOT call applyRetention; only the full Apply
// path runs it. The test seeds 6 sessions, calls Prepare, and asserts
// the session count is unchanged.
func TestApplyRetention_NotInPrepare(t *testing.T) {
	home := retentionTestHome(t)
	adapterDir := seedRetentionSessions(t, 6)

	a := makeRetentionAdapter(t, home)
	require.NoError(t, a.Prepare(adapters.InstallOpts{}),
		"Prepare must succeed")

	entries, err := os.ReadDir(adapterDir)
	require.NoError(t, err)
	assert.Equal(t, 6, len(entries),
		"Prepare must NOT trigger applyRetention (got %d, want 6)", len(entries))
}
