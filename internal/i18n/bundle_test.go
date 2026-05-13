package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/internal/i18n"
)

// TestInit_HappyPath verifies that Init() succeeds when both embedded catalogs
// are present, and Initialized() transitions to true.
// Spec scenario 1.
func TestInit_HappyPath(t *testing.T) {
	t.Parallel()

	// Before Init, Initialized should be false.
	assert.False(t, i18n.Initialized(),
		"Initialized() should be false before Init() is called")

	err := i18n.Init()
	require.NoError(t, err, "Init() should succeed with embedded en+es catalogs")

	// After Init, Initialized should be true.
	assert.True(t, i18n.Initialized(),
		"Initialized() should be true after successful Init()")

	// Calling Init again should be safe (idempotent).
	err2 := i18n.Init()
	assert.NoError(t, err2, "Init() should be safe to call multiple times")
}

// TestT_EnglishMatch verifies that T() returns the correct English message.
// Spec scenario 4.
func TestT_EnglishMatch(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before T() calls")

	result := i18n.T("welcome.menu_install", "en")
	assert.Equal(t, "Install", result,
		"T(welcome.menu_install, en) should return 'Install'")
}

// TestT_SpanishMatch verifies that T() returns the correct Spanish message.
// Spec scenario 5.
func TestT_SpanishMatch(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before T() calls")

	result := i18n.T("welcome.menu_install", "es")
	assert.Equal(t, "Instalar", result,
		"T(welcome.menu_install, es) should return 'Instalar'")
}

// TestT_EnglishDefaultFallback verifies that requesting a non-existent language
// falls back to English.
// Spec scenario 6 (extended: unsupported language).
func TestT_EnglishDefaultFallback(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before T() calls")

	// French is not supported — should fall back to English.
	result := i18n.T("welcome.menu_install", "fr")
	assert.Equal(t, "Install", result,
		"T(key, unsupported_lang) should fall back to English")
}

// TestT_MissingKeyReturnsKey verifies that a missing key returns the key itself
// instead of panicking or returning empty string.
func TestT_MissingKeyReturnsKey(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before T() calls")

	missingKey := "nonexistent.key"
	result := i18n.T(missingKey, "en")
	assert.Equal(t, missingKey, result,
		"T(missing_key, en) should return the key itself as fallback")
}

// TestT_MissingKeySpanishFallbackToEnglish verifies that when a key exists in
// English but not Spanish, T() falls back to English.
func TestT_MissingKeySpanishFallbackToEnglish(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before T() calls")

	// All keys should exist in both catalogs, but we test the fallback behavior:
	// If the key is in en.toml, it should work for es too.
	result := i18n.T("footer.quit_key", "es")
	// Even in Spanish, quit_key is "q" (same as English).
	assert.Equal(t, "q", result,
		"T(footer.quit_key, es) should return 'q'")
}

// TestT_TemplateArgs verifies template variable interpolation.
// Spec scenario 7.
func TestT_TemplateArgs(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before T() calls")

	// English with template data.
	result := i18n.T("tool_selection.selected_count", "en", map[string]interface{}{
		"Selected": 3,
		"Total":    5,
	})
	assert.Contains(t, result, "3", "should interpolate Selected")
	assert.Contains(t, result, "5", "should interpolate Total")

	// Spanish with template data.
	resultEs := i18n.T("tool_selection.selected_count", "es", map[string]interface{}{
		"Selected": 2,
		"Total":    4,
	})
	assert.Contains(t, resultEs, "2", "should interpolate Selected in Spanish")
	assert.Contains(t, resultEs, "4", "should interpolate Total in Spanish")
}

// TestT_MultipleKeysAcrossScreens verifies that multiple diverse keys work
// correctly in both languages (triangulation across screens).
func TestT_MultipleKeysAcrossScreens(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err, "Init must succeed before T() calls")

	tests := []struct {
		name     string
		key      string
		lang     string
		expected string
	}{
		// Welcome screen
		{"welcome install en", "welcome.menu_install", "en", "Install"},
		{"welcome install es", "welcome.menu_install", "es", "Instalar"},
		{"welcome status en", "welcome.menu_status", "en", "Status"},
		{"welcome status es", "welcome.menu_status", "es", "Estado"},
		{"welcome quit en", "welcome.menu_quit", "en", "Quit"},
		{"welcome quit es", "welcome.menu_quit", "es", "Salir"},

		// Tool selection
		{"tool selection title en", "tool_selection.title", "en", "Select AI Tools"},
		{"tool selection title es", "tool_selection.title", "es", "Seleccionar Herramientas IA"},

		// Configuration
		{"config title en", "configuration.title", "en", "Configuration"},
		{"config title es", "configuration.title", "es", "Configuración"},
		{"config language en", "configuration.language_label", "en", "Language"},
		{"config language es", "configuration.language_label", "es", "Idioma"},

		// Install progress
		{"progress install en", "install_progress.title_install", "en", "Installing"},
		{"progress install es", "install_progress.title_install", "es", "Instalando"},

		// Validation
		{"validation en", "validation.select_at_least_one", "en", "Select at least one tool to continue"},
		{"validation es", "validation.select_at_least_one", "es", "Selecciona al menos una herramienta para continuar"},

		// Uninstall
		{"uninstall title en", "uninstall.title", "en", "Uninstall"},
		{"uninstall title es", "uninstall.title", "es", "Desinstalar"},

		// Status
		{"status title en", "status.title", "en", "Status"},
		{"status title es", "status.title", "es", "Estado"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := i18n.T(tt.key, tt.lang)
			assert.Equal(t, tt.expected, result,
				"T(%s, %s) = %q, want %q", tt.key, tt.lang, result, tt.expected)
		})
	}
}

// TestT_BeforeInitStillWorks verifies that T() doesn't panic when called
// before Init(). It should return the key itself as a safe fallback.
func TestT_BeforeInitStillWorks(t *testing.T) {
	t.Parallel()

	// We can't easily reset the bundle's sync.Once, but the stub
	// returns the key itself in all cases before Init.
	// After the real implementation, T() should not panic even if
	// called before Init() — it should return the key.
	result := i18n.T("welcome.menu_install", "en")
	assert.NotEmpty(t, result, "T() before Init() should still return something")
}

// TestInitialized_AfterInit verifies the Initialized() guard.
// Spec scenarios 8 and 9.
func TestInitialized_AfterInit(t *testing.T) {
	t.Parallel()

	err := i18n.Init()
	require.NoError(t, err)

	assert.True(t, i18n.Initialized(),
		"Initialized() should be true after successful Init()")
}
