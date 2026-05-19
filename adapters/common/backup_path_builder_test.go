package common_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// =========================================================================
// TestBackupPathBuilder_Build_IncludesAdapterID
// =========================================================================

// TestBackupPathBuilder_Build_IncludesAdapterID verifies that Build() includes
// the adapter ID in the generated backup path, allowing different tool
// adapters to share the same base without backup collisions.
func TestBackupPathBuilder_Build_IncludesAdapterID(t *testing.T) {
	t.Parallel()

	bp := common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"claude-code",
	)

	result := bp.Build("/tmp/test-base")
	assert.Contains(t, result, "claude-code",
		"Build() should include the adapter ID in the backup path")
	assert.Contains(t, result, "test-base",
		"Build() should include the base directory path")
	assert.Contains(t, result, "backup",
		"Build() should delegate to backupPathFn for the base path")
}

// =========================================================================
// TestBackupPathBuilder_Build_UsesBackupPathFn
// =========================================================================

// TestBackupPathBuilder_Build_UsesBackupPathFn triangulates: verifies that
// Build() delegates to the backupPathFn with the provided base, using a
// different path function than the first test.
func TestBackupPathBuilder_Build_UsesBackupPathFn(t *testing.T) {
	t.Parallel()

	bp := common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "custom-backup") },
		"gemini-cli",
	)

	result := bp.Build("/another/base")
	assert.Contains(t, result, "custom-backup",
		"Build() should use the custom backupPathFn")
	assert.Contains(t, result, "gemini-cli",
		"Build() should include the adapter ID")
}

// =========================================================================
// TestBackupPathBuilder_Build_IncludesSessionSuffix
// =========================================================================

// TestBackupPathBuilder_Build_IncludesSessionSuffix verifies that Build()
// appends a session suffix to the backup path. The suffix is expected to
// be a base-36 encoded timestamp.
func TestBackupPathBuilder_Build_IncludesSessionSuffix(t *testing.T) {
	t.Parallel()

	bp := common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"test-adapter",
	)

	result := bp.Build("/base")

	// The result should have the pattern: /base/backup-test-adapter-SUFFIX
	// After the last "-" there should be a non-empty suffix.
	lastDash := strings.LastIndex(result, "-")
	assert.True(t, lastDash > 0, "Build() should include a '-' separator before the suffix")
	suffix := result[lastDash+1:]
	assert.NotEmpty(t, suffix, "Build() should have a non-empty session suffix")
}

// =========================================================================
// TestBackupPathBuilder_Build_DisjointPathsForDifferentAdapters
// REQ-BACKUP-ISOLATION-002 Scenario 1
// =========================================================================

// TestBackupPathBuilder_Build_DisjointPathsForDifferentAdapters verifies that
// two BackupPathBuilder instances with different adapter IDs produce disjoint
// backup paths, preventing backup collisions between concurrent adapter installs.
func TestBackupPathBuilder_Build_DisjointPathsForDifferentAdapters(t *testing.T) {
	t.Parallel()

	bpClaude := common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"claude",
	)
	bpOpenCode := common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"opencode",
	)

	resultA := bpClaude.Build("/tmp/home")
	resultB := bpOpenCode.Build("/tmp/home")

	// Both paths should include their respective adapter IDs.
	assert.Contains(t, resultA, "claude",
		"adapter A's backup path should contain its adapter ID")
	assert.Contains(t, resultB, "opencode",
		"adapter B's backup path should contain its adapter ID")

	// The two paths must be different.
	assert.NotEqual(t, resultA, resultB,
		"different adapters should produce different backup paths")

	// The backup directory trees must be entirely disjoint — neither path
	// is a prefix of the other, so os.RemoveAll on one cannot destroy the other.
	assert.False(t, strings.HasPrefix(resultA, resultB),
		"adapter A's backup path should not be a child of adapter B's path")
	assert.False(t, strings.HasPrefix(resultB, resultA),
		"adapter B's backup path should not be a child of adapter A's path")
}

// =========================================================================
// TestBackupPathBuilder_Build_ProducesUniquePaths
// =========================================================================

// TestBackupPathBuilder_Build_ProducesUniquePaths verifies that two
// consecutive calls to Build() produce different paths thanks to the
// time-based session suffix. This prevents backup directory collisions
// across multiple install attempts.
func TestBackupPathBuilder_Build_ProducesUniquePaths(t *testing.T) {
	t.Parallel()

	bp := common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"test-adapter",
	)

	first := bp.Build("/base")
	// Sleep to guarantee the millisecond clock advances between calls.
	time.Sleep(2 * time.Millisecond)
	second := bp.Build("/base")

	assert.NotEqual(t, first, second,
		"consecutive Build() calls should produce different paths (time-based suffix)")
}
