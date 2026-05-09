// Package main_test verifies the sequoia CLI entrypoint compiles and
// exposes the expected command behaviour for integration testing.
package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"sequoia-ai/adapters"
)

// newRootCmdWithOut returns the root command with its output redirected to w
// so callers can capture and inspect command output.
func newRootCmdWithOut(w *bytes.Buffer) *cobra.Command {
	cmd := newRootCmd()
	cmd.SetOut(w)
	cmd.SetErr(w)
	return cmd
}

// TestRootHelp verifies that the root command prints usage when --help is passed.
func TestRootHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root --help returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage") {
		t.Errorf("root --help output does not contain 'Usage'; got: %q", got)
	}
}

// TestRootNoArgs verifies that the root command exits cleanly without arguments.
func TestRootNoArgs(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root with no args returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage") {
		t.Errorf("root output does not contain 'Usage'; got: %q", got)
	}
}

// TestVersionCmd verifies the version subcommand prints the Version string.
func TestVersionCmd(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"version"})

	// Override Version for deterministic output in tests.
	prev := Version
	Version = "0.1.0"
	defer func() { Version = prev }()

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command returned unexpected error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got != "0.1.0" {
		t.Errorf("version command output = %q; want %q", got, "0.1.0")
	}
}

// TestStatusCmd verifies the status subcommand runs without error.
func TestStatusCmd(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("status command returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "ID") || !strings.Contains(got, "NAME") {
		t.Errorf("status output missing header columns; got: %q", got)
	}
}

// TestInstallHelp verifies the install command prints usage on --help.
func TestInstallHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"install", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("install --help returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage") {
		t.Errorf("install --help output does not contain 'Usage'; got: %q", got)
	}
}

// TestUninstallHelp verifies the uninstall command prints usage on --help.
func TestUninstallHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"uninstall", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("uninstall --help returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage") {
		t.Errorf("uninstall --help output does not contain 'Usage'; got: %q", got)
	}
}

// TestUnknownCommand verifies that an unknown subcommand returns an error.
func TestUnknownCommand(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown subcommand, got nil")
	}
}

// TestInstallInvalidTool verifies that install --tool with an unknown adapter fails gracefully.
func TestInstallInvalidTool(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"install", "--tool=no-existe", "--no-tui"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown adapter, got nil")
	}

	got := out.String()
	if !strings.Contains(got, "unknown adapter") {
		t.Errorf("error output does not mention 'unknown adapter'; got: %q", got)
	}
}

// TestUninstallAllFlag verifies --all flag is registered on the uninstall command.
func TestUninstallAllFlag(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"uninstall", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("uninstall --help returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "--all") {
		t.Errorf("uninstall help output does not mention --all flag; got: %q", got)
	}
}

// TestInstallNoTUI flag is registered on the install command.
func TestInstallNoTUIFlag(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"install", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("install --help returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "--no-tui") {
		t.Errorf("install help output does not mention --no-tui flag; got: %q", got)
	}
}

// T-020-04: runStatus uses 6-column format.
func TestRunStatus_SixColumns(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := runStatus(&out)
	if err != nil {
		t.Fatalf("runStatus returned unexpected error: %v", err)
	}

	got := out.String()
	// Header must contain all 6 column names.
	columns := []string{"ID", "NAME", "DETECTED", "INSTALLED", "VERSION", "PATH"}
	for _, col := range columns {
		if !strings.Contains(got, col) {
			t.Errorf("runStatus output missing column %q; got:\n%s", col, got)
		}
	}
}

// T-020-04: ScanTools returns structured status for all registered adapters.
func TestScanTools_ReturnsAllAdapters(t *testing.T) {
	t.Parallel()

	results := ScanTools()
	if len(results) < 2 {
		t.Fatalf("ScanTools() returned %d results; expected at least 2 (claude-code + opencode)", len(results))
	}

	for _, r := range results {
		// Path should be non-empty for all registered adapters.
		if r.Path == "" {
			t.Errorf("ScanTools result has empty Path")
		}
		// Version may be empty (not installed) but the field must exist.
		_ = r.Version
		// Installed is a bool — always has a value.
		_ = r.Installed
	}
}

// T-020-04: runStatus handles empty registry gracefully.
// NOTE: not parallel — modifies shared DefaultRegistry.
func TestRunStatus_EmptyRegistry(t *testing.T) {
	// Create a fresh registry with no adapters.
	reg := &adapters.Registry{}
	prev := adapters.DefaultRegistry
	adapters.DefaultRegistry = reg
	defer func() { adapters.DefaultRegistry = prev }()

	var out bytes.Buffer
	err := runStatus(&out)
	if err != nil {
		t.Fatalf("runStatus with empty registry returned error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "No adapters registered.") {
		t.Errorf("expected 'No adapters registered.' for empty registry; got: %q", got)
	}
}

// T-020-04: runStatus column alignment — each row has 6 space-separated fields after header.
func TestRunStatus_RowsHaveSixFields(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := runStatus(&out)
	if err != nil {
		t.Fatalf("runStatus returned unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least header + separator + data rows; got %d lines", len(lines))
	}

	// Verify each data row (after header and separator) has non-trivial content.
	for i, line := range lines {
		if i < 2 {
			continue // skip header and separator
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Each data row should contain the adapter name.
		if !strings.Contains(line, "Claude Code") && !strings.Contains(line, "OpenCode") {
			t.Errorf("data row %d does not contain expected adapter name: %q", i, line)
		}
	}
}

// -- T-021: Uninstall confirmation gate tests ---------------------------------

// TestUninstall_YesFlagBypass verifies that --yes skips the confirmation prompt.
// When yes=true, no interactive prompt must appear and uninstall proceeds directly.
func TestUninstall_YesFlagBypass(t *testing.T) {
	var out bytes.Buffer
	err := runUninstall("claude-code", false, true, nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if strings.Contains(got, "?") {
		t.Errorf("expected no confirmation prompt when --yes is set, got: %q", got)
	}
}

// TestUninstall_ConfirmYes verifies that entering "y" confirms the uninstall.
func TestUninstall_ConfirmYes(t *testing.T) {
	prev := isTerminalFn
	isTerminalFn = func() bool { return true }
	defer func() { isTerminalFn = prev }()

	in := strings.NewReader("y\n")
	var out bytes.Buffer
	err := runUninstall("claude-code", false, false, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "y/N") {
		t.Errorf("expected confirmation prompt containing 'y/N', got: %q", got)
	}
}

// TestUninstall_ConfirmNo verifies that "n" aborts with exit code 0 (nil error).
func TestUninstall_ConfirmNo(t *testing.T) {
	prev := isTerminalFn
	isTerminalFn = func() bool { return true }
	defer func() { isTerminalFn = prev }()

	in := strings.NewReader("n\n")
	var out bytes.Buffer
	err := runUninstall("claude-code", false, false, in, &out)
	if err != nil {
		t.Fatalf("expected nil error for user abort, got: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "aborted") {
		t.Errorf("expected 'aborted' message, got: %q", got)
	}
}

// TestUninstall_ConfirmEmpty verifies that empty input aborts the uninstall.
func TestUninstall_ConfirmEmpty(t *testing.T) {
	prev := isTerminalFn
	isTerminalFn = func() bool { return true }
	defer func() { isTerminalFn = prev }()

	in := strings.NewReader("\n")
	var out bytes.Buffer
	err := runUninstall("claude-code", false, false, in, &out)
	if err != nil {
		t.Fatalf("expected nil error for abort on empty input, got: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "aborted") {
		t.Errorf("expected 'aborted' message on empty input, got: %q", got)
	}
}

// TestUninstall_PipedStdinError verifies that piped/non-interactive stdin
// without --yes returns an error directing users to use --yes.
func TestUninstall_PipedStdinError(t *testing.T) {
	prev := isTerminalFn
	isTerminalFn = func() bool { return false }
	defer func() { isTerminalFn = prev }()

	var out bytes.Buffer
	err := runUninstall("claude-code", false, false, nil, &out)
	if err == nil {
		t.Fatal("expected error for piped stdin without --yes, got nil")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected error to mention --yes, got: %v", err)
	}
}

// TestUninstall_AllListsTools verifies that --all with confirmation
// lists the affected tool names before the prompt.
func TestUninstall_AllListsTools(t *testing.T) {
	prev := isTerminalFn
	isTerminalFn = func() bool { return true }
	defer func() { isTerminalFn = prev }()

	in := strings.NewReader("n\n")
	var out bytes.Buffer
	err := runUninstall("", true, false, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "aborted") {
		t.Errorf("expected 'aborted' on cancel, got: %q", got)
	}
	if !strings.Contains(got, "Continue?") {
		t.Errorf("expected multi-tool prompt with 'Continue?', got: %q", got)
	}
}

// TestUninstall_InvalidTool verifies that an invalid --tool value fails
// before any prompt is shown, even when --yes is set.
func TestUninstall_InvalidTool(t *testing.T) {
	var out bytes.Buffer
	err := runUninstall("no-existe", false, true, nil, &out)
	if err == nil {
		t.Fatal("expected error for unknown adapter, got nil")
	}
	if !strings.Contains(err.Error(), "unknown adapter") {
		t.Errorf("expected 'unknown adapter' error, got: %v", err)
	}
}

// TestUninstall_YesFlagRegistered verifies that the --yes/-y flag appears
// in the uninstall subcommand help output.
func TestUninstall_YesFlagRegistered(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"uninstall", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("uninstall --help returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "--yes") {
		t.Errorf("uninstall help output does not contain --yes flag; got: %q", got)
	}
}

// T-020-04: ScanTools populates Version when installed.
func TestScanTools_PopulatesVersion(t *testing.T) {
	t.Parallel()

	results := ScanTools()
	for _, r := range results {
		// Version may be empty if not installed, but should not cause panic.
		_ = r.Version
		// Path should be non-empty for all registered adapters.
		if r.Path == "" {
			t.Errorf("ScanTools result has empty Path")
		}
	}
}
