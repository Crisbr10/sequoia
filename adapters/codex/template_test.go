// Package codex provides white-box tests for OpenAI Codex template rendering.
// These tests access the unexported templateFS to verify Codex-specific
// template content (TOML format, path placeholders) is rendered correctly.
package codex

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// TestTemplateRendering_SkillTemplate verifies REQ-TEST-MOCKFS
// "Codex adapter with mockFS produces Codex-specific content".
// Renders the Codex skill template with mock data and asserts that
// Codex-specific content (agent roster format, C0 orchestrator title)
// appears in the output.
func TestTemplateRendering_SkillTemplate(t *testing.T) {
	t.Parallel()

	data := templateData{
		Version:      "9.9.9-test",
		SkillsPath:   "/home/user/.codex/skills",
		CommandsPath: "/home/user/.codex/commands",
	}
	result, err := common.RenderTemplate(&templateFS,
		"templates/skill.md.tmpl", data)
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// Codex-specific skill content markers.
	assert.Contains(t, result, "version: 9.9.9-test",
		"skill template should include version from data")
	assert.Contains(t, result, "Sequoia — Orchestrator (C0)",
		"skill template should include Codex-specific header")
	assert.Contains(t, result, "sequoia-security",
		"skill template should include agent roster in Codex format")
	assert.Contains(t, result, "sequoia-reporter",
		"skill template should include all agents")
}

// TestTemplateRendering_ConfigTemplate verifies that the Codex TOML config
// template renders correctly with path placeholders resolved. This is
// Codex-specific: no other adapter uses TOML merging.
func TestTemplateRendering_ConfigTemplate(t *testing.T) {
	t.Parallel()

	data := templateData{
		Version:      "0.5.0",
		SkillsPath:   "/custom/skills/path",
		CommandsPath: "/custom/commands/path",
	}
	result, err := common.RenderTemplate(&templateFS,
		"templates/config.toml.tmpl", data)
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// Codex-specific TOML config markers.
	assert.Contains(t, result, "[sequoia]",
		"TOML config should include [sequoia] section header")
	assert.Contains(t, result, `skills_path = "/custom/skills/path"`,
		"TOML config should include resolved skills path")
	assert.Contains(t, result, `commands_path = "/custom/commands/path"`,
		"TOML config should include resolved commands path")
}

// TestTemplateRendering_DifferentOutputs verifies REQ-TEST-MOCKFS
// "Different adapters get different template content".
// Within Codex, the skill and config templates produce distinct output.
func TestTemplateRendering_DifferentOutputs(t *testing.T) {
	t.Parallel()

	data := templateData{
		Version:      "0.1.0",
		SkillsPath:   "/tmp/skills",
		CommandsPath: "/tmp/commands",
	}

	skill, err := common.RenderTemplate(&templateFS,
		"templates/skill.md.tmpl", data)
	require.NoError(t, err)

	config, err := common.RenderTemplate(&templateFS,
		"templates/config.toml.tmpl", data)
	require.NoError(t, err)

	assert.NotEqual(t, skill, config,
		"skill and config templates must produce different output")
}

// TestTemplateRendering_SkillDoesNotContainTOML verifies that the Codex
// skill template (markdown-based) does NOT contain TOML syntax — only
// the config template produces TOML output. This prevents cross-template
// contamination.
func TestTemplateRendering_SkillDoesNotContainTOML(t *testing.T) {
	t.Parallel()

	data := templateData{
		Version:      "1.0.0",
		SkillsPath:   "/x",
		CommandsPath: "/y",
	}
	result, err := common.RenderTemplate(&templateFS,
		"templates/skill.md.tmpl", data)
	require.NoError(t, err)

	assert.NotContains(t, result, "[sequoia]",
		"skill template should NOT contain TOML section header")
	assert.NotContains(t, result, "skills_path",
		"skill template should NOT contain TOML path keys")
}

// TestTemplateRendering_TemplateNotFound verifies that rendering a
// non-existent template from Codex's templateFS returns an error.
func TestTemplateRendering_TemplateNotFound(t *testing.T) {
	t.Parallel()

	_, err := common.RenderTemplate(&templateFS,
		"templates/nonexistent.md.tmpl", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read template",
		"error should mention template reading failure")
}
