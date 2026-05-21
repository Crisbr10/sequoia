// Package common_test provides black-box tests for files.go helpers.
package common_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// TestStageFileIfNotExist_createsWhenMissing verifies that StageFileIfNotExist
// writes the file with correct content when the target does not already exist.
func TestStageFileIfNotExist_createsWhenMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := common.StageFileIfNotExist(dir, "config.yaml", []byte("key: value"))
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "key: value", string(data))
}

// TestStageFileIfNotExist_preservesWhenExists verifies that StageFileIfNotExist
// is a no-op when the target file already exists, preserving the original content.
func TestStageFileIfNotExist_preservesWhenExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "config.yaml", "original content")

	err := common.StageFileIfNotExist(dir, "config.yaml", []byte("new content"))
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "original content", string(data),
		"existing file must not be overwritten")
}

// TestStageFileIfNotExist_createsParentDirs verifies that StageFileIfNotExist
// creates missing parent directories (mode 0o755) before writing the file.
func TestStageFileIfNotExist_createsParentDirs(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "nested", "deeper")
	err := common.StageFileIfNotExist(dir, "config.yaml", []byte("nested: true"))
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(dir, "config.yaml"))
}
