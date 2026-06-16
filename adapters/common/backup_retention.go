package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DefaultMaxBackupsPerAdapter is the maximum number of backup session
// directories retained per adapter under the central backup root.
// Older sessions beyond this cap are pruned by PruneBackups.
// See REQ-BRP-04.
const DefaultMaxBackupsPerAdapter = 5

// backupHomeSubpath is the subpath appended to os.UserConfigDir() to
// produce the central backup root. It is also used as a sentinel substring
// in error wrapping so failed operations are diagnosable from the error
// string alone. See REQ-BRP-06.
const backupHomeSubpath = "sequoia/backups"

// userConfigDir returns the user's config directory. It is a package-level
// variable so tests can override it via the test hook in
// backup_retention_test.go (overrideUserConfigDir). Production callers MUST
// use BackupHomeDir() and MUST NOT touch this variable directly.
var userConfigDir = os.UserConfigDir

// BackupHomeDir returns the absolute central backup root, which is the
// join of os.UserConfigDir() with the literal "sequoia/backups" subpath.
// The directory is created with mode 0o700 on first use if it does not
// already exist. Calling BackupHomeDir() repeatedly is idempotent.
//
// Errors are wrapped with context that includes both the failing path
// and the "sequoia/backups" suffix so the failure mode is diagnosable
// from the error string alone. See REQ-BRP-01, REQ-BRP-06.
func BackupHomeDir() (string, error) {
	cfg, err := userConfigDir()
	if err != nil {
		return "", fmt.Errorf("backup home: resolve user config dir: %w", err)
	}

	home := backupRootFrom(cfg)
	if err := os.MkdirAll(home, 0o700); err != nil {
		return "", fmt.Errorf("backup home: create %q (%s): %w", home, backupHomeSubpath, err)
	}
	return home, nil
}

// backupRootFrom joins a user config dir with the central backup subpath.
// The subpath comes from the `backupHomeSubpath` constant so the path
// construction and the error-wrapping in BackupHomeDir share a single
// source of truth.
func backupRootFrom(cfg string) string {
	return filepath.Join(cfg, backupHomeSubpath)
}

// sessionDirLayout is the ISO-8601 UTC timestamp format used as the prefix
// of every session directory name produced by BackupPathBuilder.Build.
// See REQ-BRP-02.
const sessionDirLayout = "2006-01-02T15-04-05.000Z"

// PruneBackups keeps the `max` most-recent session directories for adapterID
// under the central backup root, removing the rest. "Most recent" is
// determined by the ISO-8601 timestamp prefix in the directory name —
// lexicographic order matches chronological order for the fixed-width
// prefix, so a descending lex sort yields the newest sessions first.
//
// PruneBackups returns (0, nil) when the adapter directory does not exist
// (treated as a no-op, not an error). If `max` is zero or negative, every
// session directory is removed.
//
// On per-entry removal failure, PruneBackups continues attempting the rest
// of the to-be-removed set and returns the first error encountered, with
// `removed` counting only successful removals. See REQ-BRP-04, REQ-BRP-06,
// REQ-BRP-07.
func PruneBackups(adapterID string, max int) (removed int, err error) {
	home, err := BackupHomeDir()
	if err != nil {
		return 0, err
	}
	adapterDir := filepath.Join(home, adapterID)

	entries, readErr := os.ReadDir(adapterDir)
	if os.IsNotExist(readErr) {
		return 0, nil
	}
	if readErr != nil {
		return 0, fmt.Errorf("prune backups: read %q: %w", adapterDir, readErr)
	}

	// Filter to subdirectories whose name starts with a valid ISO-8601
	// timestamp prefix; non-timestamp entries (corrupt names) are silently
	// ignored so a stray file or junk dir cannot abort the prune.
	valid := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !hasSessionPrefix(name) {
			continue
		}
		valid = append(valid, name)
	}

	// Newest first.
	sort.Sort(sort.Reverse(sort.StringSlice(valid)))

	if max < 0 {
		max = 0
	}
	if len(valid) <= max {
		return 0, nil
	}

	// The tail of `valid` (after the first `max` entries) is the
	// to-be-removed set, sorted oldest-first.
	toRemove := valid[max:]
	for _, name := range toRemove {
		full := filepath.Join(adapterDir, name)
		if rmErr := os.RemoveAll(full); rmErr != nil {
			if err == nil {
				err = fmt.Errorf("prune backups: remove %q: %w", full, rmErr)
			}
			continue
		}
		removed++
	}
	return removed, err
}

// hasSessionPrefix reports whether name starts with a parseable ISO-8601
// timestamp followed by '-' (the separator before the base-36 session
// suffix). Pure function — easy to test in isolation if needed.
func hasSessionPrefix(name string) bool {
	if len(name) < len(sessionDirLayout)+1 {
		return false
	}
	if name[len(sessionDirLayout)] != '-' {
		return false
	}
	_, err := time.Parse(sessionDirLayout, name[:len(sessionDirLayout)])
	return err == nil
}
