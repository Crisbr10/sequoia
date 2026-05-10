package cursor_test

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

var cursorTemplateFiles = []string{
	"templates/rules.md.tmpl",
	"templates/commands/sequoia-init.md",
	"templates/commands/sequoia-audit.md",
	"templates/commands/sequoia-review.md",
	"templates/commands/sequoia-fix.md",
	"templates/commands/sequoia-diff.md",
}

func TestTemplates_AllFilesExist(t *testing.T) {
	t.Parallel()
	base := ""
	for _, relPath := range cursorTemplateFiles {
		relPath := relPath
		t.Run(relPath, func(t *testing.T) {
			t.Parallel()
			_, err := os.ReadFile(filepath.Join(base, relPath))
			require.NoError(t, err)
		})
	}
}

func TestTemplates_RulesMDHasVersionPlaceholder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/rules.md.tmpl")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "{{.Version}}"),
		"rules.md.tmpl should contain {{.Version}} placeholder")
}

func TestTemplates_RulesMDHasMarkers(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/rules.md.tmpl")
	require.NoError(t, err)
	content := string(data)
	assert.True(t, strings.Contains(content, "<!-- sequoia:start -->"),
		"rules.md.tmpl should contain sequoia:start marker")
	assert.True(t, strings.Contains(content, "<!-- sequoia:end -->"),
		"rules.md.tmpl should contain sequoia:end marker")
}

func TestTemplates_RulesMDHasAllAgents(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("templates/rules.md.tmpl")
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
				"rules.md.tmpl should contain %s", agent)
		})
	}
}

func TestTemplates_GoldenFile_RulesMD(t *testing.T) {
	t.Parallel()

	// Read and render the template.
	raw, err := os.ReadFile("templates/rules.md.tmpl")
	require.NoError(t, err)

	tmpl, err := template.New("rules.md.tmpl").Parse(string(raw))
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{"Version": "0.1.0"})
	require.NoError(t, err)

	got := buf.String()

	// Read the golden file.
	goldenPath := filepath.Join("templates", "testdata", "golden", "rules.md.golden")
	golden, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	assert.Equal(t, string(golden), got, "rendered template must match golden file. To update, regenerate the golden file.")
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
