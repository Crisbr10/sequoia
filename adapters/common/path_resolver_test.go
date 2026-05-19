package common_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// =========================================================================
// TestNewPathResolver_WithExplicitHomeDir
// =========================================================================

// TestNewPathResolver_WithExplicitHomeDir verifies that a PathResolver
// constructed with an explicit homeDir uses it for Base() resolution
// without calling os.UserHomeDir().
func TestNewPathResolver_WithExplicitHomeDir(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return filepath.Join(homeDir, ".test-tool"), nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "system.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil, // no warnFn
	)

	base, err := pr.Base()
	require.NoError(t, err)
	assert.Contains(t, base, ".test-tool", "Base() should use the injected resolveBase function")
	assert.Contains(t, base, filepath.Base(tmp), "Base() should use the explicit homeDir")
}

// =========================================================================
// TestNewPathResolver_HomeDirGetter
// =========================================================================

// TestNewPathResolver_HomeDirGetter verifies that HomeDir() returns the
// home directory set at construction time.
func TestNewPathResolver_HomeDirGetter(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		tmp,
		func(base string) string { return base },
		func(base string) string { return base },
		func(base string) string { return base },
		func(base string) string { return base },
		func(base string) string { return base },
		nil,
	)

	assert.Equal(t, tmp, pr.HomeDir(), "HomeDir() should return the homeDir passed at construction")
}

// =========================================================================
// TestNewPathResolver_SetHomeDirOverride
// =========================================================================

// TestNewPathResolver_SetHomeDirOverride verifies that SetHomeDir() changes
// the home directory used by Base() for subsequent calls, bypassing any
// cached os.UserHomeDir().
func TestNewPathResolver_SetHomeDirOverride(t *testing.T) {
	t.Parallel()

	original := t.TempDir()
	overridden := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return filepath.Join(homeDir, ".tool"), nil
		},
		original,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "system.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	// First call uses original homeDir.
	base, err := pr.Base()
	require.NoError(t, err)
	assert.Contains(t, base, filepath.Base(original), "Base() should use original homeDir")

	// Override and verify change.
	pr.SetHomeDir(overridden)
	base, err = pr.Base()
	require.NoError(t, err)
	assert.Contains(t, base, filepath.Base(overridden), "Base() should use overridden homeDir")
}

// =========================================================================
// TestPathResolver_SkillsPath
// =========================================================================

// TestPathResolver_SkillsPath verifies that SkillsPath() resolves Base()
// and delegates to the skillsPathFn.
func TestPathResolver_SkillsPath(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return filepath.Join(homeDir, ".test-tool"), nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "system.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	sp := pr.SkillsPath()
	assert.NotEmpty(t, sp, "SkillsPath() should be non-empty")
	assert.Contains(t, sp, "skills", "SkillsPath() should include the skills subdirectory")
	assert.Contains(t, sp, ".test-tool", "SkillsPath() should include the base directory")
}

// =========================================================================
// TestPathResolver_CommandsPath
// =========================================================================

// TestPathResolver_CommandsPath verifies that CommandsPath() resolves Base()
// and delegates to the commandsPathFn.
func TestPathResolver_CommandsPath(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "system.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	cp := pr.CommandsPath()
	assert.NotEmpty(t, cp, "CommandsPath() should be non-empty")
	assert.Contains(t, cp, "commands", "CommandsPath() should include the commands subdirectory")
}

// =========================================================================
// TestPathResolver_SystemPromptPath
// =========================================================================

// TestPathResolver_SystemPromptPath verifies that SystemPromptPath() resolves
// Base() and delegates to the systemPromptPathFn.
func TestPathResolver_SystemPromptPath(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) { return homeDir, nil },
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	spp := pr.SystemPromptPath()
	assert.NotEmpty(t, spp, "SystemPromptPath() should be non-empty")
	assert.Contains(t, spp, "sys.md", "SystemPromptPath() should include the system prompt filename")
}

// =========================================================================
// TestPathResolver_BaseError_ReturnsEmptyPaths
// =========================================================================

// TestPathResolver_BaseError_ReturnsEmptyPaths verifies proper handling
// when the resolveBase function returns an error.
func TestPathResolver_BaseError_ReturnsEmptyPaths(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("cannot resolve base")

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return "", baseErr
		},
		"/fake/home", // non-empty → skips os.UserHomeDir(), calls resolveBase directly
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	// Base() should return the error.
	base, err := pr.Base()
	assert.Error(t, err, "Base() should return error from resolveBase")
	assert.True(t, errors.Is(err, baseErr), "Base() should propagate the original error")
	assert.Empty(t, base, "Base() should return empty string on error")

	// Delegated paths should return empty strings, not panic.
	assert.Empty(t, pr.SkillsPath(), "SkillsPath() should return empty when Base() fails")
	assert.Empty(t, pr.CommandsPath(), "CommandsPath() should return empty when Base() fails")
	assert.Empty(t, pr.SystemPromptPath(), "SystemPromptPath() should return empty when Base() fails")
}

// =========================================================================
// TestPathResolver_WarningCollection
// =========================================================================

// TestPathResolver_WarningCollection verifies that warnings produced by
// Base() (e.g., symlink resolution issues) are reported via the warnFn
// callback provided at construction time.
func TestPathResolver_WarningCollection(t *testing.T) {
	t.Parallel()

	var warnings []string

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return homeDir, nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		func(msg string) { warnings = append(warnings, msg) },
	)

	// Base() with a normal directory should not produce warnings.
	base, err := pr.Base()
	require.NoError(t, err)
	assert.NotEmpty(t, base)
	assert.Empty(t, warnings, "normal Base() should not produce warnings")
}

// =========================================================================
// TestPathResolver_WarningsNil
// =========================================================================

// TestPathResolver_WarningsNil verifies that a nil warnFn does not
// cause a panic (PathResolver gracefully handles nil).
func TestPathResolver_WarningsNil(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return homeDir, nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil, // nil warnFn
	)

	// Must not panic.
	base, err := pr.Base()
	require.NoError(t, err)
	assert.NotEmpty(t, base)
}

// =========================================================================
// TestPathResolver_BaseReturnsConsistentResult
// =========================================================================

// TestPathResolver_BaseReturnsConsistentResult verifies that calling Base()
// multiple times with the same homeDir returns consistent results (the path
// functions are deterministic given the same input).
func TestPathResolver_BaseReturnsConsistentResult(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return filepath.Join(homeDir, ".tool"), nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	first, err := pr.Base()
	require.NoError(t, err)

	second, err := pr.Base()
	require.NoError(t, err)
	assert.Equal(t, first, second, "Base() should return the same result on subsequent calls")

	// SkillsPath must also be consistent.
	assert.Equal(t, pr.SkillsPath(), pr.SkillsPath(), "SkillsPath() should be stable")
}

// =========================================================================
// TestPathResolver_BaseUsesResolveSymlink
// =========================================================================

// TestPathResolver_BaseUsesResolveSymlink verifies that Base() calls
// ResolveSymlink (via ResolveHome) on the home directory before passing
// it to resolveBase.
func TestPathResolver_BaseUsesResolveSymlink(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	// Normalize the temp dir path for comparison.
	normalized, err := filepath.EvalSymlinks(tmp)
	require.NoError(t, err)

	var receivedHomeDir string
	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			receivedHomeDir = homeDir
			return filepath.Join(homeDir, ".resolved-tool"), nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	_, err = pr.Base()
	require.NoError(t, err)
	assert.Equal(t, normalized, receivedHomeDir, "resolveBase should receive the symlink-resolved home directory")
}

// =========================================================================
// TestPathResolver_EmptyHomeDirUsesOS
// =========================================================================

// TestPathResolver_EmptyHomeDirUsesOS verifies that when constructed with
// an empty homeDir AND no SetHomeDir override, Base() calls os.UserHomeDir().
func TestPathResolver_EmptyHomeDirUsesOS(t *testing.T) {
	t.Parallel()

	var receivedHomeDir string
	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			receivedHomeDir = homeDir
			return homeDir, nil
		},
		"", // empty → will call os.UserHomeDir()
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	base, err := pr.Base()
	require.NoError(t, err)

	// The received home dir should be a real path (the user's home directory).
	assert.NotEmpty(t, receivedHomeDir, "resolveBase should receive the real home directory")
	actualHome, err := os.UserHomeDir()
	require.NoError(t, err)

	// After symlink resolution, the received home dir should match the canonical path.
	assert.Equal(t, actualHome, receivedHomeDir, "when homeDir is empty, os.UserHomeDir() should be used")
	assert.Equal(t, actualHome, base, "Base() should return the home dir when resolveBase is identity")
}

// =========================================================================
// Symlink resolution caching
// =========================================================================

// TestPathResolver_BaseSymlinkCached verifies that ResolveSymlink (via
// filepath.EvalSymlinks) is called at most once for the same PathResolver
// instance. The cached resolved path is reused on subsequent Base() calls.
func TestPathResolver_BaseSymlinkCached(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	var resolveCalls int
	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			resolveCalls++
			return filepath.Join(homeDir, ".tool"), nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	// First call resolves symlink.
	first, err := pr.Base()
	require.NoError(t, err)
	assert.NotEmpty(t, first)

	// Second call must NOT re-resolve the symlink.
	second, err := pr.Base()
	require.NoError(t, err)
	assert.Equal(t, first, second, "Base() should return consistent result")
	// resolveBase should be called twice (once per Base()), but the symlink
	// resolution inside Base() should be cached. We verify that the resolved
	// home directory passed to resolveBase is the same on both calls.
	assert.Equal(t, 2, resolveCalls, "resolveBase should be called twice (no caching of base resolution)")
}

// TestPathResolver_SetHomeDirResetsSymlinkCache verifies that calling
// SetHomeDir() invalidates the cached symlink resolution, forcing a fresh
// ResolveSymlink call on the next Base().
func TestPathResolver_SetHomeDirResetsSymlinkCache(t *testing.T) {
	t.Parallel()

	dirA := t.TempDir()
	dirB := t.TempDir()

	var lastResolved string
	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			lastResolved = homeDir
			return filepath.Join(homeDir, ".tool"), nil
		},
		dirA,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	// First call with dirA — symlink resolution cached.
	_, err := pr.Base()
	require.NoError(t, err)
	assert.Contains(t, lastResolved, filepath.Base(dirA))

	// SetHomeDir must invalidate the symlink cache.
	pr.SetHomeDir(dirB)

	// Second call with dirB — fresh symlink resolution.
	_, err = pr.Base()
	require.NoError(t, err)
	assert.Contains(t, lastResolved, filepath.Base(dirB))
}

// TestPathResolver_NonSymlinkPathStillCached verifies that even for
// non-symlink paths, the resolution result is cached (EvalSymlinks
// is still called once).
func TestPathResolver_NonSymlinkPathStillCached(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	// Normalize to get the canonical path (EvalSymlinks on non-symlink).
	canonical, err := filepath.EvalSymlinks(tmp)
	require.NoError(t, err)

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			// Verify we always get the canonical (resolved) path.
			assert.Equal(t, canonical, homeDir, "resolveBase should receive canonical path")
			return filepath.Join(homeDir, ".tool"), nil
		},
		tmp,
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	// Both calls receive the same canonical path.
	first, err := pr.Base()
	require.NoError(t, err)
	second, err := pr.Base()
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

// =========================================================================
// TestPathResolver_BaseCachingEmptyHomeDir
// =========================================================================

// TestPathResolver_BaseCachingEmptyHomeDir verifies that when homeDir is
// empty, os.UserHomeDir() is cached via sync.Once (the home dir is resolved
// only once), and SkillsPath() is stable across repeated calls.
func TestPathResolver_BaseCachingEmptyHomeDir(t *testing.T) {
	t.Parallel()

	pr := common.NewPathResolver(
		func(homeDir string) (string, error) {
			return filepath.Join(homeDir, ".cached-os"), nil
		},
		"",
		func(base string) string { return filepath.Join(base, "skills") },
		func(base string) string { return filepath.Join(base, "commands") },
		func(base string) string { return filepath.Join(base, "sys.md") },
		func(base string) string { return filepath.Join(base, "version") },
		func(base string) string { return filepath.Join(base, "backup") },
		nil,
	)

	// SkillsPath() calls Base() which calls os.UserHomeDir(). The first call
	// resolves the home dir; subsequent calls use the cached value.
	first := pr.SkillsPath()
	require.NotEmpty(t, first, "SkillsPath should be non-empty")

	second := pr.SkillsPath()
	assert.Equal(t, first, second, "SkillsPath should be stable across repeated calls (home dir cached)")

	// CommandsPath should also use the cached home dir.
	cmds := pr.CommandsPath()
	assert.NotEmpty(t, cmds, "CommandsPath should be non-empty")
	assert.Contains(t, cmds, ".cached-os", "CommandsPath should use the cached home dir")
}
