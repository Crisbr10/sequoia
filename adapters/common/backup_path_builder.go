package common

import (
	"strconv"
	"time"
)

// BackupPathBuilder generates unique backup directory paths for tool
// adapters. Each call to Build() appends a base-36 session suffix to
// prevent directory name collisions across multiple install attempts.
//
// BackupPathBuilder is constructed once per adapter and injected into
// BaseAdapter via SetBackup().
type BackupPathBuilder struct {
	backupPathFn func(base string) string
	adapterID    string
}

// NewBackupPathBuilder creates a BackupPathBuilder with the given path
// function and adapter identifier.
func NewBackupPathBuilder(
	backupPathFn func(base string) string,
	adapterID string,
) *BackupPathBuilder {
	return &BackupPathBuilder{
		backupPathFn: backupPathFn,
		adapterID:    adapterID,
	}
}

// Build generates a unique backup path by combining the base path from
// backupPathFn, the adapter ID, and a millisecond-resolution session suffix.
// Format: {backupPathFn(base)}-{adapterID}-{sessionSuffix}
func (b *BackupPathBuilder) Build(base string) string {
	sessionSuffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	return b.backupPathFn(base) + "-" + b.adapterID + "-" + sessionSuffix
}
