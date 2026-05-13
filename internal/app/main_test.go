// Package app_test contains external (black-box) tests for the TUI model,
// including integration tests that drive the full install flow.
package app_test

import (
	"os"
	"testing"

	"github.com/Crisbr10/sequoia/internal/i18n"
)

func TestMain(m *testing.M) {
	// Initialize i18n so that view rendering produces translated strings
	// instead of raw message keys. Idempotent via sync.Once.
	if err := i18n.Init(); err != nil {
		panic("i18n.Init() failed: " + err.Error())
	}
	os.Exit(m.Run())
}