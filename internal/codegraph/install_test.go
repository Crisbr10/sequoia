package codegraph

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// resetDefaults restores the package-level function variables to their defaults.
// Call this in a defer after each test that overrides them.
func resetDefaults() {
	lookPath = exec.LookPath
	runCommand = func(ctx context.Context, name string, args ...string) error {
		cmd := exec.CommandContext(ctx, name, args...)
		return cmd.Run()
	}
	runCommandWithOutput = func(ctx context.Context, name string, args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		return buf.String(), err
	}
}

func TestInstall_AlreadyInstalled(t *testing.T) {
	lookPath = func(string) (string, error) { return "/usr/bin/codegraph", nil }
	defer resetDefaults()

	var buf bytes.Buffer
	result := Install(context.Background(), &buf)

	if !result.AlreadyInstalled {
		t.Error("expected AlreadyInstalled=true")
	}
	if result.Installed {
		t.Error("expected Installed=false")
	}
	if result.Failed {
		t.Error("expected Failed=false")
	}
	if result.Message == "" {
		t.Error("expected non-empty Message")
	}
	if !strings.Contains(buf.String(), "already installed") {
		t.Errorf("expected output to contain 'already installed', got: %s", buf.String())
	}
}

func TestInstall_FreshInstallSuccess(t *testing.T) {
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, name string, args ...string) error {
		return nil
	}
	runCommandWithOutput = func(_ context.Context, name string, args ...string) (string, error) {
		return "", nil
	}
	defer resetDefaults()

	var buf bytes.Buffer
	result := Install(context.Background(), &buf)

	if !result.Installed {
		t.Error("expected Installed=true")
	}
	if result.Failed {
		t.Error("expected Failed=false")
	}
	if !strings.Contains(buf.String(), "integration ready") {
		t.Errorf("expected output to contain 'integration ready', got: %s", buf.String())
	}
}

func TestInstall_DownloadFailure_NonBlocking(t *testing.T) {
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, name string, args ...string) error {
		return errors.New("network error")
	}
	defer resetDefaults()

	var buf bytes.Buffer
	result := Install(context.Background(), &buf)

	if !result.Failed {
		t.Error("expected Failed=true on download failure")
	}
	if result.AlreadyInstalled {
		t.Error("expected AlreadyInstalled=false")
	}
	if !strings.Contains(buf.String(), "skipped") {
		t.Errorf("expected output to contain 'skipped', got: %s", buf.String())
	}
}

func TestInstall_ConfigFailure_NonBlocking(t *testing.T) {
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, name string, args ...string) error {
		return nil // download succeeds
	}
	runCommandWithOutput = func(_ context.Context, name string, args ...string) (string, error) {
		return "no agents found", errors.New("config error")
	}
	defer resetDefaults()

	var buf bytes.Buffer
	result := Install(context.Background(), &buf)

	if !result.Failed {
		t.Error("expected Failed=true on config failure")
	}
	if !strings.Contains(buf.String(), "configuration failed") {
		t.Errorf("expected output to mention config failure, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "Run manually") {
		t.Errorf("expected output to contain manual fix instructions, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "no agents found") {
		t.Errorf("expected output to include command output, got: %s", buf.String())
	}
}

func TestInstall_ConfigFailure_EmptyOutput(t *testing.T) {
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, name string, args ...string) error {
		return nil
	}
	runCommandWithOutput = func(_ context.Context, name string, args ...string) (string, error) {
		return "", errors.New("exit status 1")
	}
	defer resetDefaults()

	var buf bytes.Buffer
	result := Install(context.Background(), &buf)

	if !result.Failed {
		t.Error("expected Failed=true")
	}
	// Empty output should NOT produce an "Output:" line.
	if strings.Contains(buf.String(), "Output:") {
		t.Errorf("expected no 'Output:' line when command output is empty, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "Run manually") {
		t.Errorf("expected manual fix instructions, got: %s", buf.String())
	}
}

func TestInstall_NilWriter_NoPanic(t *testing.T) {
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, name string, args ...string) error { return nil }
	runCommandWithOutput = func(_ context.Context, name string, args ...string) (string, error) {
		return "", nil
	}
	defer resetDefaults()

	// Should not panic with nil writer.
	result := Install(context.Background(), nil)

	if result.AlreadyInstalled {
		t.Error("expected AlreadyInstalled=false")
	}
	if !result.Installed {
		t.Error("expected Installed=true")
	}
}

func TestInstall_NilWriter_AlreadyInstalled_NoPanic(t *testing.T) {
	lookPath = func(string) (string, error) { return "/usr/bin/codegraph", nil }
	defer resetDefaults()

	result := Install(context.Background(), nil)

	if !result.AlreadyInstalled {
		t.Error("expected AlreadyInstalled=true")
	}
}

func TestInstall_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	runCommand = func(_ context.Context, name string, args ...string) error { return nil }
	runCommandWithOutput = func(_ context.Context, name string, args ...string) (string, error) {
		return "", nil
	}
	defer resetDefaults()

	result := Install(ctx, nil)

	if !result.Failed {
		t.Error("expected Failed=true on cancelled context")
	}
	if !strings.Contains(result.Message, "cancelled") {
		t.Errorf("expected message to mention cancellation, got: %s", result.Message)
	}
}

// -- Status tests ---------------------------------------------------------------

func TestStatus_NotDetected(t *testing.T) {
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	defer resetDefaults()

	result := Status()

	if result.Detected {
		t.Error("expected Detected=false when codegraph not on PATH")
	}
	if result.Version != "" {
		t.Errorf("expected empty Version, got: %q", result.Version)
	}
	if result.Path != "" {
		t.Errorf("expected empty Path, got: %q", result.Path)
	}
}

func TestStatus_DetectedWithVersion(t *testing.T) {
	lookPath = func(string) (string, error) { return "/usr/local/bin/codegraph", nil }
	runCommandWithOutput = func(_ context.Context, name string, args ...string) (string, error) {
		if name == "codegraph" {
			return "0.9.7\n", nil
		}
		return "", nil
	}
	defer resetDefaults()

	result := Status()

	if !result.Detected {
		t.Error("expected Detected=true")
	}
	if result.Path != "/usr/local/bin/codegraph" {
		t.Errorf("expected Path=/usr/local/bin/codegraph, got: %q", result.Path)
	}
	if result.Version != "0.9.7" {
		t.Errorf("expected Version=0.9.7, got: %q", result.Version)
	}
}

func TestStatus_Detected_VersionFailure(t *testing.T) {
	lookPath = func(string) (string, error) { return "C:\\tools\\codegraph.exe", nil }
	runCommandWithOutput = func(_ context.Context, name string, args ...string) (string, error) {
		return "", errors.New("exit status 1")
	}
	defer resetDefaults()

	result := Status()

	if !result.Detected {
		t.Error("expected Detected=true even when --version fails")
	}
	if result.Path != "C:\\tools\\codegraph.exe" {
		t.Errorf("expected Path set, got: %q", result.Path)
	}
	if result.Version != "" {
		t.Errorf("expected empty Version on --version failure, got: %q", result.Version)
	}
}
