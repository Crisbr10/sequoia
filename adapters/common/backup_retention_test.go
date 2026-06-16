//nolint:gosec // test file: uses t.TempDir() and controlled fixtures
package common

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

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
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })

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
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })

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

	overrideUserConfigDir(t, func() (string, error) {
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

// overrideUserConfigDir swaps the package-level userConfigDir hook for the
// duration of the test, restoring the previous value on cleanup.
//
// Note: do NOT call t.TempDir() inside the override closure AND again in the
// test body — t.TempDir() returns a fresh subdir on every call, so the two
// values would not match. Capture the temp dir in a local variable first:
//
//	tmp := t.TempDir()
//	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })

func overrideUserConfigDir(t *testing.T, fn func() (string, error)) {
	t.Helper()
	orig := userConfigDir
	userConfigDir = fn
	t.Cleanup(func() { userConfigDir = orig })
}

// mustUserConfigDir returns the current value of userConfigDir, failing the
// test on error. It is a small helper to keep test bodies readable.
func mustUserConfigDir(t *testing.T) string {
	t.Helper()
	cfg, err := userConfigDir()
	require.NoError(t, err, "userConfigDir override must succeed")
	return cfg
}
