package cursor_test

import (
	"os"
	"path/filepath"
	"testing"

	"sequoia-ai/adapters/cursor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRulesMD_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "subdir", "sequoia-ai.md")
	content := "# Sequoia v0.1.0\n\nSequoia rules content.\n"
	require.NoError(t, cursor.GenerateRulesMD(p, content))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, content, string(raw))
}

func TestGenerateRulesMD_FileExists_SequoiaManaged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "sequoia-ai.md")
	oldContent := "<!-- sequoia:start -->\n# Sequoia v0.0.1\n\nOld sequoia content.\n<!-- sequoia:end -->\n"
	require.NoError(t, os.WriteFile(p, []byte(oldContent), 0o644))

	newContent := "# Sequoia v0.1.0\n\nNew sequoia content.\n"
	require.NoError(t, cursor.GenerateRulesMD(p, newContent))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, newContent, string(raw))

	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "no backup should be created when file is Sequoia-managed")
}

func TestGenerateRulesMD_OtherContent_BacksUp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "sequoia-ai.md")
	original := "# My custom Cursor rules\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	content := "# Sequoia v0.1.0\n"
	require.NoError(t, cursor.GenerateRulesMD(p, content))

	_, err := os.Stat(p + ".sequoia-backup")
	require.NoError(t, err, "backup should exist for non-Sequoia content")

	backupRaw, err := os.ReadFile(p + ".sequoia-backup")
	require.NoError(t, err)
	assert.Equal(t, original, string(backupRaw))
}

func TestGenerateRulesMD_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "sequoia-ai.md")
	content := "# Sequoia v0.1.0\n\nSame content.\n"

	require.NoError(t, cursor.GenerateRulesMD(p, content))
	raw1, err := os.ReadFile(p)
	require.NoError(t, err)

	require.NoError(t, cursor.GenerateRulesMD(p, content))
	raw2, err := os.ReadFile(p)
	require.NoError(t, err)

	assert.Equal(t, string(raw1), string(raw2))
}

func TestRemoveRulesMD_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "sequoia-ai.md")
	assert.NoError(t, cursor.RemoveRulesMD(p))
}

func TestRemoveRulesMD_WithBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "sequoia-ai.md")
	original := "# Original user rules\n"
	require.NoError(t, os.WriteFile(p+".sequoia-backup", []byte(original), 0o644))
	require.NoError(t, os.WriteFile(p, []byte("# Sequoia content\n"), 0o644))

	require.NoError(t, cursor.RemoveRulesMD(p))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(raw))

	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "backup file should be removed")
}

func TestRemoveRulesMD_NoBackup_Managed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "sequoia-ai.md")
	require.NoError(t, os.WriteFile(p, []byte("<!-- sequoia:start -->\n# Sequoia content\n<!-- sequoia:end -->\n"), 0o644))

	require.NoError(t, cursor.RemoveRulesMD(p))

	_, err := os.Stat(p)
	assert.True(t, os.IsNotExist(err), "sequoia-ai.md should be deleted")
}

func TestRemoveRulesMD_NoBackup_NotManaged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "sequoia-ai.md")
	original := "# My custom user rules\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	require.NoError(t, cursor.RemoveRulesMD(p))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(raw))
}
