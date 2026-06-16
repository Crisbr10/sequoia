package common_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// cleanupCentralHome removes the per-adapter session dir created by the test
// from the real (or override) BackupHomeDir. Best-effort: errors are
// ignored because tests run with t.TempDir() / user-config fallbacks and
// the on-disk state is incidental.
func cleanupCentralHome(t *testing.T, adapterID string) {
	t.Helper()
	cfg, err := os.UserConfigDir()
	if err != nil {
		return
	}
	adapterDir := filepath.Join(cfg, "sequoia", "backups", adapterID)
	_ = os.RemoveAll(adapterDir)
}

// =========================================================================
// TestBackupPathBuilder_Build_IncludesAdapterID
// =========================================================================

// TestBackupPathBuilder_Build_IncludesAdapterID verifies that after PR 1,
// Build() produces paths of the form <BackupHomeDir>/<adapterID>/<...>,
// no longer the per-tool backupPathFn(base) result. REQ-BRP-02.
//
// The custom backupPathFn passed to NewBackupPathBuilder is a safety-net
// fallback only; on the happy path it is NOT consulted.
func TestBackupPathBuilder_Build_IncludesAdapterID(t *testing.T) {
	t.Cleanup(func() { cleanupCentralHome(t, "claude-code") })

	bp := common.NewBackupPathBuilder(
		func(base string) string { return "SENTINEL_SHOULD_NOT_APPEAR" },
		"claude-code",
	)

	result := bp.Build("/tmp/irrelevant-base")

	// Adapter ID is a path segment in the central home.
	assert.Contains(t, result,
		string(filepath.Separator)+"claude-code"+string(filepath.Separator),
		"Build() must contain /<adapterID>/ as a path segment (got %q)", result)
	// Per-tool markers from the OLD format must NOT appear.
	assert.NotContains(t, result, "irrelevant-base",
		"Build() must not contain the per-tool base path")
	assert.NotContains(t, result, "SENTINEL_SHOULD_NOT_APPEAR",
		"Build() must not invoke backupPathFn on the happy path")
	assert.NotContains(t, result, ".sequoia-backup",
		"Build() must not contain the legacy per-tool marker")
}

// =========================================================================
// TestBackupPathBuilder_Build_UsesBackupPathFn
// =========================================================================

// TestBackupPathBuilder_Build_UsesBackupPathFn triangulates with a
// different adapter ID to confirm the new format is used uniformly across
// adapters (REQ-BRP-02 Scenario "distinct adapter IDs produce disjoint
// subtrees").
func TestBackupPathBuilder_Build_UsesBackupPathFn(t *testing.T) {
	t.Cleanup(func() { cleanupCentralHome(t, "gemini-cli") })

	bp := common.NewBackupPathBuilder(
		func(base string) string { return "SENTINEL_SHOULD_NOT_APPEAR" },
		"gemini-cli",
	)

	result := bp.Build("/another/base")

	assert.Contains(t, result,
		string(filepath.Separator)+"gemini-cli"+string(filepath.Separator),
		"Build() must contain /<adapterID>/ as a path segment (got %q)", result)
	assert.NotContains(t, result, "another/base",
		"Build() must not contain the per-tool base path")
	assert.NotContains(t, result, "SENTINEL_SHOULD_NOT_APPEAR",
		"Build() must not invoke backupPathFn on the happy path")
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
