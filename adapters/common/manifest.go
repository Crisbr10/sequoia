package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// manifestFileName is the on-disk name of the per-session manifest that
// records every ReplaceFile backup under the central backup home.
// See REQ-BRP-03.
const manifestFileName = "manifest.json"

// manifestSchemaVersion is the current value of the Version field in
// the manifest. Bump it when the schema changes in a non-backwards-
// compatible way. See REQ-BRP-03 (single source of truth for schema).
const manifestSchemaVersion = "1"

// manifestEntry is one record in a session's manifest.json. It binds
// the original (pre-replace) path to the suffix that was used for the
// backup file inside the same session directory. See REQ-BRP-03.
//
// JSON tags are stdlib (encoding/json) compatible and the only
// contract between the writer (ReplaceFile) and the reader
// (RestoreOrRemoveFile).
type manifestEntry struct {
	Version      string    `json:"version"`
	OriginalPath string    `json:"original_path"`
	Suffix       string    `json:"suffix"`
	CreatedAt    time.Time `json:"created_at"`
	AdapterID    string    `json:"adapter_id"`
}

// manifest is the per-session document stored at
// <sessionDir>/manifest.json. It lists every ReplaceFile backup the
// session performed, plus the schema version. See REQ-BRP-03.
type manifest struct {
	Version string          `json:"version"`
	Entries []manifestEntry `json:"entries"`
}

// newEmptyManifest returns a fresh manifest with the current schema
// version and an empty (but non-nil) entry slice. The slice is non-nil
// so the JSON encoder always emits `"entries": []` instead of
// `"entries": null` for the empty case.
func newEmptyManifest() manifest {
	return manifest{
		Version: manifestSchemaVersion,
		Entries: []manifestEntry{},
	}
}

// readManifest returns the manifest for the session directory at
// sessionDir. A missing manifest.json is treated as an empty manifest
// (not an error) so the first ReplaceFile in a brand-new session can
// append without having to pre-create the file.
//
// Read errors other than "not exist" are wrapped with context that
// includes the manifest path and the "manifest.json" filename for
// diagnosability.
func readManifest(sessionDir string) (manifest, error) {
	mPath := filepath.Join(sessionDir, manifestFileName)
	raw, err := os.ReadFile(mPath) //nolint:gosec // G304: mPath is the session manifest, not user input
	if os.IsNotExist(err) {
		return newEmptyManifest(), nil
	}
	if err != nil {
		return manifest{}, fmt.Errorf("manifest: read %q: %w", mPath, err)
	}

	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return manifest{}, fmt.Errorf("manifest: parse %q: %w", mPath, err)
	}
	if m.Entries == nil {
		m.Entries = []manifestEntry{}
	}
	if m.Version == "" {
		m.Version = manifestSchemaVersion
	}
	return m, nil
}

// writeManifest writes m to <sessionDir>/manifest.json atomically.
// Existing files are replaced. Errors include the manifest path and
// the "manifest.json" filename for diagnosability.
func writeManifest(sessionDir string, m manifest) error {
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		return fmt.Errorf("manifest: mkdir %q: %w", sessionDir, err)
	}

	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("manifest: marshal: %w", err)
	}

	mPath := filepath.Join(sessionDir, manifestFileName)
	if err := AtomicWriteFile(mPath, raw, 0o600); err != nil {
		return fmt.Errorf("manifest: write %q: %w", mPath, err)
	}
	return nil
}

// appendManifestEntry reads the current manifest at sessionDir (or
// starts fresh if none exists), appends entry, and writes the result
// back. The session dir is created if missing.
//
// This is the "smallest unit of work" for the ReplaceFile write path:
// one entry per backup. The function is intentionally narrow so the
// read/write contract is testable in isolation.
func appendManifestEntry(sessionDir string, entry manifestEntry) error {
	m, err := readManifest(sessionDir)
	if err != nil {
		return err
	}
	m.Entries = append(m.Entries, entry)
	return writeManifest(sessionDir, m)
}

// removeSessionDir removes a session directory and its manifest.
// Used by RestoreOrRemoveFile on success so the per-session directory
// does not accumulate (PruneBackups would clean it up later, but
// the spec says "the session directory is removed on success"
// — see REQ-BRP-03 Scenario 2).
func removeSessionDir(sessionDir string) error {
	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("manifest: remove session dir %q: %w", sessionDir, err)
	}
	return nil
}
