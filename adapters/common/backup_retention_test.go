//nolint:gosec // test file: uses t.TempDir() and controlled fixtures
package common

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackupHomeDir_ReturnsAndCreatesPath verifies REQ-BRP-01 Scenario 1
// (returns and creates the joined path with mode 0o700). The test overrides
// the package-level userConfigDir to point at a t.TempDir() so the real user
// config dir is never touched.
//
// Not parallel: this test mutates the package-level userConfigDir hook.
// Running in parallel would race with sibling tests that also override the
// hook and observe the wrong UserConfigDir value.
func TestBackupHomeDir_ReturnsAndCreatesPath(t *testing.T) {
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	home, err := BackupHomeDir()
	require.NoError(t, err, "BackupHomeDir() should succeed on a writable parent")
	require.NotEmpty(t, home, "BackupHomeDir() must return a non-empty path")

	expected := filepath.Join(tmp, "sequoia", "backups")
	assert.Equal(t, expected, home,
		"BackupHomeDir() must return <UserConfigDir>/sequoia/backups (got %q, want %q)",
		home, expected)

	info, err := os.Stat(home)
	require.NoError(t, err, "BackupHomeDir() must create the root directory")
	assert.True(t, info.IsDir(), "BackupHomeDir() result must be a directory")

	// Mode 0o700 is enforced on POSIX; Windows does not honor permission bits
	// the same way, so we only assert on non-Windows platforms.
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm(),
			"BackupHomeDir() must create the dir with mode 0o700")
	}
}

// TestBackupHomeDir_IsIdempotent verifies REQ-BRP-01 Scenario 2 (idempotency).
// Calling BackupHomeDir() twice in a row must succeed and return the same
// absolute path without altering the existing mode.
//
// Not parallel: this test mutates the package-level userConfigDir hook.
func TestBackupHomeDir_IsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	first, err := BackupHomeDir()
	require.NoError(t, err)
	second, err := BackupHomeDir()
	require.NoError(t, err, "BackupHomeDir() must be idempotent")

	assert.Equal(t, first, second,
		"BackupHomeDir() must return the same path on repeated calls")

	info, err := os.Stat(first)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm(),
			"BackupHomeDir() must not change the mode of an existing dir")
	}
}

// TestBackupHomeDir_WrapsErrorsWithContext verifies REQ-BRP-06 Scenario 1
// (error wrapping with the failing path and the "sequoia/backups" suffix).
// We point userConfigDir at a regular file so MkdirAll on a child fails.
//
// Not parallel: this test mutates the package-level userConfigDir hook.
func TestBackupHomeDir_WrapsErrorsWithContext(t *testing.T) {

	tmp := t.TempDir()
	// Create a regular file that will play the role of "UserConfigDir" — the
	// join <file>/sequoia/backups is not creatable because the parent is a file.
	blockingFile := filepath.Join(tmp, "blocking")
	require.NoError(t, os.WriteFile(blockingFile, []byte("not a dir"), 0o600))

	OverrideUserConfigDir(t, func() (string, error) {
		return blockingFile, nil
	})

	home, err := BackupHomeDir()
	require.Error(t, err, "BackupHomeDir() must fail when the parent is not a directory")
	assert.Empty(t, home, "BackupHomeDir() must return an empty path on error")
	assert.Contains(t, err.Error(), "sequoia/backups",
		"wrapped error must include the 'sequoia/backups' suffix for diagnostics")
	assert.Contains(t, err.Error(), blockingFile,
		"wrapped error must include the failing path for diagnostics")
}

// TestDefaultMaxBackupsPerAdapter_IsFive verifies REQ-BRP-04 (retention cap).
// The exported constant DefaultMaxBackupsPerAdapter must equal 5 to match the
// per-adapter retention policy documented in the spec.
func TestDefaultMaxBackupsPerAdapter_IsFive(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 5, DefaultMaxBackupsPerAdapter,
		"DefaultMaxBackupsPerAdapter must equal 5 (5-backup retention per adapter)")
}

// sessionDirName produces a session directory name in the format used by
// BackupPathBuilder.Build after PR 1: <ISO-8601-UTC>-<base36-UnixNanos>.
// The base time is offset by `offsetHours` so each fixture session is
// strictly older than the previous one (lexicographic == chronological).
func sessionDirName(t *testing.T, offsetHours int) string {
	t.Helper()
	base := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC).Add(time.Duration(offsetHours) * time.Hour)
	iso := base.Format("2006-01-02T15-04-05.000Z")
	suffix := strconv.FormatInt(base.UnixNano(), 36)
	return iso + "-" + suffix
}

// TestPruneBackups_KeepsExactlyMaxFromSevenSessions verifies REQ-BRP-04
// Scenario 1: 7 session directories with a max of 5 means exactly 2
// (the oldest) are removed and 5 (the most recent) remain.
//
// The ISO-8601 prefix sorts lexicographically the same as chronologically,
// so a descending lex sort yields the 5 most-recent sessions first.
func TestPruneBackups_KeepsExactlyMaxFromSevenSessions(t *testing.T) {
	home, adapterDir := setupAdapterFixture(t, "opencode", 7)

	removed, err := PruneBackups("opencode", 5)
	require.NoError(t, err)
	assert.Equal(t, 2, removed, "PruneBackups must remove exactly 2 of 7 sessions to reach max=5")

	// The 5 most-recent session dirs (offsets 2..6) must remain; the
	// 2 oldest (offsets 0, 1) must be gone.
	entries := readSortedDir(t, adapterDir)
	assert.Equal(t, 5, len(entries),
		"adapter dir must hold exactly 5 session dirs after pruning")

	// Reconstruct expected remaining offsets in lexicographic (== chronological)
	// descending order, then descending-sort and take top 5 to confirm.
	var all []string
	for i := 0; i < 7; i++ {
		all = append(all, sessionDirName(t, i))
	}
	sort.Sort(sort.Reverse(sort.StringSlice(all)))
	expectedKept := all[:5]
	assert.Equal(t, expectedKept, entries,
		"the 5 most-recent session dirs (descending) must remain after pruning")

	// Sanity: the BackupHomeDir root itself must still exist.
	_, err = os.Stat(home)
	require.NoError(t, err, "BackupHomeDir root must not be removed")
}

// TestPruneBackups_NoOpBelowMax verifies REQ-BRP-04 Scenario 2 and
// REQ-BRP-07: when there are fewer than max session dirs, PruneBackups
// must return (0, nil) and leave every existing directory untouched.
func TestPruneBackups_NoOpBelowMax(t *testing.T) {
	_, adapterDir := setupAdapterFixture(t, "cursor", 3)

	removed, err := PruneBackups("cursor", 5)
	require.NoError(t, err, "PruneBackups must succeed when count is below max")
	assert.Equal(t, 0, removed,
		"PruneBackups must report 0 removals when count is below max")

	entries := readSortedDir(t, adapterDir)
	assert.Equal(t, 3, len(entries),
		"all 3 session dirs must remain when count is below max")
}

// TestPruneBackups_MissingAdapterDir verifies REQ-BRP-07: PruneBackups
// for an adapter with no existing dir must return (0, nil) and create
// no new directories.
func TestPruneBackups_MissingAdapterDir(t *testing.T) {
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	// Create the root but not the adapter dir.
	_, err := BackupHomeDir()
	require.NoError(t, err)

	removed, err := PruneBackups("nonexistent-adapter", 5)
	require.NoError(t, err,
		"PruneBackups must return nil error for a missing adapter dir")
	assert.Equal(t, 0, removed,
		"PruneBackups must return 0 removed for a missing adapter dir")

	// Side-effect check: no adapter dir was created.
	_, err = os.Stat(filepath.Join(tmp, "sequoia", "backups", "nonexistent-adapter"))
	assert.True(t, os.IsNotExist(err),
		"PruneBackups must not create the adapter dir for a missing adapter")
}

// TestPruneBackups_IgnoresCorruptNames verifies REQ-BRP-07: a non-timestamp
// directory mixed in with valid session dirs must be left untouched, the
// pruning must still keep exactly `max` valid dirs, and no panic is raised.
func TestPruneBackups_IgnoresCorruptNames(t *testing.T) {
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })
	home, err := BackupHomeDir()
	require.NoError(t, err)
	adapterDir := filepath.Join(home, "x")
	require.NoError(t, os.MkdirAll(adapterDir, 0o700))

	// 7 valid sessions + 1 corrupt "garbage" entry.
	for i := 0; i < 7; i++ {
		require.NoError(t, os.MkdirAll(filepath.Join(adapterDir, sessionDirName(t, i)), 0o700))
	}
	corruptDir := filepath.Join(adapterDir, "garbage")
	require.NoError(t, os.MkdirAll(corruptDir, 0o700))

	removed, err := PruneBackups("x", 5)
	require.NoError(t, err, "PruneBackups must not error on a corrupt directory name")
	assert.Equal(t, 2, removed,
		"PruneBackups must remove exactly 2 of the 7 valid session dirs (corrupt ignored)")

	entries := readSortedDir(t, adapterDir)
	assert.Equal(t, 6, len(entries),
		"5 valid + 1 corrupt = 6 entries must remain after pruning")
	assert.Contains(t, entries, "garbage",
		"corrupt directory must be left untouched")

	// Verify the 5 most recent valid sessions are exactly the kept ones.
	var allValid []string
	for i := 0; i < 7; i++ {
		allValid = append(allValid, sessionDirName(t, i))
	}
	sort.Sort(sort.Reverse(sort.StringSlice(allValid)))
	expectedKept := allValid[:5]
	var actualValid []string
	for _, e := range entries {
		if e != "garbage" {
			actualValid = append(actualValid, e)
		}
	}
	assert.Equal(t, expectedKept, actualValid,
		"the 5 most-recent valid sessions must be the ones kept")
}

// TestPruneBackups_ContinuesOnError verifies REQ-BRP-06 Scenario 2:
// when one of the to-be-removed dirs cannot be removed, PruneBackups
// must continue and still remove the other eligible dirs; the returned
// error references the failed dir, and `removed` counts only successes.
//
// The fixture places 7 valid session dirs and marks the OLDEST (offset 0)
// read-only. With max=5, the two oldest are to be removed: offset 0 (fails)
// and offset 1 (succeeds). The remaining 5 (offsets 2..6) are kept.
//
// On Windows the chmod read-only trick does not block os.RemoveAll, so
// this test only runs on POSIX platforms.
func TestPruneBackups_ContinuesOnError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod read-only does not block os.RemoveAll on Windows")
	}

	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })
	home, err := BackupHomeDir()
	require.NoError(t, err)
	adapterDir := filepath.Join(home, "x")
	require.NoError(t, os.MkdirAll(adapterDir, 0o700))

	for i := 0; i < 7; i++ {
		name := sessionDirName(t, i)
		dir := filepath.Join(adapterDir, name)
		require.NoError(t, os.MkdirAll(dir, 0o700))
		// Mark the oldest (offset 0) read-only so its removal fails.
		if i == 0 {
			require.NoError(t, os.Chmod(dir, 0o500),
				"chmod read-only on the oldest session dir to force removal failure")
		}
	}

	removed, err := PruneBackups("x", 5)
	require.Error(t, err,
		"PruneBackups must return an error when at least one removal fails")
	assert.Equal(t, 1, removed,
		"PruneBackups must still remove the 1 non-read-only oldest dir (offset 1)")
	assert.Contains(t, err.Error(), sessionDirName(t, 0),
		"returned error must reference the directory that failed to remove")
}

// TestPruneBackups_AtExactlyMaxIsNoOp verifies that when the count of
// session dirs equals max, no removals occur.
func TestPruneBackups_AtExactlyMaxIsNoOp(t *testing.T) {
	_, adapterDir := setupAdapterFixture(t, "gemini-cli", 5)

	removed, err := PruneBackups("gemini-cli", 5)
	require.NoError(t, err)
	assert.Equal(t, 0, removed,
		"PruneBackups must report 0 removals when count is exactly max")

	entries := readSortedDir(t, adapterDir)
	assert.Equal(t, 5, len(entries),
		"all 5 session dirs must remain when count is exactly max")
}

// =========================================================================
// Test helpers
// =========================================================================

// setupAdapterFixture creates a central backup home under a temp root and
// seeds `count` session directories for the given adapter ID. The returned
// home path is the BackupHomeDir() result; adapterDir is the per-adapter
// subdirectory (the parent of every session dir).
func setupAdapterFixture(t *testing.T, adapterID string, count int) (home string, adapterDir string) {
	t.Helper()
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	h, err := BackupHomeDir()
	require.NoError(t, err)
	adapterDir = filepath.Join(h, adapterID)
	require.NoError(t, os.MkdirAll(adapterDir, 0o700))

	for i := 0; i < count; i++ {
		require.NoError(t, os.MkdirAll(filepath.Join(adapterDir, sessionDirName(t, i)), 0o700))
	}
	return h, adapterDir
}

// readSortedDir returns the entries of dir sorted in descending order
// (most-recent first, matching the PruneBackups sort). It fails the test
// if the directory cannot be read.
func readSortedDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err, "ReadDir(%q) must succeed", dir)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names
}

// OverrideUserConfigDir is defined in backup_retention.go (it must be
// exported in a non-test file so the `package common_test` external test
// package can call it from its helpers — e.g. fullInstallTestAdapter,
// installTestAdapter, warningsTestAdapter — and from the 5 direct-build
// error-path tests in base_adapter_error_test.go).
