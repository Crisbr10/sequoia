// Package styles_test provides tests for the TUI styles package.
package styles_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Crisbr10/sequoia/internal/tui/styles"
)

func init() {
	// Force lipgloss color output in non-TTY test environments.
	os.Setenv("FORCE_COLOR", "1")
	os.Setenv("CLICOLOR_FORCE", "1")
}

func TestTitle_RendersNonEmpty(t *testing.T) {
	t.Parallel()
	out := styles.Title().Render("Hello")
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Hello")
}

func TestSubtitle_RendersNonEmpty(t *testing.T) {
	t.Parallel()
	out := styles.Subtitle().Render("World")
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "World")
}

func TestBody_RendersNonEmpty(t *testing.T) {
	t.Parallel()
	out := styles.Body().Render("Body")
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Body")
}

func TestAccent_DiffersFromBody(t *testing.T) {
	t.Parallel()
	a := styles.Accent().Render("X")
	b := styles.Body().Render("X")
	assert.NotEmpty(t, a)
	assert.NotEqual(t, b, a, "Accent and Body should differ visually")
}

func TestError_DiffersFromSuccess(t *testing.T) {
	t.Parallel()
	e := styles.Error().Render("X")
	s := styles.Success().Render("X")
	assert.NotEmpty(t, e)
	assert.NotEmpty(t, s)
	assert.NotEqual(t, e, s, "Error and Success should differ visually")
}

func TestMuted_DiffersFromHighlight(t *testing.T) {
	t.Parallel()
	m := styles.Muted().Render("X")
	h := styles.Highlight().Render("X")
	assert.NotEmpty(t, m)
	assert.NotEmpty(t, h)
	assert.NotEqual(t, m, h, "Muted and Highlight should differ visually")
}
