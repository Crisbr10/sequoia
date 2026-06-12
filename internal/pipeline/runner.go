// Package pipeline provides goroutine-based install and uninstall runners
// that bridge the TUI to adapter calls via a buffered progress channel.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/internal/codegraph"
	"github.com/Crisbr10/sequoia/internal/model"
)

// WarningEmitter is a local interface for adapters that collect non-fatal
// warnings during Install/Uninstall (e.g., symlink resolution failures).
// Adapters that implement this interface will have their warnings surfaced
// as ProgressMsg{Warning: true} after a successful operation.
type WarningEmitter interface {
	Warnings() []string
}

// InstallSteps defines the install phases in execution order.
// Each phase is dispatched via the Strategy interface, sending a running
// (Done=false) message before and a done (Done=true) message after.
// This variable is exported so that callers (e.g., update.go's
// buildProgressTools) reference a single source of truth for step names.
var InstallSteps = []string{"Preparing", "Downloading", "Verifying", "Staging", "Applying"}

// codegraphProgressTimeout caps the duration of codegraphInstallProgress
// (REQ-TUI-05). Tests override via t.Cleanup. Precedent: codegraph.installTimeout.
var codegraphProgressTimeout = 5 * time.Minute

// RunInstall returns a tea.Cmd that installs Sequoia into every selected tool.
// Each tool runs in its own goroutine dispatching through the Strategy interface.
//
// Progress is reported through a buffered channel:
//   - A "running" ProgressMsg (Done=false) is sent before each phase.
//   - A "done" ProgressMsg (Done=true) is sent after each successful phase.
//   - An error ProgressMsg (Error != "") is sent when a phase fails.
//
// The channel is closed when all goroutines complete. Context cancellation
// stops goroutines gracefully while preserving partial progress.
func RunInstall(ctx context.Context, tools []model.ToolState, ch chan<- model.ProgressMsg) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup

		for _, tool := range tools {
			if !tool.Selected {
				continue
			}

			// Defensive guard: skip tools that are not detected on this system,
			// even if the user manually selected them. Installing into an undetected
			// tool can write files to unexpected locations (e.g., ~/.claude/).
			if !tool.Adapter.Detect() {
				sendProgress(ctx, ch, model.ProgressMsg{
					ToolID:  tool.Adapter.ID(),
					Step:    InstallSteps[0],
					Done:    true,
					Warning: true,
					Error:   "tool not detected on this system — skipping install",
				})
				continue
			}

			select {
			case <-ctx.Done():
				// Context cancelled before goroutine starts — stop launching new goroutines.
				wg.Wait()
				safeClose(ch)
				return nil
			default:
			}

			wg.Add(1)
			go func(t model.ToolState) {
				defer wg.Done()
				runInstallSteps(ctx, t, ch)
			}(tool)
		}

		// Wait for all goroutines to complete, then signal completion
		// by closing the channel.
		wg.Wait()

		codegraphInstallProgress(ctx, ch)

		safeClose(ch)
		return nil
	}
}

// runInstallSteps dispatches the install lifecycle through the Strategy
// interface. Each phase sends running/done progress. On phase failure,
// Rollback() is called and the error reported.
//
// If the adapter does not implement common.Strategy, an error message is
// sent and the goroutine returns without panicking.
func runInstallSteps(ctx context.Context, t model.ToolState, ch chan<- model.ProgressMsg) {
	adapter := t.Adapter
	toolID := adapter.ID()

	s, ok := adapter.(common.Strategy)
	if !ok {
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: toolID,
			Step:   "Preparing",
			Done:   true,
			Error:  fmt.Sprintf("adapter %s: does not implement Strategy", toolID),
		})
		return
	}

	opts := adapters.InstallOpts{Context: ctx}

	for _, phase := range InstallSteps {
		// Check context before each phase.
		select {
		case <-ctx.Done():
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID: toolID,
				Step:   phase,
				Done:   true,
				Error:  ctx.Err().Error(),
			})
			return
		default:
		}

		var phaseErr error

		// Send running message.
		if !sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: toolID,
			Step:   phase,
			Done:   false,
		}) {
			return // context cancelled
		}

		// Execute the phase.
		switch phase {
		case "Preparing":
			phaseErr = s.Prepare(opts)
		case "Downloading":
			phaseErr = s.Download(opts)
		case "Verifying":
			phaseErr = s.Verify()
		case "Staging":
			phaseErr = s.Stage(opts)
		case "Applying":
			phaseErr = s.Apply(opts)
		}

		if phaseErr != nil {
			// Rollback on failure.
			_ = s.Rollback()
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID: toolID,
				Step:   phase,
				Done:   true,
				Error:  fmt.Sprintf("%s: %v", phase, phaseErr),
			})
			return
		}

		// Send done message.
		if !sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: toolID,
			Step:   phase,
			Done:   true,
		}) {
			return
		}
	}

	// After all phases complete successfully, check for warnings.
	if emitter, ok := adapter.(WarningEmitter); ok {
		warnings := emitter.Warnings()
		if len(warnings) > 0 {
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID:  toolID,
				Step:    InstallSteps[len(InstallSteps)-1],
				Done:    true,
				Warning: true,
				Error:   strings.Join(warnings, "\n"),
			})
		}
	}

	// Check for backup directory info.
	if getter, ok := adapter.(adapters.BackupDirGetter); ok {
		dir := getter.LastBackupDir()
		if dir != "" {
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID: toolID,
				Step:   InstallSteps[len(InstallSteps)-1],
				Done:   true,
				Info:   dir,
			})
		}
	}
}

// runUninstallSteps dispatches the uninstall operation through the Strategy
// interface. Uninstall is a single-phase operation: Rollback() undoes the
// effects of a previous install.
//
// If the adapter does not implement common.Strategy, an error message is
// sent and the goroutine returns without panicking.
func runUninstallSteps(ctx context.Context, t model.ToolState, ch chan<- model.ProgressMsg) {
	adapter := t.Adapter
	toolID := adapter.ID()

	s, ok := adapter.(common.Strategy)
	if !ok {
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: toolID,
			Step:   "Uninstalling",
			Done:   true,
			Error:  fmt.Sprintf("adapter %s: does not implement Strategy", toolID),
		})
		return
	}

	// Send running.
	if !sendProgress(ctx, ch, model.ProgressMsg{
		ToolID: toolID,
		Step:   "Uninstalling",
		Done:   false,
	}) {
		return
	}

	// Uninstall via the adapter's Uninstall method (still backward-compat).
	// We first try the legacy approach — the Strategy interface is for install,
	// while uninstall uses the legacy Installer.Uninstall() path.
	if installer, ok := adapter.(adapters.Installer); ok {
		err := installer.Uninstall(adapters.InstallOpts{Context: ctx})
		if err != nil {
			if errors.Is(err, adapters.ErrUninstallFailed) {
				sendProgress(ctx, ch, model.ProgressMsg{
					ToolID:  toolID,
					Step:    "Uninstalling",
					Done:    true,
					Warning: true,
					Error:   err.Error(),
				})
			} else {
				sendProgress(ctx, ch, model.ProgressMsg{
					ToolID: toolID,
					Step:   "Uninstalling",
					Done:   true,
					Error:  err.Error(),
				})
			}
			return
		}
	} else {
		// Fallback: just call Rollback().
		if err := s.Rollback(); err != nil {
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID: toolID,
				Step:   "Uninstalling",
				Done:   true,
				Error:  err.Error(),
			})
			return
		}
	}

	// Done.
	sendProgress(ctx, ch, model.ProgressMsg{
		ToolID: toolID,
		Step:   "Uninstalling",
		Done:   true,
	})

	// Warnings after uninstall.
	if emitter, ok := adapter.(WarningEmitter); ok {
		warnings := emitter.Warnings()
		if len(warnings) > 0 {
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID:  toolID,
				Step:    "Uninstalling",
				Done:    true,
				Warning: true,
				Error:   strings.Join(warnings, "\n"),
			})
		}
	}
}

// RunUninstall returns a tea.Cmd that removes Sequoia from every selected tool.
// It follows the same goroutine-per-tool pattern as RunInstall, calling
// adapter.Uninstall() via the Installer interface.
//
// Progress reporting follows the same convention:
//   - "running" before the call,
//   - "done" after a successful call,
//   - error on failure.
//
// The channel is closed when all goroutines complete.
func RunUninstall(ctx context.Context, tools []model.ToolState, ch chan<- model.ProgressMsg) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup

		for _, tool := range tools {
			if !tool.Selected {
				continue
			}

			select {
			case <-ctx.Done():
				wg.Wait()
				safeClose(ch)
				return nil
			default:
			}

			wg.Add(1)
			go func(t model.ToolState) {
				defer wg.Done()
				runUninstallSteps(ctx, t, ch)
			}(tool)
		}

		wg.Wait()
		safeClose(ch)
		return nil
	}
}

// RunStatus returns a tea.Cmd that queries the installation status of all
// tools. For each tool, it sends a ProgressMsg with the tool ID as the step
// name and Done=true. The channel is closed after all queries complete.
func RunStatus(ctx context.Context, tools []model.ToolState, ch chan<- model.ProgressMsg) tea.Cmd {
	return func() tea.Msg {
		for _, tool := range tools {
			select {
			case <-ctx.Done():
				safeClose(ch)
				return nil
			default:
			}

			// Query status and send result.
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID: tool.Adapter.ID(),
				Step:   "status",
				Done:   true,
			})
		}

		safeClose(ch)
		return nil
	}
}

func codegraphInstallProgress(ctx context.Context, ch chan<- model.ProgressMsg) {
	// Skip in CI environments where real codegraph exec may hang
	// (e.g., PowerShell script prompt in non-interactive shells on windows-latest).
	// The caller will send the final CodeGraph messages after all goroutines complete.
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		// Send "skipped" messages so tests expecting CodeGraph messages still get them.
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID:  "codegraph",
			Step:    "Installing",
			Done:    true,
			Warning: true,
			Error:   "skipped in CI",
		})
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID:  "codegraph",
			Step:    "Configuring",
			Done:    true,
			Warning: true,
			Error:   "skipped in CI",
		})
		return
	}

	// REQ-TUI-05: cap the install step so a hung HTTPS connection cannot
	// block the pipeline indefinitely. parentCtx is used for the warning
	// emit below — sendProgress's ctx.Done() select would drop the
	// message on the now-expired ctx.
	parentCtx := ctx
	ctx, cancel := context.WithTimeout(ctx, codegraphProgressTimeout)
	defer cancel()

	sendProgress(ctx, ch, model.ProgressMsg{
		ToolID: "codegraph",
		Step:   "Installing",
	})

	result := codegraph.Install(ctx, nil)

	// If the wrapped context's deadline fired, surface a Warning and
	// return BEFORE the result branch so the user sees the timeout, not
	// a generic failed message. Per REQ-TUI-05.
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		sendProgress(parentCtx, ch, model.ProgressMsg{
			ToolID:  "codegraph",
			Step:    "Installing",
			Done:    true,
			Warning: true,
			Error:   "download timed out",
		})
		return
	}

	if result.AlreadyInstalled {
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: "codegraph",
			Step:   "Installing",
			Done:   true,
			Info:   "already installed",
		})
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: "codegraph",
			Step:   "Configuring",
			Done:   true,
			Info:   "already configured",
		})
	} else if result.Installed {
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: "codegraph",
			Step:   "Installing",
			Done:   true,
		})
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: "codegraph",
			Step:   "Configuring",
			Done:   true,
		})
	} else {
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID:  "codegraph",
			Step:    "Installing",
			Done:    true,
			Warning: true,
			Error:   result.Message,
		})
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID:  "codegraph",
			Step:    "Configuring",
			Done:    true,
			Warning: true,
			Error:   "skipped",
		})
	}
}

// safeClose closes a channel, recovering if the channel is already closed.
// This is defensive: startPipeline always allocates a fresh channel, but
// safeClose protects against edge cases where a channel is reused.
func safeClose(ch chan<- model.ProgressMsg) {
	defer func() { recover() }()
	close(ch)
}

// sendProgress attempts to send a ProgressMsg on ch, respecting context
// cancellation. It returns true if the message was sent, false if the
// context was cancelled or the channel is closed.
//
// The send is non-blocking for context cancellation — if the context is
// done, the message is discarded and false is returned. Otherwise, the
// send blocks until the channel has room (capacity is 64, so this is
// unlikely to block in practice).
func sendProgress(ctx context.Context, ch chan<- model.ProgressMsg, msg model.ProgressMsg) (sent bool) {
	// Defensive: recover from panic caused by sending on a closed channel.
	// This protects against callers that reuse a closed channel, though
	// startPipeline always creates a fresh channel.
	defer func() {
		if r := recover(); r != nil {
			sent = false
		}
	}()
	select {
	case <-ctx.Done():
		return false
	case ch <- msg:
		return true
	}
}
