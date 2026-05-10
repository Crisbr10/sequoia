// Package plugin defines the Plugin interface and provides a file-based
// plugin loader for extending Sequoia with custom audit phases and agents.
//
// Plugins are discovered from a directory by scanning for .sequoia-plugin.yaml
// manifest files. Each manifest defines the plugin's metadata and the audit
// agents it provides.
package plugin

// Plugin is the contract every Sequoia plugin must satisfy.
// Plugins provide additional audit agents that can be loaded at runtime
// without modifying Sequoia's core.
type Plugin interface {
	// ID returns the unique machine-readable identifier (e.g. "custom-security").
	ID() string
	// Name returns the human-readable display name.
	Name() string
	// Version returns the plugin version string (e.g. "0.1.0").
	Version() string
	// Init is called once when the plugin is loaded.
	// It should perform any one-time setup (e.g. config validation,
	// external service connection check).
	// Returning an error does not prevent the plugin from being returned
	// by Load — callers should check Init results.
	Init() error
	// Agents returns the audit agents provided by this plugin.
	Agents() []Agent
}

// Agent represents an audit agent provided by a plugin.
// Each agent is a specialized auditor that the orchestrator can delegate
// analysis to during a Sequoia audit.
type Agent struct {
	// ID is the unique machine-readable agent identifier (e.g. "custom-security-scanner").
	ID string `yaml:"id"`
	// Name is the human-readable display name.
	Name string `yaml:"name"`
	// Description explains what the agent does and when it should be activated.
	Description string `yaml:"description"`
	// SystemPrompt is the full system prompt that defines the agent's
	// behavior, expertise, and output format.
	SystemPrompt string `yaml:"system_prompt"`
}
