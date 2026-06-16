package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Crisbr10/sequoia/adapters"
)

// Strategy defines the phased install lifecycle for tool adapters.
// Each phase reports errors independently, allowing consumers to observe,
// cancel, or recover at phase boundaries.
//
// Implementations MUST support sequential phase calls:
//
//	Prepare → Download → Verify → Stage → Apply
//
// On any phase failure the consumer may call Rollback to undo completed work.
type Strategy interface {
	Prepare(opts adapters.InstallOpts) error
	Download(opts adapters.InstallOpts) error
	Verify() error
	Stage(opts adapters.InstallOpts) error
	Apply(opts adapters.InstallOpts) error
	Rollback() error
}

const (
	markerStart = "<!-- sequoia:start -->"
	markerEnd   = "<!-- sequoia:end -->"
)

// InjectMarkdownSection writes content into the Markdown file at path
// between <!-- sequoia:start --> and <!-- sequoia:end --> markers.
// If the file does not exist it is created with the section. If markers
// are already present the content between them is replaced. Otherwise
// the section is appended at the end of the file.
func InjectMarkdownSection(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("inject markdown: mkdir %s: %w", filepath.Dir(path), err)
	}

	section := markerStart + "\n" + strings.TrimRight(content, "\n") + "\n" + markerEnd + "\n"

	raw, err := os.ReadFile(path) //nolint:gosec // G304: path is a known adapter-controlled file, not user input
	if err != nil {
		if os.IsNotExist(err) {
			//nolint:gosec // G306: file is meant to be world-readable markdown config
			return os.WriteFile(path, []byte(section), 0o644)
		}
		return fmt.Errorf("inject markdown: read %s: %w", path, err)
	}

	s := string(raw)
	start := strings.Index(s, markerStart)
	end := strings.Index(s, markerEnd)

	if start != -1 && end != -1 {
		replaced := s[:start] + section + s[end+len(markerEnd):]
		// Trim a single trailing newline that WriteFile will re-add via section.
		replaced = strings.TrimRight(replaced, "\n") + "\n"
		//nolint:gosec // G306: file is meant to be world-readable markdown config
		return os.WriteFile(path, []byte(replaced), 0o644)
	}

	// Append: ensure exactly one blank line separator when existing content is non-empty.
	body := strings.TrimRight(s, "\n")
	var out string
	if body == "" {
		out = section
	} else {
		out = body + "\n\n" + section
	}
	//nolint:gosec // G306: file is meant to be world-readable markdown config
	return os.WriteFile(path, []byte(out), 0o644)
}

// RemoveMarkdownSection deletes the content between <!-- sequoia:start -->
// and <!-- sequoia:end --> markers from the file at path.
// Returns nil when the file does not exist or contains no markers.
func RemoveMarkdownSection(path string) error {
	raw, err := os.ReadFile(path) //nolint:gosec // G304: path is a known adapter-controlled file, not user input
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("remove markdown: read %s: %w", path, err)
	}

	s := string(raw)
	start := strings.Index(s, markerStart)
	end := strings.Index(s, markerEnd)
	if start == -1 || end == -1 {
		return nil
	}

	before := strings.TrimRight(s[:start], "\n")
	after := strings.TrimLeft(s[end+len(markerEnd):], "\n")

	var out string
	switch {
	case before == "" && after == "":
		out = ""
	case before == "":
		out = after
	case after == "":
		out = before + "\n"
	default:
		out = before + "\n\n" + after
	}

	//nolint:gosec // G306: file is meant to be world-readable markdown config
	return os.WriteFile(path, []byte(out), 0o644)
}

// ReplaceFile writes content to the file at path, creating a backup in
// the central backup home under <BackupHomeDir>/<adapterID>/<session>/
// if the file already exists and is not Sequoia-managed. A per-session
// manifest.json inside the same session directory records the
// original_path and the suffix used so RestoreOrRemoveFile can locate
// the correct backup during uninstall (REQ-BRP-03).
//
// Creates parent directories of the target file if needed.
//
// If BackupHomeDir() cannot resolve the central root (e.g., the user's
// config dir is unwritable), ReplaceFile falls back to the legacy
// per-tool sidecar (.sequoia-backup-<suffix> + .sequoia-session) so
// the install does not abort. The fallback is best-effort and
// documented; the happy path always uses the central home.
func ReplaceFile(adapterID, path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("replace file: mkdir %s: %w", filepath.Dir(path), err)
	}

	managed, err := isSequoiaManaged(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace file: read %s: %w", path, err)
	}

	if os.IsNotExist(err) {
		return AtomicWriteFile(path, []byte(content), 0o644)
	}

	if managed {
		return AtomicWriteFile(path, []byte(content), 0o644)
	}

	// Backup the original. Prefer the central-home + manifest layout;
	// fall back to the legacy per-tool sidecar when BackupHomeDir() fails.
	home, homeErr := BackupHomeDir()
	if homeErr != nil {
		return replaceFileLegacySidecar(path, content)
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	sessionDir := NewSessionDir(home, adapterID, suffix)
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		return fmt.Errorf("replace file: session mkdir %s: %w", sessionDir, err)
	}

	raw, err := os.ReadFile(path) //nolint:gosec // G304: path is a known adapter-controlled file, not user input
	if err != nil {
		return fmt.Errorf("replace file: backup read %s: %w", path, err)
	}

	backupName := filepath.Base(path) + ".backup"
	backupPath := filepath.Join(sessionDir, backupName)
	if err := AtomicWriteFile(backupPath, raw, 0o600); err != nil {
		return fmt.Errorf("replace file: backup write %s: %w", backupPath, err)
	}

	entry := manifestEntry{
		Version:      manifestSchemaVersion,
		OriginalPath: path,
		Suffix:       suffix,
		CreatedAt:    time.Now().UTC(),
		AdapterID:    adapterID,
	}
	if err := appendManifestEntry(sessionDir, entry); err != nil {
		return fmt.Errorf("replace file: manifest append: %w", err)
	}

	return AtomicWriteFile(path, []byte(content), 0o644)
}

// replaceFileLegacySidecar is the safety-net fallback for ReplaceFile
// when the central backup home is unavailable. It preserves the
// pre-PR-3 sidecar format (.sequoia-backup-<suffix> +
// .sequoia-session) at the target file's location. Kept private —
// callers MUST go through ReplaceFile.
func replaceFileLegacySidecar(path, content string) error {
	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	backup := path + ".sequoia-backup-" + suffix
	raw, err := os.ReadFile(path) //nolint:gosec // G304: path is a known adapter-controlled file, not user input
	if err != nil {
		return fmt.Errorf("replace file: backup read %s: %w", path, err)
	}
	if err := AtomicWriteFile(backup, raw, 0o600); err != nil {
		return fmt.Errorf("replace file: backup write %s: %w", backup, err)
	}
	if err := AtomicWriteFile(path+".sequoia-session", []byte(suffix), 0o644); err != nil {
		return fmt.Errorf("replace file: session: %w", err)
	}
	return AtomicWriteFile(path, []byte(content), 0o644)
}

// findManifestEntry is now FindManifestEntry in manifest.go (PR 3
// task 3.10 — consolidate manifest helpers). The strategy.go callers
// use the exported version.

// RestoreOrRemoveFile restores the original content of the file at
// path from a backup stored in the central backup home, or removes
// the file if it is Sequoia-managed. The adapterID identifies the
// per-adapter subdirectory under <BackupHomeDir>/<adapterID>/.
//
// RestoreOrRemoveFile scans every session directory under
// <BackupHomeDir>/<adapterID>/ for a manifest.json that contains an
// entry whose original_path matches path. The first match wins
// (sessions are scanned in directory-listing order; the newest
// session is the first candidate). The matching backup is restored
// byte-for-byte, the session directory is removed, and the
// originally targeted file is left with the pre-replace content.
// See REQ-BRP-03 Scenario 2.
//
// If no manifest entry matches (e.g., the user installed before
// the manifest format landed, or the file is unmanaged and has no
// backup), RestoreOrRemoveFile falls back to the legacy per-tool
// sidecar (path+".sequoia-session" + path+".sequoia-backup-<suffix>")
// for backwards compatibility. If the file is Sequoia-managed and
// no backup exists, it is removed. If the file does not exist or is
// not managed and has no backup, returns nil.
func RestoreOrRemoveFile(adapterID, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("restore file: stat %s: %w", path, err)
	}

	// Try the central-home + manifest restore first. If the home is
	// unavailable or no manifest entry matches, fall back to the
	// per-tool sidecar.
	if home, homeErr := BackupHomeDir(); homeErr == nil {
		entry, sessionDir, found := FindManifestEntry(home, adapterID, path)
		if found {
			backupName := filepath.Base(path) + ".backup"
			backupPath := filepath.Join(sessionDir, backupName)
			raw, readErr := os.ReadFile(backupPath) //nolint:gosec // G304: backupPath derived from session dir + basename
			if readErr != nil {
				return fmt.Errorf("restore file: backup read %s: %w", backupPath, readErr)
			}
			if err := AtomicWriteFile(path, raw, 0o644); err != nil {
				return fmt.Errorf("restore file: restore write %s: %w", path, err)
			}
			// Remove the session dir (and its manifest) — the spec says
			// the session directory is removed on successful restore.
			_ = removeSessionDir(sessionDir)
			_ = entry
			return nil
		}
	}

	// Fall back to the legacy per-tool sidecar.
	backup := findBackupPath(path)

	if backup != "" {
		raw, err := os.ReadFile(backup) //nolint:gosec // G304: backup path derived from controlled suffix, not user input
		if err != nil {
			return fmt.Errorf("restore file: backup read %s: %w", backup, err)
		}
		if err := AtomicWriteFile(path, raw, 0o644); err != nil {
			return fmt.Errorf("restore file: restore write %s: %w", path, err)
		}
		// Clean up the backup file.
		_ = os.Remove(backup)
		// Clean up the session file if it exists.
		_ = os.Remove(path + ".sequoia-session")
		return nil
	}

	managed, err := isSequoiaManaged(path)
	if err != nil {
		return fmt.Errorf("restore file: read %s: %w", path, err)
	}
	if managed {
		_ = os.Remove(path + ".sequoia-session")
		return os.Remove(path)
	}

	return nil
}

// findBackupPath returns the path of the backup to restore for the given file.
// It first checks for a .sequoia-session file with the backup suffix.
// If not found, it falls back to the legacy predictable backup name.
func findBackupPath(path string) string {
	// Try session-tracked backup first.
	sessionFile := path + ".sequoia-session"
	if data, err := os.ReadFile(sessionFile); err == nil { //nolint:gosec // G304: sessionFile derived from path, not user input
		suffix := strings.TrimSpace(string(data))
		if suffix != "" {
			backup := path + ".sequoia-backup-" + suffix
			//nolint:gosec // G703: backup path derived from controlled suffix, not user input
			if _, err := os.Stat(backup); err == nil {
				return backup
			}
		}
	}

	// Fall back to legacy predictable backup name.
	legacyBackup := path + ".sequoia-backup"
	if _, err := os.Stat(legacyBackup); err == nil {
		return legacyBackup
	}

	return ""
}

// isSequoiaManaged reports whether the file at path contains the
// sequoia marker, indicating it was previously written by Sequoia.
func isSequoiaManaged(path string) (bool, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // G304: path is a known adapter-controlled file, not user input
	if err != nil {
		return false, err
	}
	return strings.Contains(string(raw), markerStart), nil
}

// AtomicWriteFile writes data to path atomically using a temporary file and
// rename. On Windows this prevents truncated files on crash, where os.WriteFile
// truncates in place. The temporary file is cleaned up if the rename fails.
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	//nolint:gosec // G703: tmp path derived from input path with known .tmp suffix
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("atomic write: write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("atomic write: rename %s: %w", path, err)
	}
	return nil
}
