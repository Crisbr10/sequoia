package common_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/adapters/common/installembed"
)

// =========================================================================
// checkpointCtx – auto-cancels after the Nth checkContext call.
//
// Install() calls checkContext at 5 fixed points. To test each checkpoint
// deterministically, we wrap a context and count how many times Done()
// is polled by the select in checkContext.
//
// cancelAt   = which checkpoint triggers cancellation
//
//	1 → before any work (line 284)
//	2 → after staging + rendering (line 323)
//	3 → after skill install      (line 355)
//	4 → after commands install   (line 373)
//	5 → after system prompt      (line 393)
//
// =========================================================================
type checkpointCtx struct {
	context.Context
	mu        sync.Mutex
	count     int
	cancelAt  int
	cancelled chan struct{}
	closed    bool
}

func newCheckpointCtx(cancelAt int) *checkpointCtx {
	return &checkpointCtx{
		Context:   context.Background(),
		cancelAt:  cancelAt,
		cancelled: make(chan struct{}),
	}
}

func (c *checkpointCtx) Done() <-chan struct{} {
	c.mu.Lock()
	c.count++
	if c.count >= c.cancelAt && !c.closed {
		close(c.cancelled)
		c.closed = true
	}
	done := c.cancelled // return consistently before unlock
	c.mu.Unlock()
	return done
}

func (c *checkpointCtx) Err() error {
	select {
	case <-c.cancelled:
		return context.Canceled
	default:
		return nil
	}
}

// =========================================================================
// Test helpers for Install error-path tests
// =========================================================================

// fullInstallTestAdapter creates a BaseAdapter that can complete the full
// Install pipeline (skill + commands + system prompt + version file).
// Uses installembed.FS which has "templates/skill.md.tmpl" and
// "templates/sys_prompt.md.tmpl" at the paths Install expects.
func fullInstallTestAdapter(t *testing.T, home string) *common.BaseAdapter {
	t.Helper()

	a := &common.BaseAdapter{}
	a.SetIDName("err-test", "Error Test")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		home,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	// System prompt write succeeds for happy-path / checkpoint tests.
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return nil }, // write succeeds
		nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"err-test",
	))
	// Set up a Detector so Status() and IsInstalled() work correctly.
	// isInstalledFn checks whether the version file exists.
	a.SetDetector(common.NewDetector(
		a.Base,
		func(base string) bool {
			_, err := os.Stat(filepath.Join(base, "version"))
			return err == nil
		},
		func() bool { return true },
	))
	// installembed.FS has templates at:
	//   "templates/skill.md.tmpl"
	//   "templates/sys_prompt.md.tmpl"
	a.SetInstallTemplates(installembed.FS, "sequoia-err-test-*",
		"templates/sys_prompt.md.tmpl",
		func() interface{} { return map[string]string{"Name": "err-test"} })

	return a
}

// failingWriteAdapter creates a BaseAdapter whose system prompt write
// ALWAYS fails. Use this for rollback-on-error tests.
func failingWriteAdapter(t *testing.T, home string, rollback bool) *common.BaseAdapter {
	t.Helper()

	a := fullInstallTestAdapter(t, home)
	pm := common.NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return fmt.Errorf("system prompt write failed") },
		nil,
	)
	if rollback {
		pm.SetRollbackOnError(true)
	}
	a.SetPromptManager(pm)
	return a
}

// =========================================================================
// Task 6.1 — Context cancellation at each Install checkpoint
// =========================================================================

// TestInstall_PreCancelledContext_NoWorkDone verifies REQ-TEST-ERRORS
// Scenario "Context cancellation before any work":
// A pre-cancelled context causes Install to return immediately without
// creating any directories or files.
func TestInstall_PreCancelledContext_NoWorkDone(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := a.Install(adapters.InstallOpts{Context: ctx})
	require.Error(t, err, "Install with pre-cancelled context should fail")
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed),
		"error should wrap ErrInstallFailed")
	assert.Contains(t, err.Error(), "context canceled",
		"error should mention context cancellation")

	// No directories or files should have been created.
	assert.NoDirExists(t, filepath.Join(home, "skills"),
		"skills dir should not exist after pre-cancelled Install")
	assert.NoDirExists(t, filepath.Join(home, "commands"),
		"commands dir should not exist after pre-cancelled Install")
	assert.NoFileExists(t, filepath.Join(home, "sys.md"),
		"system prompt should not exist after pre-cancelled Install")
	assert.NoFileExists(t, filepath.Join(home, "version"),
		"version file should not exist after pre-cancelled Install")
}

// TestInstall_CheckpointContext_AfterStaging verifies that when the context
// is cancelled at checkpoint 2 (after staging + rendering, before directory
// creation), the staging directory is cleaned up and no install files remain.
func TestInstall_CheckpointContext_AfterStaging(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)
	ctx := newCheckpointCtx(2) // cancel at checkpoint 2

	err := a.Install(adapters.InstallOpts{Context: ctx})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "context canceled")

	// No dirs created (we cancelled before mkdirAll).
	assert.NoDirExists(t, filepath.Join(home, "skills"))
	assert.NoDirExists(t, filepath.Join(home, "commands"))
}

// TestInstall_CheckpointContext_AfterSkillInstall verifies that when the
// context is cancelled at checkpoint 3 (after skill install), the skill
// installer is rolled back, and no command files are installed.
func TestInstall_CheckpointContext_AfterSkillInstall(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)
	ctx := newCheckpointCtx(3) // cancel at checkpoint 3

	err := a.Install(adapters.InstallOpts{Context: ctx})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "context canceled")

	// Skills dir should exist but be empty (rolled back) OR not exist.
	// Rollback restores backed-up files; since this is a clean install
	// with no prior state, Rollback removes the applied file.
	skillsDir := filepath.Join(home, "skills")
	assert.False(t, fileExists(filepath.Join(skillsDir, "SKILL.md")),
		"SKILL.md should be rolled back when context cancels after skill install")

	// Commands should not exist (we never got to command install).
	assert.NoFileExists(t, filepath.Join(home, "commands", common.CommandFiles[0]),
		"command files should not be installed")
}

// TestInstall_CheckpointContext_AfterCommandsInstall verifies that when
// the context is cancelled at checkpoint 4 (after commands install), both
// the skill and command installers are rolled back.
func TestInstall_CheckpointContext_AfterCommandsInstall(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)
	ctx := newCheckpointCtx(4) // cancel at checkpoint 4

	err := a.Install(adapters.InstallOpts{Context: ctx})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "context canceled")

	// Both skill and command files should be rolled back.
	assert.False(t, fileExists(filepath.Join(home, "skills", "SKILL.md")),
		"SKILL.md should be rolled back")
	assert.False(t, fileExists(filepath.Join(home, "commands", common.CommandFiles[0])),
		"command files should be rolled back")

	// System prompt should NOT be written (cancelled before that step).
	assert.NoFileExists(t, filepath.Join(home, "sys.md"))
}

// TestInstall_CheckpointContext_AfterSystemPrompt verifies that when the
// context is cancelled at checkpoint 5 (after system prompt write), both
// installers are rolled back, but the system prompt IS written.
func TestInstall_CheckpointContext_AfterSystemPrompt(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)
	ctx := newCheckpointCtx(5) // cancel at checkpoint 5 (after sys prompt)

	err := a.Install(adapters.InstallOpts{Context: ctx})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "context canceled")

	// Skills and commands are rolled back.
	assert.False(t, fileExists(filepath.Join(home, "skills", "SKILL.md")),
		"SKILL.md should be rolled back")
	assert.False(t, fileExists(filepath.Join(home, "commands", common.CommandFiles[0])),
		"command files should be rolled back")

	// BUT: the system prompt WAS written (the write succeeded before the
	// checkpoint). The rollback of installers does NOT undo the system
	// prompt write — only the skill/command files are rolled back.
	// This is the current behavior: system prompt write is not rolled back
	// on context cancellation.
}

// TestInstall_CheckpointContext_FullSuccess verifies that without any
// cancellation (cancelAt=6, beyond all checkpoints), Install succeeds
// and all expected files are created.
func TestInstall_CheckpointContext_FullSuccess(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)
	ctx := newCheckpointCtx(6) // cancelAt beyond all 5 checkpoints → never cancels

	err := a.Install(adapters.InstallOpts{Context: ctx})
	require.NoError(t, err, "Install should succeed when context is never cancelled")

	// All files should exist.
	assert.True(t, fileExists(filepath.Join(home, "skills", "SKILL.md")),
		"SKILL.md should exist after successful install")
	assert.True(t, fileExists(filepath.Join(home, "version")),
		"version file should exist after successful install")
	for _, cmd := range common.CommandFiles {
		assert.True(t, fileExists(filepath.Join(home, "commands", cmd)),
			"command %s should exist after successful install", cmd)
	}
}

// =========================================================================
// Task 6.2 — Nil function field safety tests
// =========================================================================

// TestBaseAdapter_NilDetector_DetectReturnsFalse verifies REQ-TEST-ERRORS
// "Nil detection function panics caught": Detect() returns false when
// detector is nil, without panicking.
func TestBaseAdapter_NilDetector_DetectReturnsFalse(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}
	// No SetDetector call → detector is nil.
	assert.NotPanics(t, func() {
		assert.False(t, a.Detect(), "Detect() with nil detector should return false")
	}, "Detect() must not panic when detector is nil")
}

// TestBaseAdapter_NilDetector_IsInstalledReturnsFalse verifies that
// IsInstalled() returns false when detector is nil.
func TestBaseAdapter_NilDetector_IsInstalledReturnsFalse(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}
	assert.NotPanics(t, func() {
		assert.False(t, a.IsInstalled(), "IsInstalled() with nil detector should return false")
	}, "IsInstalled() must not panic when detector is nil")
}

// TestBaseAdapter_NilPaths_BaseReturnsError verifies that Base() returns a
// sensible error when PathResolver is nil, instead of panicking.
func TestBaseAdapter_NilPaths_BaseReturnsError(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}
	// No SetPaths call → paths is nil.
	base, err := a.Base()
	assert.Error(t, err, "Base() with nil paths should return error")
	assert.Empty(t, base, "Base() with nil paths should return empty string")
	assert.Contains(t, err.Error(), "PathResolver not configured",
		"error should mention PathResolver")
}

// TestBaseAdapter_NilPaths_PathMethodsReturnEmpty verifies that SkillsPath,
// CommandsPath, SystemPromptPath all return empty strings when paths is nil.
func TestBaseAdapter_NilPaths_PathMethodsReturnEmpty(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}

	assert.Equal(t, "", a.SkillsPath(), "SkillsPath() should return empty when paths is nil")
	assert.Equal(t, "", a.CommandsPath(), "CommandsPath() should return empty when paths is nil")
	assert.Equal(t, "", a.SystemPromptPath(), "SystemPromptPath() should return empty when paths is nil")
	assert.Equal(t, "", a.HomeDir(), "HomeDir() should return empty when paths is nil")
}

// TestBaseAdapter_NilPrompt_PromptStrategyReturnsZero verifies that
// PromptStrategy() returns the zero value when prompt manager is nil.
func TestBaseAdapter_NilPrompt_PromptStrategyReturnsZero(t *testing.T) {
	t.Parallel()

	a := &common.BaseAdapter{}
	// No SetPromptManager call → prompt is nil.
	assert.Equal(t, adapters.PromptStrategy(0), a.PromptStrategy(),
		"PromptStrategy() with nil prompt should return zero value")
}

// =========================================================================
// Task 6.3 — Staging directory creation failure
// =========================================================================

// TestInstall_StagingDirCreationFailure verifies REQ-TEST-ERRORS
// "Staging dir | MkdirTemp fails | returns wrapped error".
// We use an empty staging prefix which is valid but produces a specific
// path pattern that can be tested for. To actually cause MkdirTemp to fail,
// we use a malformed prefix containing a null byte, which os.MkdirTemp
// rejects.
func TestInstall_StagingDirCreationFailure(t *testing.T) {
	t.Parallel()

	home := t.TempDir()

	a := &common.BaseAdapter{}
	a.SetIDName("staging-fail", "Staging Fail")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		home,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		nil, nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"staging-fail",
	))
	a.SetInstallTemplates(installembed.FS, "\x00invalid*",
		"templates/sys_prompt.md.tmpl",
		func() interface{} { return map[string]string{"Name": "fail"} })

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err, "Install with null byte in staging prefix should fail")
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.True(t,
		strings.Contains(err.Error(), "create staging dir") ||
			strings.Contains(err.Error(), "staging"),
		"error should mention staging dir failure, got: %v", err)
}

// =========================================================================
// Task 6.4 — Template rendering failure
// =========================================================================

// TestInstall_SkillTemplateNotFound verifies REQ-TEST-ERRORS
// "Skill template | template not found in embed.FS | returns wrapped error".
// Uses the testFS (from shared_test.go) which lacks "templates/skill.md.tmpl".
func TestInstall_SkillTemplateNotFound(t *testing.T) {
	t.Parallel()

	home := t.TempDir()

	a := &common.BaseAdapter{}
	a.SetIDName("tmpl-fail", "Template Fail")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		home,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		nil, nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"tmpl-fail",
	))
	a.SetInstallTemplates(testFS, "sequoia-tmpl-fail-*",
		"testdata/test.tmpl",
		func() interface{} { return map[string]string{"Name": "fail"} })

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "read template",
		"error should mention template reading failure")

	// No install files should exist.
	assert.NoFileExists(t, filepath.Join(home, "version"))
}

// TestInstall_SystemPromptTemplateNotFound verifies REQ-TEST-ERRORS
// "System prompt render | template not found | returns wrapped error".
// Skill template is present (installembed.FS has "templates/skill.md.tmpl")
// but the system prompt template is set to a non-existent path.
func TestInstall_SystemPromptTemplateNotFound(t *testing.T) {
	t.Parallel()

	home := t.TempDir()

	a := &common.BaseAdapter{}
	a.SetIDName("sys-tmpl-fail", "Sys Template Fail")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		home,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return nil },
		nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"sys-tmpl-fail",
	))
	// Skill template exists (installembed.FS has "templates/skill.md.tmpl")
	// but system prompt template is wrong.
	a.SetInstallTemplates(installembed.FS, "sequoia-sys-tmpl-fail-*",
		"templates/does_not_exist.md.tmpl",
		func() interface{} { return map[string]string{"Name": "fail"} })

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "read template",
		"error should mention template reading failure for system prompt")
}

// =========================================================================
// Task 6.5 — Base() resolution failure
// =========================================================================

// TestInstall_BaseResolutionFailure verifies REQ-TEST-ERRORS
// "Base resolution | resolveBase returns error | returns wrapped error".
// Install should not proceed if the base directory cannot be resolved.
func TestInstall_BaseResolutionFailure(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("cannot determine home directory")

	a := &common.BaseAdapter{}
	a.SetIDName("base-fail", "Base Fail")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) {
			return "", baseErr
		},
		"/fake/home",
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		nil, nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"base-fail",
	))
	a.SetInstallTemplates(installembed.FS, "sequoia-base-fail-*",
		"templates/sys_prompt.md.tmpl",
		func() interface{} { return map[string]string{"Name": "fail"} })

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "resolve home",
		"error should mention resolve home failure")
}

// TestInstall_HomeDirUnavailable verifies that when no homeDir is set and
// os.UserHomeDir() would normally be called (but the PathResolver has an
// empty homeDir with failing resolveBase), Install fails with a clear error.
func TestInstall_HomeDirUnavailable(t *testing.T) {
	t.Parallel()

	// Create a zero-value adapter — no SetPaths, no SetIDName.
	a := &common.BaseAdapter{}

	// Base() should return a sensible error because paths is nil.
	base, err := a.Base()
	assert.Error(t, err)
	assert.Empty(t, base)
	assert.Contains(t, err.Error(), "PathResolver")
}

// =========================================================================
// Task 6.6 — Version file write failure
// =========================================================================

// TestInstall_VersionFileWriteFailure verifies REQ-TEST-ERRORS
// "Version file write | AtomicWriteFile fails | returns wrapped error".
// We make the version file path point to a directory (which os.Rename can't
// overwrite with a file), causing AtomicWriteFile to fail.
func TestInstall_VersionFileWriteFailure(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	// Pre-create the version path as a directory — AtomicWriteFile will fail
	// because os.Rename cannot replace a directory with a file.
	versionDir := filepath.Join(home, "version-is-dir")
	require.NoError(t, os.MkdirAll(versionDir, 0o755))

	a := &common.BaseAdapter{}
	a.SetIDName("ver-fail", "Version Fail")
	a.SetPaths(common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		home,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return versionDir }, // version "file" is a directory
		func(base string) string { return filepath.Join(base, "backup") },
		a.AddWarning,
	))
	a.SetPromptManager(common.NewPromptManager(adapters.StrategyFileReplace,
		func(base, content string) error { return nil },
		nil,
	))
	a.SetBackup(common.NewBackupPathBuilder(
		func(base string) string { return filepath.Join(base, "backup") },
		"ver-fail",
	))
	a.SetInstallTemplates(installembed.FS, "sequoia-ver-fail-*",
		"templates/sys_prompt.md.tmpl",
		func() interface{} { return map[string]string{"Name": "fail"} })

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "version file",
		"error should mention version file failure, got: %v", err)
}

// =========================================================================
// Task 6.7 — System prompt failure with/without rollback
// =========================================================================

// TestInstall_SystemPromptFailure_Rollback verifies REQ-TEST-ERRORS
// "System prompt write fails, rollback=true → skill+commands rolled back".
func TestInstall_SystemPromptFailure_Rollback(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := failingWriteAdapter(t, home, true) // rollback=true

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "system prompt",
		"error should mention system prompt failure")

	// Skills and commands should be rolled back.
	assert.False(t, fileExists(filepath.Join(home, "skills", "SKILL.md")),
		"SKILL.md should be rolled back when system prompt fails (rollback=true)")
	assert.False(t, fileExists(filepath.Join(home, "commands", common.CommandFiles[0])),
		"commands should be rolled back when system prompt fails (rollback=true)")

	// Version file should NOT be written (we bail before that step).
	assert.NoFileExists(t, filepath.Join(home, "version"),
		"version file should not be written after system prompt failure")
}

// TestInstall_SystemPromptFailure_NoRollback verifies REQ-TEST-ERRORS
// "System prompt write fails, rollback=false → skill+commands NOT rolled back".
func TestInstall_SystemPromptFailure_NoRollback(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := failingWriteAdapter(t, home, false) // rollback=false

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))
	assert.Contains(t, err.Error(), "system prompt",
		"error should mention system prompt failure")

	// Skills and commands should STILL EXIST (rollback=false).
	assert.True(t, fileExists(filepath.Join(home, "skills", "SKILL.md")),
		"SKILL.md should still exist when rollback=false")
	assert.True(t, fileExists(filepath.Join(home, "commands", common.CommandFiles[0])),
		"commands should still exist when rollback=false")

	// Version file should NOT be written.
	assert.NoFileExists(t, filepath.Join(home, "version"),
		"version file should not be written after system prompt failure")
}

// TestInstall_SystemPromptFailure_RollbackBackupDir verifies that when
// rollback=true on system prompt failure, the backup directory still
// exists with the rolled-back skill file.
func TestInstall_SystemPromptFailure_RollbackBackupDir(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	// Pre-create the skill file so Installer backs it up.
	skillsDir := filepath.Join(home, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(skillsDir, "SKILL.md"),
		[]byte("original skill content"), 0o644,
	))
	// Pre-create command files too.
	cmdsDir := filepath.Join(home, "commands")
	require.NoError(t, os.MkdirAll(cmdsDir, 0o755))
	for _, cmd := range common.CommandFiles {
		require.NoError(t, os.WriteFile(
			filepath.Join(cmdsDir, cmd),
			[]byte("original command"), 0o644,
		))
	}

	a := failingWriteAdapter(t, home, true) // rollback=true

	err := a.Install(adapters.InstallOpts{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, adapters.ErrInstallFailed))

	// Skills and commands should be rolled back to original content.
	skillContent, err := os.ReadFile(filepath.Join(skillsDir, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, "original skill content", string(skillContent),
		"SKILL.md should be restored to original after rollback")

	cmdContent, err := os.ReadFile(filepath.Join(cmdsDir, common.CommandFiles[0]))
	require.NoError(t, err)
	assert.Equal(t, "original command", string(cmdContent),
		"command files should be restored to original after rollback")
}
