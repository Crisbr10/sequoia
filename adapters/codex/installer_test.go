package codex_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/codex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeConfig_FreshFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	table := map[string]interface{}{
		"skills_path":   "/home/user/.codex/sequoia/skills/",
		"commands_path": "/home/user/.codex/sequoia/commands/",
	}

	err := codex.MergeConfig(configPath, table)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "[sequoia]")
	assert.Contains(t, content, "skills_path")
	assert.Contains(t, content, "commands_path")
}

func TestMergeConfig_ExistingSequoia_Overwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	existing := `[settings]
theme = "dark"

[sequoia]
skills_path = "/old/path"
version = "0.0.1"
`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0o644))

	table := map[string]interface{}{
		"skills_path":   "/new/path",
		"commands_path": "/new/commands",
	}

	err := codex.MergeConfig(configPath, table)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "[settings]")
	assert.Contains(t, content, "[sequoia]")
	assert.Contains(t, content, "/new/path")
	assert.NotContains(t, content, "/old/path")
	assert.NotContains(t, content, "version")
}

func TestMergeConfig_PreservesOtherContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	existing := `[models]
default = "gpt-4"

[nested]
  [nested.deep]
  key = "value"
`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0o644))

	table := map[string]interface{}{
		"skills_path": "/path",
	}

	err := codex.MergeConfig(configPath, table)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "[models]")
	assert.Contains(t, content, "gpt-4")
	assert.Contains(t, content, "[nested]")
	assert.Contains(t, content, "key")
	assert.Contains(t, content, "[sequoia]")
}

func TestMergeConfig_CreatesBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	original := "[settings]\ntheme = \"dark\"\n"
	require.NoError(t, os.WriteFile(configPath, []byte(original), 0o644))

	table := map[string]interface{}{
		"skills_path": "/path",
	}

	err := codex.MergeConfig(configPath, table)
	require.NoError(t, err)

	// The backup should have a timestamp suffix.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			require.NoError(t, err)
			assert.Equal(t, original, string(data))
			found = true
			break
		}
	}
	assert.True(t, found, "a timestamped backup should exist")

	// The old-style predictable backup name must NOT be used.
	_, err = os.Stat(configPath + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "old-style predictable backup name must not be used")
}

func TestRemoveConfig_Present(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	existing := `[settings]
theme = "dark"

[sequoia]
skills_path = "/path"
`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0o644))

	err := codex.RemoveConfig(configPath)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "[settings]")
	assert.NotContains(t, content, "[sequoia]")
}

func TestRemoveConfig_Absent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	existing := `[settings]
theme = "dark"
`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0o644))

	err := codex.RemoveConfig(configPath)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, existing, string(data))
}

func TestRemoveConfig_MissingFile(t *testing.T) {
	t.Parallel()
	configPath := filepath.Join(t.TempDir(), "nonexistent.toml")
	err := codex.RemoveConfig(configPath)
	assert.NoError(t, err)
}

func TestRemoveConfig_RestoresBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	original := "[settings]\ntheme = \"dark\"\n"

	// Simulate a session-tracked backup from MergeConfig.
	// Create a timestamped backup and the session file.
	backupPath := configPath + ".sequoia-backup-test123"
	require.NoError(t, os.WriteFile(backupPath, []byte(original), 0o644))
	require.NoError(t, os.WriteFile(configPath+".sequoia-session", []byte("test123"), 0o644))
	require.NoError(t, os.WriteFile(configPath, []byte("[sequoia]\nskills_path = \"/path\"\n"), 0o644))

	err := codex.RemoveConfig(configPath)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(data))

	// Backup and session files should be removed.
	_, err = os.Stat(backupPath)
	assert.True(t, os.IsNotExist(err), "backup file should be removed")
	_, err = os.Stat(configPath + ".sequoia-session")
	assert.True(t, os.IsNotExist(err), "session file should be removed")
}

func TestInstall_And_Uninstall_RoundTrip(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	// Create .codex dir (simulating existing codex installation).
	codexDir := filepath.Join(tmp, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	a := codex.NewAdapter(tmp)

	// Fresh install.
	err := a.Install(adapters.InstallOpts{})
	require.NoError(t, err)
	assert.True(t, a.IsInstalled())

	// Verify files exist.
	assert.DirExists(t, filepath.Join(codexDir, "sequoia", "skills"))
	assert.DirExists(t, filepath.Join(codexDir, "sequoia", "commands"))

	configPath := filepath.Join(codexDir, "config.toml")
	assert.FileExists(t, configPath)

	configData, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(configData), "[sequoia]")

	// Verify version file.
	versionFile := filepath.Join(codexDir, "sequoia", "skills", ".sequoia-version")
	assert.FileExists(t, versionFile)

	// Status should show installed with version.
	status := a.Status()
	assert.True(t, status.Installed)
	assert.NotEmpty(t, status.Version)

	// Uninstall.
	err = a.Uninstall(adapters.InstallOpts{})
	require.NoError(t, err)
	assert.False(t, a.IsInstalled())

	// Verify cleanup.
	_, err = os.Stat(filepath.Join(codexDir, "sequoia"))
	assert.True(t, os.IsNotExist(err), "sequoia dir should be removed")

	configData, err = os.ReadFile(configPath)
	require.NoError(t, err)
	assert.NotContains(t, string(configData), "[sequoia]")
}

func TestInstall_Idempotent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	codexDir := filepath.Join(tmp, ".codex")
	require.NoError(t, os.MkdirAll(codexDir, 0o755))

	a := codex.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))
	require.True(t, a.IsInstalled())

	// Second install should succeed.
	require.NoError(t, a.Install(adapters.InstallOpts{}))
	assert.True(t, a.IsInstalled())
}

// =========================================================================
// Backup collision tests (FIX-005) for Codex installer
// =========================================================================

// TestMergeConfig_BackupHasUniqueName verifies that calling MergeConfig
// twice produces two different backup files instead of overwriting.
func TestMergeConfig_BackupHasUniqueName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	table := map[string]interface{}{
		"skills_path": "/path1",
	}

	// First call with existing file.
	require.NoError(t, os.WriteFile(configPath, []byte("[settings]\ntheme = \"dark\"\n"), 0o644))
	require.NoError(t, codex.MergeConfig(configPath, table))

	// Overwrite the config (simulating external restore) for second backup.
	require.NoError(t, os.WriteFile(configPath, []byte("[settings]\ntheme = \"light\"\n"), 0o644))
	require.NoError(t, codex.MergeConfig(configPath, table))

	// Count backup files with the sequoia-backup prefix.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	backupCount := 0
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			backupCount++
		}
	}

	assert.Equal(t, 2, backupCount, "two distinct backup files should exist, not one overwritten")
}

// TestMergeConfig_ExistingBackupNotOverwritten verifies that a pre-existing
// backup file is not touched when MergeConfig creates its own backup.
func TestMergeConfig_ExistingBackupNotOverwritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	// Pre-create a file that mimics an old backup.
	oldBackup := configPath + ".sequoia-backup-old"
	require.NoError(t, os.WriteFile(oldBackup, []byte("old backup content\n"), 0o644))

	require.NoError(t, os.WriteFile(configPath, []byte("[settings]\ntheme = \"dark\"\n"), 0o644))

	table := map[string]interface{}{
		"skills_path": "/path",
	}
	require.NoError(t, codex.MergeConfig(configPath, table))

	// The old backup must remain untouched.
	data, err := os.ReadFile(oldBackup)
	require.NoError(t, err)
	assert.Equal(t, "old backup content\n", string(data))

	// A new backup with timestamp suffix must exist.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	foundNewBackup := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			foundNewBackup = true
			break
		}
	}
	assert.True(t, foundNewBackup, "a new timestamped backup should be created")
}

// TestRemoveConfig_RestoresCorrectBackup verifies the full round-trip:
// MergeConfig creates a session-tracked backup, RemoveConfig restores from it.
func TestRemoveConfig_RestoresCorrectBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	original := "[settings]\ntheme = \"dark\"\n"
	require.NoError(t, os.WriteFile(configPath, []byte(original), 0o644))

	table := map[string]interface{}{
		"skills_path": "/path",
	}
	require.NoError(t, codex.MergeConfig(configPath, table))

	// Verify a timestamped backup was created.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	foundBackup := false
	foundSession := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			foundBackup = true
		}
		if e.Name() == "config.toml.sequoia-session" {
			foundSession = true
		}
	}
	assert.True(t, foundBackup, "a timestamped backup should exist")
	assert.True(t, foundSession, "a session tracking file should exist")

	// Uninstall — RemoveConfig restores from backup.
	require.NoError(t, codex.RemoveConfig(configPath))

	// Verify original is restored.
	restored, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(restored))

	// Backup and session files should be cleaned up.
	entries, err = os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.False(t, strings.Contains(e.Name(), ".sequoia-backup-"),
			"backup file %s should have been cleaned up", e.Name())
		assert.False(t, strings.HasSuffix(e.Name(), ".sequoia-session"),
			"session file should have been cleaned up")
	}
}
