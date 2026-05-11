package common_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// TestResolveHome_NoSymlink verifies that resolving a non-symlinked path
// returns the path unchanged.
func TestResolveHome_NoSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	resolved, err := common.ResolveHome(dir)
	require.NoError(t, err)
	assert.Equal(t, dir, resolved)
}

// TestResolveHome_Symlink verifies that resolving a symlinked path
// returns the real (target) path.
func TestResolveHome_Symlink(t *testing.T) {
	t.Parallel()

	realDir := t.TempDir()
	linkDir := filepath.Join(t.TempDir(), "link")

	require.NoError(t, os.Symlink(realDir, linkDir))

	resolved, err := common.ResolveHome(linkDir)
	require.NoError(t, err)

	// resolved should be the real directory, not the symlink.
	absReal, err := filepath.Abs(realDir)
	require.NoError(t, err)
	assert.Equal(t, absReal, resolved)
	assert.NotEqual(t, linkDir, resolved)
}

// TestResolveHome_Error verifies that a non-existent path returns an error.
func TestResolveHome_Error(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "nonexistent")
	_, err := common.ResolveHome(dir)
	require.Error(t, err)
}
