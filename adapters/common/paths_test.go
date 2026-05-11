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
// returns a real path. On some platforms (macOS /var→/private/var,
// Windows short names) even TempDir paths resolve differently, so we
// verify the resolved path exists and is canonical (idempotent).
func TestResolveHome_NoSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	resolved, err := common.ResolveHome(dir)
	require.NoError(t, err)

	// ResolveHome must return a real, existing directory.
	assert.DirExists(t, resolved)

	// Double-resolving should be idempotent.
	resolved2, err := common.ResolveHome(resolved)
	require.NoError(t, err)
	assert.Equal(t, resolved, resolved2)
}

// TestResolveHome_Symlink verifies that resolving a symlinked path
// returns the real (target) path.
func TestResolveHome_Symlink(t *testing.T) {
	t.Parallel()

	realDir := t.TempDir()
	linkDir := filepath.Join(t.TempDir(), "link")

	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink creation not supported (Windows requires Developer Mode): %v", err)
	}

	resolved, err := common.ResolveHome(linkDir)
	require.NoError(t, err)

	// resolved should be the real directory, not the symlink.
	// Use EvalSymlinks on realDir too, since realDir itself may need
	// canonicalisation (e.g. macOS /var → /private/var).
	absReal, err := filepath.EvalSymlinks(realDir)
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
