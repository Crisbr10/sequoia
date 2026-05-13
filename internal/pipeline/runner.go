// Package pipeline provides goroutine-based install and uninstall runners
// that bridge the TUI to adapter calls via a buffered progress channel.
package pipeline

import (
	"context"
	"errors"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/internal/model"
)

// pipelineInstaller is a local interface that exposes the Install method
// from the concrete adapter behind model.ToolInfo.
type pipelineInstaller interface {
	Install(adapters.InstallOpts) error
}

// pipelineUninstaller is a local interface that exposes the Uninstall method
// from the concrete adapter behind model.ToolInfo.
type pipelineUninstaller interface {
	Uninstall(adapters.InstallOpts) error
}

// WarningEmitter is a local interface for adapters that collect non-fatal
// warnings during Install/Uninstall (e.g., symlink resolution failures).
// Adapters that implement this interface will have their warnings surfaced
// as ProgressMsg{Warning: true} after a successful operation.
type WarningEmitter interface {
	Warnings() []string
}

// InstallSteps defines the install steps in execution order.
// Since adapter.Install() is a single monolithic call, there is exactly
// one honest step: "Installing". This variable is exported so that
// callers (e.g., update.go's buildProgressTools) reference a single
// source of truth for step names.
var InstallSteps = []string{"Installing"}

// RunInstall returns a tea.Cmd that installs Sequoia into every selected tool.
// Each tool runs in its own goroutine calling adapter.Install().
//
// Progress is reported through a buffered channel:
//   - A "running" ProgressMsg (Done=false) is sent before the call.
//   - A "done" ProgressMsg (Done=true) is sent after a successful call.
//   - An error ProgressMsg (Error != "") is sent when the call fails.
//
// The channel is closed when all goroutines complete. Context cancellation
// stops goroutines gracefully while preserving partial progress.
//
// lang is passed to adapter.Install(opts) and adapter.Uninstall(opts)
// via adapters.InstallOpts{Language: lang} for template localization.
func RunInstall(ctx context.Context, tools []model.ToolState, ch chan<- model.ProgressMsg, lang string) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup

		for _, tool := range tools {
			if !tool.Selected {
				continue
			}

			select {
			case <-ctx.Done():
				// Context cancelled before goroutine starts — stop launching new goroutines.
				wg.Wait()
				close(ch)
				return nil
			default:
			}

			wg.Add(1)
			go func(t model.ToolState) {
				defer wg.Done()
				runInstallSteps(ctx, t, ch, lang)
			}(tool)
		}

		// Wait for all goroutines to complete, then signal completion
		// by closing the channel.
		wg.Wait()
		close(ch)
		return nil
	}
}

// runSteps executes the install/uninstall operation for a single tool,
// sending progress messages through ch.
//
// Steps:
//  1. Send a "running" ProgressMsg (Done=false, Step="Installing").
//  2. Call fn(t, opts) — the adapter performs all work internally.
//  3. On success: send a "done" ProgressMsg (Done=true).
//  4. On failure: send an error ProgressMsg (Done=true, Error set).
//
// fn is either adapter.Install or adapter.Uninstall (both have the same signature).
func runSteps(ctx context.Context, t model.ToolState, ch chan<- model.ProgressMsg, lang string, fn func(adapters.InstallOpts) error) {
	adapter := t.Adapter
	toolID := adapter.ID()
	step := InstallSteps[0] // "Installing"

	// Signal that work is starting.
	if !sendProgress(ctx, ch, model.ProgressMsg{
		ToolID: toolID,
		Step:   step,
		Done:   false,
	}) {
		return // context cancelled
	}

	// Perform the actual operation (Install or Uninstall).
	err := fn(adapters.InstallOpts{Language: lang, Context: ctx})

	// Report the result.
	if err != nil {
		// Partial failure (uninstall with warnings): mark as done with warning.
		if errors.Is(err, adapters.ErrUninstallFailed) {
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID:  toolID,
				Step:    step,
				Done:    true,
				Warning: true,
				Error:   err.Error(),
			})
			return
		}

		// Hard failure — report the error.
		sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: toolID,
			Step:   step,
			Done:   true,
			Error:  err.Error(),
		})
		return
	}

	// Success.
	sendProgress(ctx, ch, model.ProgressMsg{
		ToolID: toolID,
		Step:   step,
		Done:   true,
	})

	// After a successful install/uninstall, check if the adapter collected
	// any non-fatal warnings and surface them as a separate progress message.
	if emitter, ok := adapter.(WarningEmitter); ok {
		warnings := emitter.Warnings()
		if len(warnings) > 0 {
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID:  toolID,
				Step:    step,
				Done:    true,
				Warning: true,
				Error:   strings.Join(warnings, "\n"),
			})
		}
	}
}

// runInstallSteps extracts the Install method from the concrete adapter
// behind model.ToolInfo and calls runSteps.
func runInstallSteps(ctx context.Context, t model.ToolState, ch chan<- model.ProgressMsg, lang string) {
	a := t.Adapter.(pipelineInstaller)
	runSteps(ctx, t, ch, lang, a.Install)
}

// runUninstallSteps extracts the Uninstall method from the concrete adapter
// behind model.ToolInfo and calls runSteps.
func runUninstallSteps(ctx context.Context, t model.ToolState, ch chan<- model.ProgressMsg, lang string) {
	a := t.Adapter.(pipelineUninstaller)
	runSteps(ctx, t, ch, lang, a.Uninstall)
}

// RunUninstall returns a tea.Cmd that removes Sequoia from every selected tool.
// It follows the same goroutine-per-tool pattern as RunInstall, calling
// adapter.Uninstall() instead of Install().
//
// Progress reporting follows the same convention:
//   - "running" before the call,
//   - "done" after a successful call,
//   - error on failure.
//
// The channel is closed when all goroutines complete.
func RunUninstall(ctx context.Context, tools []model.ToolState, ch chan<- model.ProgressMsg, lang string) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup

		for _, tool := range tools {
			if !tool.Selected {
				continue
			}

			select {
			case <-ctx.Done():
				wg.Wait()
				close(ch)
				return nil
			default:
			}

			wg.Add(1)
			go func(t model.ToolState) {
				defer wg.Done()
				runUninstallSteps(ctx, t, ch, lang)
			}(tool)
		}

		wg.Wait()
		close(ch)
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
				close(ch)
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

		close(ch)
		return nil
	}
}

// sendProgress attempts to send a ProgressMsg on ch, respecting context
// cancellation. It returns true if the message was sent, false if the
// context was cancelled or the channel is closed.
//
// The send is non-blocking for context cancellation — if the context is
// done, the message is discarded and false is returned. Otherwise, the
// send blocks until the channel has room (capacity is 64, so this is
// unlikely to block in practice).
func sendProgress(ctx context.Context, ch chan<- model.ProgressMsg, msg model.ProgressMsg) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- msg:
		return true
	}
}
