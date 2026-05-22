package common

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

	// detector encapsulates tool presence detection and installation checks.
	// Set via SetDetector(). Delegates to Detect()/IsInstalled().
	detector *Detector

	// prompt handles the system prompt write/remove lifecycle.
	// Strategy, write/remove functions, and rollback flag are all encapsulated.
	// Set via SetPromptManager().
	prompt *PromptManager

	// backup generates unique backup directory paths per install session.
	// Set via SetBackup().
	backup *BackupPathBuilder

	// templateFS is the embed.FS containing this adapter's templates.
	templateFS embed.FS
	// stagingPrefix is the prefix passed to os.MkdirTemp for staging (e.g. "sequoia-claude-*").
	stagingPrefix string
	// systemPromptTemplate is the template name for the system prompt section
	// (e.g. "templates/claude-md-section.md.tmpl").
	systemPromptTemplate string
	// makeTemplateData returns the data passed to skill and system prompt templates.
	makeTemplateData func() interface{}

	// warnings collects non-fatal warnings during adapter operations
	// (e.g., symlink resolution failures). Protected by mu.
	warnings []string
	mu       sync.Mutex
	// lastBackupDir stores the backup directory path from the most recent
	// Install or Uninstall operation. Exposed via LastBackupDir() for
	// BackupDirGetter interface conformance (REQ-BUG-004).
	lastBackupDir string

	// strategyState holds per-operation state for the Strategy interface
	// phased lifecycle. Populated by Prepare/Download/Stage and consumed
	// by subsequent phases. Cleared at the start of Prepare().
	strategyState *strategyPhaseState

	// detectedOnce memoizes the Detect() result for process lifetime.
	// PATH doesn't change during execution, so the result is stable.
	// P2-001: Cache Detect() results to avoid repeated exec.LookPath calls.
	detectedOnce   sync.Once
	detectedResult bool
}

// strategyPhaseState holds the shared state across Strategy phase calls.
// It is reset when Prepare() starts a new operation.
type strategyPhaseState struct {
	base           string     // resolved config root directory
	backupDir      string     // unique backup directory for this session
	stagingDir     string     // temp directory with staged template files
	data           interface{} // template rendering data
	skillInstaller *Installer // skill Installer created in Stage()
	cmdInstaller   *Installer // command Installer created in Stage()
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

// SetPromptManager sets the PromptManager used by BaseAdapter for system prompt
// write/remove operations and rollback-on-error control. Concrete adapters MUST
// call this during construction:
//
//	a.SetPromptManager(common.NewPromptManager(strategy, writeFn, removeFn))
func (a *BaseAdapter) SetPromptManager(pm *PromptManager) {
	a.prompt = pm
}

// SetBackup sets the BackupPathBuilder used by BaseAdapter to generate unique
// backup directory paths per install session. Concrete adapters MUST call
// this during construction:
//
//	a.SetBackup(common.NewBackupPathBuilder(backupPathFn, a.ID()))
func (a *BaseAdapter) SetBackup(b *BackupPathBuilder) {
	a.backup = b
}

// BuildBackupPath generates a unique backup directory path for the given base
// directory. Delegates to the internal BackupPathBuilder. Returned by custom
// Install() implementations (e.g., Codex) that need backup paths.
func (a *BaseAdapter) BuildBackupPath(base string) string {
	return a.backup.Build(base)
}

// SetInstallTemplates sets the template embed.FS, staging prefix, system prompt
// template name, and the function that produces template data.
func (a *BaseAdapter) SetInstallTemplates(fs embed.FS, stagingPrefix, sysPromptTmpl string, makeData func() interface{}) {
	a.templateFS = fs
	a.stagingPrefix = stagingPrefix
	a.systemPromptTemplate = sysPromptTmpl
	a.makeTemplateData = makeData
}

// SetDetector sets the Detector used by BaseAdapter for tool presence
// and installation checks. Concrete adapters MUST call this during construction:
//
//	a.SetDetector(common.NewDetector(a.paths.Base, isInstalledFn, detectFn))
func (a *BaseAdapter) SetDetector(d *Detector) {
	a.detector = d
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
// Delegates to the internal PromptManager.
func (a *BaseAdapter) PromptStrategy() adapters.PromptStrategy {
	if a.prompt == nil {
		return 0 // zero value (StrategyMarkdownSections) as safe default
	}
	return a.prompt.PromptStrategy()
}

// Detect reports whether the tool is present on this machine.
// Delegates to the internal Detector. Result is cached via sync.Once
// for the process lifetime — PATH does not change during execution (P2-001).
func (a *BaseAdapter) Detect() bool {
	a.detectedOnce.Do(func() {
		if a.detector == nil {
			a.detectedResult = false
			return
		}
		a.detectedResult = a.detector.Detect()
	})
	return a.detectedResult
}

// IsInstalled reports whether Sequoia has already been installed.
// Delegates to the internal Detector.
func (a *BaseAdapter) IsInstalled() bool {
	if a.detector == nil {
		return false
	}
	return a.detector.IsInstalled()
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

// Strategy returns the Strategy implementation for this adapter.
// BaseAdapter implements Strategy directly, so it returns itself.
// Adapters that override Install() (e.g., Codex) may override this
// method to return a custom Strategy implementation.
func (a *BaseAdapter) Strategy() Strategy {
	return a
}

// Prepare sets up directories and backup paths for the install operation.
// It clears warnings from previous operations, checks context cancellation,
// resolves the base directory, creates target directories, and generates
// a unique backup path.
//
// Must be called first in the Strategy lifecycle.
func (a *BaseAdapter) Prepare(opts adapters.InstallOpts) error {
	// Reset and clear state for a new operation.
	a.clearWarnings()
	a.strategyState = nil

	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	base, err := a.Base()
	if err != nil {
		return fmt.Errorf("prepare: resolve home: %w", err)
	}

	skillsDir := a.paths.skillsPathFn(base)
	commandsDir := a.paths.commandsPathFn(base)

	backupDir := a.backup.Build(base)
	a.mu.Lock()
	a.lastBackupDir = backupDir
	a.mu.Unlock()

	for _, dir := range []string{skillsDir, commandsDir} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("prepare: create dir %q: %w", dir, err)
		}
	}

	a.strategyState = &strategyPhaseState{base: base, backupDir: backupDir}
	return nil
}

// Download renders templates and stages files to a temporary directory.
// It creates the staging directory, renders the skill template, stages
// command and config files, and checks for context cancellation.
//
// Must be called after Prepare.
func (a *BaseAdapter) Download(opts adapters.InstallOpts) error {
	if a.strategyState == nil {
		return errors.New("download: Prepare must be called first")
	}

	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	data := a.makeTemplateData()

	staging, err := os.MkdirTemp("", a.stagingPrefix)
	if err != nil {
		return fmt.Errorf("download: create staging dir: %w", err)
	}

	skillContent, err := RenderTemplate(&a.templateFS, "templates/skill.md.tmpl", data)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	if err := StageFile(staging, "SKILL.md", []byte(skillContent)); err != nil {
		return fmt.Errorf("download: stage skill: %w", err)
	}

	for _, cmd := range CommandFiles() {
		content, err := CommandFS.ReadFile("templates/commands/" + cmd)
		if err != nil {
			return fmt.Errorf("download: read command %q: %w", cmd, err)
		}
		if err := StageFile(staging, cmd, content); err != nil {
			return fmt.Errorf("download: stage command %q: %w", cmd, err)
		}
	}

	configContent, err := ConfigDefaultFS.ReadFile("templates/.sequoia-dev.yaml.default")
	if err != nil {
		return fmt.Errorf("download: read config default: %w", err)
	}
	if err := StageFileIfNotExist(a.strategyState.base, ".sequoia-dev.yaml", configContent); err != nil {
		return fmt.Errorf("download: stage config: %w", err)
	}

	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	a.strategyState.stagingDir = staging
	a.strategyState.data = data
	return nil
}

// Verify checks that the staging directory created by Download exists
// and is accessible. This phase exists so consumers that only need to
// validate staged content can run verification independently.
//
// Must be called after Download.
func (a *BaseAdapter) Verify() error {
	if a.strategyState == nil || a.strategyState.stagingDir == "" {
		return errors.New("verify: Download must be called first")
	}
	if _, err := os.Stat(a.strategyState.stagingDir); err != nil {
		return fmt.Errorf("verify: staging dir: %w", err)
	}
	return nil
}

// Stage deploys skill and command files from the staging directory to
// their target directories using the common.Installer framework.
// It checks context cancellation between skill and command installation
// and rolls back on cancellation.
//
// Must be called after Download (and optionally Verify).
func (a *BaseAdapter) Stage(opts adapters.InstallOpts) error {
	ss := a.strategyState
	if ss == nil || ss.stagingDir == "" {
		return errors.New("stage: Download must be called first")
	}

	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("stage: %w", err)
	}

	skillsDir := a.paths.skillsPathFn(ss.base)
	commandsDir := a.paths.commandsPathFn(ss.base)

	skillInstaller := NewInstaller(InstallerConfig{
		SourceDir: ss.stagingDir,
		TargetDir: skillsDir,
		BackupDir: filepath.Join(ss.backupDir, "skills"),
		Files:     []string{"SKILL.md"},
	})
	if err := skillInstaller.Run(); err != nil {
		return fmt.Errorf("stage: skill: %w", err)
	}
	ss.skillInstaller = skillInstaller

	if err := checkContext(opts.Context); err != nil {
		_ = skillInstaller.Rollback()
		return fmt.Errorf("stage: %w", err)
	}

	cmdInstaller := NewInstaller(InstallerConfig{
		SourceDir: ss.stagingDir,
		TargetDir: commandsDir,
		BackupDir: filepath.Join(ss.backupDir, "commands"),
		Files:     CommandFiles(),
	})
	if err := cmdInstaller.Run(); err != nil {
		_ = skillInstaller.Rollback()
		return fmt.Errorf("stage: commands: %w", err)
	}
	ss.cmdInstaller = cmdInstaller

	if err := checkContext(opts.Context); err != nil {
		_ = cmdInstaller.Rollback()
		_ = skillInstaller.Rollback()
		return fmt.Errorf("stage: %w", err)
	}

	return nil
}

// Apply renders and writes the system prompt and version marker file.
// If the system prompt write fails and rollback-on-error is enabled,
// previously staged skill and command files are rolled back.
//
// Must be called after Stage.
func (a *BaseAdapter) Apply(opts adapters.InstallOpts) error {
	ss := a.strategyState
	if ss == nil || ss.cmdInstaller == nil {
		return errors.New("apply: Stage must be called first")
	}

	sectionContent, err := RenderTemplate(&a.templateFS, a.systemPromptTemplate, ss.data)
	if err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	if err := a.prompt.Write(ss.base, sectionContent); err != nil {
		if a.prompt.RollbackOnError() {
			_ = ss.cmdInstaller.Rollback()
			_ = ss.skillInstaller.Rollback()
		}
		return fmt.Errorf("apply: system prompt: %w", err)
	}

	if err := checkContext(opts.Context); err != nil {
		_ = ss.cmdInstaller.Rollback()
		_ = ss.skillInstaller.Rollback()
		return fmt.Errorf("apply: %w", err)
	}

	if err := AtomicWriteFile(a.paths.versionFilePathFn(ss.base), []byte(Version+"\n"), 0o644); err != nil {
		return fmt.Errorf("apply: write version file: %w", err)
	}

	return nil
}

// Rollback undoes the effects of Stage by rolling back both the skill
// and command installers. It also cleans up the staging directory.
// Safe to call even if Stage was never called (returns nil).
func (a *BaseAdapter) Rollback() error {
	ss := a.strategyState
	if ss == nil {
		return nil
	}

	var errs []error
	if ss.cmdInstaller != nil {
		if err := ss.cmdInstaller.Rollback(); err != nil {
			errs = append(errs, fmt.Errorf("rollback: commands: %w", err))
		}
	}
	if ss.skillInstaller != nil {
		if err := ss.skillInstaller.Rollback(); err != nil {
			errs = append(errs, fmt.Errorf("rollback: skill: %w", err))
		}
	}
	if ss.stagingDir != "" {
		_ = os.RemoveAll(ss.stagingDir)
	}

	return errors.Join(errs...)
}

// Install installs Sequoia files using the phased Strategy lifecycle.
// This is a convenience wrapper that calls Prepare → Download → Verify →
// Stage → Apply in sequence, preserving backward compatibility with
// existing callers.
//
// When opts.Context is set and cancelled, each phase aborts early and
// rolls back any partial work before returning the context error.
//
// On failure, the returned error wraps adapters.ErrInstallFailed so callers
// can detect install failures with errors.Is(err, adapters.ErrInstallFailed).
func (a *BaseAdapter) Install(opts adapters.InstallOpts) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", adapters.ErrInstallFailed, err)
		}
		// Clean up staging directory if it still exists.
		if a.strategyState != nil {
			if a.strategyState.stagingDir != "" {
				_ = os.RemoveAll(a.strategyState.stagingDir)
			}
			a.strategyState = nil
		}
	}()

	if err := a.Prepare(opts); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := a.Download(opts); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := a.Verify(); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := a.Stage(opts); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if err := a.Apply(opts); err != nil {
		return fmt.Errorf("install: %w", err)
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
	for _, cmd := range CommandFiles() {
		if err := os.Remove(filepath.Join(a.paths.commandsPathFn(base), cmd)); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove command %s: %w", cmd, err))
		}
	}

	// Remove config files
	if err := os.Remove(filepath.Join(base, ".sequoia-dev.yaml")); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove config .sequoia-dev.yaml: %w", err))
	}

	// Remove or restore the system prompt.
	if err := a.prompt.Remove(base); err != nil {
		errs = append(errs, fmt.Errorf("restore system prompt: %w", err))
	}

	return errors.Join(errs...)
}
