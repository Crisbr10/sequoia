package common_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters/common"
)

// TestCommandTemplates_HaveFrontmatter verifies all embedded command files
// start with YAML frontmatter (---).
func TestCommandTemplates_HaveFrontmatter(t *testing.T) {
	t.Parallel()
	for _, cmd := range common.CommandFiles {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			data, err := common.CommandFS.ReadFile("templates/commands/" + cmd)
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(string(data), "---"),
				"%s should start with YAML frontmatter", cmd)
		})
	}
}

// TestCommandTemplates_InitHasAllowedTools verifies sequoia-init.md
// contains the allowed-tools field.
func TestCommandTemplates_InitHasAllowedTools(t *testing.T) {
	t.Parallel()
	data, err := common.CommandFS.ReadFile("templates/commands/sequoia-init.md")
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "allowed-tools:"))
}

// TestCommandTemplates_AllExist verifies all expected command files are
// present in CommandFS.
func TestCommandTemplates_AllExist(t *testing.T) {
	t.Parallel()
	for _, cmd := range common.CommandFiles {
		cmd := cmd
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()
			_, err := common.CommandFS.ReadFile("templates/commands/" + cmd)
			require.NoError(t, err, "%s should exist in CommandFS", cmd)
		})
	}
}
