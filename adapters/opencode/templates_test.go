package opencode_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var templateFiles = []string{
	"templates/skill.md.tmpl",
	"templates/agents-md-section.md.tmpl",
}

func TestTemplates_AllFilesExist(t *testing.T) {
	t.Parallel()
	for _, path := range templateFiles {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			_, err := os.ReadFile(path)
			require.NoError(t, err)
		})
	}
}

func TestTemplates_SkillHasVersionPlaceholder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/skill.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "{{.Version}}"))
}

func TestTemplates_AgentsMDSectionHasVersionPlaceholder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/agents-md-section.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "{{.Version}}"))
}

func TestTemplates_SkillHasAllAgents(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/skill.md.tmpl")
	require.NoError(t, err)
	content := string(data)
	agents := []string{
		"sequoia-security",
		"sequoia-performance",
		"sequoia-architecture",
		"sequoia-quality",
		"sequoia-experience",
		"sequoia-operations",
		"sequoia-i18n",
		"sequoia-correlator",
		"sequoia-reporter",
	}
	for _, agent := range agents {
		agent := agent
		t.Run(agent, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(content, agent), "skill.md.tmpl should contain %s", agent)
		})
	}
}

func TestTemplates_AgentsMDHasMarkers(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/agents-md-section.md.tmpl")
	require.NoError(t, err)
	content := string(data)
	assert.True(t, strings.Contains(content, "<!-- sequoia:start -->"), "agents-md-section.md.tmpl should contain sequoia:start marker")
	assert.True(t, strings.Contains(content, "<!-- sequoia:end -->"), "agents-md-section.md.tmpl should contain sequoia:end marker")
}
