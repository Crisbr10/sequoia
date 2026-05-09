package claude_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var templateFiles = []string{
	"templates/skill.md.tmpl",
	"templates/claude-md-section.md.tmpl",
	"templates/commands/sequoia-init.md",
	"templates/commands/sequoia-audit.md",
	"templates/commands/sequoia-review.md",
	"templates/commands/sequoia-fix.md",
	"templates/commands/sequoia-diff.md",
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

func TestTemplates_ClaudeMDSectionHasVersionPlaceholder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/claude-md-section.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "{{.Version}}"))
}

func TestTemplates_SkillHasName(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/skill.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "name: sequoia"))
}

func TestTemplates_CommandsHaveFrontmatter(t *testing.T) {
	t.Parallel()
	commands := []string{
		"templates/commands/sequoia-init.md",
		"templates/commands/sequoia-audit.md",
		"templates/commands/sequoia-review.md",
		"templates/commands/sequoia-fix.md",
		"templates/commands/sequoia-diff.md",
	}
	for _, path := range commands {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(string(data), "---"))
		})
	}
}

func TestTemplates_InitCommandHasAllowedTools(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/commands/sequoia-init.md")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "allowed-tools:"))
}
