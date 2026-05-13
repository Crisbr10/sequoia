package common_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

const (
	testStart = "<!-- sequoia:start -->"
	testEnd   = "<!-- sequoia:end -->"
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
	require.NoError(t, common.ReplaceFile(p, content))

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
	require.NoError(t, common.ReplaceFile(p, newContent))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.Contains(t, got, "new content")
	assert.NotContains(t, got, "old content")

	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "no backup should be created when markers are present")
}

func TestReplaceFile_OtherContent_BacksUp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	require.NoError(t, os.WriteFile(p, []byte("# User config\n"), 0o644))

	content := sequoiaBody("sequoia rules")
	require.NoError(t, common.ReplaceFile(p, content))

	// The backup should have a timestamp suffix.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			found = true
			break
		}
	}
	assert.True(t, found, "a timestamped backup should exist")
}

func TestReplaceFile_OtherContent_BackupPreservesOriginal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# User config\nsome user rules\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia rules")))

	// Find the timestamped backup and verify it preserves the original.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	var backupPath string
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			backupPath = filepath.Join(dir, e.Name())
			break
		}
	}
	require.NotEmpty(t, backupPath, "a timestamped backup should exist")

	backupRaw, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(backupRaw))
}

func TestReplaceFile_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	content := sequoiaBody("same content")

	require.NoError(t, common.ReplaceFile(p, content))
	raw1, err := os.ReadFile(p)
	require.NoError(t, err)

	require.NoError(t, common.ReplaceFile(p, content))
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
	assert.NoError(t, common.RestoreOrRemoveFile(p))
}

func TestRestoreOrRemoveFile_WithBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# Original user config\n"
	require.NoError(t, os.WriteFile(p+".sequoia-backup", []byte(original), 0o644))
	require.NoError(t, os.WriteFile(p, []byte(sequoiaBody("sequoia")), 0o644))

	require.NoError(t, common.RestoreOrRemoveFile(p))

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

	require.NoError(t, common.RestoreOrRemoveFile(p))

	_, err := os.Stat(p)
	assert.True(t, os.IsNotExist(err), "file should be deleted")
}

func TestRestoreOrRemoveFile_NoBackup_NotManaged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# User config\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	require.NoError(t, common.RestoreOrRemoveFile(p))

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
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")

	// First call sets up a user-owned file.
	require.NoError(t, os.WriteFile(p, []byte("user content v1\n"), 0o644))
	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia v1")))

	// Second call — file is now Sequoia-managed (has markers), so no backup
	// is created. Instead, we simulate the case where a file is externally
	// modified between calls to trigger a second backup.
	// Write user content back (simulating external restore/modification).
	require.NoError(t, os.WriteFile(p, []byte("user content v2\n"), 0o644))
	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia v2")))

	// Count backup files with the sequoia-backup prefix.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	backupCount := 0
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			backupCount++
		}
	}

	assert.Equal(t, 2, backupCount, "two distinct backup files should exist, not one overwritten")

	// Both backups should contain different original content.
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			require.NoError(t, err)
			assert.True(t,
				strings.Contains(string(data), "user content v1") || strings.Contains(string(data), "user content v2"),
				"backup %s should contain original user content", e.Name(),
			)
		}
	}
}

// TestReplaceFile_ExistingBackupNotOverwritten verifies that a pre-existing
// backup file (with old naming or from a different session) is not touched
// when ReplaceFile creates its own timestamped backup.
func TestReplaceFile_ExistingBackupNotOverwritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")

	// Pre-create a file that mimics an old-style backup or a user file named
	// like a backup.
	oldBackup := p + ".sequoia-backup-old"
	require.NoError(t, os.WriteFile(oldBackup, []byte("old backup content\n"), 0o644))

	// User content for the target file.
	require.NoError(t, os.WriteFile(p, []byte("user content\n"), 0o644))

	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia")))

	// The old backup must remain untouched.
	data, err := os.ReadFile(oldBackup)
	require.NoError(t, err)
	assert.Equal(t, "old backup content\n", string(data))

	// A new backup with timestamp suffix must exist.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	foundNewBackup := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			foundNewBackup = true
			// The new backup should contain the user's original content.
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			require.NoError(t, err)
			assert.Equal(t, "user content\n", string(data))
			break
		}
	}
	assert.True(t, foundNewBackup, "a new timestamped backup should be created")
}

// TestReplaceFile_BackupPermissions_Restricted verifies that ReplaceFile
// backup files use owner-only permissions (backup-permissions spec).
// Skipped on Windows because unix permission bits are no-ops there.
func TestReplaceFile_BackupPermissions_Restricted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test requires Unix semantics — skipping on Windows")
	}
	t.Parallel()

	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")

	// First backup: user content v1.
	require.NoError(t, os.WriteFile(p, []byte("user content v1\n"), 0o644))
	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia v1")))

	// Second backup: user content v2 (simulate external modification).
	require.NoError(t, os.WriteFile(p, []byte("user content v2\n"), 0o644))
	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia v2")))

	// Both timestamped backups must have 0o600.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	backupCount := 0
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			fi, err := os.Stat(filepath.Join(dir, e.Name()))
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(),
				"ReplaceFile backup %s must be owner-only (0o600)", e.Name())
			backupCount++
		}
	}
	assert.Equal(t, 2, backupCount, "two backups expected for triangulation")
}

// TestRestoreOrRemoveFile_RestoresCorrectBackup verifies the full round-trip:
// ReplaceFile creates a timestamped backup with session tracking,
// RestoreOrRemoveFile restores from that exact backup and cleans up.
func TestRestoreOrRemoveFile_RestoresCorrectBackup(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")

	original := "# My custom rules\nThese are mine.\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	// Install — ReplaceFile backs up the original.
	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia rules")))

	// Verify the file now contains Sequoia content.
	content, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Contains(t, string(content), "sequoia rules")

	// Verify a timestamped backup was created (not the old predictable name).
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	foundTimestampedBackup := false
	foundSessionFile := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".sequoia-backup-") {
			foundTimestampedBackup = true
		}
		if e.Name() == "AGENTS.md.sequoia-session" {
			foundSessionFile = true
		}
	}
	assert.True(t, foundTimestampedBackup, "a timestamped backup should exist")
	assert.True(t, foundSessionFile, "a session tracking file should exist")
	// The old-style predictable backup name should NOT exist.
	_, err = os.Stat(p + ".sequoia-backup")
	assert.True(t, os.IsNotExist(err), "old-style predictable backup name must not be used")

	// Uninstall — RestoreOrRemoveFile restores from the correct backup.
	require.NoError(t, common.RestoreOrRemoveFile(p))

	// Verify original content is restored.
	restored, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(restored))

	// Backup file and session file should be removed.
	entries, err = os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.False(t, strings.Contains(e.Name(), ".sequoia-backup-"),
			"backup file %s should have been cleaned up", e.Name())
		assert.False(t, strings.HasSuffix(e.Name(), ".sequoia-session"),
			"session file should have been cleaned up")
	}
}

// TestRestoreOrRemoveFile_MultipleBackupsOnlyRestoresLatest verifies that
// when multiple backups exist (from multiple installs), only the session-
// tracked backup is restored.
func TestRestoreOrRemoveFile_MultipleBackupsOnlyRestoresLatest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")

	// First "session": old backup with old naming convention.
	oldBackup := p + ".sequoia-backup"
	require.NoError(t, os.WriteFile(oldBackup, []byte("old backup content\n"), 0o644))

	// Current session: write user content and call ReplaceFile.
	original := "current user content\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))
	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia")))

	// Restore — should use the session-tracked backup, not the old one.
	require.NoError(t, common.RestoreOrRemoveFile(p))

	// The restored content should be the current user content, not the old backup.
	restored, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Equal(t, original, string(restored))

	// The old backup file should still exist (wasn't from current session).
	_, err = os.Stat(oldBackup)
	assert.NoError(t, err, "old backup from different session should remain untouched")
}
