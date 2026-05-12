package codex_test

import (
	"testing"

	"github.com/Crisbr10/sequoia/adapters/codex"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeSection_EmptyContent(t *testing.T) {
	t.Parallel()

	section := map[string]interface{}{
		"skills_path":   "/home/user/.codex/sequoia/skills/",
		"commands_path": "/home/user/.codex/sequoia/commands/",
	}

	result, err := codex.MergeSection("", section)
	require.NoError(t, err)
	assert.Contains(t, result, "[sequoia]")
	assert.Contains(t, result, "skills_path")
	assert.Contains(t, result, "commands_path")
}

func TestMergeSection_PreservesExistingKeys(t *testing.T) {
	t.Parallel()

	existing := `[settings]
theme = "dark"
editor = "vscode"

[models]
default = "gpt-4"
`

	section := map[string]interface{}{
		"skills_path": "/home/user/.codex/sequoia/skills/",
	}

	result, err := codex.MergeSection(existing, section)
	require.NoError(t, err)

	assert.Contains(t, result, "[settings]")
	assert.Contains(t, result, "theme")
	assert.Contains(t, result, "[models]")
	assert.Contains(t, result, "gpt-4")
	assert.Contains(t, result, "[sequoia]")
	assert.Contains(t, result, "skills_path")
}

func TestMergeSection_OverwriteExistingSequoia(t *testing.T) {
	t.Parallel()

	existing := `[settings]
theme = "dark"

[sequoia]
skills_path = "/old/path"
version = "0.0.1"
`

	section := map[string]interface{}{
		"skills_path":   "/new/path",
		"commands_path": "/new/commands",
	}

	result, err := codex.MergeSection(existing, section)
	require.NoError(t, err)

	assert.Contains(t, result, "[settings]")
	assert.Contains(t, result, "[sequoia]")
	assert.Contains(t, result, "/new/path")
	assert.Contains(t, result, "/new/commands")
	// Old sequoia keys should be gone.
	assert.NotContains(t, result, "version")
	assert.NotContains(t, result, "/old/path")
}

func TestMergeSection_InvalidTOML(t *testing.T) {
	t.Parallel()

	invalid := "this is not valid toml [[["
	section := map[string]interface{}{
		"skills_path": "/path",
	}

	_, err := codex.MergeSection(invalid, section)
	assert.Error(t, err, "invalid TOML should return an error")
}

func TestMergeSection_RoundTripIdempotent(t *testing.T) {
	t.Parallel()

	existing := `[settings]
theme = "dark"
`

	section := map[string]interface{}{
		"skills_path": "/home/user/.codex/sequoia/skills/",
	}

	result1, err := codex.MergeSection(existing, section)
	require.NoError(t, err)

	result2, err := codex.MergeSection(result1, section)
	require.NoError(t, err)

	// Second merge should not duplicate [sequoia].
	// Count occurrences of [sequoia].
	assert.Equal(t, stringOccurrences(result1, "[sequoia]"), stringOccurrences(result2, "[sequoia]"),
		"merge should be idempotent — no duplicate [sequoia] tables")
}

func TestRemoveSection_Present(t *testing.T) {
	t.Parallel()

	existing := `[settings]
theme = "dark"

[sequoia]
skills_path = "/path"
`

	result, err := codex.RemoveSection(existing)
	require.NoError(t, err)

	assert.NotContains(t, result, "[sequoia]")
	assert.NotContains(t, result, "skills_path")
	assert.Contains(t, result, "[settings]")
	assert.Contains(t, result, "theme")
}

func TestRemoveSection_Absent(t *testing.T) {
	t.Parallel()

	existing := `[settings]
theme = "dark"
`

	result, err := codex.RemoveSection(existing)
	require.NoError(t, err)

	assert.Contains(t, result, "[settings]")
	assert.Contains(t, result, "theme")
	assert.NotContains(t, result, "[sequoia]")
}

func TestRemoveSection_EmptyContent(t *testing.T) {
	t.Parallel()

	result, err := codex.RemoveSection("")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRemoveSection_InvalidTOML(t *testing.T) {
	t.Parallel()

	invalid := "this is not valid toml [[["
	_, err := codex.RemoveSection(invalid)
	assert.Error(t, err, "invalid TOML should return an error")
}

func TestRemoveSection_RoundTrip(t *testing.T) {
	t.Parallel()

	existing := `[settings]
theme = "dark"

[other]
key = "value"
`

	section := map[string]interface{}{
		"skills_path": "/path",
	}

	merged, err := codex.MergeSection(existing, section)
	require.NoError(t, err)

	removed, err := codex.RemoveSection(merged)
	require.NoError(t, err)

	// After merge+remove, the settings and other sections should still be preserved.
	assert.Contains(t, removed, "[settings]")
	assert.Contains(t, removed, "[other]")
	assert.NotContains(t, removed, "[sequoia]")
}

// stringOccurrences counts non-overlapping occurrences of substr in s.
func stringOccurrences(s, substr string) int {
	n := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			n++
			i += len(substr) - 1
		}
	}
	return n
}
