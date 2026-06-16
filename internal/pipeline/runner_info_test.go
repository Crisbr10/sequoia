//nolint:gosec // test file: all os.* operations use t.TempDir() test fixtures, not production paths
package pipeline_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/model"
	"github.com/Crisbr10/sequoia/internal/pipeline"
)

// TestRunInstall_BackupInfo_MigrationNote verifies that when an adapter
// implements BackupDirGetter and reports a non-empty LastBackupDir, the
// pipeline emits a ProgressMsg whose Info field combines the backup
// directory with a one-line note about pre-existing scattered backups
// from prior sequoia versions. REQ-BRP-05.
//
// The note helps users locate the new central-home location while
// reassuring them that older per-tool .sequoia-backup-* files remain
// untouched at their original locations.
func TestRunInstall_BackupInfo_MigrationNote(t *testing.T) {
	t.Parallel()

	b := &backupAdapter{
		testAdapter: testAdapter{id: "migrate-tool", name: "Migrate Tool"},
		backupDir:   "C:/Users/me/AppData/Roaming/sequoia/backups/migrate-tool/2026-06-16T05-00-00.000Z-abc",
	}
	tools := []model.ToolState{
		{Adapter: b, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)
	adapterMsgs := filterCodegraphMessages(msgs)

	// 10 phase messages + 1 backup info = 11.
	require.Len(t, adapterMsgs, 11,
		"Should receive 11 messages: 10 phase messages + 1 backup info")

	// Last message: backup info with migration note.
	last := adapterMsgs[len(adapterMsgs)-1]
	assert.Equal(t, "migrate-tool", last.ToolID)
	assert.True(t, last.Done)
	assert.NotEmpty(t, last.Info, "backup info message must have Info set")

	// The Info must contain the backup dir so the user can find it.
	assert.Contains(t, last.Info, b.backupDir,
		"Info must contain the central-home backup directory path")

	// The Info must also include the migration note about pre-existing
	// scattered backups (REQ-BRP-05). The exact wording is a product
	// decision; the test asserts on a stable, descriptive substring
	// rather than the full sentence.
	assert.True(t,
		strings.Contains(last.Info, "pre-existing") ||
			strings.Contains(last.Info, "scattered") ||
			strings.Contains(last.Info, "prior sequoia"),
		"Info must include a one-line note about pre-existing scattered backups from prior sequoia versions, got: %q",
		last.Info)

	assert.False(t, last.Warning, "backup info should not be a warning")
	assert.Empty(t, last.Error, "backup info should not have Error set")

	// Sanity: the Info must not be just the path; it must include the note.
	// (i.e., the pipeline is appending to the path, not replacing it.)
	assert.Greater(t, len(last.Info), len(b.backupDir),
		"Info should be longer than just the backup path (the note is appended)")
}

// TestRunInstall_BackupInfo_NoMigrationNoteWhenEmpty verifies that when
// BackupDirGetter returns an empty string, the pipeline does not emit
// an Info message at all (no migration note, no backup dir). The
// migration note is part of the backup-info path, not a separate message.
func TestRunInstall_BackupInfo_NoMigrationNoteWhenEmpty(t *testing.T) {
	t.Parallel()

	b := &backupAdapter{
		testAdapter: testAdapter{id: "empty-backup", name: "Empty Backup"},
		backupDir:   "",
	}
	tools := []model.ToolState{
		{Adapter: b, Selected: true},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan model.ProgressMsg, model.ProgressChannelBufferSize)
	cmd := pipeline.RunInstall(ctx, tools, ch)
	require.NotNil(t, cmd)

	cmd()
	msgs := collectProgress(ch)
	adapterMsgs := filterCodegraphMessages(msgs)

	// 10 phase messages, no backup info when dir is empty.
	require.Len(t, adapterMsgs, 10,
		"Should receive 10 messages when backup dir is empty (no migration note)")
	assert.Empty(t, adapterMsgs[len(adapterMsgs)-1].Info,
		"last message should not have Info when backup dir is empty")
	assert.NotContains(t, strings.Join([]string{
		adapterMsgs[len(adapterMsgs)-1].Info,
	}, ""), "pre-existing",
		"no migration note should be emitted when backup dir is empty")
}

// Reference: adapters.BackupDirGetter
var _ adapters.BackupDirGetter = (*backupAdapter)(nil)
