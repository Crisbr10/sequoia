// Package main is the entrypoint for the sequoia CLI tool.
// It uses Cobra to provide install, status, uninstall, and version commands.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Crisbr10/sequoia/adapters"
	"github.com/Crisbr10/sequoia/adapters/common"
	"github.com/Crisbr10/sequoia/internal/app"

	// Register all adapters via their init() functions (database/sql pattern).
	_ "github.com/Crisbr10/sequoia/adapters/claude"
	_ "github.com/Crisbr10/sequoia/adapters/codex"
	_ "github.com/Crisbr10/sequoia/adapters/cursor"
	_ "github.com/Crisbr10/sequoia/adapters/gemini"
	_ "github.com/Crisbr10/sequoia/adapters/opencode"
)

// Version is the Sequoia CLI version. Set at build time via:
//
//	-ldflags "-X github.com/Crisbr10/sequoia/cmd/sequoia.Version=0.1.2" (GoReleaser)
//	go install auto-detects it via debug.ReadBuildInfo
//
// Falls back to "0.1.0-dev" only for local go build.
var Version = "0.1.0-dev"

func init() {
	if Version != "0.1.0-dev" {
		return // ldflags already set the version (GoReleaser)
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	// go install @v0.1.2 embeds the module version
	if v := info.Main.Version; v != "" && v != "(devel)" {
		Version = v
		return
	}
	// Check VCS info (go build from tagged commit)
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			Version = "unknown-" + s.Value[:8]
			return
		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root := newRootCmd()
	root.SetContext(ctx)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// newRootCmd creates the root command with all subcommands attached.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sequoia",
		Short: "Sequoia — Code Audit Framework installer and manager",
		Long: `Sequoia is a comprehensive technical audit framework that integrates
into AI-assisted coding tools (Claude Code, OpenCode, and more).

Running sequoia with no arguments launches the interactive TUI.

  sequoia           Launch interactive TUI (install, status, uninstall)
  sequoia install   Install Sequoia into supported AI tools
  sequoia status    Show installation status for all detected tools
  sequoia uninstall Remove Sequoia from one or all tools
  sequoia version   Print the CLI version`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if isTerminalFn() {
				return runTUI("")
			}
			return cmd.Help()
		},
	}

	root.AddCommand(
		newInstallCmd(),
		newStatusCmd(),
		newUninstallCmd(),
		newVersionCmd(),
	)

	return root
}

// newInstallCmd creates the 'install' subcommand.
func newInstallCmd() *cobra.Command {
	var (
		toolID string
		noTUI  bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Sequoia into a supported AI tool",
		Long: `Install Sequoia skills, commands, and system prompt into one or more
supported AI tools (Claude Code, OpenCode).

By default, all detected tools are shown interactively via the TUI.
Use --no-tui to skip the interactive interface and install directly.
Use --tool to target a specific adapter by ID.

Examples:
  sequoia install                    # Interactive TUI
  sequoia install --no-tui           # Install into all detected tools
  sequoia install --tool=claude-code # Install only into Claude Code`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// TUI mode: if stdin is a terminal and --no-tui is not set,
			// launch the Bubbletea app.
			if !noTUI && isTerminal() {
				return runTUI(toolID)
			}
			// Headless mode.
			return runInstall(cmd.Context(), toolID, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&toolID, "tool", "", "Target adapter ID (e.g. claude-code, opencode)")
	cmd.Flags().BoolVar(&noTUI, "no-tui", false, "Skip interactive TUI and install directly")

	return cmd
}

// newStatusCmd creates the 'status' subcommand.
func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show installation status for all adapters",
		Long: `Detect supported AI tools on this machine and report whether
Sequoia is installed for each one.

The output includes the adapter ID, display name, detected status,
and installation path.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd.OutOrStdout())
		},
	}

	return cmd
}

// newUninstallCmd creates the 'uninstall' subcommand.
func newUninstallCmd() *cobra.Command {
	var (
		toolID  string
		all     bool
		yesFlag bool
	)

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove Sequoia from a supported AI tool",
		Long: `Remove Sequoia files and system prompt sections from one or all
supported AI tools. Backups created during installation are restored
where applicable.

Use --yes/-y to skip the confirmation prompt. When stdin is not a
terminal (e.g. piped), --yes is required.

Examples:
  sequoia uninstall --tool=claude-code  # Remove from Claude Code
  sequoia uninstall --all               # Remove from all tools
  sequoia uninstall --all --yes         # Remove from all tools without prompt`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUninstall(cmd.Context(), toolID, all, yesFlag, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&toolID, "tool", "", "Target adapter ID (e.g. claude-code, opencode)")
	cmd.Flags().BoolVar(&all, "all", false, "Remove Sequoia from all installed tools")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// newVersionCmd creates the 'version' subcommand.
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Long:  "Print the Sequoia CLI version number.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), Version)
			return nil
		},
	}

	return cmd
}

// isTerminalFn wraps isTerminal so it can be overridden in tests.
// Defaults to os.Stdin.Stat() check. Tests override it to simulate
// terminal vs. piped stdin without touching os.Stdin.
var isTerminalFn = isTerminal

// -- Command handlers ---------------------------------------------------------

// runInstall installs Sequoia into a specific adapter or all detected adapters.
// ctx is the signal-aware context from main() — cancellation stops the install early.
func runInstall(ctx context.Context, toolID string, out io.Writer) error {
	targets := targetAdapters(toolID)
	if len(targets) == 0 {
		if toolID != "" {
			return fmt.Errorf("unknown adapter %q — use 'sequoia status' to list available adapters", toolID)
		}
		_, _ = fmt.Fprintln(out, "No supported AI tools detected on this machine.")
		_, _ = fmt.Fprintln(out, "Currently supported: Claude Code, OpenCode, Cursor IDE, Gemini CLI, OpenAI Codex")
		return nil
	}

	for _, a := range targets {
		_, _ = fmt.Fprintf(out, "Installing Sequoia for %s ...\n", a.Name())
		if a.IsInstalled() {
			_, _ = fmt.Fprintf(out, "  Sequoia is already installed. Reinstalling ...\n")
		}
		if err := a.Install(adapters.InstallOpts{Context: ctx}); err != nil {
			return fmt.Errorf("install %s: %w", a.ID(), err)
		}
		_, _ = fmt.Fprintf(out, "  Done! Use /sequoia-init inside %s to get started.\n", a.Name())
	}

	return nil
}

// runStatus prints the installation status for all registered adapters
// in a 6-column fixed-width table: ID, NAME, DETECTED, INSTALLED, VERSION, PATH.
// If the user's home directory is a symlink, a note is included showing the
// real resolved path.
func runStatus(out io.Writer) error {
	all := adapters.DefaultRegistry.All()
	if len(all) == 0 {
		_, _ = fmt.Fprintln(out, "No adapters registered.")
		return nil
	}

	// Header with column widths: ID(14) NAME(14) DETECTED(9) INSTALLED(10) VERSION(10) PATH(55).
	_, _ = fmt.Fprintf(out, "%-14s %-14s %-9s %-10s %-10s %-55s\n",
		"ID", "NAME", "DETECTED", "INSTALLED", "VERSION", "PATH")
	_, _ = fmt.Fprintln(out, strings.Repeat("-", 117))

	for _, a := range all {
		s := a.Status()
		detected := "no"
		if a.Detect() {
			detected = "yes"
		}
		installed := "no"
		if s.Installed {
			installed = "yes"
		}
		_, _ = fmt.Fprintf(out, "%-14s %-14s %-9s %-10s %-10s %-55s\n",
			a.ID(), a.Name(), detected, installed, s.Version, s.Path)
	}

	// Symlink detection: if the home directory is a symlink, note the real path.
	if home, err := os.UserHomeDir(); err == nil {
		if common.IsSymlink(home) {
			resolved, err := common.ResolveHome(home)
			if err == nil {
				_, _ = fmt.Fprintf(out, "\nHome directory is a symlink: %s → %s\n", home, resolved)
			}
		}
	}

	return nil
}

// ScanTools returns structured installation status for all registered adapters.
// Each result includes name, path, installed state, and Sequoia version.
// The returned slice includes adapters even if not detected — callers filter.
func ScanTools() []adapters.AdapterStatus {
	all := adapters.DefaultRegistry.All()
	results := make([]adapters.AdapterStatus, 0, len(all))
	for _, a := range all {
		results = append(results, a.Status())
	}
	return results
}

// runUninstall removes Sequoia from a specific adapter or all installed adapters.
//
// When yes is false and stdin is a terminal, it displays a confirmation prompt
// listing the affected tools and waits for "y"/"Y" input. When stdin is piped
// and yes is false, it returns an error directing users to --yes. When yes is
// true, the confirmation prompt is skipped entirely.
// ctx is the signal-aware context from main() — cancellation stops uninstall early.
func runUninstall(ctx context.Context, toolID string, all bool, yes bool, in io.Reader, out io.Writer) error {
	targets := targetAdapters(toolID)
	if all && toolID == "" {
		targets = adapters.DefaultRegistry.All()
	}
	if len(targets) == 0 {
		if toolID != "" {
			return fmt.Errorf("unknown adapter %q — use 'sequoia status' to list available adapters", toolID)
		}
		_, _ = fmt.Fprintln(out, "No adapters to uninstall from.")
		return nil
	}

	// Confirmation gate: skip if --yes, error if piped, prompt otherwise.
	if !yes {
		if !isTerminalFn() {
			return fmt.Errorf("stdin is not a terminal; use --yes to skip confirmation")
		}

		// Build the confirmation prompt.
		if len(targets) == 1 {
			_, _ = fmt.Fprintf(out, "Remove Sequoia from %s? [y/N]: ", targets[0].Name())
		} else {
			_, _ = fmt.Fprintln(out, "This will remove Sequoia from:")
			for _, a := range targets {
				_, _ = fmt.Fprintf(out, "  %s\n", a.Name())
			}
			_, _ = fmt.Fprint(out, "Continue? [y/N]: ")
		}

		var response string
		_, _ = fmt.Fscanln(in, &response)
		if response != "y" && response != "Y" {
			_, _ = fmt.Fprintln(out, "Uninstall aborted.")
			return nil
		}
	}

	for _, a := range targets {
		if !a.IsInstalled() {
			_, _ = fmt.Fprintf(out, "Sequoia is not installed for %s — skipping.\n", a.Name())
			continue
		}
		_, _ = fmt.Fprintf(out, "Removing Sequoia from %s ...\n", a.Name())
		if err := a.Uninstall(adapters.InstallOpts{Context: ctx}); err != nil {
			return fmt.Errorf("uninstall %s: %w", a.ID(), err)
		}
		_, _ = fmt.Fprintf(out, "  Done.\n")
	}

	return nil
}

// runTUI launches the interactive TUI installer using Bubbletea.
// It creates the root model, configures the program with alt-screen and mouse
// support, and blocks until the user quits.
func runTUI(toolID string) error {
	p := tea.NewProgram(
		app.NewModel(toolID, Version),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

// targetAdapters returns adapters matching toolID, or all detected adapters if toolID is empty.
func targetAdapters(toolID string) []adapters.ToolAdapter {
	if toolID != "" {
		a, err := adapters.DefaultRegistry.Get(toolID)
		if err != nil {
			return nil
		}
		return []adapters.ToolAdapter{a}
	}

	var detected []adapters.ToolAdapter
	for _, a := range adapters.DefaultRegistry.All() {
		if a.Detect() {
			detected = append(detected, a)
		}
	}
	return detected
}
