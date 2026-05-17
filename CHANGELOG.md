# Changelog

All notable changes to Sequoia will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.6] — 2026-05-17

### Fixed

- **Data race in version tests**: `TestVersionCmd` and `TestVersionCmd_DevVersionResolves` no longer run in parallel — both access the shared global `Version` variable. Fixes `-race` failures on macOS and Linux CI.
- **golangci-lint CI**: Downgraded `golangci-lint-action` from v7 to v6. v7 dropped support for golangci-lint v1.x, causing `invalid version string 'v1.64'` errors.

### Removed

- **i18n package (`internal/i18n/`)**: 7 files, ~815 lines deleted. English is the only supported language.
- **`go-i18n/v2` dependency** from `go.mod`.
- **Language selector** from TUI Configuration screen (only persistence backend remains).
- **`Language` type and `LangEN`/`LangES` constants** from `model/types.go`.
- **`i18n.Init()`** from application initialization.
- **`lang` parameter** from all TUI view functions, `RenderConfirmPrompt`, pipeline helpers, and `InstallOpts`.
- **`RenderTemplateLang`** from adapters; all call sites use `RenderTemplate`.
- **`adapters/interface_test.go`** — language-specific adapter tests removed.
- **P7 i18n agent** documentation and specs (`docs/agents/sequoia-i18n.md`, `docs/sequoia/sequoia-phases/07-i18n.md`).
- **Legacy roadmap and task files** (`futuras-implementaciones/`).
- **`openspec/specs/agent-p7-i18n/` and `openspec/specs/i18n-engine/`** specs.

### Changed

- **All TUI screens** now use hardcoded English strings directly.
- **15 golden test files** regenerated for English-only output.
- **All adapter templates** (Claude, Codex, Gemini, OpenCode) simplified — no language injection.
- **Documentation** updated to reflect 7-agent architecture (P1–P6 + M1–M2), no i18n phase.
- **Sequoia Health Score** formula updated (7 phases instead of 8).
- **Install script** (`scripts/install.ps1`) updated for v1.0.6.

### Added

- **SDD change archive**: `remove-i18n-english-only` with full proposal, specs, design, tasks, and verify report.
- **New specs**: `skill-documentation`, `test-infrastructure` in `openspec/specs/`.

## [1.0.5] — 2026-05-16

### Fixed

- **TUI display fixes**: uninstall cursor navigation, progress channel recreation, closed-channel panic defense.
- **Adapter stability**: `.sequoia-version` file check for `IsInstalled` (instead of `sequoia-ai.md`).
- **CI fixes**: cross-platform test stability, golangci-lint pinning, goreleaser changelog git source.

### Added

- **Backup directory path** surfaced in install/uninstall progress UI.

## [0.1.0] — 2026-05-10

### Added

- **5 AI tool adapters**: Claude Code, OpenCode, Cursor IDE, Gemini CLI, OpenAI Codex
- **Interactive TUI installer** (Bubbletea) with 8 screens:
  - Welcome, Tool Selection, Configuration, Install Progress, Complete, Error, Status, Uninstall
- **Headless CLI installer** (Cobra) with 4 subcommands: install, status, uninstall, version
- **Common installer framework** with atomic Prepare → Apply → Verify → Rollback pipeline
- **Four prompt strategies**: MarkdownSections, FileReplace, ConfigMerge, TOMLMerge
- **Adapter registry** with self-registration pattern (database/sql `init()`)
- **Cross-platform support**: macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64)
- **One-line installers**: `curl | bash` (Unix) and `irm | iex` (Windows PowerShell)
- **SHA-256 checksum verification** in install scripts
- **GoReleaser build pipeline** with Homebrew formula and Scoop manifest
- **Strict TDD test suite**: 327+ tests, 90%+ code coverage across 11 packages
- **Cross-platform CI**: GitHub Actions matrix on ubuntu, macos, windows
- **Plugin system foundation**: `ToolAdapter` interface, registry, factory
- **Adapter template**: Copy-paste boilerplate in `adapters/_template/`
- **Comprehensive documentation**: Getting started, architecture, CLI reference, FAQ, contributing guide
- **9 specialized audit agents**: C0 (orchestrator), P1–P6 (domain agents), M1–M2 (correlator, reporter)
