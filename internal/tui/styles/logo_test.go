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

func TestLogo_ReturnsSameStringOnRepeatedCalls(t *testing.T) {
	// RED: Logo() currently regenerates the figure on every call.
	// After caching, two calls should return the identical string.
	first := styles.Logo()
	second := styles.Logo()
	assert.Equal(t, first, second, "consecutive Logo() calls should return identical strings (cached)")
}

func TestLogo_CachingIsGoroutineSafe(t *testing.T) {
	// RED: Logo() must be safe for concurrent use after caching is added.
	const goroutines = 10
	results := make([]string, goroutines)
	done := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			results[i] = styles.Logo()
			done <- struct{}{}
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}

	first := results[0]
	for i := 1; i < goroutines; i++ {
		assert.Equal(t, first, results[i],
			"all goroutines should receive the same cached logo (goroutine %d differs)", i)
	}
}
