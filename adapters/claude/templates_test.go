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
