package common_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/adapters/common/installembed"
)

// =========================================================================
// Task 7.1 — Mock FS template rendering tests
//
// These tests use installembed.FS (an embed.FS with "templates/skill.md.tmpl"
// and "templates/sys_prompt.md.tmpl") to verify that RenderTemplate produces
// correct output with mock data, and returns an error when the template file
// is missing from the mock FS.
//
// NOTE: testing/fstest.MapFS cannot be used directly because RenderTemplate
// accepts *embed.FS, not fs.FS. The installembed package provides an
// embed.FS as a workaround. If RenderTemplate is generalized to accept
// fs.FS in a future refactor, these tests can switch to MapFS.
// =========================================================================

// TestMockFS_RenderSkillTemplate verifies REQ-TEST-MOCKFS
// Scenario "Install with mock filesystem — template rendering succeeds".
// Renders the skill template from installembed.FS with mock data and
// asserts the output contains the expected mock content.
func TestMockFS_RenderSkillTemplate(t *testing.T) {
	t.Parallel()

	result, err := common.RenderTemplate(&installembed.FS,
		"templates/skill.md.tmpl",
		map[string]string{"Name": "MockAdapter"})
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	assert.Contains(t, result, "MockAdapter",
		"rendered skill template should contain mock data")
	assert.Contains(t, result, "skill template for testing",
		"rendered skill template should contain template content")
}

// TestMockFS_RenderSystemPromptTemplate triangulates: renders the system
// prompt template and verifies it produces different output from the skill
// template.
func TestMockFS_RenderSystemPromptTemplate(t *testing.T) {
	t.Parallel()

	result, err := common.RenderTemplate(&installembed.FS,
		"templates/sys_prompt.md.tmpl",
		map[string]string{"Name": "SysPromptTest"})
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	assert.Contains(t, result, "SysPromptTest",
		"rendered system prompt should contain mock data")
	assert.Contains(t, result, "system prompt for",
		"rendered system prompt should contain template content")
}

// TestMockFS_RenderSystemPrompt_DiffersFromSkill verifies that skill and
// system prompt templates produce distinct output (different adapter
// templates produce different content per REQ-TEST-MOCKFS).
func TestMockFS_RenderSystemPrompt_DiffersFromSkill(t *testing.T) {
	t.Parallel()

	data := map[string]string{"Name": "DiffTest"}

	skillResult, err := common.RenderTemplate(&installembed.FS,
		"templates/skill.md.tmpl", data)
	require.NoError(t, err)

	sysResult, err := common.RenderTemplate(&installembed.FS,
		"templates/sys_prompt.md.tmpl", data)
	require.NoError(t, err)

	assert.NotEqual(t, skillResult, sysResult,
		"skill and system prompt templates must produce different output")
}

// TestMockFS_TemplateNotFound verifies REQ-TEST-MOCKFS
// "template file is missing from mock FS → returns error".
func TestMockFS_TemplateNotFound(t *testing.T) {
	t.Parallel()

	_, err := common.RenderTemplate(&installembed.FS,
		"templates/does_not_exist.md.tmpl", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read template",
		"error should mention template reading failure")
}

// =========================================================================
// Task 7.3 — Full Install pipeline with mock filesystem (integration-style)
//
// TestMockFS_InstallFullPipeline verifies the complete pipeline using
// installembed.FS: staging → skill install → command install → system
// prompt → version file. This tests that the refactored BaseAdapter's
// Install method works correctly end-to-end with a mock filesystem.
// =========================================================================

// TestMockFS_InstallFullPipeline verifies REQ-TEST-MOCKFS
// Scenario "Install with mock filesystem — complete pipeline succeeds".
// Uses fullInstallTestAdapter which is configured with installembed.FS and
// verifies every expected output file is created with correct content.
func TestMockFS_InstallFullPipeline(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)

	err := a.Install(adapters.InstallOpts{})
	require.NoError(t, err, "full pipeline install should succeed with mock FS")

	// Verify skill file was installed.
	skillsDir := a.SkillsPath()
	assert.NotEmpty(t, skillsDir)
	assert.FileExists(t, filepath.Join(skillsDir, "SKILL.md"),
		"SKILL.md should exist after successful install")

	// Verify command files were installed.
	cmdsDir := a.CommandsPath()
	assert.NotEmpty(t, cmdsDir)
	for _, cmd := range common.CommandFiles {
		assert.FileExists(t, filepath.Join(cmdsDir, cmd),
			"command %s should exist after install", cmd)
	}

	// Verify system prompt was written (write function is a no-op mock
	// in fullInstallTestAdapter, so we check the file based on path setup).
	sysPath := a.SystemPromptPath()
	assert.NotEmpty(t, sysPath)

	// Verify version file was written.
	versionPath := filepath.Join(home, "version")
	assert.FileExists(t, versionPath, "version file should exist after install")

	// Verify version file content is the current version.
	data, err := os.ReadFile(versionPath)
	require.NoError(t, err)
	versionContent := strings.TrimSpace(string(data))
	assert.Equal(t, common.Version, versionContent,
		"version file should contain the current Sequoia version")
}

// TestMockFS_InstallFullPipeline_StatusReportsInstalled verifies that
// Status() reports the adapter as installed after a successful Install.
func TestMockFS_InstallFullPipeline_StatusReportsInstalled(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	a := fullInstallTestAdapter(t, home)

	err := a.Install(adapters.InstallOpts{})
	require.NoError(t, err)

	status := a.Status()
	assert.True(t, status.Installed,
		"Status should report installed after successful Install")
	assert.Equal(t, common.Version, status.Version,
		"Status should report correct version after Install")
	assert.Contains(t, status.Path, "skills",
		"Status path should point to skills directory")
}
