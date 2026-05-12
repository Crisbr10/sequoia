package gemini_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var geminiTemplateFiles = []string{
	"templates/skill.md.tmpl",
	"templates/gemini-md-section.md.tmpl",
}

func TestTemplates_AllFilesExist(t *testing.T) {
	t.Parallel()
	for _, relPath := range geminiTemplateFiles {
		relPath := relPath
		t.Run(relPath, func(t *testing.T) {
			t.Parallel()
			_, err := os.ReadFile(relPath)
			require.NoError(t, err)
		})
	}
}

func TestTemplates_SkillHasVersionPlaceholder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/skill.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "{{.Version}}"),
		"skill.md.tmpl should contain {{.Version}} placeholder")
}

func TestTemplates_SkillHasName(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/skill.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "name: sequoia"),
		"skill.md.tmpl should contain 'name: sequoia'")
}

func TestTemplates_GeminiMDSectionHasVersionPlaceholder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/gemini-md-section.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "{{.Version}}"),
		"gemini-md-section.md.tmpl should contain {{.Version}} placeholder")
}

func TestTemplates_GeminiMDSectionHasNoMarkers(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/gemini-md-section.md.tmpl")
	require.NoError(t, err)
	content := string(data)
	assert.NotContains(t, content, "<!-- sequoia:start -->",
		"gemini-md-section.md.tmpl should NOT contain start marker (added by InjectSection)")
	assert.NotContains(t, content, "<!-- sequoia:end -->",
		"gemini-md-section.md.tmpl should NOT contain end marker (added by InjectSection)")
}

func TestTemplates_GoldenFile_GeminiMDSection(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("templates/gemini-md-section.md.tmpl")
	require.NoError(t, err)

	tmpl, err := template.New("gemini-md-section.md.tmpl").Parse(string(raw))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{"Version": "0.1.0"})
	require.NoError(t, err)

	got := buf.String()

	goldenPath := filepath.Join("templates", "testdata", "golden", "gemini-md-section.md.golden")
	golden, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	got = strings.ReplaceAll(got, "\r\n", "\n")
	want := strings.ReplaceAll(string(golden), "\r\n", "\n")
	assert.Equal(t, want, got, "rendered template must match golden file.")
}
