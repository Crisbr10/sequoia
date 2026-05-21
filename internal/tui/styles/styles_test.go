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
	_ = os.Setenv("FORCE_COLOR", "1")
	_ = os.Setenv("CLICOLOR_FORCE", "1")
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

// TestStyleFunctions_ZeroAllocations verifies that after caching is implemented,
// each style function performs zero heap allocations per call.
// RED: Currently each call creates a new lipgloss.NewStyle() (~200 bytes heap).
func TestStyleFunctions_ZeroAllocations(t *testing.T) {
	// Warm up: first call may allocate during lazy initialization.
	_ = styles.Title()
	_ = styles.Subtitle()
	_ = styles.Body()
	_ = styles.Accent()
	_ = styles.Error()
	_ = styles.Success()
	_ = styles.Muted()
	_ = styles.Highlight()

	t.Run("Title", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Title()
		})
		assert.Equal(t, float64(0), allocs, "Title() should perform 0 allocations after caching")
	})
	t.Run("Subtitle", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Subtitle()
		})
		assert.Equal(t, float64(0), allocs, "Subtitle() should perform 0 allocations after caching")
	})
	t.Run("Body", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Body()
		})
		assert.Equal(t, float64(0), allocs, "Body() should perform 0 allocations after caching")
	})
	t.Run("Accent", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Accent()
		})
		assert.Equal(t, float64(0), allocs, "Accent() should perform 0 allocations after caching")
	})
	t.Run("Error", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Error()
		})
		assert.Equal(t, float64(0), allocs, "Error() should perform 0 allocations after caching")
	})
	t.Run("Success", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Success()
		})
		assert.Equal(t, float64(0), allocs, "Success() should perform 0 allocations after caching")
	})
	t.Run("Muted", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Muted()
		})
		assert.Equal(t, float64(0), allocs, "Muted() should perform 0 allocations after caching")
	})
	t.Run("Highlight", func(t *testing.T) {
		allocs := testing.AllocsPerRun(100, func() {
			_ = styles.Highlight()
		})
		assert.Equal(t, float64(0), allocs, "Highlight() should perform 0 allocations after caching")
	})
}
