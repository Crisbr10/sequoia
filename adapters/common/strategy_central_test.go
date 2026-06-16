//nolint:gosec // test file: uses t.TempDir() and overrideUserConfigDir hook, no real user paths
package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =========================================================================
// Test helpers for PR 3 (REQ-BRP-03) — central home + manifest round-trip
// =========================================================================

// adapterSessionDirs returns the session directory names under
// <root>/<adapterID>/. The names are sorted descending (most recent
// first), matching the PruneBackups sort. Used to assert that
// ReplaceFile created exactly one session dir per call. A missing
// adapter subdirectory is reported as an empty slice (no error) so
// tests can assert "no session was created" without pre-creating the
// adapter dir.
func adapterSessionDirs(t *testing.T, home, adapterID string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, adapterID))
	if os.IsNotExist(err) {
		return nil
	}
	require.NoError(t, err, "ReadDir on the adapter subdir must succeed")
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// readManifestJSON unmarshals <sessionDir>/manifest.json into a
// manifest struct. Fails the test on read/parse error.
func readManifestJSON(t *testing.T, sessionDir string) manifest {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(sessionDir, manifestFileName))
	require.NoError(t, err, "manifest.json must exist at <sessionDir>/manifest.json")
	var m manifest
	require.NoError(t, json.Unmarshal(raw, &m), "manifest.json must be valid JSON")
	return m
}

// =========================================================================
// Task 3.3 (RED) — ReplaceFile writes to central home + manifest
// =========================================================================

// TestReplaceFile_WritesToCentralHome_WithManifest verifies REQ-BRP-03
// Scenario 1: ReplaceFile writes the backup under the central home at
// <root>/<adapterID>/<session>/<basename>.backup, AND creates a
// manifest.json inside the same session directory that records the
// original_path and the adapter_id.
//
// The test redirects UserConfigDir to a temp dir so the real user
// config is never touched. The signature is the new PR 3 signature:
// ReplaceFile(adapterID, path, content).
//
// Not parallel: the userConfigDir override is package-level.
func TestReplaceFile_WritesToCentralHome_WithManifest(t *testing.T) {
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	home, err := BackupHomeDir()
	require.NoError(t, err, "BackupHomeDir() must succeed on a writable parent")

	// User-owned file (not Sequoia-managed) — ReplaceFile must back it up.
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "AGENTS.md")
	original := "# User config\nsome rules\n"
	require.NoError(t, os.WriteFile(targetPath, []byte(original), 0o644))

	const adapterID = "opencode"
	require.NoError(t, ReplaceFile(adapterID, targetPath, sequoiaBody("sequoia rules")))

	// 1. Exactly one session dir was created under the central home.
	sessions := adapterSessionDirs(t, home, adapterID)
	require.Len(t, sessions, 1,
		"ReplaceFile must create exactly one session dir under the central home (got %d)", len(sessions))
	sessionDir := filepath.Join(home, adapterID, sessions[0])

	// 2. The backup is at <sessionDir>/AGENTS.md.backup with byte-equal
	//    content to the original (REQ-BRP-03 Scenario 1).
	backupPath := filepath.Join(sessionDir, "AGENTS.md.backup")
	backupRaw, err := os.ReadFile(backupPath)
	require.NoError(t, err, "backup file must exist at <sessionDir>/<basename>.backup")
	assert.Equal(t, original, string(backupRaw),
		"backup must be byte-equal to the original user content")

	// 3. The target file is now the Sequoia content (the file was replaced).
	replaced, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, sequoiaBody("sequoia rules"), string(replaced),
		"target file must be replaced with the new content")

	// 4. A manifest.json exists in the session dir and records the entry.
	m := readManifestJSON(t, sessionDir)
	assert.Equal(t, manifestSchemaVersion, m.Version,
		"manifest must carry the current schema version")
	require.Len(t, m.Entries, 1,
		"manifest must hold exactly one entry after one ReplaceFile call")
	entry := m.Entries[0]
	assert.Equal(t, targetPath, entry.OriginalPath,
		"manifest entry must record the original (pre-replace) path")
	assert.NotEmpty(t, entry.Suffix,
		"manifest entry must have a non-empty suffix")
	assert.Equal(t, adapterID, entry.AdapterID,
		"manifest entry must record the adapter ID")
	assert.Equal(t, manifestSchemaVersion, entry.Version,
		"manifest entry must carry the schema version")
	assert.True(t, time.Since(entry.CreatedAt) < 5*time.Second,
		"manifest entry created_at must be recent (got %s)", entry.CreatedAt)
}

// TestReplaceFile_NoBackupWhenManaged verifies that when the target
// file is already Sequoia-managed (has the markers), ReplaceFile does
// NOT create a new backup or a new session dir. This is the existing
// "no backup on re-install" behavior, preserved under the central
// home. REQ-BRP-03.
//
// Not parallel: the userConfigDir override is package-level.
func TestReplaceFile_NoBackupWhenManaged(t *testing.T) {
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })
	home, err := BackupHomeDir()
	require.NoError(t, err)

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "AGENTS.md")
	require.NoError(t, os.WriteFile(targetPath, []byte(sequoiaBody("old sequoia content")), 0o644))

	const adapterID = "opencode"
	require.NoError(t, ReplaceFile(adapterID, targetPath, sequoiaBody("new sequoia content")))

	// No session dir should exist — the file is managed.
	sessions := adapterSessionDirs(t, home, adapterID)
	assert.Empty(t, sessions,
		"ReplaceFile on a managed file must NOT create a session dir (got %d)", len(sessions))

	// The target file is updated in place.
	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Contains(t, string(got), "new sequoia content",
		"managed file must be updated to the new content")
}

// TestReplaceFile_BackupPermissionsOwnerOnly verifies that the
// backup file is owner-only (0o600), matching the existing PR 1
// behavior. Skipped on Windows because POSIX permission bits are
// not enforced there. REQ-BRP-03 / REQ-BRP-04.
func TestReplaceFile_BackupPermissionsOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission test requires Unix semantics — skipping on Windows")
	}

	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })
	home, err := BackupHomeDir()
	require.NoError(t, err)

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "AGENTS.md")
	require.NoError(t, os.WriteFile(targetPath, []byte("# User content\n"), 0o644))

	const adapterID = "opencode"
	require.NoError(t, ReplaceFile(adapterID, targetPath, sequoiaBody("sequoia")))

	sessions := adapterSessionDirs(t, home, adapterID)
	require.Len(t, sessions, 1)
	backupPath := filepath.Join(home, adapterID, sessions[0], "AGENTS.md.backup")
	fi, err := os.Stat(backupPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(),
		"backup must be owner-only (0o600) — got %s", fi.Mode().Perm())
}

// TestReplaceFile_NoBackupWhenFileMissing verifies that ReplaceFile
// on a non-existent target file creates neither a backup nor a session
// dir. The target is created with the new content. This is the
// "clean install" path. REQ-BRP-03.
//
// Not parallel: the userConfigDir override is package-level.
func TestReplaceFile_NoBackupWhenFileMissing(t *testing.T) {
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })
	home, err := BackupHomeDir()
	require.NoError(t, err)

	targetPath := filepath.Join(t.TempDir(), "AGENTS.md")
	const adapterID = "opencode"
	require.NoError(t, ReplaceFile(adapterID, targetPath, sequoiaBody("sequoia")))

	sessions := adapterSessionDirs(t, home, adapterID)
	assert.Empty(t, sessions,
		"ReplaceFile on a missing file must NOT create a session dir")

	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Contains(t, string(got), "sequoia",
		"missing-file ReplaceFile must create the target with the new content")
}

// TestReplaceFile_TwoCallsProduceTwoSessions verifies that two
// ReplaceFile calls in sequence produce two distinct session dirs
// under the central home, each with its own backup + manifest. This
// is the "rapid consecutive install" scenario.
//
// Not parallel: the userConfigDir override is package-level.
func TestReplaceFile_TwoCallsProduceTwoSessions(t *testing.T) {
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })
	home, err := BackupHomeDir()
	require.NoError(t, err)

	const adapterID = "opencode"
	// Two different target paths (two different files) with two
	// different original contents.
	target1 := filepath.Join(t.TempDir(), "AGENTS.md")
	target2 := filepath.Join(t.TempDir(), "CLAUDE.md")
	require.NoError(t, os.WriteFile(target1, []byte("# user one\n"), 0o644))
	require.NoError(t, os.WriteFile(target2, []byte("# user two\n"), 0o644))

	require.NoError(t, ReplaceFile(adapterID, target1, sequoiaBody("sequoia one")))
	require.NoError(t, ReplaceFile(adapterID, target2, sequoiaBody("sequoia two")))

	sessions := adapterSessionDirs(t, home, adapterID)
	require.Len(t, sessions, 2,
		"two ReplaceFile calls must produce two session dirs (got %d: %v)", len(sessions), sessions)

	// Each session must have its own manifest with exactly one entry.
	for _, s := range sessions {
		m := readManifestJSON(t, filepath.Join(home, adapterID, s))
		require.Len(t, m.Entries, 1,
			"each session's manifest must hold exactly one entry (session %q)", s)
	}
}

// sequoiaBody is duplicated here (in package common) so the new
// internal tests don't need to import a helper from common_test.
func sequoiaBody(body string) string {
	return markerStart + "\n" + body + "\n" + markerEnd + "\n"
}

// =========================================================================
// Sanity: a fresh manifest on disk parses back to the same fields
// =========================================================================

// TestManifest_PersistedToDisk covers the "manifest.json actually
// landed on disk" assertion explicitly so a regression that writes
// the manifest to memory but not to disk is caught.
func TestManifest_PersistedToDisk(t *testing.T) {
	tmp := t.TempDir()
	sessionDir := filepath.Join(tmp, "session-1")
	require.NoError(t, appendManifestEntry(sessionDir, manifestEntry{
		Version:      "1",
		OriginalPath: "/abs/path",
		Suffix:       "z9",
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		AdapterID:    "x",
	}))

	// File must exist on disk.
	_, err := os.Stat(filepath.Join(sessionDir, manifestFileName))
	require.NoError(t, err, "manifest.json must exist on disk after append")

	// The JSON must contain the schema's exact field names so the
	// RestoreOrRemoveFile reader can find them.
	raw, err := os.ReadFile(filepath.Join(sessionDir, manifestFileName))
	require.NoError(t, err)
	for _, key := range []string{"version", "original_path", "suffix", "created_at", "adapter_id"} {
		assert.True(t, strings.Contains(string(raw), `"`+key+`"`),
			"manifest.json on disk must contain the JSON key %q (got: %s)", key, string(raw))
	}
}

// =========================================================================
// Task 3.5/3.6 — RestoreOrRemoveFile reads from central home via manifest
// =========================================================================

// TestRestoreOrRemoveFile_RestoresFromCentralHome verifies the
// manifest-based round-trip: after ReplaceFile writes a backup to
// the central home, RestoreOrRemoveFile finds the manifest entry
// matching the original path, restores the file byte-for-byte, and
// removes the session directory. REQ-BRP-03 Scenario 2.
//
// Not parallel: the userConfigDir override is package-level.
func TestRestoreOrRemoveFile_RestoresFromCentralHome(t *testing.T) {
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })
	home, err := BackupHomeDir()
	require.NoError(t, err)

	// Install: user content gets backed up to the central home.
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "AGENTS.md")
	original := "# My custom rules\nThese are mine.\n"
	require.NoError(t, os.WriteFile(targetPath, []byte(original), 0o644))

	const adapterID = "opencode"
	require.NoError(t, ReplaceFile(adapterID, targetPath, sequoiaBody("sequoia rules")))

	// Sanity: target was replaced and the session dir exists.
	sessions := adapterSessionDirs(t, home, adapterID)
	require.Len(t, sessions, 1, "ReplaceFile must have created exactly one session")
	sessionDir := filepath.Join(home, adapterID, sessions[0])

	// Uninstall: RestoreOrRemoveFile must restore from the manifest.
	require.NoError(t, RestoreOrRemoveFile(adapterID, targetPath))

	// Target is restored byte-for-byte.
	restored, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(restored),
		"target must be restored to the original content")

	// Session directory is removed (spec: "the session directory is
	// removed on success").
	_, statErr := os.Stat(sessionDir)
	assert.True(t, os.IsNotExist(statErr),
		"session directory must be removed after successful restore (got statErr=%v)", statErr)
}

// TestRestoreOrRemoveFile_NoManifestEntryNoOp verifies that when
// there is no manifest entry for the target path AND no legacy
// per-tool sidecar AND the file is not Sequoia-managed, the function
// is a no-op (returns nil, file untouched). REQ-BRP-03.
//
// Not parallel: the userConfigDir override is package-level.
func TestRestoreOrRemoveFile_NoManifestEntryNoOp(t *testing.T) {
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "AGENTS.md")
	original := "# User config\n"
	require.NoError(t, os.WriteFile(targetPath, []byte(original), 0o644))

	require.NoError(t, RestoreOrRemoveFile("opencode", targetPath),
		"RestoreOrRemoveFile must be a no-op when there is no backup to restore")

	got, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(got),
		"file must be untouched when no backup is available")
}

// TestRestoreOrRemoveFile_ManagedFileRemovedWhenNoBackup verifies
// that when there is no backup (neither manifest nor sidecar) and the
// file IS Sequoia-managed, the file is removed. This preserves the
// pre-PR-3 contract for the "reinstall" scenario.
//
// Not parallel: the userConfigDir override is package-level.
func TestRestoreOrRemoveFile_ManagedFileRemovedWhenNoBackup(t *testing.T) {
	tmp := t.TempDir()
	overrideUserConfigDir(t, func() (string, error) { return tmp, nil })

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "AGENTS.md")
	require.NoError(t, os.WriteFile(targetPath, []byte(sequoiaBody("sequoia content")), 0o644))

	require.NoError(t, RestoreOrRemoveFile("opencode", targetPath),
		"RestoreOrRemoveFile must remove a managed file when no backup is available")

	_, err := os.Stat(targetPath)
	assert.True(t, os.IsNotExist(err),
		"managed file must be removed when no backup is available")
}
