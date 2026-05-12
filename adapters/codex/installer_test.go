package codex_test

import (
	"os"
	"path/filepath"
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

	backupPath := configPath + ".sequoia-backup"
	backupData, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(backupData))
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
	require.NoError(t, os.WriteFile(configPath+".sequoia-backup", []byte(original), 0o644))
	require.NoError(t, os.WriteFile(configPath, []byte("[sequoia]\nskills_path = \"/path\"\n"), 0o644))

	err := codex.RemoveConfig(configPath)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(data))

	_, err = os.Stat(configPath + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "backup file should be removed")
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
