// Package sequoia validates CI pipeline artifacts (T-033 sub-tasks).
//
// Strict TDD: tests written BEFORE the release.yml and script updates.
package sequoia

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// releaseWorkflowSchema mirrors the GitHub Actions workflow structure
// relevant to the Sequoia release CI pipeline.
type releaseWorkflowSchema struct {
	Name string                `yaml:"name"`
	On   releaseWorkflowOn     `yaml:"on"`
	Jobs map[string]releaseJob `yaml:"jobs"`
}

type releaseWorkflowOn struct {
	Push releaseWorkflowPush `yaml:"push"`
}

type releaseWorkflowPush struct {
	Tags []string `yaml:"tags"`
}

type releaseJob struct {
	Name   string           `yaml:"name"`
	RunsOn string           `yaml:"runs-on"`
	If     string           `yaml:"if,omitempty"`
	Steps  []releaseJobStep `yaml:"steps"`
}

type releaseJobStep struct {
	Name string                 `yaml:"name,omitempty"`
	Uses string                 `yaml:"uses,omitempty"`
	With map[string]interface{} `yaml:"with,omitempty"`
	Run  string                 `yaml:"run,omitempty"`
	Env  map[string]string      `yaml:"env,omitempty"`
}

// TestReleaseWorkflow validates .github/workflows/release.yml
// triggers on v* tags and runs GoReleaser.
func TestReleaseWorkflow(t *testing.T) {
	content, err := os.ReadFile(".github/workflows/release.yml")
	require.NoError(t, err, "release.yml must exist")

	var wf releaseWorkflowSchema
	require.NoError(t, yaml.Unmarshal(content, &wf), "release.yml must be valid YAML")

	t.Run("has a name", func(t *testing.T) {
		assert.NotEmpty(t, wf.Name, "workflow must have a name")
	})

	t.Run("triggers on v* tags", func(t *testing.T) {
		require.NotEmpty(t, wf.On.Push.Tags,
			"push.tags must be defined to trigger on tag push")
		// Accept v* glob patterns (e.g. "v*") or semver regex patterns
		// (e.g. "v[0-9]+.[0-9]+.[0-9]+"), as long as they match version tags.
		foundVPattern := false
		for _, tag := range wf.On.Push.Tags {
			if strings.HasPrefix(tag, "v") {
				// Matches "v*", "v[0-9]+.*", etc.
				foundVPattern = true
				break
			}
		}
		assert.True(t, foundVPattern,
			"must trigger on version tags starting with 'v' (e.g. 'v*' or 'v[0-9]+.[0-9]+.[0-9]+')")
	})

	t.Run("has at least one job", func(t *testing.T) {
		require.NotEmpty(t, wf.Jobs, "at least one job required")
	})

	t.Run("jobs run on ubuntu-latest", func(t *testing.T) {
		for _, job := range wf.Jobs {
			assert.Equal(t, "ubuntu-latest", job.RunsOn,
				"release jobs should run on ubuntu-latest")
		}
	})

	t.Run("includes goreleaser action", func(t *testing.T) {
		foundGoreleaser := false
		for _, job := range wf.Jobs {
			for _, step := range job.Steps {
				if strings.Contains(step.Uses, "goreleaser/goreleaser-action") {
					foundGoreleaser = true
				}
			}
		}
		assert.True(t, foundGoreleaser,
			"workflow must use goreleaser/goreleaser-action")
	})

	t.Run("uses GITHUB_TOKEN for release", func(t *testing.T) {
		foundToken := false
		for _, job := range wf.Jobs {
			for _, step := range job.Steps {
				if step.Env != nil {
					if token, ok := step.Env["GITHUB_TOKEN"]; ok {
						assert.Contains(t, token, "secrets.GITHUB_TOKEN",
							"GITHUB_TOKEN must reference ${{ secrets.GITHUB_TOKEN }}")
						foundToken = true
					}
				}
			}
		}
		assert.True(t, foundToken,
			"release step must pass GITHUB_TOKEN via secrets")
	})
}
