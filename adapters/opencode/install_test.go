//nolint:gosec // test file: all os.* operations use t.TempDir() test fixtures, not production paths
package opencode_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/adapters/opencode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall_CreatesAllFiles(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	assert.FileExists(t, filepath.Join(a.SkillsPath(), "SKILL.md"))

	for _, cmd := range []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-diff.md",
	} {
		assert.FileExists(t, filepath.Join(a.CommandsPath(), cmd))
	}

	assert.FileExists(t, a.SystemPromptPath())
}

func TestInstall_IsInstalledAfterInstall(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))
	assert.True(t, a.IsInstalled())
}

func TestInstall_SkillContainsVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	raw, err := os.ReadFile(filepath.Join(a.SkillsPath(), "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), common.Version)
}

func TestInstall_AgentsMDHasMarkers(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	raw, err := os.ReadFile(a.SystemPromptPath())
	require.NoError(t, err)
	content := string(raw)
	assert.Contains(t, content, "<!-- sequoia:start -->")
	assert.Contains(t, content, "<!-- sequoia:end -->")
}

func TestInstall_Idempotent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))
	require.NoError(t, a.Install(adapters.InstallOpts{}))
	assert.True(t, a.IsInstalled())

	raw, err := os.ReadFile(a.SystemPromptPath())
	require.NoError(t, err)
	content := string(raw)

	// Markers should appear exactly once each.
	assert.Equal(t, 1, strings.Count(content, "<!-- sequoia:start -->"),
		"sequoia:start marker should appear exactly once")
	assert.Equal(t, 1, strings.Count(content, "<!-- sequoia:end -->"),
		"sequoia:end marker should appear exactly once")
}

func TestInstall_PreservesExistingAgentsMD(t *testing.T) {
	// PR 3: the backup moved to the central backup home and is now a
	// <sessionDir>/AGENTS.md.backup inside <root>/opencode/. This test
	// is re-enabled by the central-home round-trip test in
	// adapters/common/strategy_central_test.go; the opencode-specific
	// check below is kept as a "smoke" assertion that the per-tool
	// directory does NOT contain a legacy sidecar.
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	// Write existing content to AGENTS.md before installing.
	agentsMDPath := a.SystemPromptPath()
	require.NoError(t, os.MkdirAll(filepath.Dir(agentsMDPath), 0o755))
	originalContent := "# Existing\n"
	require.NoError(t, os.WriteFile(agentsMDPath, []byte(originalContent), 0o644))

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	// After PR 3, the per-tool directory must NOT contain a legacy
	// .sequoia-backup-* sidecar (the backup moved to the central home).
	dir := filepath.Dir(agentsMDPath)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.False(t, strings.Contains(e.Name(), ".sequoia-backup-"),
			"per-tool dir must not contain a legacy .sequoia-backup-* sidecar (found %s)", e.Name())
	}

	// The AGENTS.md itself must be the Sequoia-managed content now.
	got, err := os.ReadFile(agentsMDPath)
	require.NoError(t, err)
	assert.NotEqual(t, originalContent, string(got),
		"AGENTS.md must be replaced with Sequoia content after install")
	_ = originalContent // originalContent is now used to assert "not equal" above; keep the variable for readability
}

func TestUninstall_RemovesAllFiles(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))
	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	assert.NoFileExists(t, filepath.Join(a.SkillsPath(), "SKILL.md"))

	for _, cmd := range []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-diff.md",
	} {
		assert.NoFileExists(t, filepath.Join(a.CommandsPath(), cmd))
	}

	assert.False(t, a.IsInstalled())
}

func TestUninstall_CleansAgentsMD(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))
	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	assert.False(t, a.IsInstalled())
}

func TestUninstall_PreservesOtherAgentsMD(t *testing.T) {
	// PR 3: the round-trip (Install + Uninstall) restores the
	// pre-install content from the central-home + manifest backup.
	// The test now exercises the full manifest-based round-trip.
	// The previous assertion (per-tool .sequoia-backup-* sidecar)
	// is no longer relevant; see adapters/common/strategy_central_test.go
	// for the canonical central-home round-trip test.
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	agentsMDPath := a.SystemPromptPath()
	require.NoError(t, os.MkdirAll(filepath.Dir(agentsMDPath), 0o755))
	originalContent := "# My content\n"
	require.NoError(t, os.WriteFile(agentsMDPath, []byte(originalContent), 0o644))

	require.NoError(t, a.Install(adapters.InstallOpts{}))
	require.NoError(t, a.Uninstall(adapters.InstallOpts{}))

	// After uninstall, the AGENTS.md must be restored to the
	// pre-install content (the manifest-based restore succeeded).
	raw, err := os.ReadFile(agentsMDPath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(raw),
		"AGENTS.md must be restored from the central-home + manifest backup after uninstall")
}

func TestStatus_AfterInstall(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	status := a.Status()
	assert.True(t, status.Installed)
	assert.NotEmpty(t, status.Path)
}

func TestVerify_AllFilesReadable(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install(adapters.InstallOpts{}))

	// Skill file.
	skillPath := filepath.Join(a.SkillsPath(), "SKILL.md")
	assert.FileExists(t, skillPath)
	raw, err := os.ReadFile(skillPath)
	require.NoError(t, err)
	assert.NotEmpty(t, raw)

	// All command files.
	for _, cmd := range []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-diff.md",
	} {
		cmdPath := filepath.Join(a.CommandsPath(), cmd)
		assert.FileExists(t, cmdPath)
		raw, err := os.ReadFile(cmdPath)
		require.NoError(t, err)
		assert.NotEmpty(t, raw, "command file %s should not be empty", cmd)
	}

	// AGENTS.md.
	raw, err = os.ReadFile(a.SystemPromptPath())
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
}
