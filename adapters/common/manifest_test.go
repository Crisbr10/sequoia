//nolint:gosec // test file: uses in-memory JSON round-trip fixtures, not production paths
package common

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestManifestEntry_JSONRoundTrip verifies REQ-BRP-03 Scenario 1 — a
// manifestEntry marshals to JSON and back losslessly, preserving every
// field. The schema is:
//
//	{
//	  "version":       "1",
//	  "original_path": "<abs path>",
//	  "suffix":        "<base36 nanos>",
//	  "created_at":    "<ISO-8601 UTC>",
//	  "adapter_id":    "<adapter id>"
//	}
//
// The fields are exported via encoding/json tags so ReplaceFile and
// RestoreOrRemoveFile can write and read the manifest without exposing
// the internal struct layout.
func TestManifestEntry_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := manifestEntry{
		Version:      "1",
		OriginalPath: "/home/alice/.config/opencode/AGENTS.md",
		Suffix:       "k2j9m4abc",
		CreatedAt:    time.Date(2026, 6, 15, 15, 30, 45, 123_000_000, time.UTC),
		AdapterID:    "opencode",
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err, "manifestEntry must marshal cleanly")

	var decoded manifestEntry
	require.NoError(t, json.Unmarshal(raw, &decoded),
		"manifestEntry must unmarshal cleanly from its own JSON shape")

	assert.Equal(t, original.Version, decoded.Version,
		"version must round-trip")
	assert.Equal(t, original.OriginalPath, decoded.OriginalPath,
		"original_path must round-trip")
	assert.Equal(t, original.Suffix, decoded.Suffix,
		"suffix must round-trip")
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt),
		"created_at must round-trip (got %s, want %s)", decoded.CreatedAt, original.CreatedAt)
	assert.Equal(t, original.AdapterID, decoded.AdapterID,
		"adapter_id must round-trip")
}

// TestManifest_AppendAndRead verifies that appendManifestEntry on a
// freshly-written manifest (or on a new manifest that does not exist
// yet) produces a manifest with one entry, and that readManifest
// recovers it. The round-trip exercises the write path (manifest.json
// is created in the session dir) and the read path (manifest.json is
// loaded back into memory).
func TestManifest_AppendAndRead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Sanity: the session dir exists. writeManifest creates it if missing,
	// so we don't pre-create it here — the test exercises the create path.

	entry := manifestEntry{
		Version:      "1",
		OriginalPath: "/home/alice/.config/opencode/AGENTS.md",
		Suffix:       "k2j9m4",
		CreatedAt:    time.Date(2026, 6, 15, 15, 30, 45, 0, time.UTC),
		AdapterID:    "opencode",
	}

	require.NoError(t, appendManifestEntry(dir, entry),
		"appendManifestEntry must succeed on a fresh session dir")

	m, err := readManifest(dir)
	require.NoError(t, err, "readManifest must succeed after appendManifestEntry")
	assert.Equal(t, "1", m.Version,
		"manifest version must be initialized on first write")
	require.Len(t, m.Entries, 1, "manifest must hold exactly one entry after one append")
	assert.Equal(t, entry, m.Entries[0],
		"the appended entry must be the only one returned by readManifest")
}

// TestManifest_AppendPreservesExistingEntries verifies that a second
// appendManifestEntry call does not overwrite the first entry — both
// are present after the second call. This is the round-trip for
// multi-file sessions (even though the spec uses one entry per session
// for the system prompt, the data structure must support multiple).
func TestManifest_AppendPreservesExistingEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	first := manifestEntry{
		Version:      "1",
		OriginalPath: "/home/alice/.config/opencode/AGENTS.md",
		Suffix:       "k2j9m4",
		CreatedAt:    time.Date(2026, 6, 15, 15, 30, 45, 0, time.UTC),
		AdapterID:    "opencode",
	}
	second := manifestEntry{
		Version:      "1",
		OriginalPath: "/home/alice/.config/opencode/SKILL.md",
		Suffix:       "abcd",
		CreatedAt:    time.Date(2026, 6, 15, 15, 30, 46, 0, time.UTC),
		AdapterID:    "opencode",
	}

	require.NoError(t, appendManifestEntry(dir, first))
	require.NoError(t, appendManifestEntry(dir, second))

	m, err := readManifest(dir)
	require.NoError(t, err)
	require.Len(t, m.Entries, 2,
		"manifest must hold both entries after two appends")
	assert.Equal(t, first, m.Entries[0], "first entry must be preserved")
	assert.Equal(t, second, m.Entries[1], "second entry must follow first")
}

// TestManifest_ReadMissingReturnsEmpty verifies that readManifest on a
// session dir that does not yet have a manifest.json returns a fresh
// manifest (no error, no entries). This matches the "no manifest yet"
// case for the very first ReplaceFile in a new install — the caller
// must not have to pre-create the file.
func TestManifest_ReadMissingReturnsEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No manifest.json here.

	m, err := readManifest(dir)
	require.NoError(t, err, "readManifest must not error on a missing file")
	assert.Equal(t, "1", m.Version, "fresh manifest must have default version")
	assert.Empty(t, m.Entries, "fresh manifest must have zero entries")
}
