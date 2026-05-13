package screens_test

import (
	"os"
	"testing"

	"github.com/Crisbr10/sequoia/internal/i18n"
)

func TestMain(m *testing.M) {
	// Initialize i18n before any screen test runs so that i18n.T() returns
	// real translations instead of raw keys.
	if err := i18n.Init(); err != nil {
		// If English catalog fails, let the test suite fail immediately
		// rather than producing confusing key-as-string output.
		panic("i18n.Init() failed in screens_test TestMain: " + err.Error())
	}

	os.Exit(m.Run())
}
