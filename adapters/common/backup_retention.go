package common

import (
	"fmt"
	"os"
	"path/filepath"
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
// It is a small pure helper extracted so the path-prefix construction
// has a single source of truth and is trivially testable.
func backupRootFrom(cfg string) string {
	return filepath.Join(cfg, "sequoia", "backups")
}
