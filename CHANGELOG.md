# Changelog

All notable changes to Sequoia will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
