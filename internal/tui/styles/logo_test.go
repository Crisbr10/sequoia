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

// TestLogo_QDiffersFromO verifies that Q has a dedicated tail line (╚══╝)
// below the main 6-row body. This tail is absent for every other letter,
// making Q unmistakably different from O at a glance.
func TestLogo_QDiffersFromO(t *testing.T) {
	t.Parallel()
	logo := styles.Logo()

	// ╚══╝ appears only on Q's exclusive tail line (row 6).
	// It is NOT a substring of the longer ╚═════╝ / ╚══════╝ sequences
	// used in other rows, so this check is uniquely identifying.
	assert.Contains(t, logo, "╚══╝",
		"logo must contain Q's tail (╚══╝) on a dedicated line below the main body")

	// The logo must have at least 8 content lines: 6 letter rows + Q tail + tagline.
	lines := strings.Split(strings.TrimSpace(logo), "\n")
	assert.GreaterOrEqual(t, len(lines), 8,
		"logo must include the Q tail row in addition to 6 main body rows and the tagline")
}
