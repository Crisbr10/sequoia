## Sequoia v0.1.0

First public release of the multi-agent code audit framework. CLI installer, interactive TUI, and adapters for five AI coding tools.

### What's Included

- **5 AI tool adapters**: Claude Code, OpenCode, Cursor IDE, Gemini CLI, OpenAI Codex
- **Interactive TUI installer**: 8-screen Bubbletea interface with auto-detection, multi-select, real-time progress, and error recovery
- **Headless CLI**: Cobra-powered install, status, uninstall, version commands
- **4 prompt strategies**: Markdown section injection, file replace with backup, config merge, TOML merge
- **Cross-platform**: macOS (Intel + Apple Silicon), Linux (amd64 + arm64), Windows (amd64)
- **Plugin system**: File-based plugin discovery via `.sequoia-plugin.yaml`
- **Strict TDD**: 327+ tests across 17 packages, 90%+ coverage

### Install

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.ps1 | iex
```

Or use a package manager:

```bash
# Homebrew
brew tap Crisbr10/sequoia
brew install sequoia

# Scoop
scoop bucket add crisbr10 https://github.com/Crisbr10/scoop-sequoia
scoop install sequoia
```

### Supported Tools

| Tool | Strategy |
|------|----------|
| Claude Code | Markdown section injection |
| OpenCode | File replace with backup |
| Cursor IDE | File replace with backup |
| Gemini CLI | Config merge |
| OpenAI Codex | TOML merge |

### Quick Start

```bash
sequoia install              # Interactive TUI — select your tools
sequoia status               # Check what's installed
sequoia uninstall --all      # Remove everything
```

### Known Limitations

- E2E audit tests (T-012, T-017) require full Claude Code / OpenCode sessions and haven't been automated
- TUI language selection is present but Spanish translations are not complete
- Plugin system has the interface and loader but example plugins need real audit logic

### Documentation

- [Getting Started](https://github.com/Crisbr10/sequoia/blob/main/docs/getting-started.md)
- [Architecture](https://github.com/Crisbr10/sequoia/blob/main/docs/architecture.md)
- [CLI Reference](https://github.com/Crisbr10/sequoia/blob/main/docs/cli-reference.md)
- [FAQ](https://github.com/Crisbr10/sequoia/blob/main/docs/faq.md)
- [Contributing](https://github.com/Crisbr10/sequoia/blob/main/CONTRIBUTING.md)
