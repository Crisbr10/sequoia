package common_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// TestResolveSymlink_ResolvableSymlink verifies that a real, resolvable
// symlink returns the resolved (target) path and no warning.
func TestResolveSymlink_ResolvableSymlink(t *testing.T) {
	t.Parallel()

	realDir := t.TempDir()
	linkDir := filepath.Join(t.TempDir(), "link-dir")

	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink creation not supported (Windows may require Developer Mode): %v", err)
	}

	resolved, warning := common.ResolveSymlink(linkDir)

	// Warning must be empty for a resolvable symlink.
	assert.Empty(t, warning, "resolvable symlink should not produce a warning")

	// Resolved path must equal the target, not the link itself.
	absReal, err := filepath.EvalSymlinks(realDir)
	require.NoError(t, err)
	assert.Equal(t, absReal, resolved, "resolved path should be the real directory")
	assert.NotEqual(t, linkDir, resolved, "resolved path should differ from the symlink path")
}

// TestResolveSymlink_BrokenSymlink verifies that a dangling symlink
// (target does not exist) returns the original path with a warning.
func TestResolveSymlink_BrokenSymlink(t *testing.T) {
	t.Parallel()

	nonexistentTarget := filepath.Join(t.TempDir(), "does-not-exist")
	linkDir := filepath.Join(t.TempDir(), "broken-link")

	if err := os.Symlink(nonexistentTarget, linkDir); err != nil {
		t.Skipf("symlink creation not supported (Windows may require Developer Mode): %v", err)
	}

	resolved, warning := common.ResolveSymlink(linkDir)

	// Must return the original (unresolved) path.
	assert.Equal(t, linkDir, resolved, "should return the original path when symlink cannot be resolved")

	// Must emit a warning containing the unresolved path.
	assert.NotEmpty(t, warning, "broken symlink should produce a warning")
	assert.Contains(t, warning, linkDir, "warning should contain the unresolved path")
}

// TestResolveSymlink_NormalDirectory verifies that a regular directory
// (not a symlink) returns the canonical path with no warning.
func TestResolveSymlink_NormalDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	resolved, warning := common.ResolveSymlink(dir)

	// No warning for a normal directory.
	assert.Empty(t, warning, "normal directory should not produce a warning")

	// Resolved should be the canonical path.
	absDir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	assert.Equal(t, absDir, resolved, "resolved path should be the canonical directory path")
}

// TestResolveSymlink_NonexistentPath verifies that a path that does not
// exist returns the original path and no warning (no false positives).
func TestResolveSymlink_NonexistentPath(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nonexistent-subdir")

	resolved, warning := common.ResolveSymlink(path)

	// Must return the original path when it does not exist.
	assert.Equal(t, path, resolved, "should return the original path for nonexistent paths")

	// Must NOT emit a warning for nonexistent paths (not a symlink, just doesn't exist).
	assert.Empty(t, warning, "nonexistent path should not produce a symlink warning")
}
