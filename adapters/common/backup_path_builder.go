package common

import (
	"path/filepath"
	"strconv"
	"time"
)

// BackupPathBuilder generates unique backup directory paths for tool
// adapters. Each call to Build() produces a fresh session directory under
// the central backup root returned by BackupHomeDir().
//
// BackupPathBuilder is constructed once per adapter and injected into
// BaseAdapter via SetBackup().
//
// The legacy `backupPathFn` is retained as a safety-net fallback: if
// BackupHomeDir() cannot resolve the central root (e.g., the user's
// config directory is unwritable), Build() falls back to the per-tool
// path so installs can still proceed. The happy path never consults it.
type BackupPathBuilder struct {
	backupPathFn func(base string) string
	adapterID    string
}

// NewBackupPathBuilder creates a BackupPathBuilder with the given path
// function and adapter identifier.
//
// The path function is a legacy per-tool path constructor used as a
// fallback only. Production callers can pass any non-nil closure; the
// happy-path Build() ignores it.
func NewBackupPathBuilder(
	backupPathFn func(base string) string,
	adapterID string,
) *BackupPathBuilder {
	return &BackupPathBuilder{
		backupPathFn: backupPathFn,
		adapterID:    adapterID,
	}
}

// Build generates a unique backup session directory path of the form:
//
//	<BackupHomeDir>/<adapterID>/<ISO8601-UTC>-<base36-UnixNanos>/
//
// The ISO-8601 prefix sorts lexicographically the same as chronologically,
// so a descending lex sort yields the newest sessions first (used by
// PruneBackups). The base-36 nanos suffix guarantees uniqueness across
// rapid consecutive calls. See REQ-BRP-02, REQ-BRP-07.
//
// If BackupHomeDir() fails (e.g., the user's config dir is unwritable),
// Build() falls back to a per-tool path of the form
// `<base>/.sequoia-backup/<adapterID>/<sessionSuffix>` so the install
// does not abort. The fallback is best-effort and uses a hard-coded
// path shape (no longer consulting the per-adapter `backupPathFn`),
// because when BackupHomeDir() fails the per-adapter `backupPath` (if
// it tried to delegate to BackupHomeDir) would fail too. Keeping the
// fallback independent of BackupHomeDir makes it a true last resort.
func (b *BackupPathBuilder) Build(base string) string {
	now := time.Now()
	sessionSuffix := strconv.FormatInt(now.UnixNano(), 36)

	home, err := BackupHomeDir()
	if err != nil {
		// Safety-net fallback: independent of BackupHomeDir (which
		// already failed) and independent of the per-adapter backupPath
		// closure (which may itself delegate to BackupHomeDir). The
		// fallback shape is `<base>/.sequoia-backup/<adapterID>/<suffix>`
		// — same shape the legacy code path used to produce.
		return filepath.Join(base, ".sequoia-backup", b.adapterID, sessionSuffix)
	}

	isoPrefix := now.UTC().Format(sessionDirLayout)
	return filepath.Join(home, b.adapterID, isoPrefix+"-"+sessionSuffix)
}
