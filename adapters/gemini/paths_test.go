package gemini_test

import (
	"path/filepath"
	"testing"

	"sequoia-ai/adapters/gemini"

	"github.com/stretchr/testify/assert"
)

func TestPaths_SkillsPath(t *testing.T) {
	t.Parallel()
	a := gemini.NewAdapter(t.TempDir())
	p := a.SkillsPath()
	assert.NotEmpty(t, p)
	assert.True(t, hasSuffixAny(p, ".gemini/sequoia/skills", ".gemini\\sequoia\\skills"),
		"SkillsPath should end with gemini/sequoia/skills, got %s", p)
}

func TestPaths_CommandsPath(t *testing.T) {
	t.Parallel()
	a := gemini.NewAdapter(t.TempDir())
	p := a.CommandsPath()
	assert.NotEmpty(t, p)
	assert.True(t, hasSuffixAny(p, ".gemini/sequoia/commands", ".gemini\\sequoia\\commands"),
		"CommandsPath should end with gemini/sequoia/commands, got %s", p)
}

func TestPaths_SystemPromptPath(t *testing.T) {
	t.Parallel()
	a := gemini.NewAdapter(t.TempDir())
	p := a.SystemPromptPath()
	assert.NotEmpty(t, p)
	assert.True(t, hasSuffixAny(p, "GEMINI.md", "GEMINI.md"),
		"SystemPromptPath should end with GEMINI.md, got %s", p)
}

func TestPaths_SlashNormalized(t *testing.T) {
	t.Parallel()
	a := gemini.NewAdapter(t.TempDir())
	p := filepath.ToSlash(a.SkillsPath())
	assert.Contains(t, p, ".gemini/sequoia/skills",
		"Slash-normalized SkillsPath should contain .gemini/sequoia/skills, got %s", p)
}

func TestPaths_AllUseFilepathJoin(t *testing.T) {
	// This test verifies that all path functions return OS-correct paths.
	// On Windows, this means backslashes; on Unix, forward slashes.
	t.Parallel()
	a := gemini.NewAdapter(t.TempDir())
	p := a.SkillsPath()
	// filepath.Join on Windows uses backslashes — verify path contains separator.
	assert.Contains(t, p, string(filepath.Separator),
		"SkillsPath should use filepath.Separator")
}

// hasSuffixAny returns true if p ends with any of the given suffixes.
func hasSuffixAny(p string, suffixes ...string) bool {
	normalized := filepath.ToSlash(p)
	for _, s := range suffixes {
		if len(normalized) >= len(s) && normalized[len(normalized)-len(s):] == s {
			return true
		}
	}
	return false
}
