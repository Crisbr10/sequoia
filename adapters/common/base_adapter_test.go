package common_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	a.SetHomeDir("") // not used — we inject home via ResolveBase
	a.ResolveBase(func(_ string) (string, error) {
		return home, nil
	})

	// Path functions that place files under the fake home.
	skillsDir := filepath.Join(home, "skills")
	cmdsDir := filepath.Join(home, "commands")
	versionFile := filepath.Join(home, "sequoia-version")
	a.SetPathFns(
		func(base string) string { return skillsDir },
		func(base string) string { return cmdsDir },
		func(base string) string { return filepath.Join(home, "system.md") },
		func(base string) string { return versionFile },
		func(base string) string { return filepath.Join(home, "backup") },
	)

	// No-op system prompt removal for tests.
	a.SetStrategy(adapters.StrategyFileReplace,
		nil,
		func(_ string) error { return nil },
	)

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
	a.ResolveBase(func(_ string) (string, error) {
		return home, nil
	})
	a.SetPathFns(
		func(base string) string { return filepath.Join(home, "skills") },
		func(base string) string { return filepath.Join(home, "commands") },
		func(base string) string { return filepath.Join(home, "sys.md") },
		func(base string) string { return filepath.Join(home, "version") },
		func(base string) string { return filepath.Join(home, "backup") },
	)
	a.SetStrategy(adapters.StrategyFileReplace,
		func(base, content string) error { return fmt.Errorf("system prompt write failed") },
		nil,
	)
	// Use testFS (from shared_test.go) which only has testdata/test.tmpl.
	// RenderTemplate will fail looking for "templates/skill.md.tmpl",
	// simulating a template failure.
	// We pass stagingPrefix="" to trigger os.MkdirTemp("", ""), which
	// will fail — but we actually WANT the install to fail before that,
	// at the template rendering step.
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

	// Install should fail because templateFS is nil.
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
	a.SetHomeDir(home)
	a.ResolveBase(func(homeDir string) (string, error) {
		return homeDir, nil
	})
	a.SetPathFns(
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
	)
	a.SetStrategy(adapters.StrategyFileReplace,
		func(base, content string) error { return nil }, // no-op for test
		nil,
	)
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
