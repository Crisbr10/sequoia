package main

import "os"

// isTerminal reports whether fd 0 (stdin) is a terminal.
// When it returns true the TUI can render interactive screens;
// when false it degrades to headless mode automatically.
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
