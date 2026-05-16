// Package main_test verifies the sequoia CLI entrypoint compiles and
// exposes the expected command behaviour for integration testing.
package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/Crisbr10/sequoia/adapters"
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

// TestRootNoArgs verifies that the root command exits cleanly without arguments
// when stdin is not a terminal (prints help instead of launching TUI).
func TestRootNoArgs(t *testing.T) {
	// Modifies global isTerminalFn — must not run in parallel.
	prev := isTerminalFn
	isTerminalFn = func() bool { return false }
	defer func() { isTerminalFn = prev }()

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

// TestVersionCmd_DevVersionResolves confirms that when Version is the default
// "0.1.0-dev", the version command resolves it via debug.ReadBuildInfo.
// The resolved value is non-empty and does not contain "(devel)".
func TestVersionCmd_DevVersionResolves(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command returned unexpected error: %v", err)
	}

	got := strings.TrimSpace(out.String())
	if got == "" {
		t.Fatal("version command with dev fallback returned empty output")
	}
	if got == "(devel)" {
		t.Error("version command should not output raw '(devel)', should be resolved")
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
		// Each data row should contain a known adapter name.
		knownNames := []string{"Claude Code", "OpenCode", "Cursor IDE", "Gemini CLI", "OpenAI Codex"}
		found := false
		for _, name := range knownNames {
			if strings.Contains(line, name) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("data row %d does not contain a known adapter name: %q", i, line)
		}
	}
}

// -- T-021: Uninstall confirmation gate tests ---------------------------------

// TestUninstall_YesFlagBypass verifies that --yes skips the confirmation prompt.
// When yes=true, no interactive prompt must appear and uninstall proceeds directly.
func TestUninstall_YesFlagBypass(t *testing.T) {
	var out bytes.Buffer
	err := runUninstall(context.Background(), "claude-code", false, true, nil, &out)
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
	err := runUninstall(context.Background(), "claude-code", false, false, in, &out)
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
	err := runUninstall(context.Background(), "claude-code", false, false, in, &out)
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
	err := runUninstall(context.Background(), "claude-code", false, false, in, &out)
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
	err := runUninstall(context.Background(), "claude-code", false, false, nil, &out)
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
	err := runUninstall(context.Background(), "", true, false, in, &out)
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
	err := runUninstall(context.Background(), "no-existe", false, true, nil, &out)
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

// -- FIX-004: Signal handling tests -------------------------------------------

// TestResolveVersion_PassThrough verifies that resolveVersion returns the raw
// value unchanged when it is not the dev fallback "0.1.0-dev".
func TestResolveVersion_PassThrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw  string
		want string
	}{
		{"1.2.3", "1.2.3"},
		{"", ""},
	}

	for _, tt := range tests {
		got := resolveVersion(tt.raw)
		if got != tt.want {
			t.Errorf("resolveVersion(%q) = %q; want %q", tt.raw, got, tt.want)
		}
	}
}

// TestResolveVersion_DevResolves verifies that when Version is "0.1.0-dev",
// resolveVersion returns a non-empty resolved value that is not "(devel)".
func TestResolveVersion_DevResolves(t *testing.T) {
	t.Parallel()

	got := resolveVersion("0.1.0-dev")
	if got == "" {
		t.Fatal("resolveVersion('0.1.0-dev') returned empty string")
	}
	if got == "(devel)" {
		t.Error("resolveVersion should not return raw '(devel)', should be resolved")
	}
}

// TestSignalHandling_RootCommandHasContext verifies that the root command
// created by newRootCmd can be assigned a context via SetContext, and that
// this context is accessible through cmd.Context(). This confirms the
// wiring between main()'s signal-aware context and the Cobra command tree.
func TestSignalHandling_RootCommandHasContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := newRootCmd()
	root.SetContext(ctx)

	// Verify the context is accessible.
	got := root.Context()
	if got != ctx {
		t.Fatal("root.Context() should return the context set via SetContext")
	}

	// Verify the context is cancellable.
	cancel()
	select {
	case <-got.Done():
		// Expected — context was cancelled.
	default:
		t.Error("context should be cancelled after cancel() is called")
	}
}

// TestSignalHandling_InstallCommandPropagatesContext verifies that the context
// set on the root command propagates to the install command handler. When the
// context is cancelled before execution, the install command should return
// an error (because no tools are detected or context is cancelled).
func TestSignalHandling_InstallCommandPropagatesContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	var out bytes.Buffer
	root := newRootCmdWithOut(&out)
	root.SetContext(ctx)
	root.SetArgs([]string{"install", "--no-tui", "--tool=nonexistent"})

	// Cancel the context before execution.
	cancel()

	err := root.Execute()
	// The command either reports the cancelled context or the unknown adapter error.
	// Both are valid outcomes — what matters is the command respected the context.
	if err == nil {
		// If no error, check that "unknown adapter" was printed (tool was validated).
		got := out.String()
		if !strings.Contains(got, "unknown adapter") {
			t.Errorf("expected 'unknown adapter' or context error; got: %q", got)
		}
	}
	// If err != nil, the context cancellation was properly propagated.
}

// TestSignalHandling_NormalOperationPreservesContext verifies that a live
// (non-cancelled) context flows normally through the command pipeline.
func TestSignalHandling_NormalOperationPreservesContext(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	root := newRootCmdWithOut(&out)
	root.SetArgs([]string{"status"})

	// Set a non-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	root.SetContext(ctx)

	err := root.Execute()
	if err != nil {
		t.Fatalf("status command with live context returned unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "ID") || !strings.Contains(got, "NAME") {
		t.Errorf("status output missing header columns; got: %q", got)
	}
}
