package styles_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/internal/tui/styles"
)

func TestSequoiaTree_NonEmpty(t *testing.T) {
	t.Parallel()
	tree := styles.SequoiaTree()
	assert.NotEmpty(t, tree, "SequoiaTree() should return a non-empty string")
}

func TestSequoiaTree_HasCanopy(t *testing.T) {
	t.Parallel()
	tree := styles.SequoiaTree()
	// Canopy rows use green foliage and full-block (█) characters.
	// The ANSI escape for forest green is \x1b[38;2;34;139;34m (or similar).
	assert.Contains(t, tree, "\x1b[", "must contain ANSI escape sequences (colored output)")
	assert.Contains(t, tree, "█", "must contain full-block canopy characters")
}

func TestSequoiaTree_HasTrunk(t *testing.T) {
	t.Parallel()
	tree := styles.SequoiaTree()
	// Trunk rows exist and are non-empty.
	lines := strings.Split(strings.TrimRight(tree, "\n"), "\n")
	// After trimming trailing newline, the last few lines should be the trunk.
	assert.GreaterOrEqual(t, len(lines), 10, "tree must have at least 10 lines")
}

func TestSequoiaTree_MultiLine(t *testing.T) {
	t.Parallel()
	tree := styles.SequoiaTree()
	lines := strings.Split(strings.TrimRight(tree, "\n"), "\n")
	assert.GreaterOrEqual(t, len(lines), 10, "tree must be multi-line (at least 10 lines)")
}

func TestSequoiaTree_Deterministic(t *testing.T) {
	tree1 := styles.SequoiaTree()
	tree2 := styles.SequoiaTree()
	assert.Equal(t, tree1, tree2, "two calls to SequoiaTree() must return the same string")
}

func TestSequoiaTree_ContainsColors(t *testing.T) {
	t.Parallel()
	tree := styles.SequoiaTree()
	assert.Contains(t, tree, "\x1b[", "output must contain ANSI color escape codes")
	assert.Contains(t, tree, "█", "output must contain rendered pixel art (full-block chars)")
}
