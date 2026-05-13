// Package i18n provides internationalization support for the Sequoia TUI installer.
// It wraps go-i18n/v2 with embedded TOML catalogs and a minimal public API:
// Init() to load catalogs, T() to look up messages, and Initialized() to gate
// language-dependent UI features.
package i18n

import (
	"fmt"
	"log"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/Crisbr10/sequoia/internal/i18n/translations"
)

// bundle holds the go-i18n/v2 bundle with loaded message catalogs.
// It is populated by Init() via sync.Once.
var (
	bundle      *i18n.Bundle
	initOnce    sync.Once
	initErr     error
	initialized bool
)

// Init loads the embedded English and Spanish TOML catalogs into the global
// i18n bundle. It MUST be called once at application startup before any T()
// calls. The function is idempotent via sync.Once — subsequent calls after the
// first successful init are no-ops.
//
// Missing or corrupt English catalog is fatal (returns error). Missing Spanish
// catalog is non-fatal (logs a warning, continues with English-only).
func Init() error {
	initOnce.Do(func() {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

		// Load English catalog (required).
		enData, err := translations.FS.ReadFile("en.toml")
		if err != nil {
			initErr = fmt.Errorf("i18n: failed to read en.toml: %w", err)
			return
		}
		bundle.MustParseMessageFileBytes(enData, "en.toml")

		// Load Spanish catalog (non-fatal).
		esData, err := translations.FS.ReadFile("es.toml")
		if err != nil {
			log.Printf("[i18n] warning: Spanish catalog not found (%v), continuing with English only", err)
		} else {
			if _, err := bundle.ParseMessageFileBytes(esData, "es.toml"); err != nil {
				log.Printf("[i18n] warning: failed to parse es.toml: %v", err)
			}
		}

		initialized = true
	})
	return initErr
}

// T looks up the localized message for key in the given language.
// If the key is missing in the target language, it falls back to English.
// If the key is missing in both languages, the key itself is returned
// (no crash, no empty string).
//
// Optional data parameters provide template variables for interpolation.
// The last data argument (if any) is used as the template data map.
func T(key, lang string, data ...interface{}) string {
	if bundle == nil {
		return key
	}

	localizer := i18n.NewLocalizer(bundle, lang, "en")

	var templateData interface{}
	if len(data) > 0 {
		templateData = data[0]
	}

	cfg := &i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	}

	msg, err := localizer.Localize(cfg)
	if err != nil {
		// Fallback: return the key itself so the UI never breaks on a
		// missing translation.
		return key
	}

	// Handle the case where Localize returns the message ID (not found).
	if msg == "" {
		return key
	}

	return msg
}

// Initialized reports whether Init() completed successfully and the bundle is
// ready for T() lookups. Use this to gate language-dependent UI features
// (e.g., the language selector on the Configuration screen).
func Initialized() bool {
	return initialized
}
