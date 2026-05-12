// Package pipeline provides goroutine-based install and uninstall runners
// that bridge the TUI to adapter calls via a buffered progress channel.
package pipeline

import (
	"context"
	"errors"
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

// defaultStepNames defines the install steps in execution order.
// These names MUST match the step names used by screens.ProgressTool
// so that ApplyProgressMsg can correlate progress messages.
var defaultStepNames = []string{"Skills", "Commands", "System Prompt"}

// RunInstall returns a tea.Cmd that installs Sequoia into every selected tool.
// Each tool runs in its own goroutine calling adapter.Install().
//
// Progress is reported through a buffered channel:
//   - A "running" ProgressMsg (Done=false) is sent before each step.
//   - A "done" ProgressMsg (Done=true) is sent after each step succeeds.
//   - An error ProgressMsg (Error != "") is sent when a step fails.
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

// runSteps executes the standard install/uninstall steps (Skills, Commands,
// System Prompt) for a single tool, sending progress messages through ch.
//
// Steps:
//  1. Send "running" ProgressMsg for each step.
//  2. Call fn() — the adapter performs all steps internally.
//  3. On success: send "done" ProgressMsg for each step.
//  4. On failure: send error ProgressMsg for the step where the failure was
//     detected (the first step in the sequence), then stop.
//
// fn is either adapter.Install or adapter.Uninstall (both have the same signature).
func runSteps(ctx context.Context, t model.ToolState, ch chan<- model.ProgressMsg, lang string, fn func(adapters.InstallOpts) error) {
	adapter := t.Adapter
	toolID := adapter.ID()

	// Signal that each step is starting.
	for _, step := range defaultStepNames {
		if !sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: toolID,
			Step:   step,
			Done:   false,
		}) {
			return // context cancelled
		}
	}

	// Perform the actual operation (Install or Uninstall).
	err := fn(adapters.InstallOpts{Language: lang, Context: ctx})

	// Report the result.
	if err != nil {
		// Partial failure (uninstall with warnings): some files removed,
		// some not. Treat as "done with warnings" rather than hard error.
		if errors.Is(err, adapters.ErrUninstallFailed) {
			// Mark all steps as done but include a warning note.
			for _, step := range defaultStepNames {
				sendProgress(ctx, ch, model.ProgressMsg{
					ToolID:  toolID,
					Step:    step,
					Done:    true,
					Warning: true,
					Error:   err.Error(),
				})
			}
			return
		}

		// Hard failure — report the first step as errored with the message.
		if len(defaultStepNames) > 0 {
			sendProgress(ctx, ch, model.ProgressMsg{
				ToolID: toolID,
				Step:   defaultStepNames[0],
				Done:   true,
				Error:  err.Error(),
			})
		}
		return
	}

	// Success — mark all steps as done.
	for _, step := range defaultStepNames {
		if !sendProgress(ctx, ch, model.ProgressMsg{
			ToolID: toolID,
			Step:   step,
			Done:   true,
		}) {
			return // context cancelled
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
//   - "running" before each step,
//   - "done" after each successful step,
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
