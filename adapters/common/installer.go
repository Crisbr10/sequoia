// Package common provides shared infrastructure for tool adapters,
// including the Installer lifecycle (Prepare → Apply → Verify → Rollback).
package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InstallerConfig holds everything the Installer needs to run.
type InstallerConfig struct {
	// SourceDir is the directory containing the template files to install.
	SourceDir string
	// TargetDir is the destination directory where files will be placed.
	TargetDir string
	// BackupDir is where existing files are backed up during Prepare.
	BackupDir string
	// Files is the list of filenames (relative to SourceDir/TargetDir) to install.
	Files []string
}

// Installer manages the four-phase install lifecycle for a ToolAdapter.
// The expected call order is: Prepare → Apply → Verify.
// On any failure in Apply or Verify, the caller should invoke Rollback.
type Installer struct {
	config  InstallerConfig
	applied []string // tracks which files were successfully applied, for rollback
}

// NewInstaller creates an Installer from the given config.
func NewInstaller(cfg InstallerConfig) *Installer {
	return &Installer{config: cfg}
}

// Run executes the full Prepare → Apply → Verify cycle.
// On Apply or Verify failure it calls Rollback (best-effort) and returns the
// original error. This is the convenience wrapper that most adapters use
// instead of calling each phase individually.
func (i *Installer) Run() error {
	if err := i.Prepare(); err != nil {
		return err
	}
	if err := i.Apply(); err != nil {
		_ = i.Rollback()
		return err
	}
	if err := i.Verify(); err != nil {
		_ = i.Rollback()
		return err
	}
	return nil
}

// Prepare validates paths, checks write permissions, and backs up existing files.
// It MUST be called before Apply.
// Returns an error if TargetDir is not writable or any backup fails.
func (i *Installer) Prepare() error {
	cfg := i.config

	// Verify TargetDir is writable by creating and immediately removing a temp file.
	probe := filepath.Join(cfg.TargetDir, ".sequoia-probe")
	f, err := os.Create(probe)
	if err != nil {
		return fmt.Errorf("prepare: target directory not writable: %w", err)
	}
	f.Close()
	if err := os.Remove(probe); err != nil {
		return fmt.Errorf("prepare: could not remove write-probe file: %w", err)
	}

	// Back up any files that already exist in TargetDir.
	for _, name := range cfg.Files {
		src := filepath.Join(cfg.TargetDir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			// File not present — nothing to back up.
			continue
		} else if err != nil {
			return fmt.Errorf("prepare: stat %q: %w", src, err)
		}

		// Ensure BackupDir exists before the first copy.
		if err := os.MkdirAll(cfg.BackupDir, 0o755); err != nil {
			return fmt.Errorf("prepare: create backup dir: %w", err)
		}

		dst := filepath.Join(cfg.BackupDir, name)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("prepare: backup %q: %w", name, err)
		}
	}

	return nil
}

// Apply copies files from SourceDir to TargetDir.
// It MUST be called after a successful Prepare.
// Returns an error if any file copy fails.
// Apply does NOT call Rollback on failure — the caller is responsible.
func (i *Installer) Apply() error {
	cfg := i.config
	i.applied = i.applied[:0] // reset from any previous call

	for _, name := range cfg.Files {
		src := filepath.Join(cfg.SourceDir, name)
		dst := filepath.Join(cfg.TargetDir, name)

		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("apply: copy %q: %w", name, err)
		}
		i.applied = append(i.applied, name)
	}

	return nil
}

// Verify checks that all installed files exist and are readable.
// It MUST be called after a successful Apply.
// Returns an error if any expected file is missing or unreadable.
func (i *Installer) Verify() error {
	cfg := i.config

	for _, name := range cfg.Files {
		path := filepath.Join(cfg.TargetDir, name)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("verify: %s: %w", name, err)
		}
	}

	return nil
}

// Rollback restores files backed up during Prepare, then removes the backup
// directory. It also removes any files listed in i.applied that were NOT
// backed up (clean installs with no prior state).
// Safe to call even if Apply was never called.
// Returns the first error encountered, but continues restoring all files.
func (i *Installer) Rollback() error {
	cfg := i.config
	var firstErr error

	// Build a set of backed-up file names for quick lookup.
	backedUp := make(map[string]bool)
	if _, err := os.Stat(cfg.BackupDir); err == nil {
		for _, name := range cfg.Files {
			backupPath := filepath.Join(cfg.BackupDir, name)
			if _, err := os.Stat(backupPath); err == nil {
				backedUp[name] = true
			}
		}
	}

	// Restore backed-up files → TargetDir.
	for name := range backedUp {
		src := filepath.Join(cfg.BackupDir, name)
		dst := filepath.Join(cfg.TargetDir, name)
		if err := copyFile(src, dst); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("rollback: restore %q: %w", name, err)
			}
		}
	}

	// Remove files that were applied but had no prior backup (clean install case).
	for _, name := range i.applied {
		if backedUp[name] {
			continue // already restored above
		}
		target := filepath.Join(cfg.TargetDir, name)
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			if firstErr == nil {
				firstErr = fmt.Errorf("rollback: remove %q: %w", name, err)
			}
		}
	}

	// Remove the backup directory.
	if _, err := os.Stat(cfg.BackupDir); err == nil {
		if err := os.RemoveAll(cfg.BackupDir); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("rollback: remove backup dir: %w", err)
		}
	}

	return firstErr
}

// copyFile copies the file at src to dst, creating or overwriting dst.
// Parent directories of dst must already exist.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source %q: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dest %q: %w", dst, err)
	}
	defer func() {
		// Surface close error only when no write error occurred.
		if cerr := out.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close dest %q: %w", dst, cerr)
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %q → %q: %w", src, dst, err)
	}

	return nil
}
