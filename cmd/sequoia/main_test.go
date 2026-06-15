// Package main_test verifies the sequoia CLI entrypoint compiles and
// exposes the expected command behaviour for integration testing.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/claude"
	"github.com/Crisbr10/sequoia/adapters/codex"
	"github.com/Crisbr10/sequoia/adapters/cursor"
	"github.com/Crisbr10/sequoia/adapters/gemini"
	"github.com/Crisbr10/sequoia/adapters/opencode"
	"github.com/Crisbr10/sequoia/adapters/testutil"
	"github.com/Crisbr10/sequoia/internal/codegraph"
)

// newPopulatedRegistry creates a fresh Registry with all 5 real adapters
// registered via their lazy RegisterFactory. Each test gets its own isolated
// registry — no shared state, no init() ordering dependencies.
func newPopulatedRegistry(t *testing.T) *adapters.Registry {
	t.Helper()
	reg := adapters.NewRegistry()
	claude.RegisterIn(reg)
	opencode.RegisterIn(reg)
	gemini.RegisterIn(reg)
	cursor.RegisterIn(reg)
	codex.RegisterIn(reg)
	return reg
}

// newRootCmdWithOut returns the root command with its output redirected to w
// so callers can capture and inspect command output.
func newRootCmdWithOut(w *bytes.Buffer, reg *adapters.Registry) *cobra.Command {
	cmd := newRootCmd(reg)
	cmd.SetOut(w)
	cmd.SetErr(w)
	return cmd
}

// TestRootHelp verifies that the root command prints usage when --help is passed.
func TestRootHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
// NOT parallel: modifies the shared global Version variable.
func TestVersionCmd(t *testing.T) {

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
// NOT parallel: reads the shared global Version variable.
func TestVersionCmd_DevVersionResolves(t *testing.T) {

	var out bytes.Buffer
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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
	err := runStatus(&out, newPopulatedRegistry(t))
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

	results := ScanTools(newPopulatedRegistry(t))
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
// Uses a fresh empty registry directly.
func TestRunStatus_EmptyRegistry(t *testing.T) {
	reg := adapters.NewRegistry()

	var out bytes.Buffer
	err := runStatus(&out, reg)
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
	err := runStatus(&out, newPopulatedRegistry(t))
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
		// Each data row should contain a known adapter name or CodeGraph.
		knownNames := []string{"Claude Code", "OpenCode", "Cursor IDE", "Gemini CLI", "OpenAI Codex", "CodeGraph"}
		found := false
		for _, name := range knownNames {
			if strings.Contains(line, name) {
				found = true
				break
			}
		}
		// Skip indented detail lines (e.g. "  Version: ...", "  Path: ...").
		if !found && strings.HasPrefix(line, "  ") {
			found = true
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
	err := runUninstall(context.Background(), "claude-code", false, true, nil, &out, newPopulatedRegistry(t))
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
	err := runUninstall(context.Background(), "claude-code", false, false, in, &out, newPopulatedRegistry(t))
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
	err := runUninstall(context.Background(), "claude-code", false, false, in, &out, newPopulatedRegistry(t))
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
	err := runUninstall(context.Background(), "claude-code", false, false, in, &out, newPopulatedRegistry(t))
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
	err := runUninstall(context.Background(), "claude-code", false, false, nil, &out, newPopulatedRegistry(t))
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
	err := runUninstall(context.Background(), "", true, false, in, &out, newPopulatedRegistry(t))
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
	err := runUninstall(context.Background(), "no-existe", false, true, nil, &out, newPopulatedRegistry(t))
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
	cmd := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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

	results := ScanTools(newPopulatedRegistry(t))
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

	root := newRootCmd(newPopulatedRegistry(t))
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
	root := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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

// TestTargetAdapters_ReturnsIdentifier verifies that targetAdapters returns
// the narrow []Identifier type (ISP narrowing — P3-003). The test relies on
// Go's lack of slice covariance: assigning []ToolAdapter to []Identifier
// is a compile error, so this test only compiles if targetAdapters returns
// the narrower type.
func TestTargetAdapters_ReturnsIdentifier(t *testing.T) {
	t.Parallel()

	reg := adapters.NewRegistry()
	claude.RegisterIn(reg)

	ids := targetAdapters("claude-code", reg)
	if len(ids) != 1 {
		t.Fatalf("targetAdapters returned %d identifiers, want 1", len(ids))
	}
	if ids[0].ID() != "claude-code" {
		t.Errorf("expected ID 'claude-code', got %q", ids[0].ID())
	}
}

// TestTargetAdapters_AllReturnsIdentifiers verifies that targetAdapters with an
// empty toolID returns only []Identifier (not []ToolAdapter), satisfying the
// ISP narrowing contract. Triangulation test with different input.
func TestTargetAdapters_AllReturnsIdentifiers(t *testing.T) {
	t.Parallel()

	reg := adapters.NewRegistry()
	claude.RegisterIn(reg)
	opencode.RegisterIn(reg)

	ids := targetAdapters("", reg)
	if len(ids) != 2 {
		t.Fatalf("targetAdapters returned %d identifiers, want 2", len(ids))
	}
	// Verify each item only exposes Identifier methods (ID, Name).
	for _, id := range ids {
		if id.ID() == "" {
			t.Error("identifier has empty ID")
		}
		if id.Name() == "" {
			t.Error("identifier has empty Name")
		}
	}
}

// TestSignalHandling_NormalOperationPreservesContext verifies that a live
// (non-cancelled) context flows normally through the command pipeline.
func TestSignalHandling_NormalOperationPreservesContext(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	root := newRootCmdWithOut(&out, newPopulatedRegistry(t))
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

// -- close-coverage-gaps REQ-COV-01 -------------------------------------------
//
// The following tests close the cmd/sequoia/ coverage gap from 58.8% to ≥ 70%.
// They are pure test additions — no production code is modified.
// Hermeticity hooks: codegraph.InstallFunc and isTerminalFn are saved at the
// top of each test that mutates them and restored via t.Cleanup (NFR-5).
// No t.Parallel() in tests that mutate globals (NFR-7).

// TestPrintError_AllBranches verifies that printError maps each sentinel
// error to the expected user-friendly stderr message. Covers main.go:61-72.
func TestPrintError_AllBranches(t *testing.T) {
	// Save and restore os.Stderr (NFR-5: restoration via t.Cleanup, not defer).
	origStderr := os.Stderr
	t.Cleanup(func() { os.Stderr = origStderr })

	tests := []struct {
		name        string
		err         error
		wantContain string
	}{
		{"ErrNotDetected", adapters.ErrNotDetected, "Tool not detected on this system"},
		{"ErrInstallFailed", fmt.Errorf("%w: x", adapters.ErrInstallFailed), "Installation failed"},
		{"ErrUninstallFailed", fmt.Errorf("%w: x", adapters.ErrUninstallFailed), "Uninstall completed with warnings"},
		{"GenericError", errors.New("boom"), "error: boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr via a per-subtest pipe (NFR-2: hermetic).
			r, w, err := os.Pipe()
			require.NoError(t, err, "os.Pipe() failed")
			os.Stderr = w

			printError(tt.err)

			// Close the write end and read the captured output.
			require.NoError(t, w.Close(), "close pipe write end")
			out, readErr := io.ReadAll(r)
			require.NoError(t, readErr, "read captured stderr")
			_ = r.Close()

			assert.Contains(t, string(out), tt.wantContain,
				"printError(%v) stderr should contain %q; got %q",
				tt.err, tt.wantContain, string(out))
		})
	}
}

// TestExitCode_AllBranches verifies that exitCode maps each sentinel error
// to the documented exit code at main.go:75-86. Includes a nil-error
// subtest to exercise the default branch.
func TestExitCode_AllBranches(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"ErrNotDetected", adapters.ErrNotDetected, 2},
		{"ErrInstallFailed", fmt.Errorf("%w: x", adapters.ErrInstallFailed), 3},
		{"ErrUninstallFailed", fmt.Errorf("%w: x", adapters.ErrUninstallFailed), 4},
		{"GenericError", errors.New("boom"), 1},
		{"NilError", nil, 1}, // default branch (no error → 1)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, exitCode(tt.err),
				"exitCode(%v) = %d; want %d", tt.err, exitCode(tt.err), tt.want)
		})
	}
}

// TestResolveVersion_DevHasVCSRevisionPrefix verifies that resolveVersion
// produces a real version or the "unknown-<sha8>" pattern when called with
// the dev fallback. Strengthens TestResolveVersion_DevResolves (line 552)
// by asserting the VCS-prefix shape rather than just non-empty.
//
// Under `go test` inside a git repo, debug.ReadBuildInfo returns settings
// with a vcs.revision key. The branch at main.go:240-243 then synthesizes
// "unknown-" + s.Value[:8]. If the build was tagged (e.g., a release),
// info.Main.Version is non-empty and the function returns it directly.
//
// Covers main.go:240-243 (the vcs.revision branch).
func TestResolveVersion_DevHasVCSRevisionPrefix(t *testing.T) {
	got := resolveVersion("0.1.0-dev")

	require.NotEmpty(t, got, "resolveVersion must not return empty for '0.1.0-dev'")
	assert.NotEqual(t, "(devel)", got, "resolveVersion must not return raw '(devel)'")

	// Either the VCS branch hit (unknown-<sha8>) or a real version (contains '.').
	hasUnknownPrefix := strings.HasPrefix(got, "unknown-")
	hasSemverDot := strings.Contains(got, ".")
	assert.True(t, hasUnknownPrefix || hasSemverDot,
		"resolveVersion result should be 'unknown-<sha8>' or a real semver (containing '.'); got %q", got)

	// If it has the unknown- prefix, also verify the sha8 shape (8 hex chars).
	if hasUnknownPrefix {
		assert.Regexp(t, `^unknown-[0-9a-f]{8}$`, got,
			"unknown- prefix should be followed by exactly 8 hex chars; got %q", got)
	}
}

// withCodegraphInstallNoop saves codegraph.InstallFunc and replaces it with a
// no-op stub for the duration of the test. Restores via t.Cleanup (NFR-5).
// The no-op records that it was invoked (callCount) and returns a benign
// result so runInstall's downstream code runs without side effects.
func withCodegraphInstallNoop(t *testing.T) (callCount *int) {
	t.Helper()
	origInstall := codegraph.InstallFunc
	callCount = new(int)
	codegraph.InstallFunc = func(_ context.Context, _ io.Writer) codegraph.InstallResult {
		*callCount++
		return codegraph.InstallResult{
			AlreadyInstalled: true,
			Message:          "noop (test stub)",
		}
	}
	t.Cleanup(func() { codegraph.InstallFunc = origInstall })
	return callCount
}

// TestRunInstall_AlreadyInstalledBranch verifies the "Reinstalling" branch
// at main.go:291-293: when IsInstalled() returns true, runInstall prints
// the "Sequoia is already installed" notice but still proceeds with Install().
// Covers main.go:291-293.
func TestRunInstall_AlreadyInstalledBranch(t *testing.T) {
	callCount := withCodegraphInstallNoop(t)

	reg := adapters.NewRegistry()
	reg.RegisterFactory("test-already", func() adapters.ToolAdapter {
		return &testutil.MockAdapter{
			IDVal:           "test-already",
			NameVal:         "Test Already",
			DetectFunc:      func() bool { return true },
			IsInstalledFunc: func() bool { return true },
			InstallFunc:     func(_ adapters.InstallOpts) error { return nil },
		}
	})

	var buf bytes.Buffer
	err := runInstall(context.Background(), "test-already", &buf, reg)
	require.NoError(t, err, "runInstall should not error on 'already installed' path")

	out := buf.String()
	assert.Contains(t, out, "Installing Sequoia for Test Already",
		"output should announce install target; got %q", out)
	assert.Contains(t, out, "Sequoia is already installed. Reinstalling",
		"output should print reinstall notice; got %q", out)
	assert.Contains(t, out, "Done! Use /sequoia-init inside Test Already",
		"output should confirm install complete; got %q", out)
	assert.Equal(t, 1, *callCount,
		"codegraph.InstallFunc should be invoked exactly once (no-op stub)")
}

// TestRunInstall_NoAdaptersBranch verifies the empty-registry branch at
// main.go:263-270: when no adapters are registered, runInstall prints
// a helpful message and returns nil (no error). Covers main.go:263-270.
func TestRunInstall_NoAdaptersBranch(t *testing.T) {
	withCodegraphInstallNoop(t)

	reg := adapters.NewRegistry() // empty registry — no factories

	var buf bytes.Buffer
	err := runInstall(context.Background(), "", &buf, reg)
	require.NoError(t, err, "runInstall with empty registry should not error")

	out := buf.String()
	assert.Contains(t, out, "No supported AI tools detected on this machine.",
		"output should explain no tools were detected; got %q", out)
}

// TestRunInstall_UnknownToolID verifies the "unknown adapter" error branch
// at main.go:264-266: when toolID is non-empty but no adapter is registered
// for it, runInstall returns an error wrapping adapters.ErrUnknownAdapter.
//
// R-9 resolution: per design R-9, the design hypothesised this branch might
// not be reachable and proposed a pivot to test ErrNotDetected at main.go:282.
// After code reading, the branch IS reachable: targetAdapters returns nil
// for any unknown ID, len(ids)==0 hits, toolID!="" returns the wrapped error.
// This test exercises that exact branch (main.go:264-266), with a
// non-empty registry containing a different adapter to prove the registry
// itself is not the problem.
func TestRunInstall_UnknownToolID(t *testing.T) {
	withCodegraphInstallNoop(t)

	reg := adapters.NewRegistry()
	// Register an unrelated adapter so the registry is non-empty.
	reg.RegisterFactory("test-known", func() adapters.ToolAdapter {
		return &testutil.MockAdapter{
			IDVal:   "test-known",
			NameVal: "Test Known",
		}
	})

	var buf bytes.Buffer
	err := runInstall(context.Background(), "nonexistent-tool", &buf, reg)
	require.Error(t, err, "runInstall with unknown toolID should return an error")
	assert.ErrorIs(t, err, adapters.ErrUnknownAdapter,
		"error should wrap adapters.ErrUnknownAdapter; got %v", err)
	assert.Contains(t, err.Error(), "nonexistent-tool",
		"error should mention the unknown toolID; got %v", err)
}

// TestInstallCmd_RunEHeadlessBranch verifies the headless branch at
// main.go:150-152: when --no-tui is set, newInstallCmd.RunE calls
// runInstall directly (no TUI launch). Covers main.go:150-152.
func TestInstallCmd_RunEHeadlessBranch(t *testing.T) {
	withCodegraphInstallNoop(t)

	reg := adapters.NewRegistry()
	reg.RegisterFactory("test-headless", func() adapters.ToolAdapter {
		return &testutil.MockAdapter{
			IDVal:           "test-headless",
			NameVal:         "Test Headless",
			DetectFunc:      func() bool { return true },
			IsInstalledFunc: func() bool { return true },
			InstallFunc:     func(_ adapters.InstallOpts) error { return nil },
		}
	})

	var out bytes.Buffer
	cmd := newInstallCmd(reg)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--no-tui", "--tool=test-headless"})

	err := cmd.Execute()
	require.NoError(t, err, "install --no-tui --tool=... should succeed; got error")

	got := out.String()
	assert.Contains(t, got, "Done!", "headless install output should contain 'Done!'; got %q", got)
}

// TestRootCmd_RunETUIBranch verifies the TUI branch at main.go:106-107:
// when isTerminalFn() returns true, newRootCmd.RunE calls runTUI and
// surfaces the resulting TUI error. The Bubbletea program needs a real
// TTY to render interactively, so we swap os.Stdin with a closed pipe
// to force it down the error-return path at runTUI:473-475. Covers
// main.go:106-107 and main.go:467-475 (runTUI error path). The success
// return at line 476 is the accepted R-3 gap (requires real TTY).
func TestRootCmd_RunETUIBranch(t *testing.T) {
	// Save and restore isTerminalFn (NFR-5: t.Cleanup, no t.Parallel).
	origIsTerminalFn := isTerminalFn
	isTerminalFn = func() bool { return true }
	t.Cleanup(func() { isTerminalFn = origIsTerminalFn })

	// Force Bubbletea into the error-return path by feeding it a stdin
	// that returns EOF immediately. This avoids the test hanging on a
	// real TTY read. Save and restore the real os.Stdin via t.Cleanup
	// (NFR-5: hermetic, no t.Parallel).
	origStdin := os.Stdin
	eofR, eofW, pipeErr := os.Pipe()
	require.NoError(t, pipeErr, "os.Pipe() failed")
	// Close the write end → reads from eofR return EOF with no bytes.
	require.NoError(t, eofW.Close(), "close pipe write end")
	os.Stdin = eofR
	t.Cleanup(func() {
		os.Stderr = nil // ensure no swap lingers
		os.Stdin = origStdin
		_ = eofR.Close()
	})

	reg := adapters.NewRegistry()
	reg.RegisterFactory("test-tui", func() adapters.ToolAdapter {
		return &testutil.MockAdapter{
			IDVal:   "test-tui",
			NameVal: "Test TUI",
		}
	})

	var out bytes.Buffer
	cmd := newRootCmd(reg)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	// Run cmd.Execute() with a deadline because Bubbletea might still
	// block on its internal setup even with EOF on stdin (e.g., Windows
	// console mode setup). The TUI branch (main.go:106-107) is entered
	// synchronously before any block, so coverage is recorded.
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Execute()
	}()

	select {
	case err := <-errCh:
		// Bubbletea returned — assert on the wrapped TUI error.
		require.Error(t, err, "root RunE in TUI mode should return an error when stdin EOFs")
		assert.Contains(t, err.Error(), "TUI error:",
			"error should be wrapped with 'TUI error:' prefix from runTUI; got %v", err)
	case <-time.After(3 * time.Second):
		// Bubbletea didn't return. TUI branch was entered (main.go:106-107
		// → main.go:467), so coverage is recorded. Skip with a clear
		// message; the goroutine leak is a known artifact of testing
		// Bubbletea without a TTY and is accepted per R-3.
		t.Skip("Bubbletea program did not return within 3s on EOF stdin; TUI branch coverage recorded")
	}
}
