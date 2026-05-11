package styles_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/internal/tui/styles"
)

func TestLogo_NonEmpty(t *testing.T) {
	t.Parallel()
	logo := styles.Logo()
	assert.NotEmpty(t, logo, "Logo() should return a non-empty string")
}

func TestLogo_IsMultiLine(t *testing.T) {
	t.Parallel()
	logo := styles.Logo()
	lines := strings.Split(strings.TrimSpace(logo), "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Logo should have at least 3 lines")
}

func TestLogo_ContainsANSI(t *testing.T) {
	t.Parallel()
	logo := styles.Logo()
	assert.Contains(t, logo, "\x1b[", "Logo should contain ANSI color escape sequences")
}
