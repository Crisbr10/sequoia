//nolint:gosec // test file: uses in-memory JSON round-trip fixtures, not production paths
package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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

// TestManifest_TopLevelCreatedAtRoundTrip verifies REQ-BRP-03 — the
// top-level created_at field of the manifest must round-trip through
// JSON serialization. The design's locked schema is
// {version, created_at, entries:[...]}; if a writer drops created_at,
// downstream tooling (audit, diff) loses the session timestamp.
func TestManifest_TopLevelCreatedAtRoundTrip(t *testing.T) {
	t.Parallel()

	original := newEmptyManifest()
	require.True(t, !original.CreatedAt.IsZero(),
		"newEmptyManifest must stamp a non-zero session creation time")
	original.Entries = []manifestEntry{
		{
			Version:      "1",
			OriginalPath: "/home/alice/.config/opencode/AGENTS.md",
			Suffix:       "k2j9m4",
			CreatedAt:    time.Date(2026, 6, 15, 15, 30, 45, 0, time.UTC),
			AdapterID:    "opencode",
		},
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err, "marshal manifest")

	// The serialized form MUST contain the top-level "created_at" key.
	// We assert this by string check (cheap) so a regression that drops
	// the JSON tag surfaces clearly, not via a field-equals check that
	// could pass if both sides are zero.
	require.Contains(t, string(raw), `"created_at"`,
		"manifest JSON must include the top-level created_at field per REQ-BRP-03 / design schema")

	var decoded manifest
	require.NoError(t, json.Unmarshal(raw, &decoded), "unmarshal manifest")
	assert.True(t, original.CreatedAt.Equal(decoded.CreatedAt),
		"top-level created_at must round-trip (got %s, want %s)",
		decoded.CreatedAt, original.CreatedAt)
	assert.Equal(t, "1", decoded.Version,
		"version must round-trip alongside created_at")
	require.Len(t, decoded.Entries, 1,
		"entries must round-trip alongside created_at")
	assert.Equal(t, original.Entries[0], decoded.Entries[0],
		"entry payload must round-trip unchanged")
}

// TestManifest_ReadLegacyMissingCreatedAt verifies backward
// compatibility: a manifest.json written before the top-level
// created_at was added must still load via readManifest without
// error; the field is stamped with the read time as a sensible
// default so downstream code can rely on a non-zero CreatedAt.
//
// The test writes a manifest JSON literal that omits created_at,
// then calls readManifest and asserts CreatedAt is non-zero.
func TestManifest_ReadLegacyMissingCreatedAt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Hand-written legacy manifest: no top-level created_at.
	legacy := []byte(`{"version":"1","entries":[]}` + "\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "manifest.json"), legacy, 0o600))

	m, err := readManifest(dir)
	require.NoError(t, err, "readManifest must tolerate legacy manifests without created_at")
	assert.False(t, m.CreatedAt.IsZero(),
		"readManifest must stamp a non-zero CreatedAt for legacy manifests so the in-memory representation is uniform")
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

// TestManifest_ReadCorruptReturnsError verifies that readManifest on
// a session dir that has a manifest.json with invalid JSON returns a
// non-nil error. This protects against a half-written or tampered
// manifest from silently producing an empty result.
func TestManifest_ReadCorruptReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write a clearly-invalid JSON document.
	require.NoError(t, os.WriteFile(filepath.Join(dir, manifestFileName),
		[]byte("not-valid-json{{"), 0o600))

	_, err := readManifest(dir)
	require.Error(t, err, "readManifest must return an error for invalid JSON")
	assert.Contains(t, err.Error(), "manifest: parse",
		"error must wrap with the 'manifest: parse' context")
}

// TestWriteManifest_MkdirFailureReturnsError verifies that writeManifest
// returns a non-nil error when the parent of sessionDir is not
// creatable. We use a regular file as the parent so MkdirAll on its
// child fails on all platforms.
func TestWriteManifest_MkdirFailureReturnsError(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	blockingFile := filepath.Join(tmp, "blocker")
	require.NoError(t, os.WriteFile(blockingFile, []byte("not a dir"), 0o600))

	// Session dir would live under the blocker — MkdirAll fails.
	sessionDir := filepath.Join(blockingFile, "session")

	err := writeManifest(sessionDir, newEmptyManifest())
	require.Error(t, err, "writeManifest must return an error when the parent is not a directory")
	assert.Contains(t, err.Error(), "manifest: mkdir",
		"error must wrap with the 'manifest: mkdir' context")
}

// TestRemoveSessionDir_FailureReturnsError verifies that removeSessionDir
// returns a non-nil error when the underlying os.RemoveAll fails. We
// trigger this by making the session dir path a child of a regular
// file (so RemoveAll on a non-existent child under a file fails on
// POSIX). On Windows, os.RemoveAll on a missing path is a no-op
// (returns nil), so this test only runs on POSIX.
func TestRemoveSessionDir_FailureReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.RemoveAll on a missing path is a no-op on Windows; cannot exercise the error path here")
	}
	t.Parallel()

	tmp := t.TempDir()
	blockingFile := filepath.Join(tmp, "blocker")
	require.NoError(t, os.WriteFile(blockingFile, []byte("not a dir"), 0o600))

	// The session dir lives under the blocker — RemoveAll on a child
	// of a regular file fails on POSIX.
	sessionDir := filepath.Join(blockingFile, "session")

	err := removeSessionDir(sessionDir)
	require.Error(t, err, "removeSessionDir must return an error when the parent is not a directory")
	assert.Contains(t, err.Error(), "manifest: remove session dir",
		"error must wrap with the 'manifest: remove session dir' context")
}
