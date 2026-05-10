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

func TestStrategyConfigMerge_Inject_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "subdir", "GEMINI.md")
	s := gemini.NewStrategy(p)
	require.NoError(t, s.Inject("hello sequoia"))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.Contains(t, got, "<!-- sequoia:start -->")
	assert.Contains(t, got, "hello sequoia")
	assert.Contains(t, got, "<!-- sequoia:end -->")
}

func TestStrategyConfigMerge_Inject_MarkersAbsent(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "GEMINI.md")
	require.NoError(t, os.WriteFile(p, []byte("existing content\n"), 0o644))

	s := gemini.NewStrategy(p)
	require.NoError(t, s.Inject("new section"))

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.Contains(t, got, "existing content")
	assert.Contains(t, got, "new section")
}

func TestStrategyConfigMerge_Inject_Idempotent(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "GEMINI.md")
	s := gemini.NewStrategy(p)
	require.NoError(t, s.Inject("content"))

	raw1, err := os.ReadFile(p)
	require.NoError(t, err)
	first := string(raw1)

	require.NoError(t, s.Inject("content"))
	raw2, err := os.ReadFile(p)
	require.NoError(t, err)
	second := string(raw2)

	assert.Equal(t, first, second)
}

func TestStrategyConfigMerge_Remove_FileNotExist(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "missing.md")
	s := gemini.NewStrategy(p)
	assert.NoError(t, s.Remove())
}

func TestStrategyConfigMerge_Remove_MarkersPresent(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "GEMINI.md")
	content := "# Header\n\n" + testStart + "\nsequoia\n" + testEnd + "\n\nAfter\n"
	require.NoError(t, os.WriteFile(p, []byte(content), 0o644))

	s := gemini.NewStrategy(p)
	require.NoError(t, s.Remove())

	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)
	assert.NotContains(t, got, testStart)
	assert.NotContains(t, got, testEnd)
	assert.Contains(t, got, "# Header")
	assert.Contains(t, got, "After")
}

func TestStrategyConfigMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "GEMINI.md")
	original := "# My Config\n\nUser preferences here.\n"
	require.NoError(t, os.WriteFile(p, []byte(original), 0o644))

	s := gemini.NewStrategy(p)

	// Inject
	require.NoError(t, s.Inject("sequoia v0.1.0 rules"))
	raw, err := os.ReadFile(p)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "sequoia v0.1.0 rules")

	// Remove
	require.NoError(t, s.Remove())
	raw, err = os.ReadFile(p)
	require.NoError(t, err)
	got := string(raw)

	assert.NotContains(t, got, testStart)
	assert.NotContains(t, got, testEnd)
	assert.Equal(t, original, strings.TrimRight(got, "\n")+"\n")
}
