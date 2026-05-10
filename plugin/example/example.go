// Package example provides a minimal "Hello World" plugin that demonstrates
// how to implement the plugin.Plugin interface.
//
// Use this as a reference when building your own Sequoia plugins.
package example

import "sequoia-ai/plugin"

// helloPlugin is a simple example plugin that registers a single audit agent.
type helloPlugin struct{}

// NewHelloPlugin returns a new instance of the Hello World example plugin.
func NewHelloPlugin() plugin.Plugin {
	return &helloPlugin{}
}

// ID returns the unique machine-readable identifier for this plugin.
func (p *helloPlugin) ID() string { return "hello-world" }

// Name returns the human-readable display name.
func (p *helloPlugin) Name() string { return "Hello World Plugin" }

// Version returns the plugin version string.
func (p *helloPlugin) Version() string { return "0.1.0" }

// Init performs one-time plugin setup. For this example, it always succeeds.
func (p *helloPlugin) Init() error { return nil }

// Agents returns the audit agents provided by this plugin.
func (p *helloPlugin) Agents() []plugin.Agent {
	return []plugin.Agent{
		{
			ID:           "hello-world-greeter",
			Name:         "Hello World Greeter",
			Description:  "A friendly agent that greets every codebase it audits. Demonstrates the plugin agent contract.",
			SystemPrompt: "You are the Hello World Greeter. Before starting your audit, greet the developer with a warm 'Hello, World!' and then proceed with your analysis.",
		},
	}
}
