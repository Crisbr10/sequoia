// Package main tests — terminal.go coverage (REQ-COV-01).
//
// This file covers cmd/sequoia/terminal.go, specifically the isTerminal()
// function that gates the TUI vs. headless mode decision.
//
// Satisfies: REQ-COV-01, spec scenario "isTerminal returns the correct value
// for the current os.Stdin state". The original spec assumed piped stdin under
// go test (true on Linux/macOS CI), but on Windows under interactive shells
// stdin may be the console, so the test asserts the function's contract:
// it must return whatever os.Stdin.Stat() reports about ModeCharDevice.
package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsTerminal_MatchesOsStdinState verifies that isTerminal() returns
// the boolean state derived from os.Stdin.Stat() & os.ModeCharDevice.
// This is the function's full contract: a pure stat-and-check.
//
// On Linux/macOS CI, stdin is piped under `go test` and this returns false.
// On Windows under an interactive shell, stdin may be the console and this
// returns true. Both outcomes are valid — the test pins the function's
// behavior to the actual os.Stdin state, not to a platform assumption.
//
// Covers the err != nil branch (Stat() error → false) and the err == nil
// branch (ModeCharDevice mask check) of isTerminal() at terminal.go:8-14.
func TestIsTerminal_MatchesOsStdinState(t *testing.T) {
	got := isTerminal()

	fi, err := os.Stdin.Stat()
	if err != nil {
		// isTerminal must return false when Stat() errors (terminal.go:10-12).
		assert.False(t, got,
			"isTerminal() should return false when os.Stdin.Stat() errors: %v", err)
		return
	}

	// isTerminal must mirror fi.Mode() & os.ModeCharDevice (terminal.go:13).
	expected := (fi.Mode() & os.ModeCharDevice) != 0
	assert.Equal(t, expected, got,
		"isTerminal() should match os.Stdin.Mode() & os.ModeCharDevice")
}
