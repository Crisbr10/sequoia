package common_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/adapters/common/installembed"
)

// =========================================================================
// Compile-time check: BaseAdapter satisfies Strategy
// =========================================================================
var _ common.Strategy = (*common.BaseAdapter)(nil)

// =========================================================================
// strategyTestAdapter creates a BaseAdapter for Strategy phase testing.
// =========================================================================
func strategyTestAdapter(t *testing.T, home string) *common.BaseAdapter {
	t.Helper()

	a := &common.BaseAdapter{}
	a.SetIDName("strategy-test", "Strategy Test")
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
		func(base, content string) error { return nil },
		nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"strategy-test",
	))
	a.SetDetector(common.NewDetector(
		a.Base,
		func(base string) bool {
			_, err := os.Stat(filepath.Join(base, "version"))
			return err == nil
		},
		func() bool { return true },
	))
	a.SetInstallTemplates(installembed.FS, "sequoia-strategy-*",
		"templates/sys_prompt.md.tmpl",
		func() interface{} { return map[string]string{"Name": "strategy-test"} })

	return a
}

// =========================================================================
// TestStrategy_PrepareSucceeds — RED
// =========================================================================
func TestStrategy_PrepareSucceeds(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	err := a.Prepare(adapters.InstallOpts{})
	t.Logf("Prepare error: %v", err)
	require.NoError(t, err, "Prepare should succeed with valid adapter")

	// Skills and commands dirs should be created.
	assert.DirExists(t, filepath.Join(home, "skills"), "Prepare should create skills dir")
	assert.DirExists(t, filepath.Join(home, "commands"), "Prepare should create commands dir")
}

// =========================================================================
// TestStrategy_PreparePreCancelledContext — RED
// =========================================================================
func TestStrategy_PreparePreCancelledContext(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := a.Prepare(adapters.InstallOpts{Context: ctx})
	require.Error(t, err, "Prepare with cancelled context should fail")
	assert.Contains(t, err.Error(), "context canceled", "error should mention context cancellation")

	// No directories created.
	assert.NoDirExists(t, filepath.Join(home, "skills"))
	assert.NoDirExists(t, filepath.Join(home, "commands"))
}

// =========================================================================
// TestStrategy_DownloadSucceeds — RED
// =========================================================================
func TestStrategy_DownloadSucceeds(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	require.NoError(t, a.Prepare(adapters.InstallOpts{}),
		"Prepare must succeed before Download")

	err := a.Download(adapters.InstallOpts{})
	require.NoError(t, err, "Download should succeed after Prepare")
}

// =========================================================================
// TestStrategy_DownloadWithoutPrepare — RED
// =========================================================================
func TestStrategy_DownloadWithoutPrepare(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	err := a.Download(adapters.InstallOpts{})
	require.Error(t, err, "Download without Prepare should fail")
	assert.Contains(t, err.Error(), "Prepare must be called first")
}

// =========================================================================
// TestStrategy_VerifySucceeds — RED
// =========================================================================
func TestStrategy_VerifySucceeds(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	require.NoError(t, a.Prepare(adapters.InstallOpts{}))
	require.NoError(t, a.Download(adapters.InstallOpts{}))

	err := a.Verify()
	require.NoError(t, err, "Verify should succeed after Download")
}

// =========================================================================
// TestStrategy_VerifyWithoutDownload — RED
// =========================================================================
func TestStrategy_VerifyWithoutDownload(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	// No Download called — Verify should fail.
	err := a.Verify()
	require.Error(t, err, "Verify without Download should fail")
}

// =========================================================================
// TestStrategy_FullSequenceSucceeds — RED
// =========================================================================
func TestStrategy_FullSequenceSucceeds(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	opts := adapters.InstallOpts{}

	require.NoError(t, a.Prepare(opts), "Prepare")
	require.NoError(t, a.Download(opts), "Download")
	require.NoError(t, a.Verify(), "Verify")
	require.NoError(t, a.Stage(opts), "Stage")
	require.NoError(t, a.Apply(opts), "Apply")

	// Files should be installed.
	assert.FileExists(t, filepath.Join(home, "skills", "SKILL.md"), "SKILL.md should be installed")
	assert.FileExists(t, filepath.Join(home, "version"), "version file should be installed")

	for _, cmd := range common.CommandFiles() {
		assert.FileExists(t, filepath.Join(home, "commands", cmd),
			"command %s should be installed", cmd)
	}
}

// =========================================================================
// TestStrategy_FullSequenceFailureRollback — RED
// =========================================================================
func TestStrategy_FullSequenceFailureRollback(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	opts := adapters.InstallOpts{}

	require.NoError(t, a.Prepare(opts))
	require.NoError(t, a.Download(opts))
	require.NoError(t, a.Verify())
	require.NoError(t, a.Stage(opts))

	// Make Apply fail by writing the version path as a directory.
	versionDir := filepath.Join(home, "version")
	require.NoError(t, os.MkdirAll(versionDir, 0o755))

	err := a.Apply(opts)
	require.Error(t, err, "Apply should fail when version path is a directory")

	// Rollback should clean up.
	rbErr := a.Rollback()
	assert.NoError(t, rbErr, "Rollback should succeed after Apply failure")
}

// =========================================================================
// TestStrategy_InstallBackwardCompat — RED
// =========================================================================
func TestStrategy_InstallBackwardCompat(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	err := a.Install(adapters.InstallOpts{})
	require.NoError(t, err, "Install should still work via backward-compat")

	// Files should be installed.
	assert.FileExists(t, filepath.Join(home, "skills", "SKILL.md"))
	assert.FileExists(t, filepath.Join(home, "version"))

	// Sentinel check.
	assert.False(t, errors.Is(err, adapters.ErrInstallFailed),
		"successful Install should not wrap ErrInstallFailed")
}

// =========================================================================
// TestStrategy_InstallFailsWrapsSentinel — RED
// =========================================================================
func TestStrategy_InstallFailsWrapsSentinel(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	// Make a pre-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := a.Install(adapters.InstallOpts{Context: ctx})
	require.Error(t, err, "Install with cancelled context should fail")
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed),
		"failed Install should wrap ErrInstallFailed")
}

// =========================================================================
// TestStrategy_StageCancelledContext — RED
// =========================================================================
func TestStrategy_StageCancelledContext(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	opts := adapters.InstallOpts{}

	require.NoError(t, a.Prepare(opts))
	require.NoError(t, a.Download(opts))
	require.NoError(t, a.Verify())

	// Cancel context before Stage.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := a.Stage(adapters.InstallOpts{Context: ctx})
	require.Error(t, err, "Stage with cancelled context should fail")
	assert.Contains(t, err.Error(), "context canceled")
}

// =========================================================================
// TestStrategy_RollbackCleansStaging — RED
// =========================================================================
func TestStrategy_RollbackCleansStaging(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	opts := adapters.InstallOpts{}

	require.NoError(t, a.Prepare(opts))
	require.NoError(t, a.Download(opts))
	require.NoError(t, a.Verify())
	require.NoError(t, a.Stage(opts))

	// Rollback should clean up.
	err := a.Rollback()
	assert.NoError(t, err, "Rollback should succeed")
}

// =========================================================================
// TestStrategy_InstallSentinelOnPhaseError — RED
// =========================================================================
func TestStrategy_InstallSentinelOnPhaseError(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := strategyTestAdapter(t, home)

	// Make Apply fail.
	versionDir := filepath.Join(home, "version")
	require.NoError(t, os.MkdirAll(versionDir, 0o755))

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err, "Install should fail when version path is a dir")
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed),
		"should wrap ErrInstallFailed even when phase fails")
}
