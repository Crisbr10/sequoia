package codex_test

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

var codexTemplateFiles = []string{
	"templates/config.toml.tmpl",
	"templates/skill.md.tmpl",
	"templates/commands/sequoia-init.md",
	"templates/commands/sequoia-audit.md",
	"templates/commands/sequoia-review.md",
	"templates/commands/sequoia-fix.md",
	"templates/commands/sequoia-diff.md",
}

func TestTemplates_AllFilesExist(t *testing.T) {
	t.Parallel()
	base := ""
	for _, relPath := range codexTemplateFiles {
		relPath := relPath
		t.Run(relPath, func(t *testing.T) {
			t.Parallel()
			_, err := os.ReadFile(filepath.Join(base, relPath))
			require.NoError(t, err)
		})
	}
}

func TestTemplates_ConfigHasSequoiaTable(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/config.toml.tmpl")
	require.NoError(t, err)
	content := string(data)
	assert.True(t, strings.Contains(content, "[sequoia]"),
		"config.toml.tmpl should contain [sequoia] table header")
	assert.True(t, strings.Contains(content, "{{.SkillsPath}}"),
		"config.toml.tmpl should contain {{.SkillsPath}} placeholder")
	assert.True(t, strings.Contains(content, "{{.CommandsPath}}"),
		"config.toml.tmpl should contain {{.CommandsPath}} placeholder")
}

func TestTemplates_SkillHasVersionPlaceholder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/skill.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "{{.Version}}"),
		"skill.md.tmpl should contain {{.Version}} placeholder")
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
		"sequoia-correlator",
		"sequoia-reporter",
	}
	for _, agent := range agents {
		agent := agent
		t.Run(agent, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(content, agent),
				"skill.md.tmpl should contain %s", agent)
		})
	}
}

func TestTemplates_GoldenFile_Config(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("templates/config.toml.tmpl")
	require.NoError(t, err)

	tmpl, err := template.New("config.toml.tmpl").Parse(string(raw))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"SkillsPath":   "/home/user/.codex/sequoia/skills/",
		"CommandsPath": "/home/user/.codex/sequoia/commands/",
	})
	require.NoError(t, err)

	got := buf.String()

	goldenPath := filepath.Join("templates", "testdata", "golden", "config.toml.golden")
	golden, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	// Normalize line endings: golden files may have \r\n on Windows after checkout,
	// but template execution always produces \n.
	got = strings.ReplaceAll(got, "\r\n", "\n")
	want := strings.ReplaceAll(string(golden), "\r\n", "\n")
	assert.Equal(t, want, got, "rendered template must match golden file. To update, regenerate the golden file.")
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
