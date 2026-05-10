package gemini_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sequoia-ai/adapters/gemini"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testStart = "<!-- sequoia:start -->"
	testEnd   = "<!-- sequoia:end -->"
)

func tmpFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "GEMINI.md")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	return p
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	require.NoError(t, err)
	return string(b)
}

func TestInjectSection_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "subdir", "GEMINI.md")
	require.NoError(t, gemini.InjectSection(p, "hello sequoia\n"))

	got := readFile(t, p)
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "hello sequoia")
	assert.Contains(t, got, testEnd)
	// File must be exactly the section — no content outside the markers.
	stripped := strings.ReplaceAll(got, testStart, "")
	stripped = strings.ReplaceAll(stripped, testEnd, "")
	stripped = strings.ReplaceAll(stripped, "hello sequoia", "")
	assert.Equal(t, strings.TrimSpace(stripped), "")
}

func TestInjectSection_MarkersAbsent(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "existing content\n")
	require.NoError(t, gemini.InjectSection(p, "new section"))

	got := readFile(t, p)
	assert.Contains(t, got, "existing content")
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "new section")
	assert.Contains(t, got, testEnd)
	// Existing content must come before the marker.
	assert.Less(t, strings.Index(got, "existing content"), strings.Index(got, testStart))
}

func TestInjectSection_MarkersPresent(t *testing.T) {
	t.Parallel()
	initial := "# Header\n\n" + testStart + "\nold content\n" + testEnd + "\n"
	p := tmpFile(t, initial)
	require.NoError(t, gemini.InjectSection(p, "new content"))

	got := readFile(t, p)
	assert.Contains(t, got, "# Header")
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "new content")
	assert.Contains(t, got, testEnd)
	assert.NotContains(t, got, "old content")
}

func TestInjectSection_Idempotent(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "# Header\n")
	require.NoError(t, gemini.InjectSection(p, "my content"))
	first := readFile(t, p)
	require.NoError(t, gemini.InjectSection(p, "my content"))
	second := readFile(t, p)
	assert.Equal(t, first, second)
}

func TestInjectSection_PreservesExistingContent(t *testing.T) {
	t.Parallel()
	original := "# Existing\n\nSome important notes here.\n"
	p := tmpFile(t, original)
	require.NoError(t, gemini.InjectSection(p, "sequoia rules"))

	got := readFile(t, p)
	assert.Contains(t, got, "# Existing")
	assert.Contains(t, got, "Some important notes here.")
	// Inject again — original content must still be present.
	require.NoError(t, gemini.InjectSection(p, "sequoia rules updated"))
	got2 := readFile(t, p)
	assert.Contains(t, got2, "# Existing")
	assert.Contains(t, got2, "Some important notes here.")
	assert.Contains(t, got2, "sequoia rules updated")
}

func TestInjectSection_MultipleMarkerPairs(t *testing.T) {
	t.Parallel()
	initial := testStart + "\nfirst\n" + testEnd + "\n\n" + testStart + "\nsecond\n" + testEnd + "\n"
	p := tmpFile(t, initial)
	require.NoError(t, gemini.InjectSection(p, "replaced"))

	got := readFile(t, p)
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

func TestInjectSection_EmptyFile(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "")
	require.NoError(t, gemini.InjectSection(p, "sequoia content"))

	got := readFile(t, p)
	assert.Contains(t, got, testStart)
	assert.Contains(t, got, "sequoia content")
	assert.Contains(t, got, testEnd)
}

func TestRemoveSection_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "missing.md")
	assert.NoError(t, gemini.RemoveSection(p))
}

func TestRemoveSection_MarkersAbsent(t *testing.T) {
	t.Parallel()
	original := "# Config\n\nsome content\n"
	p := tmpFile(t, original)
	require.NoError(t, gemini.RemoveSection(p))
	assert.Equal(t, original, readFile(t, p))
}

func TestRemoveSection_MarkersPresent(t *testing.T) {
	t.Parallel()
	content := "# Header\n\nBefore content.\n\n" + testStart + "\nsequoia stuff\n" + testEnd + "\n\nAfter content.\n"
	p := tmpFile(t, content)
	require.NoError(t, gemini.RemoveSection(p))

	got := readFile(t, p)
	assert.NotContains(t, got, testStart)
	assert.NotContains(t, got, testEnd)
	assert.NotContains(t, got, "sequoia stuff")
	assert.Contains(t, got, "# Header")
	assert.Contains(t, got, "Before content.")
	assert.Contains(t, got, "After content.")
}

func TestRemoveSection_CleansBlanks(t *testing.T) {
	t.Parallel()
	content := "# Header\n\n" + testStart + "\nsequoia\n" + testEnd + "\n"
	p := tmpFile(t, content)
	require.NoError(t, gemini.RemoveSection(p))

	got := readFile(t, p)
	assert.NotContains(t, got, testStart)
	// No triple (or more) consecutive newlines.
	assert.NotContains(t, got, "\n\n\n")
}

func TestRemoveSection_OnlyMarkers(t *testing.T) {
	t.Parallel()
	content := testStart + "\nsequoia\n" + testEnd + "\n"
	p := tmpFile(t, content)
	require.NoError(t, gemini.RemoveSection(p))

	got := readFile(t, p)
	assert.Empty(t, got)
}
