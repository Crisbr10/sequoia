package opencode_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Crisbr10/sequoia/adapters/opencode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	installerStart = "<!-- sequoia:start -->"
	installerEnd   = "<!-- sequoia:end -->"
)

func sequoiaContent(body string) string {
	return installerStart + "\n" + body + "\n" + installerEnd + "\n"
}

func TestGenerateAgentsMD_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "subdir", "AGENTS.md")
	content := sequoiaContent("hello sequoia")
	require.NoError(t, opencode.GenerateAgentsMD(p, content))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.Contains(t, got, installerStart)
	assert.Contains(t, got, installerEnd)
	assert.Contains(t, got, "hello sequoia")
}

func TestGenerateAgentsMD_MarkersPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.WriteFile(p, []byte(sequoiaContent("old content")), 0o644))

	newContent := sequoiaContent("new content")
	require.NoError(t, opencode.GenerateAgentsMD(p, newContent))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.Contains(t, got, "new content")
	assert.NotContains(t, got, "old content")

	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "no backup should be created when markers are present")
}

func TestGenerateAgentsMD_OtherContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.WriteFile(p, []byte("# User config\n"), 0o644))

	content := sequoiaContent("sequoia rules")
	require.NoError(t, opencode.GenerateAgentsMD(p, content))

	_, err := os.Stat(p + ".sequoia-backup")
	require.NoError(t, err, "backup should exist")

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Contains(t, string(raw), installerStart)
}

func TestGenerateAgentsMD_OtherContent_BackupPreservesOriginal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# User config\nsome user rules\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	require.NoError(t, opencode.GenerateAgentsMD(p, sequoiaContent("sequoia rules")))

	backupRaw, err := os.ReadFile(p + ".sequoia-backup")
	require.NoError(t, err)
	assert.Equal(t, original, string(backupRaw))
}

func TestGenerateAgentsMD_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	content := sequoiaContent("same content")

	require.NoError(t, opencode.GenerateAgentsMD(p, content))
	raw1, err := os.ReadFile(p)
	require.NoError(t, err)

	require.NoError(t, opencode.GenerateAgentsMD(p, content))
	raw2, err := os.ReadFile(p)
	require.NoError(t, err)

	assert.Equal(t, string(raw1), string(raw2))
}

func TestRemoveAgentsMD_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "AGENTS.md")
	assert.NoError(t, opencode.RemoveAgentsMD(p))
}

func TestRemoveAgentsMD_WithBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# Original user config\n"
	require.NoError(t, os.WriteFile(p+".sequoia-backup", []byte(original), 0o644))
	require.NoError(t, os.WriteFile(p, []byte(sequoiaContent("sequoia")), 0o644))

	require.NoError(t, opencode.RemoveAgentsMD(p))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(raw))

	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "backup file should be removed")
}

func TestRemoveAgentsMD_WithoutBackup_MarkerPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.WriteFile(p, []byte(sequoiaContent("sequoia")), 0o644))

	require.NoError(t, opencode.RemoveAgentsMD(p))

	_, err := os.Stat(p)
	assert.True(t, os.IsNotExist(err), "AGENTS.md should be deleted")
}

func TestRemoveAgentsMD_NoMarkers_NoBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# User config\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	require.NoError(t, opencode.RemoveAgentsMD(p))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(raw))
}
