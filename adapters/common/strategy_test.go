package common_test

import (
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

	_, err := os.Stat(p + ".sequoia-backup")
	require.NoError(t, err, "backup should exist")
}

func TestReplaceFile_OtherContent_BackupPreservesOriginal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "AGENTS.md")
	original := "# User config\nsome user rules\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	require.NoError(t, common.ReplaceFile(p, sequoiaBody("sequoia rules")))

	backupRaw, err := os.ReadFile(p + ".sequoia-backup")
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
