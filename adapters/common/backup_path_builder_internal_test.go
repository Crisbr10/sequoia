//nolint:gosec // test file: uses t.TempDir() and controlled fixtures
package common

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sessionSuffixPattern matches the base-36 session suffix (digits + lowercase
// letters, length 8+ for UnixNano). It is used by the new Build() tests to
// assert the result ends in /<adapterID>/<ISO8601>-<base36>/.
var sessionSuffixPattern = regexp.MustCompile(`^[0-9a-z]+$`)

// TestBackupPathBuilder_Build_UsesCentralHome verifies REQ-BRP-02 and
// REQ-BRP-07: after PR 1, BackupPathBuilder.Build must produce paths of the
// form <BackupHomeDir>/<adapterID>/<ISO8601>-<sessionSuffix>/, no longer
// the per-tool backupPathFn result.
//
// Not parallel: this test mutates the package-level userConfigDir hook.
func TestBackupPathBuilder_Build_UsesCentralHome(t *testing.T) {
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	home, err := BackupHomeDir()
	require.NoError(t, err)

	// The custom backupPathFn is a safety-net fallback only — it is NOT
	// consulted on the happy path. Pass a sentinel value that would be
	// obvious if it leaked into the result.
	bp := NewBackupPathBuilder(
		func(base string) string { return "SENTINEL_FALLBACK_SHOULD_NOT_APPEAR" },
		"claude-code",
	)

	result := bp.Build("/tmp/irrelevant-base")

	// Path must start with the central home (joined with the test temp dir).
	expectedPrefix := home + string(filepath.Separator)
	assert.True(t, strings.HasPrefix(result, expectedPrefix),
		"Build() result must start with BackupHomeDir() (got %q, expected prefix %q)",
		result, expectedPrefix)
	assert.NotContains(t, result, "SENTINEL_FALLBACK_SHOULD_NOT_APPEAR",
		"Build() must not use the backupPathFn on the happy path")

	// Path must NOT contain any per-tool marker (the old format was
	// {backupPathFn(base)}-{adapterID}-{sessionSuffix}, with the base
	// appearing as a substring; the new format drops the base entirely).
	assert.NotContains(t, result, "irrelevant-base",
		"Build() result must not contain the per-tool base path")
	assert.NotContains(t, result, ".sequoia-backup",
		"Build() result must not contain the legacy per-tool marker")

	// Path must end with /claude-code/<ISO8601>-<base36>/ — the new
	// spec-mandated format. Split off the basename and assert shape.
	base := filepath.Base(result)
	parts := strings.Split(base, "-")
	require.GreaterOrEqual(t, len(parts), 2,
		"Build() session dir name must contain at least one '-' separating ISO8601 from suffix (got %q)", base)
	suffix := parts[len(parts)-1]
	assert.True(t, sessionSuffixPattern.MatchString(suffix),
		"the trailing token of the session dir name must be a base-36 suffix (got %q)", suffix)

	// Adapter ID must be a path segment in the result.
	assert.Contains(t, result, string(filepath.Separator)+"claude-code"+string(filepath.Separator),
		"Build() result must contain /<adapterID>/ as a path segment (got %q)", result)
}

// TestBackupPathBuilder_Build_DisjointForDifferentAdapters verifies that
// the new Build() implementation, like the old one, produces disjoint
// paths for different adapter IDs (REQ-BRP-02 Scenario "distinct adapter
// IDs produce disjoint subtrees").
//
// Not parallel: this test mutates the package-level userConfigDir hook.
func TestBackupPathBuilder_Build_DisjointForDifferentAdapters(t *testing.T) {
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	_, err := BackupHomeDir()
	require.NoError(t, err)

	bpClaude := NewBackupPathBuilder(func(base string) string { return "x" }, "claude-code")
	bpOpenCode := NewBackupPathBuilder(func(base string) string { return "x" }, "opencode")

	pathA := bpClaude.Build("/tmp/irrelevant")
	pathB := bpOpenCode.Build("/tmp/irrelevant")

	assert.NotEqual(t, pathA, pathB,
		"different adapter IDs must produce different paths")
	assert.False(t, strings.HasPrefix(pathA, pathB),
		"pathA must not be a child of pathB (got A=%q, B=%q)", pathA, pathB)
	assert.False(t, strings.HasPrefix(pathB, pathA),
		"pathB must not be a child of pathA (got A=%q, B=%q)", pathA, pathB)
	assert.Contains(t, pathA, "claude-code",
		"pathA must contain its adapter ID (got %q)", pathA)
	assert.Contains(t, pathB, "opencode",
		"pathB must contain its adapter ID (got %q)", pathB)
}

// TestBackupPathBuilder_Build_UniqueAcrossCalls verifies that two
// consecutive Build() calls produce different paths (session suffix is
// time-based and never collides for a single process).
//
// Not parallel: this test mutates the package-level userConfigDir hook.
func TestBackupPathBuilder_Build_UniqueAcrossCalls(t *testing.T) {
	tmp := t.TempDir()
	OverrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	_, err := BackupHomeDir()
	require.NoError(t, err)

	bp := NewBackupPathBuilder(func(base string) string { return "x" }, "test-adapter")

	first := bp.Build("/base")
	// Sleep to guarantee the underlying clock advances — UnixNano has
	// nanosecond resolution and consecutive calls within the same ns
	// would otherwise collide.
	time.Sleep(2 * time.Millisecond)
	second := bp.Build("/base")

	assert.NotEqual(t, first, second,
		"consecutive Build() calls must produce different paths")
}

// TestBackupPathBuilder_Build_SafetyNetSkipsBackupPathFn verifies
// PR 3 task 3.11: the safety-net fallback in Build() no longer calls
// the per-adapter `backupPathFn`. We point userConfigDir at a
// read-only file so BackupHomeDir() fails (MkdirAll on a child of a
// regular file fails on all platforms), then assert the returned
// fallback path does NOT contain the sentinel we put in `backupPathFn`.
//
// Not parallel: the userConfigDir override is package-level.
func TestBackupPathBuilder_Build_SafetyNetSkipsBackupPathFn(t *testing.T) {
	tmp := t.TempDir()
	// Create a regular file that will play the role of "UserConfigDir" —
	// joining <file>/sequoia/backups is not creatable because the parent
	// is a file. This forces BackupHomeDir() to fail.
	blockingFile := filepath.Join(tmp, "blocking")
	require.NoError(t, os.WriteFile(blockingFile, []byte("not a dir"), 0o600))

	OverrideUserConfigDir(t, func() (string, error) {
		return blockingFile, nil
	})

	const sentinel = "SENTINEL_FN_PATH_SAFETY_NET_SHOULD_NOT_CONSULT"
	bp := NewBackupPathBuilder(
		func(string) string { return sentinel },
		"claude-code",
	)

	result := bp.Build("/some/per-tool/base")

	// The result MUST use the hard-coded `<base>/.sequoia-backup/<adapterID>/<suffix>`
	// shape, NOT the per-adapter backupPathFn result. The sentinel must
	// not appear anywhere in the fallback path.
	assert.NotContains(t, result, sentinel,
		"safety-net fallback must NOT consult the per-adapter backupPathFn (got %q)", result)
	assert.Contains(t, result, string(filepath.Separator)+"claude-code"+string(filepath.Separator),
		"safety-net fallback must include the adapter ID as a path segment (got %q)", result)
	assert.Contains(t, result, ".sequoia-backup",
		"safety-net fallback must use the legacy per-tool marker (got %q)", result)
}
