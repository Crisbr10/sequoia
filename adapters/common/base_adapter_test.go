package common_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
)

// uninstallTestAdapter creates a minimal BaseAdapter for testing Uninstall()
// with temp directories. The adapter writes to a temp "home" directory that
// simulates the tool's config root.
func uninstallTestAdapter(t *testing.T, home string) *common.BaseAdapter {
	t.Helper()

	a := &common.BaseAdapter{}
	a.SetIDName("test-adapter", "Test Adapter")

	skillsDir := filepath.Join(home, "skills")
	cmdsDir := filepath.Join(home, "commands")
	versionFile := filepath.Join(home, "sequoia-version")

	a.SetPaths(common.NewPathResolver(
		func(_ string) (string, error) { return home, nil },
		"",
		func(base string) string { return skillsDir },
		func(base string) string { return cmdsDir },
		func(base string) string { return filepath.Join(home, "system.md") },
		func(base string) string { return versionFile },
		func(base string) string { return filepath.Join(home, "backup") },
		a.AddWarning,
	))

	// No-op system prompt removal for tests.
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		nil,
		func(_ string) error { return nil },
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(home, "backup") },
		"test-adapter",
	))

	return a
}

// createUninstallFiles creates the expected file structure that Uninstall()
// would remove: SKILL.md in skillsDir, version file, and command files.
func createUninstallFiles(t *testing.T, skillsDir, cmdsDir, versionFile string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(skillsDir, 0o755))
	require.NoError(t, os.MkdirAll(cmdsDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte("skill"), 0o644))
	require.NoError(t, os.WriteFile(versionFile, []byte("0.1.0\n"), 0o644))
	for _, cmd := range common.CommandFiles {
		require.NoError(t, os.WriteFile(filepath.Join(cmdsDir, cmd), []byte("cmd"), 0o644))
	}
}

// =========================================================================
// TestUninstall_NoErrorWhenFilesMissing
// =========================================================================

// TestUninstall_NoErrorWhenFilesMissing verifies that Uninstall returns nil
// when no Sequoia files exist (missing files are not errors during cleanup).
func TestUninstall_NoErrorWhenFilesMissing(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := uninstallTestAdapter(t, home)
	// Do NOT create any files — they're all missing.
	err := a.Uninstall(adapters.InstallOpts{})

	assert.NoError(t, err, "Uninstall should not error when files are already missing")
}

// =========================================================================
// TestUninstall_CollectsErrors
// =========================================================================

// TestUninstall_CollectsErrors verifies that when file removal fails
// (e.g., path is a non-empty directory instead of a file), Uninstall
// collects and returns all errors via a joined error, rather than
// silently discarding them.
func TestUninstall_CollectsErrors(t *testing.T) {
	t.Parallel()

	home := t.TempDir()

	// Determine paths — must match what the adapter uses.
	skillsDir := filepath.Join(home, "skills")
	cmdsDir := filepath.Join(home, "commands")
	versionFile := filepath.Join(home, "sequoia-version")

	// Create real files for skills and commands.
	createUninstallFiles(t, skillsDir, cmdsDir, versionFile)

	// Make the "version file" path be a non-empty directory so os.Remove
	// fails (os.Remove on a non-empty directory returns an error).
	require.NoError(t, os.Remove(versionFile))
	require.NoError(t, os.MkdirAll(versionFile, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(versionFile, "blocker.txt"), []byte("x"), 0o644))

	a := uninstallTestAdapter(t, home)
	err := a.Uninstall(adapters.InstallOpts{})

	// The error should NOT be nil — at least the version file removal failed.
	require.Error(t, err, "Uninstall should return an error when file removal fails")

	// The error should mention which file failed.
	t.Logf("Uninstall error: %v", err)
	assert.Contains(t, err.Error(), "version file", "error should mention the version file removal")
}

// =========================================================================
// TestUninstall_ReturnsSentinelError
// =========================================================================

// TestUninstall_ReturnsSentinelError verifies that an uninstall failure
// wraps the adapters.ErrUninstallFailed sentinel error so callers can
// detect it with errors.Is.
func TestUninstall_ReturnsSentinelError(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	cmdsDir := filepath.Join(home, "commands")
	versionFile := filepath.Join(home, "sequoia-version")

	createUninstallFiles(t, skillsDir, cmdsDir, versionFile)

	// Make the version file path a non-empty directory so os.Remove fails.
	require.NoError(t, os.Remove(versionFile))
	require.NoError(t, os.MkdirAll(versionFile, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(versionFile, "blocker.txt"), []byte("x"), 0o644))

	a := uninstallTestAdapter(t, home)
	err := a.Uninstall(adapters.InstallOpts{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrUninstallFailed),
		"error should wrap ErrUninstallFailed, got: %v", err)
}

// =========================================================================
// TestUninstall_PartialFailure
// =========================================================================

// TestUninstall_PartialFailure verifies that when some files are removable
// and others are not, Uninstall returns an error that describes which files
// could not be removed. The removable files should actually be deleted.
func TestUninstall_PartialFailure(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	skillsDir := filepath.Join(home, "skills")
	cmdsDir := filepath.Join(home, "commands")
	versionFile := filepath.Join(home, "sequoia-version")

	createUninstallFiles(t, skillsDir, cmdsDir, versionFile)

	// Make only the version file path a non-empty directory — skills and
	// commands should still be removable.
	require.NoError(t, os.Remove(versionFile))
	require.NoError(t, os.MkdirAll(versionFile, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(versionFile, "blocker.txt"), []byte("x"), 0o644))

	a := uninstallTestAdapter(t, home)
	err := a.Uninstall(adapters.InstallOpts{})

	require.Error(t, err, "Uninstall should return an error due to partial failure")

	// The removable files should be gone.
	_, statErr := os.Stat(filepath.Join(skillsDir, "SKILL.md"))
	assert.True(t, os.IsNotExist(statErr), "SKILL.md should have been removed successfully")

	// The error should mention the failed file.
	assert.Contains(t, err.Error(), "version file", "error should reference the failed version file removal")
}

// =========================================================================
// TestInstall_ReturnsSentinelError
// =========================================================================

// installTestAdapter creates a minimal BaseAdapter for Install testing
// that is guaranteed to fail (no templates set).
func installTestAdapter(t *testing.T, home string) *common.BaseAdapter {
	t.Helper()

	a := &common.BaseAdapter{}
	a.SetIDName("test-adapter", "Test Adapter")
	a.SetPaths(common.NewPathResolver(
		func(_ string) (string, error) { return home, nil },
		"",
		func(base string) string { return filepath.Join(home, "skills") },
		func(base string) string { return filepath.Join(home, "commands") },
		func(base string) string { return filepath.Join(home, "sys.md") },
		func(base string) string { return filepath.Join(home, "version") },
		func(base string) string { return filepath.Join(home, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return fmt.Errorf("system prompt write failed") },
		nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(home, "backup") },
		"test-adapter",
	))
	// Use testFS (from shared_test.go) which only has testdata/test.tmpl.
	// RenderTemplate will fail looking for "templates/skill.md.tmpl",
	// simulating a template failure.
	a.SetInstallTemplates(testFS, "sequoia-test-*",
		"templates/skill.md.tmpl",
		func() interface{} { return map[string]string{"Name": "test"} })

	return a
}

// TestInstall_ReturnsSentinelError verifies that a failed Install wraps
// adapters.ErrInstallFailed so callers can detect it with errors.Is.
func TestInstall_ReturnsSentinelError(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := installTestAdapter(t, home)

	// Install should fail because templateFS lacks skill template.
	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err, "Install should fail when templates are missing")

	assert.True(t, errors.Is(err, adapters.ErrInstallFailed),
		"error should wrap ErrInstallFailed, got: %v", err)
}

// =========================================================================
// TestBaseAdapter_AddWarning_Warnings
// =========================================================================

// TestBaseAdapter_AddWarning_Warnings verifies that AddWarning appends to the
// internal warnings slice and that Warnings returns a copy (defensive).
func TestBaseAdapter_AddWarning_Warnings(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}

	// Start with no warnings.
	assert.Empty(t, a.Warnings(), "warnings should start empty")

	// Add a warning.
	a.AddWarning("test warning one")
	warns := a.Warnings()
	assert.Len(t, warns, 1, "should have 1 warning after adding one")
	assert.Equal(t, "test warning one", warns[0])

	// Add another warning.
	a.AddWarning("test warning two")
	warns = a.Warnings()
	assert.Len(t, warns, 2, "should have 2 warnings after adding second")

	// Warnings() must return a copy — mutating the returned slice
	// must not affect the internal slice.
	warns[0] = "mutated"
	warns2 := a.Warnings()
	assert.Equal(t, "test warning one", warns2[0], "Warnings() must return a defensive copy")
	assert.Len(t, warns2, 2)
}

// TestBaseAdapter_AddWarning_ThreadSafety verifies concurrent AddWarning
// and Warnings calls do not race. Run with: go test -race
func TestBaseAdapter_AddWarning_ThreadSafety(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}
	const numGoroutines = 50
	done := make(chan struct{})

	// Launch writers and readers concurrently.
	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			a.AddWarning(fmt.Sprintf("warning-%d", n))
			done <- struct{}{}
		}(i)
		go func() {
			_ = a.Warnings()
			done <- struct{}{}
		}()
	}

	// Wait for all goroutines.
	for i := 0; i < numGoroutines*2; i++ {
		<-done
	}

	// Final warnings should be non-empty.
	assert.NotEmpty(t, a.Warnings(), "should have accumulated warnings")
}

// =========================================================================
// TestBaseAdapter_WarningsClearedOnInstall
// =========================================================================

// warningsTestAdapter creates a minimal BaseAdapter for Install testing
// with a working template setup so Install succeeds.
func warningsTestAdapter(t *testing.T, home string) *common.BaseAdapter {
	t.Helper()

	a := &common.BaseAdapter{}
	a.SetIDName("warn-adapter", "Warning Adapter")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		home,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return nil }, // no-op for test
		nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(home, "backup") },
		"warn-adapter",
	))
	a.SetInstallTemplates(testFS, "sequoia-warn-*",
		"testdata/test.tmpl",
		func() interface{} { return map[string]string{"Name": "warn", "Version": "0.1.0"} })
	return a
}

// TestBaseAdapter_WarningsClearedOnInstall verifies that warnings are cleared
// when Install() starts (even if Install later fails).
func TestBaseAdapter_WarningsClearedOnInstall(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := warningsTestAdapter(t, home)

	// Pre-populate warnings.
	a.AddWarning("stale warning from previous run")

	// Install should fail (templateFS is testFS without "templates/skill.md")
	// but warnings should be cleared at the start nonetheless.
	_ = a.Install(adapters.InstallOpts{})

	// The stale warning must be gone.
	assert.Empty(t, a.Warnings(), "warnings should be cleared at start of Install")
}

// =========================================================================
// TestBaseAdapter_BaseCachesUserHomeDir
// =========================================================================

// TestBaseAdapter_BaseCachesUserHomeDir verifies that SkillsPath/CommandsPath
// return stable results across repeated calls (home dir is cached via
// sync.Once in PathResolver).
func TestBaseAdapter_BaseCachesUserHomeDir(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}
	a.SetIDName("cache-test", "Cache Test")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) {
			return filepath.Join(homeDir, ".cache-test"), nil
		},
		"", // empty → uses os.UserHomeDir()
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))

	// Calling SkillsPath() triggers base(), which calls os.UserHomeDir().
	// Repeated calls should return the same path (cached).
	first := a.SkillsPath()
	require.NotEmpty(t, first, "SkillsPath should be non-empty")

	second := a.SkillsPath()
	assert.Equal(t, first, second, "SkillsPath should be stable across calls (home dir cached)")

	// CommandsPath should also use the cached home.
	cmds := a.CommandsPath()
	assert.NotEmpty(t, cmds, "CommandsPath should be non-empty")
	assert.Contains(t, cmds, ".cache-test", "CommandsPath should use the cached home dir")
}

// TestBaseAdapter_HomeDirOverrideBypassesCache verifies that when
// SetHomeDir is used (via SetPaths with explicit homeDir), Base() uses
// the explicit path without calling os.UserHomeDir().
func TestBaseAdapter_HomeDirOverrideBypassesCache(t *testing.T) {
	t.Parallel()

	raw := t.TempDir()
	// Normalize the temp dir path: on Windows, t.TempDir() may return paths
	// using 8.3 short names (C:\Users\RUNNER~1\...) while ResolveSymlink
	// inside base() resolves them to long names (C:\Users\runneradmin\...).
	// Normalize before comparison to avoid short-vs-long mismatch.
	tmp, err := filepath.EvalSymlinks(raw)
	require.NoError(t, err)

	a := &common.BaseAdapter{}
	a.SetIDName("override-test", "Override Test")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) {
			return filepath.Join(homeDir, ".override"), nil
		},
		tmp, // explicit homeDir
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))

	// SkillsPath should use the explicit homeDir.
	sp := a.SkillsPath()
	assert.Contains(t, sp, tmp, "SkillsPath should use the explicit homeDir via override")
	assert.Contains(t, sp, ".override", "SkillsPath should include the resolved base")
}

// =========================================================================
// TestBackupIsolation_FreshInstallProducesIdenticalOutput
// REQ-BACKUP-ISOLATION-003 Scenario 1
// =========================================================================

// TestBackupIsolation_FreshInstallProducesIdenticalOutput verifies that a
// fresh install with the new namespaced backup paths produces identical
// output files as before the backup isolation change. The backup directory
// structure should be the only difference; installed file content must be
// byte-for-byte identical regardless of backup path changes.
func TestBackupIsolation_FreshInstallProducesIdenticalOutput(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)

	err := a.Install(adapters.InstallOpts{})
	require.NoError(t, err, "fresh install with namespaced backup paths should succeed")

	// Verify SKILL.md at target has expected rendered content.
	skillsDir := a.SkillsPath()
	skillBytes, err := os.ReadFile(filepath.Join(skillsDir, "SKILL.md"))
	require.NoError(t, err)
	skillContent := strings.ReplaceAll(string(skillBytes), "\r\n", "\n")
	// installembed.FS template "templates/skill.md.tmpl" is:
	//   {{.Name}} skill template for testing
	// With data {"Name": "err-test"}, the rendered output is:
	//   err-test skill template for testing
	assert.Contains(t, skillContent, "err-test",
		"installed SKILL.md should contain adapter name")
	assert.Contains(t, skillContent, "skill template for testing",
		"installed SKILL.md should contain template content")

	// Verify command files at target have the correct static content.
	cmdsDir := a.CommandsPath()
	for _, cmd := range common.CommandFiles {
		cmdPath := filepath.Join(cmdsDir, cmd)
		assert.FileExists(t, cmdPath,
			"command %s should exist at target after install", cmd)

		cmdBytes, err := os.ReadFile(cmdPath)
		require.NoError(t, err, "command file %s should be readable", cmd)
		assert.NotEmpty(t, cmdBytes,
			"command file %s should have non-empty content", cmd)

		// Command files come from CommandFS (static, no rendering).
		// Verify they match the embedded template content byte-for-byte.
		expectedBytes, err := common.CommandFS.ReadFile("templates/commands/" + cmd)
		require.NoError(t, err, "CommandFS should contain %s", cmd)
		// Normalize line endings for cross-platform comparison.
		got := strings.ReplaceAll(string(cmdBytes), "\r\n", "\n")
		want := strings.ReplaceAll(string(expectedBytes), "\r\n", "\n")
		assert.Equal(t, want, got,
			"installed command %s should match CommandFS content byte-for-byte", cmd)
	}

	// Verify version file exists and has correct content.
	versionPath := filepath.Join(home, "version")
	assert.FileExists(t, versionPath, "version file should exist after successful install")
	versionBytes, err := os.ReadFile(versionPath)
	require.NoError(t, err)
	assert.Equal(t, common.Version, strings.TrimSpace(string(versionBytes)),
		"version file should contain the current Sequoia version")
}

// =========================================================================
// TestBackupIsolation_NamespacedBackupStructure
// REQ-BACKUP-ISOLATION-003 Scenario 2
// =========================================================================

// TestBackupIsolation_NamespacedBackupStructure verifies that after a
// successful install (no rollback), the backup directory uses the namespaced
// structure {base}-{adapterID}-{suffix} with type-specific subdirectories
// for skills and commands. It also verifies that no backup cleanup is
// performed after a successful install.
func TestBackupIsolation_NamespacedBackupStructure(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)

	// Pre-create files at target so the installer backs them up.
	// This lets us verify the backup directory has real content.
	skillsDir := a.SkillsPath()
	cmdsDir := a.CommandsPath()
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))
	require.NoError(t, os.MkdirAll(cmdsDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(skillsDir, "SKILL.md"),
		[]byte("original skill content"), 0o644,
	))
	for _, cmd := range common.CommandFiles {
		require.NoError(t, os.WriteFile(
			filepath.Join(cmdsDir, cmd),
			[]byte("original "+cmd), 0o644,
		))
	}

	err := a.Install(adapters.InstallOpts{})
	require.NoError(t, err, "install should succeed with pre-existing target files")

	// --- Verify backup directory uses namespaced structure ---
	backupDir := a.LastBackupDir()
	require.NotEmpty(t, backupDir, "LastBackupDir should be set after successful install")

	// Backup path should contain the adapter ID.
	assert.Contains(t, backupDir, "err-test",
		"backup path should contain the adapter ID")

	// Backup path should have format: {base}-{adapterID}-{sessionSuffix}
	// The session suffix is a base-36 timestamp, so after "err-test-" there
	// should be another segment (the suffix).
	adapterIdx := strings.LastIndex(backupDir, "err-test")
	require.True(t, adapterIdx >= 0, "backup path should contain adapter ID")
	suffixStart := adapterIdx + len("err-test")
	assert.True(t, suffixStart < len(backupDir),
		"backup path should have a session suffix after the adapter ID")
	// The character right after the adapter ID should be "-".
	assert.Equal(t, byte('-'), backupDir[suffixStart],
		"backup path should separate adapter ID and suffix with '-'")

	// --- Verify type-specific backup subdirectories ---
	skillBackupDir := filepath.Join(backupDir, "skills")
	cmdBackupDir := filepath.Join(backupDir, "commands")

	assert.True(t, fileExists(skillBackupDir),
		"skills backup subdirectory should exist under namespaced backup dir")
	assert.True(t, fileExists(cmdBackupDir),
		"commands backup subdirectory should exist under namespaced backup dir")

	// --- Verify backed-up files have original content ---
	skillBackupPath := filepath.Join(skillBackupDir, "SKILL.md")
	assert.True(t, fileExists(skillBackupPath),
		"SKILL.md should be backed up under skills subdirectory")
	skillBackupBytes, err := os.ReadFile(skillBackupPath)
	require.NoError(t, err)
	assert.Equal(t, "original skill content", string(skillBackupBytes),
		"backed-up SKILL.md should have original content")

	for _, cmd := range common.CommandFiles {
		cmdBackupPath := filepath.Join(cmdBackupDir, cmd)
		assert.True(t, fileExists(cmdBackupPath),
			"command %s should be backed up under commands subdirectory", cmd)
		cmdBackupBytes, err := os.ReadFile(cmdBackupPath)
		require.NoError(t, err)
		assert.Equal(t, "original "+cmd, string(cmdBackupBytes),
			"backed-up command %s should have original content", cmd)
	}

	// --- Verify target files have NEW (installed) content ---
	// SKILL.md should be overwritten with new rendered content.
	newSkillBytes, err := os.ReadFile(filepath.Join(skillsDir, "SKILL.md"))
	require.NoError(t, err)
	newSkillContent := strings.ReplaceAll(string(newSkillBytes), "\r\n", "\n")
	assert.Contains(t, newSkillContent, "err-test",
		"installed SKILL.md should have new content, not original")
	assert.NotEqual(t, "original skill content", strings.TrimSpace(newSkillContent),
		"installed SKILL.md should differ from original backed-up content")

	// --- Verify no backup cleanup was performed ---
	// After a successful install, the backup directory should remain intact.
	assert.True(t, fileExists(backupDir),
		"backup directory should NOT be cleaned up after successful install")
	assert.True(t, fileExists(skillBackupPath),
		"backed-up SKILL.md should still exist after successful install (no cleanup)")
}
