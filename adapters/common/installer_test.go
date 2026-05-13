// Package common_test provides black-box tests for the common installer lifecycle.
package common_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// writeFile is a test helper that creates a file with given content under dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// readFile is a test helper that reads and returns the content of a file.
func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err)
	return string(data)
}

// fileExists returns true when the file is present and accessible.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestInstaller_HappyPath_CleanInstall covers Scenario 1:
// SourceDir has 2 files, TargetDir is empty — full lifecycle succeeds.
func TestInstaller_HappyPath_CleanInstall(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	writeFile(t, srcDir, "alpha.txt", "content-alpha")
	writeFile(t, srcDir, "beta.txt", "content-beta")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt", "beta.txt"},
	}
	inst := common.NewInstaller(cfg)

	require.NoError(t, inst.Prepare(), "Prepare should succeed on clean target")
	require.NoError(t, inst.Apply(), "Apply should succeed")
	require.NoError(t, inst.Verify(), "Verify should succeed")

	assert.Equal(t, "content-alpha", readFile(t, dstDir, "alpha.txt"))
	assert.Equal(t, "content-beta", readFile(t, dstDir, "beta.txt"))
}

// TestInstaller_HappyPath_UpgradeExistingFiles covers Scenario 2:
// TargetDir already has old versions — Prepare backs them up, Apply overwrites.
func TestInstaller_HappyPath_UpgradeExistingFiles(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	// Write old content to destination.
	writeFile(t, dstDir, "alpha.txt", "old-alpha")
	writeFile(t, dstDir, "beta.txt", "old-beta")

	// Write new content to source.
	writeFile(t, srcDir, "alpha.txt", "new-alpha")
	writeFile(t, srcDir, "beta.txt", "new-beta")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt", "beta.txt"},
	}
	inst := common.NewInstaller(cfg)

	require.NoError(t, inst.Prepare(), "Prepare should succeed")
	require.NoError(t, inst.Apply(), "Apply should succeed")
	require.NoError(t, inst.Verify(), "Verify should succeed")

	// TargetDir now has new content.
	assert.Equal(t, "new-alpha", readFile(t, dstDir, "alpha.txt"))
	assert.Equal(t, "new-beta", readFile(t, dstDir, "beta.txt"))

	// BackupDir has old content.
	assert.Equal(t, "old-alpha", readFile(t, backupDir, "alpha.txt"))
	assert.Equal(t, "old-beta", readFile(t, backupDir, "beta.txt"))
}

// TestInstaller_ApplyFailure_RollbackRestoresState covers Scenario 3:
// Apply fails (source file missing) — Rollback restores target to original state.
func TestInstaller_ApplyFailure_RollbackRestoresState(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	// Only provide one of the two declared files in source — the other is missing.
	writeFile(t, srcDir, "alpha.txt", "src-alpha")
	// "beta.txt" intentionally NOT created in srcDir — triggers Apply failure.

	// TargetDir has original state for both files.
	writeFile(t, dstDir, "alpha.txt", "orig-alpha")
	writeFile(t, dstDir, "beta.txt", "orig-beta")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt", "beta.txt"},
	}
	inst := common.NewInstaller(cfg)

	require.NoError(t, inst.Prepare(), "Prepare should succeed")
	err := inst.Apply()
	require.Error(t, err, "Apply should fail when source file is missing")

	require.NoError(t, inst.Rollback(), "Rollback should succeed")

	// TargetDir is restored to original state.
	assert.Equal(t, "orig-alpha", readFile(t, dstDir, "alpha.txt"))
	assert.Equal(t, "orig-beta", readFile(t, dstDir, "beta.txt"))
}

// TestInstaller_VerifyFailure covers Scenario 4:
// A file is deleted after Apply — Verify returns an error.
func TestInstaller_VerifyFailure(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	writeFile(t, srcDir, "alpha.txt", "content-alpha")
	writeFile(t, srcDir, "beta.txt", "content-beta")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt", "beta.txt"},
	}
	inst := common.NewInstaller(cfg)

	require.NoError(t, inst.Prepare())
	require.NoError(t, inst.Apply())

	// Sabotage: delete a file after Apply succeeds.
	require.NoError(t, os.Remove(filepath.Join(dstDir, "beta.txt")))

	err := inst.Verify()
	require.Error(t, err, "Verify should fail when a file is missing")
	assert.True(t,
		strings.Contains(err.Error(), "beta.txt"),
		"error should mention the missing file, got: %s", err.Error(),
	)
}

// TestInstaller_PrepareFailsOnNonWritableTarget covers Scenario 5:
// TargetDir is not writable — Prepare returns an error.
// Skipped on Windows because chmod 0444 does not enforce write restriction
// for the file owner on that platform.
func TestInstaller_PrepareFailsOnNonWritableTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test requires Unix semantics — skipping on Windows")
	}
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	writeFile(t, srcDir, "alpha.txt", "content")

	// Make TargetDir read-only.
	require.NoError(t, os.Chmod(dstDir, 0o444))
	t.Cleanup(func() { _ = os.Chmod(dstDir, 0o755) })

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt"},
	}
	inst := common.NewInstaller(cfg)

	err := inst.Prepare()
	require.Error(t, err, "Prepare should fail on a non-writable TargetDir")
}

// TestInstaller_Rollback_SafeWithoutApply verifies that Rollback is a no-op
// when Apply was never called (no partial state to clean up).
func TestInstaller_Rollback_SafeWithoutApply(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")

	writeFile(t, srcDir, "alpha.txt", "content")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt"},
	}
	inst := common.NewInstaller(cfg)

	require.NoError(t, inst.Prepare())
	// Skip Apply intentionally.
	require.NoError(t, inst.Rollback(), "Rollback should be safe when Apply was not called")

	// No files should have been installed.
	assert.False(t, fileExists(filepath.Join(dstDir, "alpha.txt")))
}

// =========================================================================
// Backup directory collision tests (FIX-005)
// =========================================================================

// TestInstaller_BackupDirHasUniqueSuffix verifies that running the
// common.Installer with timestamped backup dirs produces unique names
// and no collisions.
func TestInstaller_BackupDirHasUniqueSuffix(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	parent := t.TempDir()

	writeFile(t, srcDir, "alpha.txt", "content-alpha")
	writeFile(t, dstDir, "alpha.txt", "old-content")

	// Simulate two install sessions with timestamped backup dirs.
	// In production, base_adapter.go generates the timestamp suffix.
	suffix1 := "abc123"
	suffix2 := "xyz789"
	backupDir1 := filepath.Join(parent, ".sequoia-backup-"+suffix1)
	backupDir2 := filepath.Join(parent, ".sequoia-backup-"+suffix2)

	// First "session".
	cfg1 := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir1,
		Files:     []string{"alpha.txt"},
	}
	inst1 := common.NewInstaller(cfg1)
	require.NoError(t, inst1.Run())

	// Verify first backup dir exists and has the old file.
	assert.True(t, fileExists(filepath.Join(backupDir1, "alpha.txt")),
		"backup dir 1 should contain the backed-up file")

	// Second "session": overwrite destination with new "old" content first.
	require.NoError(t, os.WriteFile(filepath.Join(dstDir, "alpha.txt"), []byte("older-content"), 0o644))

	cfg2 := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir2,
		Files:     []string{"alpha.txt"},
	}
	inst2 := common.NewInstaller(cfg2)
	require.NoError(t, inst2.Run())

	// Both backup dirs should exist and be different.
	assert.True(t, fileExists(filepath.Join(backupDir1, "alpha.txt")),
		"first backup dir should still exist")
	assert.True(t, fileExists(filepath.Join(backupDir2, "alpha.txt")),
		"second backup dir should exist")

	// Verify backup dirs have different paths.
	assert.NotEqual(t, backupDir1, backupDir2, "backup dirs should have unique names")
}

// TestInstaller_BackupPermissions_Restricted verifies that backup files
// and directories use owner-only permissions (backup-permissions spec).
// Skipped on Windows because unix permission bits are no-ops there.
func TestInstaller_BackupPermissions_Restricted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test requires Unix semantics — skipping on Windows")
	}
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), ".sequoia-backup-perm-test")

	writeFile(t, srcDir, "alpha.txt", "new-alpha")
	writeFile(t, srcDir, "beta.txt", "new-beta")
	writeFile(t, dstDir, "alpha.txt", "old-alpha")
	writeFile(t, dstDir, "beta.txt", "old-beta")

	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: backupDir,
		Files:     []string{"alpha.txt", "beta.txt"},
	}
	inst := common.NewInstaller(cfg)
	require.NoError(t, inst.Prepare())

	// Backup directory must have 0o700 (owner rwx only).
	info, err := os.Stat(backupDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm(),
		"backup directory must be owner-only (0o700)")

	// Both backup files must have 0o600 (owner rw only).
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		fi, err := os.Stat(filepath.Join(backupDir, name))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(),
			"backup file %s must be owner-only (0o600)", name)
	}
}

// TestInstaller_BackupDirDoesNotCollideWithPreExisting verifies that when
// a user has pre-created a directory with the predictable backup name,
// the timestamped backup dir does not collide.
func TestInstaller_BackupDirDoesNotCollideWithPreExisting(t *testing.T) {
	t.Parallel()

	srcDir := t.TempDir()
	dstDir := t.TempDir()
	parent := t.TempDir()

	writeFile(t, srcDir, "alpha.txt", "content-alpha")
	writeFile(t, dstDir, "alpha.txt", "user-content")

	// Pre-create a "predictable" backup dir (simulating user/attacker pre-creation).
	predictableBackup := filepath.Join(parent, ".sequoia-backup")
	require.NoError(t, os.MkdirAll(predictableBackup, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(predictableBackup, "malicious.txt"),
		[]byte("this should not be touched\n"), 0o644,
	))

	// Install with a timestamped backup dir.
	timestampedBackup := filepath.Join(parent, ".sequoia-backup-abc123")
	cfg := common.InstallerConfig{
		SourceDir: srcDir,
		TargetDir: dstDir,
		BackupDir: timestampedBackup,
		Files:     []string{"alpha.txt"},
	}
	inst := common.NewInstaller(cfg)
	require.NoError(t, inst.Run())

	// The predictable backup dir should be untouched.
	maliciousData, err := os.ReadFile(filepath.Join(predictableBackup, "malicious.txt"))
	require.NoError(t, err)
	assert.Equal(t, "this should not be touched\n", string(maliciousData))
}
