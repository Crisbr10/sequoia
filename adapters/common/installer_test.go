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

	"sequoia-ai/adapters/common"
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
