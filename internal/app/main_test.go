// Package app_test contains external (black-box) tests for the TUI model,
// including integration tests that drive the full install flow.
package app_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
