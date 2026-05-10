package opencode_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sequoia-ai/adapters/common"
	"sequoia-ai/adapters/opencode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall_CreatesAllFiles(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install())

	assert.FileExists(t, filepath.Join(a.SkillsPath(), "SKILL.md"))

	for _, cmd := range []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-fix.md",
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

	require.NoError(t, a.Install())
	assert.True(t, a.IsInstalled())
}

func TestInstall_SkillContainsVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install())

	raw, err := os.ReadFile(filepath.Join(a.SkillsPath(), "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), common.Version)
}

func TestInstall_AgentsMDHasMarkers(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install())

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

	require.NoError(t, a.Install())
	require.NoError(t, a.Install())
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
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	// Write existing content to AGENTS.md before installing.
	agentsMDPath := a.SystemPromptPath()
	require.NoError(t, os.MkdirAll(filepath.Dir(agentsMDPath), 0o755))
	originalContent := "# Existing\n"
	require.NoError(t, os.WriteFile(agentsMDPath, []byte(originalContent), 0o644))

	require.NoError(t, a.Install())

	// Backup should contain the original content.
	backupPath := agentsMDPath + ".sequoia-backup"
	raw, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(raw))
}

func TestUninstall_RemovesAllFiles(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install())
	require.NoError(t, a.Uninstall())

	assert.NoFileExists(t, filepath.Join(a.SkillsPath(), "SKILL.md"))

	for _, cmd := range []string{
		"sequoia-init.md",
		"sequoia-audit.md",
		"sequoia-review.md",
		"sequoia-fix.md",
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

	require.NoError(t, a.Install())
	require.NoError(t, a.Uninstall())

	assert.False(t, a.IsInstalled())
}

func TestUninstall_PreservesOtherAgentsMD(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	agentsMDPath := a.SystemPromptPath()
	require.NoError(t, os.MkdirAll(filepath.Dir(agentsMDPath), 0o755))
	originalContent := "# My content\n"
	require.NoError(t, os.WriteFile(agentsMDPath, []byte(originalContent), 0o644))

	require.NoError(t, a.Install())
	require.NoError(t, a.Uninstall())

	raw, err := os.ReadFile(agentsMDPath)
	require.NoError(t, err)
	assert.Equal(t, originalContent, string(raw))
}

func TestStatus_AfterInstall(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install())

	status := a.Status()
	assert.True(t, status.Installed)
	assert.NotEmpty(t, status.Path)
}

func TestVerify_AllFilesReadable(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	a := opencode.NewAdapter(tmp)

	require.NoError(t, a.Install())

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
		"sequoia-fix.md",
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
