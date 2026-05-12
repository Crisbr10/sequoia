package example_test

import (
	"testing"

	"github.com/Crisbr10/sequoia/plugin/example"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamplePlugin_ImplementsPlugin(t *testing.T) {
	t.Parallel()

	p := example.NewHelloPlugin()
	assert.NotNil(t, p)
}

func TestExamplePlugin_ID(t *testing.T) {
	t.Parallel()

	p := example.NewHelloPlugin()
	assert.Equal(t, "hello-world", p.ID())
}

func TestExamplePlugin_Name(t *testing.T) {
	t.Parallel()

	p := example.NewHelloPlugin()
	assert.Equal(t, "Hello World Plugin", p.Name())
}

func TestExamplePlugin_Version(t *testing.T) {
	t.Parallel()

	p := example.NewHelloPlugin()
	assert.Equal(t, "0.1.0", p.Version())
}

func TestExamplePlugin_Init(t *testing.T) {
	t.Parallel()

	p := example.NewHelloPlugin()
	assert.NoError(t, p.Init())
}

func TestExamplePlugin_Agents(t *testing.T) {
	t.Parallel()

	p := example.NewHelloPlugin()
	agents := p.Agents()
	require.Len(t, agents, 1)

	a := agents[0]
	assert.Equal(t, "hello-world-greeter", a.ID)
	assert.Equal(t, "Hello World Greeter", a.Name)
	assert.Contains(t, a.Description, "greets")
	assert.Contains(t, a.SystemPrompt, "Hello")
}
