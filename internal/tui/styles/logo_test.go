package styles_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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
