//nolint:gosec // test file: all os.* operations use t.TempDir() test fixtures, not production paths
package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
)

// TestAdapter_Install_BackupDirUsesCentralHome verifies that after
// codex.Adapter.Install() (which overrides BaseAdapter.Install), the
// recorded LastBackupDir starts with the central home returned by
// common.BackupHomeDir() and is set to a non-empty value. REQ-BRP-02.
//
// Before PR 2 task 2.7, the custom Codex Install did NOT set
// lastBackupDir, so the BackupDirGetter interface reported an empty
// path and the TUI Info message never showed the backup location.
//
// The test uses the real os.UserConfigDir() (matching the existing
// codex test pattern) and cleans up the per-adapter subdir under
// the central home on test teardown. Worst case: a single session
// dir under the real user config if the cleanup fails.
func TestAdapter_Install_BackupDirUsesCentralHome(t *testing.T) {
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	// Compute the central home for the assertion and the cleanup hook.
	userConfig, err := os.UserConfigDir()
	require.NoError(t, err, "os.UserConfigDir must succeed for the test environment")
	centralHome := filepath.Join(userConfig, "sequoia", "backups")

	// Best-effort cleanup of any prior codex backup subdirs the test
	// might leave behind if it fails between Build and cleanup.
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(centralHome, "codex")) })

	a := newAdapter(tmp)
	require.NoError(t, a.Install(adapters.InstallOpts{}),
		"Codex custom Install should succeed against a temp home")

	// LastBackupDir must be set and live under the central home.
	backupDir := a.LastBackupDir()
	require.NotEmpty(t, backupDir,
		"Codex.Install must set LastBackupDir so the TUI can surface the central-home backup location (REQ-BRP-02)")

	assert.True(t, strings.HasPrefix(backupDir, centralHome+string(filepath.Separator)),
		"LastBackupDir %q must start with the central home %q",
		backupDir, centralHome)

	// Must NOT use the per-tool path or the legacy .sequoia-backup marker.
	assert.NotContains(t, backupDir, ".sequoia-backup",
		"LastBackupDir must not use the legacy per-tool .sequoia-backup marker")
	assert.NotContains(t, backupDir, ".codex",
		"LastBackupDir must not live under the per-tool codex config dir")
}
