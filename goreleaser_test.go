// Package sequoia validates the GoReleaser configuration (T-033).
//
// This file follows the Strict TDD cycle:
//   RED → test written first (goreleaser.yaml does not exist yet → will fail)
//   GREEN → goreleaser.yaml created to pass all assertions
//
// Because .goreleaser.yaml is purely structural (YAML config, no branching logic),
// triangulation is skipped per strict-tdd.md rules.
package sequoia

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// goreleaserSchema mirrors the subset of GoReleaser v2 config that T-033 requires.
type goreleaserSchema struct {
	Version  int `yaml:"version"`
	Builds   []goreleaserBuild
	Archives []goreleaserArchive
	Checksum goreleaserChecksum
	Release  goreleaserRelease
	Changelog goreleaserChangelog
	Brews    []goreleaserBrew   `yaml:"brews"`
	Scoops   []goreleaserScoop  `yaml:"scoops"`
}

type goreleaserBuild struct {
	ID     string   `yaml:"id"`
	Main   string   `yaml:"main"`
	Binary string   `yaml:"binary"`
	Goos   []string `yaml:"goos"`
	Goarch []string `yaml:"goarch"`
	Ldflags []string `yaml:"ldflags"`
	Env    []string `yaml:"env"`
}

type goreleaserArchive struct {
	ID           string           `yaml:"id"`
	NameTemplate string           `yaml:"name_template"`
	Formats      []string         `yaml:"formats"`
	FormatOverrides []goreleaserFormatOverride `yaml:"format_overrides"`
}

type goreleaserFormatOverride struct {
	Goos   string   `yaml:"goos"`
	Goarch string   `yaml:"goarch,omitempty"`
	Formats []string `yaml:"formats"`
}

type goreleaserChecksum struct {
	NameTemplate string `yaml:"name_template"`
}

type goreleaserRelease struct {
	Draft      bool   `yaml:"draft"`
	Discussion goreleaserDiscussion `yaml:"discussion"`
}

type goreleaserDiscussion struct {
	Category string `yaml:"category"`
}

type goreleaserChangelog struct {
	Use   string            `yaml:"use"`
	Groups []goreleaserChangelogGroup `yaml:"groups"`
	Sort  string            `yaml:"sort"`
	Filters goreleaserChangelogFilters `yaml:"filters"`
}

type goreleaserChangelogGroup struct {
	Title  string `yaml:"title"`
	Regexp string `yaml:"regexp"`
	Order  int    `yaml:"order"`
}

type goreleaserChangelogFilters struct {
	Exclude []string `yaml:"exclude"`
}

type goreleaserBrew struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Homepage     string `yaml:"homepage"`
	License      string `yaml:"license"`
	Repository   goreleaserRepo `yaml:"repository"`
	URLTemplate  string `yaml:"url_template"`
	Install      string `yaml:"install"`
}

type goreleaserScoop struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Homepage     string `yaml:"homepage"`
	License      string `yaml:"license"`
	Repository   goreleaserRepo `yaml:"repository"`
	URLTemplate  string `yaml:"url_template"`
}

type goreleaserRepo struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
}

// requiredBuildTargets encodes the 5 OS/arch combos from T-033 acceptance criteria.
// goreleaser uses `goos` + `goarch` arrays; we validate these pairs are covered.
var requiredBuildTargets = []struct{ goos, goarch string }{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"windows", "amd64"},
}

// TestGoReleaserConfig validates .goreleaser.yaml against T-033 requirements.
func TestGoReleaserConfig(t *testing.T) {
	content, err := os.ReadFile(".goreleaser.yaml")
	require.NoError(t, err, ".goreleaser.yaml must exist")

	var cfg goreleaserSchema
	require.NoError(t, yaml.Unmarshal(content, &cfg), ".goreleaser.yaml must be valid YAML")

	t.Run("version is 2", func(t *testing.T) {
		assert.Equal(t, 2, cfg.Version, "GoReleaser v2 syntax required")
	})

	t.Run("has at least one build", func(t *testing.T) {
		require.NotEmpty(t, cfg.Builds, "at least one build definition required")
	})

	t.Run("build targets cover all 5 OS/arch combos", func(t *testing.T) {
		build := cfg.Builds[0]
		require.NotEmpty(t, build.Goos, "build.goos must not be empty")
		require.NotEmpty(t, build.Goarch, "build.goarch must not be empty")

		goosSet := toSet(build.Goos)
		goarchSet := toSet(build.Goarch)

		for _, target := range requiredBuildTargets {
			assert.True(t, goosSet[target.goos],
				"build.goos must include %s", target.goos)
			assert.True(t, goarchSet[target.goarch],
				"build.goarch must include %s", target.goarch)
		}
	})

	t.Run("build main points to cmd/sequoia", func(t *testing.T) {
		assert.Equal(t, "./cmd/sequoia/", cfg.Builds[0].Main,
			"main must point to cmd/sequoia")
	})

	t.Run("binary name is sequoia", func(t *testing.T) {
		assert.Equal(t, "sequoia", cfg.Builds[0].Binary,
			"binary name must be sequoia")
	})

	t.Run("ldflags inject version", func(t *testing.T) {
		require.NotEmpty(t, cfg.Builds[0].Ldflags,
			"ldflags must be set for version injection")
		hasVersionFlag := false
		for _, f := range cfg.Builds[0].Ldflags {
			if contains(f, ".Version=") {
				hasVersionFlag = true
				break
			}
		}
		assert.True(t, hasVersionFlag,
			"ldflags must inject version into main package")
	})

	t.Run("archives configured", func(t *testing.T) {
		require.NotEmpty(t, cfg.Archives, "archives section must exist")
		archive := cfg.Archives[0]

		assert.Contains(t, archive.NameTemplate, ".Version",
			"archive name template should include version")
		assert.Contains(t, archive.NameTemplate, ".Os",
			"archive name template should include OS")
		assert.Contains(t, archive.NameTemplate, ".Arch",
			"archive name template should include Arch")

		assert.Contains(t, archive.Formats, "tar.gz",
			"tar.gz format required for linux/darwin")

		// Verify format_overrides: windows → zip, others → tar.gz
		hasWindowsOverride := false
		for _, override := range archive.FormatOverrides {
			if override.Goos == "windows" {
				hasWindowsOverride = true
				assert.Contains(t, override.Formats, "zip",
					"windows override should use zip format")
			}
		}
		assert.True(t, hasWindowsOverride, "format_overrides must include windows → zip")
	})

	t.Run("checksum enabled", func(t *testing.T) {
		assert.Contains(t, cfg.Checksum.NameTemplate, "checksums",
			"checksum name template should reference checksums")
	})

	t.Run("release configured as draft on GitHub", func(t *testing.T) {
		assert.True(t, cfg.Release.Draft,
			"release should be draft mode to allow review before publishing")
	})

	t.Run("changelog uses conventional commits", func(t *testing.T) {
		assert.Equal(t, "github-native", cfg.Changelog.Use,
			"changelog should use github-native")
		assert.NotEmpty(t, cfg.Changelog.Groups,
			"changelog should group commits (feat, fix, docs, etc.)")
	})

	t.Run("homebrew formula defined", func(t *testing.T) {
		require.NotEmpty(t, cfg.Brews, "homebrew formula must be configured")
		brew := cfg.Brews[0]
		assert.NotEmpty(t, brew.Name)
		assert.NotEmpty(t, brew.Description)
		assert.Contains(t, brew.Homepage, "github.com")
		assert.Equal(t, "MIT", brew.License, "license must be MIT")
		assert.NotEmpty(t, brew.Install, "brew install command must be defined")
	})

	t.Run("scoop manifest defined", func(t *testing.T) {
		require.NotEmpty(t, cfg.Scoops, "scoop manifest must be configured")
		scoop := cfg.Scoops[0]
		assert.NotEmpty(t, scoop.Name)
		assert.NotEmpty(t, scoop.Description)
		assert.Contains(t, scoop.Homepage, "github.com")
		assert.Equal(t, "MIT", scoop.License, "license must be MIT")
		assert.NotEmpty(t, scoop.URLTemplate, "scoop url_template must be defined")
	})
}

// toSet converts a string slice to a set for O(1) lookups.
func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

// contains checks if substr is within s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexInString(s, substr) >= 0
}

func indexInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
