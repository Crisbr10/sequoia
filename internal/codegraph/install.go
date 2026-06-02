// Package codegraph provides CodeGraph installation utilities.
// CodeGraph is a complementary tool to Sequoia — not an AI adapter.
package codegraph

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
)

// errCancelled is returned when the context is cancelled before an operation completes.
var errCancelled = fmt.Errorf("cancelled by context")

// InstallResult reports what happened during CodeGraph installation.
type InstallResult struct {
	AlreadyInstalled bool
	Installed        bool
	Failed           bool
	Message          string
}

// StatusResult reports the current state of CodeGraph on this machine.
type StatusResult struct {
	// Detected reports whether the "codegraph" binary is on PATH.
	Detected bool
	// Version is the CodeGraph version string (e.g. "0.9.7"), or "" if not detected.
	Version string
	// Path is the full path to the codegraph binary, or "" if not detected.
	Path string
}

// Status checks whether CodeGraph is installed on this machine.
// It uses exec.LookPath to find the binary and "codegraph --version" to get the version.
// This function never returns an error — failures are reported via Detected=false.
func Status() StatusResult {
	path, err := lookPath("codegraph")
	if err != nil {
		return StatusResult{}
	}

	version := ""
	// Best-effort version detection; failure is non-fatal.
	output, err := runCommandWithOutput(context.Background(), "codegraph", "--version")
	if err == nil {
		version = strings.TrimSpace(output)
	}

	return StatusResult{
		Detected: true,
		Version:  version,
		Path:     path,
	}
}

var lookPath = exec.LookPath

var runCommand = func(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

// runCommandWithOutput runs a command and captures its combined stdout+stderr output.
// Returns the error (if any) and the captured output as a string.
var runCommandWithOutput = func(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// InstallFunc attempts to install and configure CodeGraph as a non-blocking convenience step.
// It mirrors the logic in scripts/install.sh and scripts/install.ps1:
//  1. Check if "codegraph" is on PATH → skip if yes
//  2. Download CodeGraph via official installer
//  3. Run "codegraph install --target=auto --location=global --yes" to configure agents
//  4. Any failure produces a warning; the function never returns an error
//
// out receives progress messages (may be nil to suppress output).
var InstallFunc = func(ctx context.Context, out io.Writer) InstallResult {
	// Check for cancellation before starting.
	if ctx.Err() != nil {
		return InstallResult{Failed: true, Message: "CodeGraph install skipped: " + errCancelled.Error()}
	}

	if out != nil {
		_, _ = fmt.Fprintln(out, "Checking CodeGraph installation...")
	}

	if _, err := lookPath("codegraph"); err == nil {
		if out != nil {
			_, _ = fmt.Fprintln(out, "CodeGraph is already installed.")
		}
		return InstallResult{
			AlreadyInstalled: true,
			Message:          "CodeGraph is already installed.",
		}
	}

	if out != nil {
		_, _ = fmt.Fprintln(out, "Installing CodeGraph for enhanced code intelligence...")
	}

	var installErr error
	switch runtime.GOOS {
	case "windows":
		installErr = runCommand(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-Command",
			`irm https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.ps1 | iex`)
	default:
		installErr = runCommand(ctx, "sh", "-c", "curl -fsSL https://raw.githubusercontent.com/colbymchenry/codegraph/main/install.sh | sh")
	}

	if installErr != nil {
		if ctx.Err() != nil {
			return InstallResult{Failed: true, Message: "CodeGraph install skipped: " + errCancelled.Error()}
		}
		if out != nil {
			_, _ = fmt.Fprintf(out, "CodeGraph installation skipped (network issue or unsupported platform).\n")
			_, _ = fmt.Fprintf(out, "Sequoia works fine without it — /sequoia-dev will fall back to file-based exploration.\n")
		}
		return InstallResult{
			Failed:  true,
			Message: fmt.Sprintf("CodeGraph download failed: %v", installErr),
		}
	}

	if out != nil {
		_, _ = fmt.Fprintln(out, "CodeGraph installed. Auto-configuring agents...")
	}

	cfgOutput, cfgErr := runCommandWithOutput(ctx, "codegraph", "install", "--target=auto", "--location=global", "--yes")
	if cfgErr != nil {
		detail := strings.TrimSpace(cfgOutput)
		if out != nil {
			_, _ = fmt.Fprintln(out, "CodeGraph agent configuration failed (non-blocking).")
			if detail != "" {
				_, _ = fmt.Fprintf(out, "  Output: %s\n", detail)
			}
			_, _ = fmt.Fprintln(out, "  Run manually: codegraph install --target=auto --location=global --yes")
		}
		return InstallResult{
			Failed:  true,
			Message: fmt.Sprintf("CodeGraph installed but agent configuration failed: %v", cfgErr),
		}
	}

	if out != nil {
		_, _ = fmt.Fprintln(out, "CodeGraph integration ready.")
	}
	return InstallResult{
		Installed: true,
		Message:   "CodeGraph installed and configured successfully.",
	}
}

// Install delegates to InstallFunc. Exposed as a variable so tests can override it.
func Install(ctx context.Context, out io.Writer) InstallResult {
	return InstallFunc(ctx, out)
}
