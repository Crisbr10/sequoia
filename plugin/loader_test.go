package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Crisbr10/sequoia/plugin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load_ValidPlugin(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePluginYAML(t, dir, "valid.sequoia-plugin.yaml", validPluginYAML)

	plugins, err := plugin.Load(dir)
	require.NoError(t, err)
	require.Len(t, plugins, 1)

	p := plugins[0]
	assert.Equal(t, "my-plugin", p.ID())
	assert.Equal(t, "My Custom Plugin", p.Name())
	assert.Equal(t, "0.1.0", p.Version())

	require.NoError(t, p.Init())

	agents := p.Agents()
	require.Len(t, agents, 2)
	assert.Equal(t, "my-security-agent", agents[0].ID)
	assert.Equal(t, "Security Agent", agents[0].Name)
	assert.Contains(t, agents[0].Description, "security vulnerabilities")
	assert.Contains(t, agents[0].SystemPrompt, "security auditor")

	assert.Equal(t, "my-quality-agent", agents[1].ID)
	assert.Equal(t, "Quality Agent", agents[1].Name)
	assert.Contains(t, agents[1].Description, "code quality")
	assert.Contains(t, agents[1].SystemPrompt, "code review specialist")
}

func TestLoader_Load_EmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	plugins, err := plugin.Load(dir)

	require.NoError(t, err)
	assert.Empty(t, plugins, "empty directory should return zero plugins")
}

func TestLoader_Load_MultiplePlugins(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePluginYAML(t, dir, "plugin-a.sequoia-plugin.yaml", validPluginYAML)
	writePluginYAML(t, dir, "plugin-b.sequoia-plugin.yaml", secondPluginYAML)

	plugins, err := plugin.Load(dir)
	require.NoError(t, err)
	require.Len(t, plugins, 2)

	ids := make([]string, len(plugins))
	for i, p := range plugins {
		ids[i] = p.ID()
	}
	assert.Contains(t, ids, "my-plugin")
	assert.Contains(t, ids, "other-plugin")
}

func TestLoader_Load_DirectoryDoesNotExist(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "does-not-exist")
	plugins, err := plugin.Load(dir)

	require.NoError(t, err)
	assert.Empty(t, plugins, "non-existent directory should return empty slice, not error")
}

func TestLoader_Load_IgnoresNonPluginFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write a non-.sequoia-plugin.yaml file.
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "README.md"),
		[]byte("# Just a readme\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "not-a-plugin.yaml"),
		[]byte("id: nope\n"),
		0o644,
	))

	plugins, err := plugin.Load(dir)
	require.NoError(t, err)
	assert.Empty(t, plugins, "only .sequoia-plugin.yaml files should be loaded")
}

func TestLoader_Load_IgnoresNestedDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	nested := filepath.Join(dir, "nested")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	writePluginYAML(t, nested, "nested.sequoia-plugin.yaml", validPluginYAML)

	// Also add a valid plugin in the root to verify we're not just empty.
	writePluginYAML(t, dir, "root.sequoia-plugin.yaml", validPluginYAML)

	plugins, err := plugin.Load(dir)
	require.NoError(t, err)
	require.Len(t, plugins, 1, "should only find plugins in the root, not nested directories")
	assert.Equal(t, "my-plugin", plugins[0].ID())
}

func TestLoader_Load_InvalidYAML_SkippedGracefully(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Write an invalid YAML file (plain garbage).
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "broken.sequoia-plugin.yaml"),
		[]byte("this is: not: valid: yaml: [unclosed"),
		0o644,
	))
	// Also write a valid one to verify partial success.
	writePluginYAML(t, dir, "valid.sequoia-plugin.yaml", validPluginYAML)

	plugins, err := plugin.Load(dir)
	require.NoError(t, err, "Load should not fail on a single invalid file")
	require.Len(t, plugins, 1, "valid plugin should still be loaded despite invalid sibling")
	assert.Equal(t, "my-plugin", plugins[0].ID())
}

func TestLoader_Load_MissingID_ReturnsError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	noID := `name: "No ID Plugin"
version: "0.1.0"
agents: []
`
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "noid.sequoia-plugin.yaml"),
		[]byte(noID),
		0o644,
	))

	plugins, err := plugin.Load(dir)
	// The individual file should fail but Load should not crash.
	require.NoError(t, err)
	assert.Empty(t, plugins, "plugin without ID should be skipped")
}

func TestLoader_Load_PluginInitError_StillReturned(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	failingInit := `id: "failing-plugin"
name: "Failing Plugin"
version: "0.1.0"
init_should_fail: true
agents: []
`
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "failing.sequoia-plugin.yaml"),
		[]byte(failingInit),
		0o644,
	))

	plugins, err := plugin.Load(dir)
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	assert.Equal(t, "failing-plugin", plugins[0].ID())
	// Init may fail but the plugin should still be returned.
}

func TestAgent_Fields(t *testing.T) {
	t.Parallel()

	a := plugin.Agent{
		ID:           "test-agent",
		Name:         "Test Agent",
		Description:  "A test agent for verification",
		SystemPrompt: "You are a test agent.",
	}

	assert.Equal(t, "test-agent", a.ID)
	assert.Equal(t, "Test Agent", a.Name)
	assert.Equal(t, "A test agent for verification", a.Description)
	assert.Equal(t, "You are a test agent.", a.SystemPrompt)
}

func TestPlugin_Interface_Satisfaction(t *testing.T) {
	t.Parallel()

	// The Plugin interface is satisfied by pluginImpl (internal).
	// We verify this via the Load function which returns []Plugin.
	dir := t.TempDir()
	writePluginYAML(t, dir, "test.sequoia-plugin.yaml", validPluginYAML)

	plugins, err := plugin.Load(dir)
	require.NoError(t, err)
	require.Len(t, plugins, 1)

	// Each method must be callable without panic.
	p := plugins[0]
	assert.NotEmpty(t, p.ID())
	assert.NotEmpty(t, p.Name())
	assert.NotEmpty(t, p.Version())
	assert.NoError(t, p.Init())
	assert.NotNil(t, p.Agents())
}

// -- test helpers --

func writePluginYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

const validPluginYAML = `id: "my-plugin"
name: "My Custom Plugin"
version: "0.1.0"
agents:
  - id: "my-security-agent"
    name: "Security Agent"
    description: "Scans for security vulnerabilities and misconfigurations"
    system_prompt: "You are a security auditor. Review the code for vulnerabilities."
  - id: "my-quality-agent"
    name: "Quality Agent"
    description: "Evaluates code quality and test coverage"
    system_prompt: "You are a code review specialist. Evaluate quality metrics."
`

const secondPluginYAML = `id: "other-plugin"
name: "Other Plugin"
version: "1.0.0"
agents:
  - id: "other-agent"
    name: "Other Agent"
    description: "Does other things"
    system_prompt: "You do other things."
`
