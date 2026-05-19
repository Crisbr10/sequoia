// Package claude provides white-box tests for Claude Code template rendering.
// These tests access the unexported templateFS to verify Claude-specific
// template content is rendered correctly.
package claude

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// TestTemplateRendering_SkillTemplate verifies REQ-TEST-MOCKFS
// "Claude adapter with mockFS produces Claude-specific content".
// Renders the Claude skill template with mock data and asserts that
// Claude-specific content (Sequoia orchestrator, agent roster, command
// reference) appears in the output.
func TestTemplateRendering_SkillTemplate(t *testing.T) {
	t.Parallel()

	data := templateData{Version: "9.9.9-test"}
	result, err := common.RenderTemplate(&templateFS,
		"templates/skill.md.tmpl", data)
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// Claude-specific skill content markers.
	assert.Contains(t, result, "version: 9.9.9-test",
		"skill template should include version from data")
	assert.Contains(t, result, "Sequoia — Orchestrator Skill",
		"skill template should include Claude-specific section header")
	assert.Contains(t, result, "/sequoia init",
		"skill template should include Sequoia commands")
	assert.Contains(t, result, "Health Score",
		"skill template should include Health Score documentation")
}

// TestTemplateRendering_SystemPromptTemplate verifies that the Claude
// system prompt template renders Claude-specific content (markdown
// sections format, Claude-specific agent table).
func TestTemplateRendering_SystemPromptTemplate(t *testing.T) {
	t.Parallel()

	data := templateData{Version: "1.2.3-test"}
	result, err := common.RenderTemplate(&templateFS,
		"templates/claude-md-section.md.tmpl", data)
	require.NoError(t, err)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	// Claude-specific system prompt markers.
	assert.Contains(t, result, "Sequoia v1.2.3-test",
		"system prompt should include version from data")
	assert.Contains(t, result, "Available Commands",
		"system prompt should include commands table")
	assert.Contains(t, result, "P1 Security",
		"system prompt should include agent coverage")
	assert.Contains(t, result, "~/.claude/skills/sequoia/SKILL.md",
		"system prompt should include Claude-specific skill path")
}

// TestTemplateRendering_DifferentOutputs verifies REQ-TEST-MOCKFS
// "Different adapters get different template content".
// Within Claude, the skill and system prompt templates produce
// distinct output.
func TestTemplateRendering_DifferentOutputs(t *testing.T) {
	t.Parallel()

	data := templateData{Version: "0.1.0"}

	skill, err := common.RenderTemplate(&templateFS,
		"templates/skill.md.tmpl", data)
	require.NoError(t, err)

	sys, err := common.RenderTemplate(&templateFS,
		"templates/claude-md-section.md.tmpl", data)
	require.NoError(t, err)

	assert.NotEqual(t, skill, sys,
		"skill and system prompt templates must produce different output")
}

// TestTemplateRendering_TemplateNotFound verifies that rendering a
// non-existent template from Claude's templateFS returns an error.
func TestTemplateRendering_TemplateNotFound(t *testing.T) {
	t.Parallel()

	_, err := common.RenderTemplate(&templateFS,
		"templates/nonexistent.md.tmpl", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read template",
		"error should mention template reading failure")
}
