# Sequoia Plugin System

Sequoia supports loading custom audit agents at runtime through a file-based
plugin system. Plugins are defined via YAML manifest files and discovered
automatically from a configured directory.

## Quick Start

Create a `.sequoia-plugin.yaml` file:

```yaml
id: "my-plugin"
name: "My Custom Plugin"
version: "0.1.0"
agents:
  - id: "my-security-agent"
    name: "Security Agent"
    description: "Scans for security vulnerabilities and misconfigurations"
    system_prompt: "You are a security auditor. Review the code for vulnerabilities."
```

Place it in your plugin directory (default: `~/.sequoia/plugins/`), and Sequoia
will load it automatically at startup.

## Manifest Format

Each plugin is defined by a single `.sequoia-plugin.yaml` file. Only files
in the root of the plugin directory are scanned (no recursive discovery).

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique machine-readable identifier (kebab-case, e.g. `my-plugin`) |
| `name` | Yes | Human-readable display name |
| `version` | Yes | Plugin version (SemVer, e.g. `0.1.0`) |
| `agents` | No | List of audit agents provided by this plugin |

### Agent Definition

Each agent in the `agents` list has the following fields:

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique agent identifier (e.g. `my-security-scanner`) |
| `name` | Yes | Human-readable agent name |
| `description` | Yes | What the agent does and when it should be activated |
| `system_prompt` | Yes | Full system prompt defining behavior and output format |

## Programmatic Plugins (Go API)

For plugins written in Go, implement the `plugin.Plugin` interface:

```go
type Plugin interface {
    ID() string
    Name() string
    Version() string
    Init() error
    Agents() []Agent
}
```

See `plugin/example/example.go` for a complete working example.

## Loading Plugins

```go
import "sequoia-ai/plugin"

plugins, err := plugin.Load("/path/to/plugins")
if err != nil {
    // Handle directory read errors.
}
for _, p := range plugins {
    fmt.Printf("Loaded: %s (%s)\n", p.Name(), p.Version())
    if err := p.Init(); err != nil {
        // Init failure — plugin may still be usable but should be logged.
        fmt.Printf("Warning: %s init failed: %v\n", p.Name(), err)
    }
    for _, a := range p.Agents() {
        fmt.Printf("  Agent: %s — %s\n", a.Name, a.Description)
    }
}
```

## Error Handling

Malformed manifest files (invalid YAML, missing required fields) are silently
skipped during loading. A plugin without an `id` field is not loaded. Errors
reading individual manifests are logged but do not prevent other plugins from
being loaded.

When a plugin's `Init()` method returns an error, the plugin is still returned
by `Load`. Callers should always check `Init()` results and decide whether to
use or skip the plugin based on the error.

## Conventions

- Plugin IDs should be kebab-case and unique across all loaded plugins
- Agent IDs should include the plugin ID as a prefix (e.g. `my-plugin-scanner`)
- System prompts should be self-contained and include output format instructions
- Keep manifests small — the system prompt is the heavy part
- Version your plugins and document breaking changes

## Example: Hello World Plugin

The `plugin/example/` directory contains a minimal plugin that registers a
single "Hello World Greeter" agent. Use it as a reference when building
your own plugins.

```go
package example

import "sequoia-ai/plugin"

type helloPlugin struct{}

func NewHelloPlugin() plugin.Plugin {
    return &helloPlugin{}
}

func (p *helloPlugin) ID() string      { return "hello-world" }
func (p *helloPlugin) Name() string    { return "Hello World Plugin" }
func (p *helloPlugin) Version() string { return "0.1.0" }
func (p *helloPlugin) Init() error     { return nil }

func (p *helloPlugin) Agents() []plugin.Agent {
    return []plugin.Agent{
        {
            ID:           "hello-world-greeter",
            Name:         "Hello World Greeter",
            Description:  "A friendly agent that greets every codebase.",
            SystemPrompt: "You are the Hello World Greeter. Greet the developer and proceed with your analysis.",
        },
    }
}
```
