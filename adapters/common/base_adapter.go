package common

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Crisbr10/sequoia/adapters"
)

// BaseAdapter provides shared Install, Uninstall, Status, and path methods
// for tool adapters. Concrete adapters embed BaseAdapter and set the function
// fields to customize path resolution, detection, and system prompt handling.
//
// The Install/Uninstall flow follows the same 8-step pattern for all adapters;
// only the system prompt strategy and path layout differ.
type BaseAdapter struct {
	// adapterID is the unique machine-readable identifier (e.g. "claude-code").
	adapterID string
	// adapterName is the human-readable display name (e.g. "Claude Code").
	adapterName string
	// homeDir overrides os.UserHomeDir() for testing.
	homeDir string

	// resolveBase returns the tool's config root directory from the user home.
	// For example: ~/.claude/, ~/.gemini/, ~/.config/opencode/.
	resolveBase func(homeDir string) (string, error)

	// Path functions — each adapter provides its own directory layout.
	skillsPathFn       func(base string) string
	commandsPathFn     func(base string) string
	systemPromptPathFn func(base string) string
	versionFilePathFn  func(base string) string
	backupPathFn       func(base string) string

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
}

// SetIDName sets the adapter's unique ID and human-readable name.
func (a *BaseAdapter) SetIDName(id, name string) {
	a.adapterID = id
	a.adapterName = name
}

// SetHomeDir overrides the user home directory (for testing).
func (a *BaseAdapter) SetHomeDir(dir string) {
	a.homeDir = dir
}

// HomeDir returns the current home directory override (empty string means
// production — os.UserHomeDir() is used by the resolveBase function).
func (a *BaseAdapter) HomeDir() string {
	return a.homeDir
}

// ID returns the unique machine-readable identifier.
func (a *BaseAdapter) ID() string { return a.adapterID }

// Name returns the human-readable display name.
func (a *BaseAdapter) Name() string { return a.adapterName }

// ResolveBase sets the base directory resolution function.
func (a *BaseAdapter) ResolveBase(fn func(homeDir string) (string, error)) {
	a.resolveBase = fn
}

// SetPathFns sets all five path functions at once.
func (a *BaseAdapter) SetPathFns(skills, commands, systemPrompt, versionFile, backup func(base string) string) {
	a.skillsPathFn = skills
	a.commandsPathFn = commands
	a.systemPromptPathFn = systemPrompt
	a.versionFilePathFn = versionFile
	a.backupPathFn = backup
}

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

// base resolves and returns the tool's config root directory.
func (a *BaseAdapter) base() (string, error) {
	return a.resolveBase(a.homeDir)
}

// SkillsPath returns the absolute path to the skills directory.
func (a *BaseAdapter) SkillsPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return a.skillsPathFn(base)
}

// CommandsPath returns the absolute path to the commands directory.
func (a *BaseAdapter) CommandsPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return a.commandsPathFn(base)
}

// SystemPromptPath returns the absolute path to the system prompt file.
func (a *BaseAdapter) SystemPromptPath() string {
	base, err := a.base()
	if err != nil {
		return ""
	}
	return a.systemPromptPathFn(base)
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
	base, err := a.base()
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
		base, err := a.base()
		if err == nil {
			data, err := os.ReadFile(a.versionFilePathFn(base))
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
func (a *BaseAdapter) Install(opts adapters.InstallOpts) error {
	// Check for early cancellation before doing any work.
	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	_ = opts.Language

	base, err := a.base()
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
	skillContent, err := RenderTemplate(a.templateFS, "templates/skill.md.tmpl", data)
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

	skillsDir := a.skillsPathFn(base)
	commandsDir := a.commandsPathFn(base)
	// Append a unique session suffix to the backup dir to avoid name collisions
	// with pre-existing directories.
	sessionSuffix := strconv.FormatInt(time.Now().UnixMilli(), 36)
	backupDir := a.backupPathFn(base) + "-" + sessionSuffix

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
		BackupDir: backupDir,
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
		BackupDir: backupDir,
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
	sectionContent, err := RenderTemplate(a.templateFS, a.systemPromptTemplate, data)
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
	if err := os.WriteFile(a.versionFilePathFn(base), []byte(Version+"\n"), 0o644); err != nil {
		return fmt.Errorf("install: write version file: %w", err)
	}

	return nil
}

// Uninstall removes Sequoia files using the common pattern.
// The system prompt strategy is delegated to removeSystemPrompt.
//
// When opts.Context is set and cancelled, Uninstall returns early without
// modifying any files.
func (a *BaseAdapter) Uninstall(opts adapters.InstallOpts) error {
	// Check for early cancellation before doing any work.
	if err := checkContext(opts.Context); err != nil {
		return fmt.Errorf("uninstall: %w", err)
	}

	_ = opts.Language

	base, err := a.base()
	if err != nil {
		return fmt.Errorf("uninstall: resolve home: %w", err)
	}

	// Remove skill file, version marker, and command files (best-effort).
	_ = os.Remove(filepath.Join(a.skillsPathFn(base), "SKILL.md"))
	_ = os.Remove(a.versionFilePathFn(base))
	for _, cmd := range CommandFiles {
		_ = os.Remove(filepath.Join(a.commandsPathFn(base), cmd))
	}

	// Remove or restore the system prompt.
	return a.removeSystemPrompt(base)
}
