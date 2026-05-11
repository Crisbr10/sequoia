package styles_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/internal/tui/styles"
)

func TestLogo_IsMultiLine(t *testing.T) {
	t.Parallel()
	logo := styles.Logo()
	lines := strings.Split(strings.TrimSpace(logo), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Logo should have at least 3 lines")
}

func TestLogo_ContainsName(t *testing.T) {
	t.Parallel()
	logo := styles.Logo()
	assert.NotEmpty(t, logo)
	assert.Contains(t, logo, "Sequoia", "Logo should contain the project name")
}

func TestLogo_QDiffersFromO(t *testing.T) {
	t.Parallel()
	logo := styles.Logo()
	lines := strings.Split(logo, "\n")

	// Find the line containing the Q and O glyph tails (╚════ marks line 4).
	var line4 string
	for _, line := range lines {
		if strings.Contains(line, "\u255A\u2550\u2550\u2550\u2550") {
			line4 = line
			break
		}
	}
	require.NotEmpty(t, line4, "should find line 4 containing Q and O glyph tails")

	// After the fix, only Q retains the ▄▄ tail; O has spaces instead.
	// Currently (pre-fix) both Q and O have ▄▄, so the count is 2.
	// After fix: count == 1 (only Q).
	qGlyph := "██║▄▄ ██║"
	count := strings.Count(line4, qGlyph)
	assert.Equal(t, 1, count, "only Q should have ▄▄ tail on line 4; O should use spaces")
}
