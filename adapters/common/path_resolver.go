package common

import (
	"errors"
	"os"
	"sync"
)

// PathResolver resolves tool config directory paths from the user home
// directory. It encapsulates home directory resolution (with sync.Once
// caching), symlink resolution, and the 5 path functions that map a base
// directory to tool-specific subdirectories.
//
// PathResolver is constructed once per adapter and SHOULD NOT be modified
// after construction except via SetHomeDir() (used for testing).
//
// Warnings produced during symlink resolution are reported via the warnFn
// callback. If warnFn is nil, warnings are silently discarded.
type PathResolver struct {
	resolveBase        func(homeDir string) (string, error)
	skillsPathFn       func(base string) string
	commandsPathFn     func(base string) string
	systemPromptPathFn func(base string) string
	versionFilePathFn  func(base string) string
	backupPathFn       func(base string) string

	// homeDir overrides os.UserHomeDir(). When empty AND no explicit
	// override via SetHomeDir, os.UserHomeDir() is called and cached.
	homeDir string

	// Caching for os.UserHomeDir() when homeDir is empty.
	cachedHomeOnce sync.Once
	cachedHomeDir  string
	cachedHomeErr  error

	// warnFn is called with non-fatal warning messages (e.g., symlink
	// resolution issues). If nil, warnings are discarded. The callback
	// is responsible for any necessary synchronization.
	warnFn func(msg string)
}

// NewPathResolver creates a PathResolver with the given configuration.
//
// Parameters:
//   - resolveBase: maps homeDir → tool config root (e.g. ~/.claude/)
//   - homeDir: explicit home directory override (empty = use os.UserHomeDir)
//   - skillsPathFn, commandsPathFn, systemPromptPathFn, versionFilePathFn,
//     backupPathFn: map base → tool-specific paths
//   - warnFn: optional callback for non-fatal warnings (nil = discard warnings)
func NewPathResolver(
	resolveBase func(homeDir string) (string, error),
	homeDir string,
	skillsPathFn func(base string) string,
	commandsPathFn func(base string) string,
	systemPromptPathFn func(base string) string,
	versionFilePathFn func(base string) string,
	backupPathFn func(base string) string,
	warnFn func(string),
) *PathResolver {
	return &PathResolver{
		resolveBase:        resolveBase,
		homeDir:            homeDir,
		skillsPathFn:       skillsPathFn,
		commandsPathFn:     commandsPathFn,
		systemPromptPathFn: systemPromptPathFn,
		versionFilePathFn:  versionFilePathFn,
		backupPathFn:       backupPathFn,
		warnFn:             warnFn,
	}
}

// HomeDir returns the currently configured home directory. An empty string
// means os.UserHomeDir() will be used at Base() resolution time.
func (p *PathResolver) HomeDir() string {
	return p.homeDir
}

// SetHomeDir overrides the home directory. After calling SetHomeDir, the
// next call to Base() will use the new directory and skip os.UserHomeDir().
// This is primarily used for testing.
func (p *PathResolver) SetHomeDir(dir string) {
	p.homeDir = dir
}

// Base resolves and returns the tool's config root directory. If homeDir
// was set explicitly (non-empty), it is used directly. Otherwise,
// os.UserHomeDir() is called once and cached via sync.Once.
//
// The home directory is resolved via ResolveSymlink before being passed
// to resolveBase. Any symlink warnings are appended to the shared warnings
// slice.
func (p *PathResolver) Base() (string, error) {
	homeDir := p.homeDir
	if homeDir == "" {
		p.cachedHomeOnce.Do(func() {
			p.cachedHomeDir, p.cachedHomeErr = os.UserHomeDir()
		})
		if p.cachedHomeErr != nil {
			return "", p.cachedHomeErr
		}
		homeDir = p.cachedHomeDir
	}

	resolved, warning := ResolveSymlink(homeDir)
	if warning != "" {
		p.addWarning(warning)
	}

	return p.resolveBase(resolved)
}

// SkillsPath returns the absolute path to the skills directory.
// Returns "" if Base() fails.
func (p *PathResolver) SkillsPath() string {
	base, err := p.Base()
	if err != nil {
		return ""
	}
	return p.skillsPathFn(base)
}

// CommandsPath returns the absolute path to the commands directory.
// Returns "" if Base() fails.
func (p *PathResolver) CommandsPath() string {
	base, err := p.Base()
	if err != nil {
		return ""
	}
	return p.commandsPathFn(base)
}

// SystemPromptPath returns the absolute path to the system prompt file.
// Returns "" if Base() fails.
func (p *PathResolver) SystemPromptPath() string {
	base, err := p.Base()
	if err != nil {
		return ""
	}
	return p.systemPromptPathFn(base)
}

// addWarning reports a non-fatal warning via the warnFn callback.
// If warnFn is nil, the warning is silently discarded.
func (p *PathResolver) addWarning(msg string) {
	if p.warnFn != nil {
		p.warnFn(msg)
	}
}

// ErrHomeUnavailable is returned when os.UserHomeDir() fails and no
// explicit homeDir was set via SetHomeDir().
var ErrHomeUnavailable = errors.New("home directory unavailable")
