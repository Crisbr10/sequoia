// Package pipeline — timeout test for codegraphInstallProgress (REQ-TUI-05).
//
// This file lives in `package pipeline` (not `pipeline_test`) because it must
// override the unexported `codegraphProgressTimeout` package var. The existing
// runner_test.go is in `package pipeline_test` (external) and exercises the
// happy path under the default 5-minute timeout.
//
// CRITICAL setup requirements (from design D2 + risks in obs #914):
//   - Save AND restore BOTH codegraph.InstallFunc AND codegraphProgressTimeout
//     via t.Cleanup. Missing either override causes the test to block for the
//     full 5-minute default timeout and waste CI minutes.
//   - Save AND restore CI / GITHUB_ACTIONS env vars: the existing
//     init() in runner_test.go sets them to "true" to suppress real exec, but
//     our test needs to BYPASS the skip path to actually exercise the
//     install mock.
//   - Use t.Deadline() guard: if the test deadline is too tight (<500ms),
//     skip rather than block CI.
//
// This test does NOT call t.Parallel() — package-level vars are racy under
// parallel execution and the override pattern is per-test sequential.
package pipeline

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Crisbr10/sequoia/internal/codegraph"
	"github.com/Crisbr10/sequoia/internal/model"
)

// TestCodegraphInstallProgress_TimeoutFires verifies REQ-TUI-05:
//
//	codegraphInstallProgress SHALL wrap the install call in a 5-min
//	context.WithTimeout, and on ctx.Done() SHALL emit a Warning
//	ProgressMsg with Error="download timed out" and return.
//
// RED signal: this test references `codegraphProgressTimeout`, which does
// not exist on main. The build fails — that IS the RED.
//
// GREEN signal: after the production var + timeout wrap are added, the test
// passes — the mock blocks on ctx.Done(), the 50ms deadline fires, the
// function emits the warning and returns.
func TestCodegraphInstallProgress_TimeoutFires(t *testing.T) {
	// t.Deadline() guard — skip if the test runner's deadline is too tight.
	// (per design risk: a missed override would block for 5 minutes.)
	if deadline, ok := t.Deadline(); ok && time.Until(deadline) < 500*time.Millisecond {
		t.Skip("test deadline too tight for timeout test (need >= 500ms)")
	}

	// Save and restore ALL overrides. Each missing override has a different
	// failure mode — see design D2 + risks in obs #914.
	origTimeout := codegraphProgressTimeout
	origInstall := codegraph.InstallFunc
	origCI := os.Getenv("CI")
	origGA := os.Getenv("GITHUB_ACTIONS")
	t.Cleanup(func() {
		codegraphProgressTimeout = origTimeout
		codegraph.InstallFunc = origInstall
		os.Setenv("CI", origCI)
		os.Setenv("GITHUB_ACTIONS", origGA)
	})

	// Override 1: shrink the timeout from 5min (production default) to 50ms.
	codegraphProgressTimeout = 50 * time.Millisecond

	// Bypass the CI skip path inside codegraphInstallProgress so we actually
	// reach the codegraph.Install call site and exercise the timeout.
	os.Unsetenv("CI")
	os.Unsetenv("GITHUB_ACTIONS")

	// Override 2: replace codegraph.Install with a mock that blocks until
	// ctx fires (simulates a hung HTTPS download). When ctx fires at 50ms,
	// the mock returns a sentinel failed result — but the production code
	// must detect the timeout BEFORE branching on result and emit its own
	// "download timed out" warning.
	codegraph.InstallFunc = func(ctx context.Context, out io.Writer) codegraph.InstallResult {
		<-ctx.Done()
		return codegraph.InstallResult{
			Failed:  true,
			Message: "blocked-by-mock",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan model.ProgressMsg, 8)

	// Run codegraphInstallProgress in a goroutine and assert it returns
	// within a generous safety window. The override is 50ms; we give it
	// 500ms of slack for goroutine scheduling and the post-timeout warning
	// emit. If the function takes longer, the timeout-wrap is missing.
	done := make(chan struct{})
	start := time.Now()
	go func() {
		codegraphInstallProgress(ctx, ch)
		close(done)
	}()

	select {
	case <-done:
		elapsed := time.Since(start)
		if elapsed > 500*time.Millisecond {
			t.Fatalf("codegraphInstallProgress took %s (expected <= 500ms)", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("codegraphInstallProgress did not return within 2s — timeout wrap missing")
	}

	// Drain the channel looking for the Warning "download timed out" message.
	// The 100ms drain window is enough for the buffered channel (cap 8) to
	// be fully consumed after the function has returned.
	var foundWarning bool
	drain := time.After(100 * time.Millisecond)
drainLoop:
	for {
		select {
		case msg := <-ch:
			if msg.Warning && msg.Error == "download timed out" {
				foundWarning = true
			}
		case <-drain:
			break drainLoop
		}
	}

	if !foundWarning {
		t.Fatal(`expected Warning ProgressMsg with Error="download timed out", got none`)
	}
}
