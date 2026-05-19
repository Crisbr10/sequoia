package common

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Crisbr10/sequoia/adapters"
)

// BaseAdapter provides shared Install, Uninstall, Status, and path methods
// for tool adapters. Concrete adapters embed BaseAdapter and configure it
// via setters for detection, prompt strategy, and templates.
//
// Path resolution is delegated to an internal *PathResolver set via SetPaths().
//
// The Install/Uninstall flow follows the same 8-step pattern for all adapters;
// only the system prompt strategy and path layout differ.
type BaseAdapter struct {
	// adapterID is the unique machine-readable identifier (e.g. "claude-code").
	adapterID string
	// adapterName is the human-readable display name (e.g. "Claude Code").
	adapterName string

	// paths resolves tool config directory paths. Set via SetPaths().
	paths *PathResolver

	// isInstalledFn checks whether Sequoia is already installed for this tool.
	// Receives the resolved base directory.
	isInstalledFn func(base string) bool

	// detectFn reports whether the tool is present on the host.
	detectFn func() bool

	// promptStrategy is the PromptStrategy constant for this adapter.
	promptStrategy adapters.PromptStrategy

	// --- Install customization ---

	// templateFS is the embed.FS containing this adapter's templates.
	templateFS embed.FS
	// stagingPrefix is the prefix passed to os.MkdirTemp for staging (e.g. "sequoia-claude-*").
	stagingPrefix string
	// systemPromptTemplate is the template name for the system prompt section
	// (e.g. "templates/claude-md-section.md.tmpl").
	systemPromptTemplate string
	// makeTemplateData returns the data passed to skill and system prompt templates.
	makeTemplateData func() interface{}

	// writeSystemPrompt writes the rendered system prompt content. Strategy varies:
	//   - StrategyMarkdownSections → InjectMarkdownSection
	//   - StrategyFileReplace → ReplaceFile
	writeSystemPrompt func(base, content string) error

	// removeSystemPrompt removes or restores the system prompt. Strategy varies:
	//   - StrategyMarkdownSections → RemoveMarkdownSection
	//   - StrategyFileReplace → RestoreOrRemoveFile
	removeSystemPrompt func(base string) error
	// rollbackOnSystemPromptError controls whether the shared Install() rolls
	// back skill and command installers when the system prompt step fails.
	// opencode and cursor need this; claude and gemini do not.
	rollbackOnSystemPromptError bool

	// warnings collects non-fatal warnings during adapter operations
	// (e.g., symlink resolution failures). Protected by mu.
	warnings []string
	mu       sync.Mutex
	// lastBackupDir stores the backup directory path from the most recent
	// Install or Uninstall operation. Exposed via LastBackupDir() for
	// BackupDirGetter interface conformance (REQ-BUG-004).
	lastBackupDir string
}

// SetIDName sets the adapter's unique ID and human-readable name.
func (a *BaseAdapter) SetIDName(id, name string) {
	a.adapterID = id
	a.adapterName = name
}

// SetPaths sets the PathResolver used by BaseAdapter for all path operations.
// Concrete adapters MUST call this during construction:
//
//	a.SetPaths(common.NewPathResolver(resolveBase, homeDir, skillsFn, commandsFn,
//	    systemPromptFn, versionFileFn, backupFn, a.AddWarning))
func (a *BaseAdapter) SetPaths(p *PathResolver) {
	a.paths = p
}

// SetHomeDir overrides the user home directory (for testing).
// Delegates to the internal PathResolver if configured.
func (a *BaseAdapter) SetHomeDir(dir string) {
	if a.paths != nil {
		a.paths.SetHomeDir(dir)
	}
}

// HomeDir returns the current home directory override.
func (a *BaseAdapter) HomeDir() string {
	if a.paths != nil {
		return a.paths.HomeDir()
	}
	return ""
}

// ID returns the unique machine-readable identifier.
func (a *BaseAdapter) ID() string { return a.adapterID }

// Name returns the human-readable display name.
func (a *BaseAdapter) Name() string { return a.adapterName }

// SetStrategy sets the prompt strategy and its write/remove functions.
func (a *BaseAdapter) SetStrategy(strategy adapters.PromptStrategy, write func(base, content string) error, remove func(base string) error) {
	a.promptStrategy = strategy
	a.writeSystemPrompt = write
	a.removeSystemPrompt = remove
}

// SetInstallTemplates sets the template embed.FS, staging prefix, system prompt
// template name, and the function that produces template data.
func (a *BaseAdapter) SetInstallTemplates(fs embed.FS, stagingPrefix, sysPromptTmpl string, makeData func() interface{}) {
	a.templateFS = fs
	a.stagingPrefix = stagingPrefix
	a.systemPromptTemplate = sysPromptTmpl
	a.makeTemplateData = makeData
}

// SetIsInstalledFn sets the function that checks if Sequoia is installed.
func (a *BaseAdapter) SetIsInstalledFn(fn func(base string) bool) {
	a.isInstalledFn = fn
}

// SetDetectFn sets the function that detects whether the tool is present.
func (a *BaseAdapter) SetDetectFn(fn func() bool) {
	a.detectFn = fn
}

// SetRollbackOnSystemPromptError enables or disables rollback of skill and
// command installers when the system prompt step fails during Install().
func (a *BaseAdapter) SetRollbackOnSystemPromptError(v bool) {
	a.rollbackOnSystemPromptError = v
}

// AddWarning appends a non-fatal warning message. Thread-safe.
func (a *BaseAdapter) AddWarning(msg string) {
	a.mu.Lock()
	a.warnings = append(a.warnings, msg)
	a.mu.Unlock()
}

// Warnings returns a copy of all accumulated warning messages. Thread-safe.
func (a *BaseAdapter) Warnings() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string{}, a.warnings...)
}

// LastBackupDir returns the backup directory path from the most recent
// Install or Uninstall operation. Implements adapters.BackupDirGetter.
func (a *BaseAdapter) LastBackupDir() string {
	return a.lastBackupDir
}

// clearWarnings removes all accumulated warnings. Caller must hold a.mu or
// call from a context where no concurrent access is possible (e.g., start of
// Install/Uninstall before any goroutine shares the adapter).
func (a *BaseAdapter) clearWarnings() {
	a.warnings = a.warnings[:0]
}

// Base resolves and returns the tool's config root directory.
// Delegates to the internal PathResolver.
func (a *BaseAdapter) Base() (string, error) {
	if a.paths == nil {
		return "", errors.New("base adapter: PathResolver not configured")
	}
	return a.paths.Base()
}

// SkillsPath returns the absolute path to the skills directory.
func (a *BaseAdapter) SkillsPath() string {
	if a.paths == nil {
		return ""
	}
	return a.paths.SkillsPath()
}

// CommandsPath returns the absolute path to the commands directory.
func (a *BaseAdapter) CommandsPath() string {
	if a.paths == nil {
		return ""
	}
	return a.paths.CommandsPath()
}

// SystemPromptPath returns the absolute path to the system prompt file.
func (a *BaseAdapter) SystemPromptPath() string {
	if a.paths == nil {
		return ""
	}
	return a.paths.SystemPromptPath()
}

// PromptStrategy returns the injection strategy used by this adapter.
func (a *BaseAdapter) PromptStrategy() adapters.PromptStrategy {
	return a.promptStrategy
}

// Detect reports whether the tool is present on this machine.
func (a *BaseAdapter) Detect() bool {
	return a.detectFn()
}

// IsInstalled reports whether Sequoia has already been installed.
func (a *BaseAdapter) IsInstalled() bool {
	base, err := a.Base()
	if err != nil {
		return false
	}
	return a.isInstalledFn(base)
}

// Status returns the current installation status.
func (a *BaseAdapter) Status() adapters.AdapterStatus {
	installed := a.IsInstalled()
	version := ""
	if installed {
		base, err := a.Base()
		if err == nil {
			data, err := os.ReadFile(a.paths.versionFilePathFn(base))
			if err == nil {
				version = strings.TrimSpace(string(data))
			}
		}
	}
	return adapters.AdapterStatus{
		Installed: installed,
		Version:   version,
		Path:      a.SkillsPath(),
	}
}

// checkContext returns ctx.Err() if the context is done, nil otherwise.
// A nil context is treated as not cancelled (backwards-compatible).
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Install installs Sequoia files using the common 9-step pattern.
// The system prompt strategy is delegated to writeSystemPrompt.
//
// When opts.Context is set and cancelled, Install aborts early and rolls
// back any partial work before returning the context error.
//
// On failure, the returned error wraps adapters.ErrInstallFailed so callers
// can detect install failures with errors.Is(err, adapters.ErrInstallFailed).
func (a *BaseAdapter) Install(opts adapters.InstallOpts) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", adapters.ErrInstallFailed, err)
		}
	}()
	// Clear warnings from any previous operation.
	a.clearWarnings()

	// Check for early cancellation before doing any work.
	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	base, err := a.Base()
	if err != nil {
		return fmt.Errorf("install: resolve home: %w", err)
	}

	data := a.makeTemplateData()

	// Stage rendered templates to a temp dir for common.Installer.
	staging, err := os.MkdirTemp("", a.stagingPrefix)
	if err != nil {
		return fmt.Errorf("install: create staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	// Render and stage the skill file.
	skillContent, err := RenderTemplate(&a.templateFS, "templates/skill.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := StageFile(staging, "SKILL.md", []byte(skillContent)); err != nil {
		return fmt.Errorf("install: stage skill: %w", err)
	}

	// Stage command files (static — no rendering needed). Uses the shared CommandFS.
	for _, cmd := range CommandFiles {
		content, err := CommandFS.ReadFile("templates/commands/" + cmd)
		if err != nil {
			return fmt.Errorf("install: read command %q: %w", cmd, err)
		}
		if err := StageFile(staging, cmd, content); err != nil {
			return fmt.Errorf("install: stage command %q: %w", cmd, err)
		}
	}

	// Check cancellation before creating directories.
	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	skillsDir := a.paths.skillsPathFn(base)
	commandsDir := a.paths.commandsPathFn(base)
	// Append a unique session suffix to the backup dir to avoid name collisions
	// with pre-existing directories.
	sessionSuffix := strconv.FormatInt(time.Now().UnixMilli(), 36)
	backupDir := a.paths.backupPathFn(base) + "-" + a.ID() + "-" + sessionSuffix
	a.lastBackupDir = backupDir

	// Create target directories before Prepare (Prepare probes for write access).
	for _, dir := range []string{skillsDir, commandsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("install: create dir %q: %w", dir, err)
		}
	}

	// Install skill file via the common framework.
	skillInstaller := NewInstaller(InstallerConfig{
		SourceDir: staging,
		TargetDir: skillsDir,
		BackupDir: filepath.Join(backupDir, "skills"),
		Files:     []string{"SKILL.md"},
	})
	if err := skillInstaller.Run(); err != nil {
		return fmt.Errorf("install: skill: %w", err)
	}

	// Check cancellation after skills install — roll back if needed.
	if err := checkContext(opts.Context); err != nil {
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: %w", err)
	}

	// Install command files via the common framework.
	cmdInstaller := NewInstaller(InstallerConfig{
		SourceDir: staging,
		TargetDir: commandsDir,
		BackupDir: filepath.Join(backupDir, "commands"),
		Files:     CommandFiles,
	})
	if err := cmdInstaller.Run(); err != nil {
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: commands: %w", err)
	}

	// Check cancellation after commands install — roll back if needed.
	if err := checkContext(opts.Context); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: %w", err)
	}

	// Render and write the system prompt.
	sectionContent, err := RenderTemplate(&a.templateFS, a.systemPromptTemplate, data)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := a.writeSystemPrompt(base, sectionContent); err != nil {
		if a.rollbackOnSystemPromptError {
			_ = cmdInstaller.Rollback()
			_ = skillInstaller.Rollback()
		}
		return fmt.Errorf("install: system prompt: %w", err)
	}

	// Check cancellation after system prompt — roll back if needed.
	if err := checkContext(opts.Context); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("install: %w", err)
	}

	// Write the version marker file.
	if err := AtomicWriteFile(a.paths.versionFilePathFn(base), []byte(Version+"\n"), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files using the common pattern.
// The system prompt strategy is delegated to removeSystemPrompt.
//
// When opts.Context is set and cancelled, Uninstall returns early without
// modifying any files.
//
// Errors from individual file removals are collected via errors.Join.
// Missing files are not treated as errors (os.IsNotExist is checked).
// On failure, the returned error wraps adapters.ErrUninstallFailed so
// callers can detect uninstall failures with errors.Is.
func (a *BaseAdapter) Uninstall(opts adapters.InstallOpts) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", adapters.ErrUninstallFailed, err)
		}
	}()
	// Clear warnings from any previous operation.
	a.clearWarnings()

	// Check for early cancellation before doing any work.
	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}

	base, err := a.Base()
	if err != nil {
		return fmt.Errorf("uninstall: resolve home: %w", err)
	}

	// Collect errors from individual file removals instead of discarding them.
	var errs []error

	if err := os.Remove(filepath.Join(a.paths.skillsPathFn(base), "SKILL.md")); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove skill file: %w", err))
	}
	if err := os.Remove(a.paths.versionFilePathFn(base)); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove version file: %w", err))
	}
	for _, cmd := range CommandFiles {
		if err := os.Remove(filepath.Join(a.paths.commandsPathFn(base), cmd)); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove command %s: %w", cmd, err))
		}
	}

	// Remove or restore the system prompt.
	if err := a.removeSystemPrompt(base); err != nil {
		errs = append(errs, fmt.Errorf("restore system prompt: %w", err))
	}

	return errors.Join(errs...)
}
