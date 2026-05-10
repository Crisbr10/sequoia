package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// pluginManifest is the on-disk YAML structure for .sequoia-plugin.yaml files.
type pluginManifest struct {
	ID      string       `yaml:"id"`
	Name    string       `yaml:"name"`
	Version string       `yaml:"version"`
	Agents  []agentDef   `yaml:"agents"`
}

type agentDef struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	SystemPrompt string `yaml:"system_prompt"`
}

// pluginImpl is the concrete implementation of the Plugin interface
// backed by a parsed YAML manifest.
type pluginImpl struct {
	manifest pluginManifest
}

func (p *pluginImpl) ID() string      { return p.manifest.ID }
func (p *pluginImpl) Name() string    { return p.manifest.Name }
func (p *pluginImpl) Version() string { return p.manifest.Version }

// Init performs plugin initialization.
// For v0.1.0, init is a no-op that always succeeds.
func (p *pluginImpl) Init() error {
	return nil
}

func (p *pluginImpl) Agents() []Agent {
	result := make([]Agent, len(p.manifest.Agents))
	for i, def := range p.manifest.Agents {
		result[i] = Agent{
			ID:           def.ID,
			Name:         def.Name,
			Description:  def.Description,
			SystemPrompt: def.SystemPrompt,
		}
	}
	return result
}

// Load discovers plugins from the given directory by scanning for
// .sequoia-plugin.yaml manifest files. Only the root of dir is scanned
// (no recursive descent). Invalid or unparseable manifest files are
// skipped gracefully without causing the entire load to fail.
//
// If dir does not exist, Load returns an empty slice and nil error.
func Load(dir string) ([]Plugin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("plugin load: read dir %q: %w", dir, err)
	}

	var plugins []Plugin
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".sequoia-plugin.yaml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		manifest, err := parseManifest(path)
		if err != nil {
			// Skip invalid files gracefully.
			continue
		}

		if manifest.ID == "" {
			// Plugin without an ID is not loadable.
			continue
		}

		plugins = append(plugins, &pluginImpl{manifest: *manifest})
	}

	return plugins, nil
}

// parseManifest reads and parses a .sequoia-plugin.yaml file.
func parseManifest(path string) (*pluginManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", path, err)
	}

	var m pluginManifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %q: %w", path, err)
	}

	return &m, nil
}
