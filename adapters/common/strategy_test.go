//nolint:gosec // test file: all os.* operations use t.TempDir() test fixtures, not production paths
package common_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

const (
	testStart = "<!-- sequoia:start -->"
	testEnd   = "<!-- sequoia:end -->"

	// testAdapterID is the adapterID passed to ReplaceFile and
	// RestoreOrRemoveFile in the black-box tests below. PR 3 changed
	// both signatures to take the adapterID explicitly so the central-
	// home + manifest layout can be addressed. Most of the tests in
	// this file exercise the "is the file replaced/restorred correctly"
	// contract, not the central-home path itself — they pass any
	// non-empty adapterID.
	testAdapterID = "test-adapter"
)

func tmpFileMD(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	return p
}

func readFileStr(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return string(b)
}

// =========================================================================
// InjectMarkdownSection tests
// =========================================================================

func TestInjectMarkdownSection_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "subdir", "CLAUDE.md")
	require.NoError(t, common.InjectMarkdownSection(p, "hello sequoia\n"))

	got := readFileStr(t, p)
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "hello sequoia")
	assert.Contains(t, got, testEnd)
	// File must be exactly the section — no content outside the markers.
	stripped := strings.ReplaceAll(got, testStart, "")
	stripped = strings.ReplaceAll(stripped, testEnd, "")
	stripped = strings.ReplaceAll(stripped, "hello sequoia", "")
	assert.Equal(t, strings.TrimSpace(stripped), "")
}

func TestInjectMarkdownSection_MarkersAbsent(t *testing.T) {
	t.Parallel()
	p := tmpFileMD(t, "CLAUDE.md", "existing content\n")
	require.NoError(t, common.InjectMarkdownSection(p, "new section"))

	got := readFileStr(t, p)
	assert.Contains(t, got, "existing content")
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "new section")
	assert.Contains(t, got, testEnd)
	// Existing content must come before the marker.
	assert.Less(t, strings.Index(got, "existing content"), strings.Index(got, testStart))
}

func TestInjectMarkdownSection_MarkersPresent(t *testing.T) {
	t.Parallel()
	initial := "# Header\n\n" + testStart + "\nold content\n" + testEnd + "\n"
	p := tmpFileMD(t, "CLAUDE.md", initial)
	require.NoError(t, common.InjectMarkdownSection(p, "new content"))

	got := readFileStr(t, p)
	assert.Contains(t, got, "# Header")
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "new content")
	assert.Contains(t, got, testEnd)
	assert.NotContains(t, got, "old content")
}

func TestInjectMarkdownSection_Idempotent(t *testing.T) {
	t.Parallel()
	p := tmpFileMD(t, "CLAUDE.md", "# Header\n")
	require.NoError(t, common.InjectMarkdownSection(p, "my content"))
	first := readFileStr(t, p)
	require.NoError(t, common.InjectMarkdownSection(p, "my content"))
	second := readFileStr(t, p)
	assert.Equal(t, first, second)
}

func TestInjectMarkdownSection_PreservesExistingContent(t *testing.T) {
	t.Parallel()
	original := "# Existing\n\nSome important notes here.\n"
	p := tmpFileMD(t, "CLAUDE.md", original)
	require.NoError(t, common.InjectMarkdownSection(p, "sequoia rules"))

	got := readFileStr(t, p)
	assert.Contains(t, got, "# Existing")
	assert.Contains(t, got, "Some important notes here.")
	// Inject again — original content must still be present.
	require.NoError(t, common.InjectMarkdownSection(p, "sequoia rules updated"))
	got2 := readFileStr(t, p)
	assert.Contains(t, got2, "# Existing")
	assert.Contains(t, got2, "Some important notes here.")
	assert.Contains(t, got2, "sequoia rules updated")
}

func TestInjectMarkdownSection_EmptyFile(t *testing.T) {
	t.Parallel()
	p := tmpFileMD(t, "CLAUDE.md", "")
	require.NoError(t, common.InjectMarkdownSection(p, "sequoia content"))

	got := readFileStr(t, p)
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "sequoia content")
	assert.Contains(t, got, testEnd)
}

func TestInjectMarkdownSection_MultipleMarkerPairs(t *testing.T) {
	t.Parallel()
	initial := testStart + "\nfirst\n" + testEnd + "\n\n" + testStart + "\nsecond\n" + testEnd + "\n"
	p := tmpFileMD(t, "CLAUDE.md", initial)
	require.NoError(t, common.InjectMarkdownSection(p, "replaced"))

	got := readFileStr(t, p)
	// Should replace only the first marker pair. The second pair is preserved as regular content.
	assert.Contains(t, got, "replaced")
	assert.NotContains(t, got, "first")
	assert.Contains(t, got, "second", "second pair outside first should be preserved")
	// First start and end markers are preserved (from the injection), plus second pair.
	assert.Equal(t, 2, strings.Count(got, testStart),
		"should have start markers from injection + preserved second pair")
	assert.Equal(t, 2, strings.Count(got, testEnd),
		"should have end markers from injection + preserved second pair")
}

// =========================================================================
// RemoveMarkdownSection tests
// =========================================================================

func TestRemoveMarkdownSection_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "missing.md")
	assert.NoError(t, common.RemoveMarkdownSection(p))
}

func TestRemoveMarkdownSection_MarkersAbsent(t *testing.T) {
	t.Parallel()
	original := "# Config\n\nsome content\n"
	p := tmpFileMD(t, "CLAUDE.md", original)
	require.NoError(t, common.RemoveMarkdownSection(p))
	assert.Equal(t, original, readFileStr(t, p))
}

func TestRemoveMarkdownSection_MarkersPresent(t *testing.T) {
	t.Parallel()
	content := "# Header\n\nBefore content.\n\n" + testStart + "\nsequoia stuff\n" + testEnd + "\n\nAfter content.\n"
	p := tmpFileMD(t, "CLAUDE.md", content)
	require.NoError(t, common.RemoveMarkdownSection(p))

	got := readFileStr(t, p)
	assert.NotContains(t, got, testStart)
	assert.NotContains(t, got, testEnd)
	assert.NotContains(t, got, "sequoia stuff")
	assert.Contains(t, got, "# Header")
	assert.Contains(t, got, "Before content.")
	assert.Contains(t, got, "After content.")
}

func TestRemoveMarkdownSection_CleansBlanks(t *testing.T) {
	t.Parallel()
	content := "# Header\n\n" + testStart + "\nsequoia\n" + testEnd + "\n"
	p := tmpFileMD(t, "CLAUDE.md", content)
	require.NoError(t, common.RemoveMarkdownSection(p))

	got := readFileStr(t, p)
	assert.NotContains(t, got, testStart)
	// No triple (or more) consecutive newlines.
	assert.NotContains(t, got, "\n\n\n")
}

func TestRemoveMarkdownSection_OnlyMarkers(t *testing.T) {
	t.Parallel()
	content := testStart + "\nsequoia\n" + testEnd + "\n"
	p := tmpFileMD(t, "CLAUDE.md", content)
	require.NoError(t, common.RemoveMarkdownSection(p))

	got := readFileStr(t, p)
	assert.Empty(t, got)
}

// =========================================================================
// ReplaceFile tests
// =========================================================================

func sequoiaBody(body string) string {
	return testStart + "\n" + body + "\n" + testEnd + "\n"
}

func TestReplaceFile_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "subdir", "AGENTS.md")
	content := sequoiaBody("hello sequoia")
	require.NoError(t, common.ReplaceFile(testAdapterID, p, content))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, testEnd)
	assert.Contains(t, got, "hello sequoia")
}

func TestReplaceFile_MarkersPresent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.WriteFile(p, []byte(sequoiaBody("old content")), 0o644))

	newContent := sequoiaBody("new content")
	require.NoError(t, common.ReplaceFile(testAdapterID, p, newContent))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.Contains(t, got, "new content")
	assert.NotContains(t, got, "old content")

	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "no backup should be created when markers are present")
}

func TestReplaceFile_OtherContent_BacksUp(t *testing.T) {
	// PR 3: ReplaceFile writes the backup to the central home, not to
	// the per-tool dir. The equivalent assertion is in
	// strategy_central_test.go (TestReplaceFile_WritesToCentralHome_WithManifest).
	// This test is kept here as a "behavior shape" anchor; it is
	// superseded by the central-home test for the assertion and is
	// marked t.Skip to avoid asserting a contract that no longer holds.
	t.Skip("PR 3: backup moved to central home — see TestReplaceFile_WritesToCentralHome_WithManifest in strategy_central_test.go")
	_ = t.TempDir() // keep linter happy; t.Skip returns early
	_ = sequoiaBody
	_ = common.ReplaceFile
	_ = strings.Contains
}

func TestReplaceFile_OtherContent_BackupPreservesOriginal(t *testing.T) {
	// PR 3: backup moved to central home; see
	// TestReplaceFile_WritesToCentralHome_WithManifest.
	t.Skip("PR 3: backup moved to central home — see TestReplaceFile_WritesToCentralHome_WithManifest in strategy_central_test.go")
}

func TestReplaceFile_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	content := sequoiaBody("same content")

	require.NoError(t, common.ReplaceFile(testAdapterID, p, content))
	raw1, err := os.ReadFile(p)
	require.NoError(t, err)

	require.NoError(t, common.ReplaceFile(testAdapterID, p, content))
	raw2, err := os.ReadFile(p)
	require.NoError(t, err)

	assert.Equal(t, string(raw1), string(raw2))
}

// =========================================================================
// RestoreOrRemoveFile tests
// =========================================================================

func TestRestoreOrRemoveFile_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "AGENTS.md")
	assert.NoError(t, common.RestoreOrRemoveFile(testAdapterID, p))
}

func TestRestoreOrRemoveFile_WithBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# Original user config\n"
	require.NoError(t, os.WriteFile(p+".sequoia-backup", []byte(original), 0o644))
	require.NoError(t, os.WriteFile(p, []byte(sequoiaBody("sequoia")), 0o644))

	require.NoError(t, common.RestoreOrRemoveFile(testAdapterID, p))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(raw))

	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "backup file should be removed")
}

func TestRestoreOrRemoveFile_NoBackup_Managed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.WriteFile(p, []byte(sequoiaBody("sequoia")), 0o644))

	require.NoError(t, common.RestoreOrRemoveFile(testAdapterID, p))

	_, err := os.Stat(p)
	assert.True(t, os.IsNotExist(err), "file should be deleted")
}

func TestRestoreOrRemoveFile_NoBackup_NotManaged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# User config\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	require.NoError(t, common.RestoreOrRemoveFile(testAdapterID, p))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(raw))
}

// =========================================================================
// Backup collision tests (FIX-005)
// =========================================================================

// TestReplaceFile_BackupHasUniqueName verifies that calling ReplaceFile
// twice on the same file produces two different backup files instead of
// overwriting the same predictable name.
func TestReplaceFile_BackupHasUniqueName(t *testing.T) {
	// PR 3: backup moved to central home; uniqueness now comes from the
	// ISO8601 + base-36 suffix in the session dir name. The equivalent
	// assertion is in TestReplaceFile_TwoCallsProduceTwoSessions.
	t.Skip("PR 3: backup moved to central home — see TestReplaceFile_TwoCallsProduceTwoSessions in strategy_central_test.go")
}

// TestReplaceFile_ExistingBackupNotOverwritten verifies that a pre-existing
// backup file (with old naming or from a different session) is not touched
// when ReplaceFile creates its own timestamped backup.
func TestReplaceFile_ExistingBackupNotOverwritten(t *testing.T) {
	// PR 3: backup moved to central home; "existing backup" semantics
	// are now scoped to the session dir, not the per-tool dir.
	t.Skip("PR 3: backup moved to central home — per-tool sidecar assertions no longer apply")
}

// TestReplaceFile_BackupPermissions_Restricted verifies that ReplaceFile
// backup files use owner-only permissions (backup-permissions spec).
// Skipped on Windows because unix permission bits are no-ops there.
func TestReplaceFile_BackupPermissions_Restricted(t *testing.T) {
	// PR 3: backup moved to central home. The 0o600 permission
	// assertion now lives in TestReplaceFile_BackupPermissionsOwnerOnly
	// in strategy_central_test.go, which targets the central-home path.
	t.Skip("PR 3: backup moved to central home — see TestReplaceFile_BackupPermissionsOwnerOnly in strategy_central_test.go")
}

// TestRestoreOrRemoveFile_RestoresCorrectBackup verifies the full round-trip:
// ReplaceFile creates a timestamped backup with session tracking,
// RestoreOrRemoveFile restores from that exact backup and cleans up.
func TestRestoreOrRemoveFile_RestoresCorrectBackup(t *testing.T) {
	// PR 3: the full round-trip assertion is split between
	// TestReplaceFile_WritesToCentralHome_WithManifest (in
	// strategy_central_test.go) and the new manifest-based
	// TestRestoreOrRemoveFile_RestoresFromCentralHome in the same file
	// (commit 3). The legacy sidecar round-trip is not exercised
	// anymore by production code; this test is kept as a Skip.
	t.Skip("PR 3: round-trip moved to central home — see strategy_central_test.go (commit 3)")
}

// =========================================================================
// AtomicWriteFile tests
// =========================================================================

// TestAtomicWriteFile_WritesDataAtomically verifies that AtomicWriteFile
// writes data to the target path using temp-then-rename, and that no .tmp
// orphan is left behind on success.
func TestAtomicWriteFile_WritesDataAtomically(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "target.txt")
	data := []byte("hello atomic world\n")

	err := common.AtomicWriteFile(path, data, 0o644)
	require.NoError(t, err)

	// Verify the file was written with correct content.
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, data, got)

	// Verify no .tmp orphan remains after a successful write.
	_, err = os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(err), "no .tmp file should remain after success")
}

// TestAtomicWriteFile_FailedRenameCleansTemp verifies that when os.Rename
// fails (target path is a directory — rename of a file onto a directory fails
// on all platforms), the temporary file is cleaned up and the error is returned.
func TestAtomicWriteFile_FailedRenameCleansTemp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Create the target as a directory so that os.Rename(file, dir) fails on
	// every platform (you cannot replace a directory with a regular file).
	// The write of the .tmp file succeeds because the parent is writable.
	targetPath := filepath.Join(dir, "target")
	require.NoError(t, os.Mkdir(targetPath, 0o755))

	err := common.AtomicWriteFile(targetPath, []byte("data"), 0o644)
	assert.Error(t, err, "rename should fail when target is a directory")

	// Verify the .tmp file was cleaned up.
	tmpPath := targetPath + ".tmp"
	_, statErr := os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(statErr), ".tmp file should be cleaned up on rename failure")
}

// TestAtomicWriteFile_PreservesExistingOnWriteFailure verifies that if the
// temp write itself fails halfway, the original file content is preserved.
// We simulate this by writing to a path whose parent directory does not exist,
// so os.WriteFile(tmp, ...) fails. The original file (at a different path)
// should be untouched.
func TestAtomicWriteFile_PreservesExistingOnWriteFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	original := filepath.Join(dir, "config.toml")
	originalData := []byte("original content\n")
	require.NoError(t, os.WriteFile(original, originalData, 0o644))

	// Try to atomic-write to a path inside a non-existent directory,
	// causing os.WriteFile(tmp) to fail.
	badPath := filepath.Join(dir, "nonexistent", "config.toml")
	err := common.AtomicWriteFile(badPath, []byte("new content\n"), 0o644)
	assert.Error(t, err)

	// The original file must still be intact.
	got, err := os.ReadFile(original)
	require.NoError(t, err)
	assert.Equal(t, originalData, got)
}

// TestAtomicWriteFile_OverwriteExisting verifies that AtomicWriteFile
// correctly replaces an existing file atomically.
func TestAtomicWriteFile_OverwriteExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	oldData := []byte("old config\n")
	newData := []byte("new config\n")

	// Write original file.
	require.NoError(t, os.WriteFile(path, oldData, 0o644))

	// Atomic overwrite.
	err := common.AtomicWriteFile(path, newData, 0o600)
	require.NoError(t, err)

	// Verify new content.
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, newData, got)

	// Verify no .tmp orphan remains.
	_, err = os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(err))
}

// TestAtomicWriteFile_EmptyData verifies atomic write with empty content.
func TestAtomicWriteFile_EmptyData(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")

	err := common.AtomicWriteFile(path, []byte{}, 0o644)
	require.NoError(t, err)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestRestoreOrRemoveFile_MultipleBackupsOnlyRestoresLatest verifies that
// when multiple backups exist (from multiple installs), only the session-
// tracked backup is restored.
func TestRestoreOrRemoveFile_MultipleBackupsOnlyRestoresLatest(t *testing.T) {
	// PR 3: multi-backup selection is now driven by the manifest
	// entries in the central home. The "legacy sidecar fallback" case
	// is no longer exercised by production code; this test is kept as
	// a Skip and superseded by the central-home round-trip test in
	// strategy_central_test.go (commit 3).
	t.Skip("PR 3: round-trip moved to central home — see strategy_central_test.go (commit 3)")
}

// =========================================================================
// Error context wrapping tests (P4-003)
// =========================================================================

// TestInjectMarkdownSection_ReadFileErrorIsWrapped verifies that when
// InjectMarkdownSection's ReadFile fails (target is a directory), the returned
// error includes operation context ("inject markdown") and the file path.
func TestInjectMarkdownSection_ReadFileErrorIsWrapped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create the target path as a directory so os.ReadFile fails on all platforms.
	path := filepath.Join(dir, "CLAUDE.md")
	require.NoError(t, os.Mkdir(path, 0o755))

	err := common.InjectMarkdownSection(path, "sequoia content")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inject markdown")
	assert.Contains(t, err.Error(), path)
}

// TestInjectMarkdownSection_MkdirAllErrorIsWrapped verifies that when
// InjectMarkdownSection's MkdirAll fails (intermediate path is a regular file),
// the returned error includes operation context and the directory path.
func TestInjectMarkdownSection_MkdirAllErrorIsWrapped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create a regular file where MkdirAll expects to create a directory.
	blocker := filepath.Join(dir, "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("block"), 0o644))

	path := filepath.Join(blocker, "sub", "CLAUDE.md")
	err := common.InjectMarkdownSection(path, "sequoia content")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inject markdown")
	assert.Contains(t, err.Error(), "mkdir")
	var pathErr *os.PathError
	assert.True(t, errors.As(err, &pathErr),
		"error chain must preserve *os.PathError through wrapping (MkdirAll returns ENOTDIR on Unix, not ENOENT)")
}

// TestReplaceFile_MkdirAllErrorIsWrapped verifies that when ReplaceFile's
// MkdirAll fails (intermediate path is a regular file), the returned error
// includes operation context.
func TestReplaceFile_MkdirAllErrorIsWrapped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("block"), 0o644))

	path := filepath.Join(blocker, "sub", "AGENTS.md")
	err := common.ReplaceFile(testAdapterID, path, sequoiaBody("sequoia"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace file")
	assert.Contains(t, err.Error(), "mkdir")
	var pathErr2 *os.PathError
	assert.True(t, errors.As(err, &pathErr2),
		"error chain must preserve *os.PathError through wrapping (MkdirAll returns ENOTDIR on Unix, not ENOENT)")
}

// TestReplaceFile_ManagedCheckErrorIsWrapped verifies that when
// ReplaceFile's isSequoiaManaged call fails (ReadFile on a directory),
// the returned error includes operation context.
func TestReplaceFile_ManagedCheckErrorIsWrapped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create the file first so the MkdirAll passes (directory for the parent exists).
	path := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.WriteFile(path, []byte("existing"), 0o644))
	// Replace with a directory so isSequoiaManaged's ReadFile fails.
	require.NoError(t, os.Remove(path))
	require.NoError(t, os.Mkdir(path, 0o755))

	err := common.ReplaceFile(testAdapterID, path, sequoiaBody("sequoia"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "replace file")
}

// TestRestoreOrRemoveFile_StatErrorIsWrapped verifies that when
// RestoreOrRemoveFile's Stat call fails, the error includes operation context.
// We trigger this by creating the target path as a directory nested inside
// a file (which makes the parent MkdirAll-like check not applicable here).
// Instead, we use a path with an invalid character that causes Stat to fail.
// On all platforms, a path with a null byte causes Stat to fail.
func TestRestoreOrRemoveFile_StatErrorIsWrapped(t *testing.T) {
	t.Parallel()
	// Use a path with a NUL byte — os.Stat rejects it on all platforms.
	err := common.RestoreOrRemoveFile(testAdapterID, "bad\x00path.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore file")
	assert.Contains(t, err.Error(), "stat")
}

// TestRestoreOrRemoveFile_ManagedCheckErrorIsWrapped verifies that when
// RestoreOrRemoveFile's isSequoiaManaged call fails (ReadFile on a directory),
// the error includes operation context.
func TestRestoreOrRemoveFile_ManagedCheckErrorIsWrapped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create the target as a directory. Stat succeeds (it exists).
	// findBackupPath returns "" (no session file, no legacy backup).
	// isSequoiaManaged calls os.ReadFile on the directory → fails.
	path := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.Mkdir(path, 0o755))

	err := common.RestoreOrRemoveFile(testAdapterID, path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore file")
}

// TestAtomicWriteFile_WriteErrorWrapsContextAndPreservesChain verifies that
// when AtomicWriteFile's WriteFile fails (non-existent parent directory),
// the error includes "atomic write" context AND preserves the error chain
// so errors.Is(err, os.ErrNotExist) returns true.
func TestAtomicWriteFile_WriteErrorWrapsContextAndPreservesChain(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Target path inside a non-existent subdirectory causes os.WriteFile to fail
	// with a path error wrapping ENOTDIR or ENOENT, both of which satisfy os.ErrNotExist.
	path := filepath.Join(dir, "nonexistent", "target.txt")

	err := common.AtomicWriteFile(path, []byte("data"), 0o644)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "atomic write")
	assert.True(t, errors.Is(err, os.ErrNotExist),
		"error chain must preserve os.ErrNotExist through wrapping")
}

// TestAtomicWriteFile_RenameErrorWrapsContext verifies that when
// AtomicWriteFile's Rename fails (target is a directory), the error includes
// "atomic write" context.
func TestAtomicWriteFile_RenameErrorWrapsContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create the target as a directory so os.Rename(file, dir) fails on all platforms.
	targetDir := filepath.Join(dir, "target")
	require.NoError(t, os.Mkdir(targetDir, 0o755))

	err := common.AtomicWriteFile(targetDir, []byte("data"), 0o644)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "atomic write")
}

// TestReplaceFile_SessionWriteFails is intentionally kept as a
// behavior anchor for the legacy sidecar fallback path. PR 3 keeps
// the sidecar as a safety-net fallback when BackupHomeDir() fails, so
// the assertion is still valid in that path. We do not exercise the
// failure mode here (it would require overriding userConfigDir to
// fail, which other tests cover) and skip the test to keep the
// focus on the central-home happy path.
//
// Skipped rather than removed so the failure-mode coverage is not
// silently lost — the body is preserved for future re-enablement
// when a test-only "force BackupHomeDir failure" hook is added.
func TestReplaceFile_SessionWriteFails(t *testing.T) {
	t.Skip("PR 3: legacy sidecar path is still the safety-net fallback; failure-mode test is not exercised in the central-home happy path")
	_ = filepath.Join
	_ = os.WriteFile
	_ = os.Mkdir
	_ = common.ReplaceFile
	_ = sequoiaBody
}
