// Package sequoia validates install scripts (T-033 sub-tasks).
//
// Strict TDD: tests written BEFORE script updates.
// Verifies scripts reference correct GitHub repo and goreleaser artifact URLs.
package sequoia

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstallShRepoRefs validates that install.sh references:
// 1. The correct GitHub repo (Crisbr10/sequoia)
// 2. Download URLs matching goreleaser artifact naming (sequoia_{OS}_{ARCH}.tar.gz)
// 3. Checksums URL pointing to checksums.txt
func TestInstallShRepoRefs(t *testing.T) {
	content, err := os.ReadFile("scripts/install.sh")
	require.NoError(t, err, "scripts/install.sh must exist")

	script := string(content)

	t.Run("default REPO is Crisbr10/sequoia", func(t *testing.T) {
		// The script sets REPO="${REPO:-DEFAULT}". Check the DEFAULT value.
		assert.Contains(t, script, "Crisbr10/sequoia",
			"default REPO must be Crisbr10/sequoia, not sequoia-ai/sequoia-ai")
		assert.NotContains(t, script, "sequoia-ai/sequoia-ai",
			"install.sh must NOT reference old repo sequoia-ai/sequoia-ai")
	})

	t.Run("download URL uses goreleaser naming", func(t *testing.T) {
		// The goreleaser archive name_template:
		//   {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
		// So the tarball should be: sequoia_${OS}_${ARCH}.tar.gz
		assert.Contains(t, script, "sequoia_${OS}_${ARCH}.tar.gz",
			"download URL must use goreleaser naming: sequoia_OS_ARCH.tar.gz")
		assert.Contains(t, script, "releases/download",
			"must use GitHub releases download URL")
	})

	t.Run("checksum URL references checksums.txt", func(t *testing.T) {
		assert.Contains(t, script, "checksums.txt",
			"must reference the goreleaser-generated checksums.txt")
	})

	t.Run("REPO default comment matches", func(t *testing.T) {
		// The helper comments in the script header should reference the correct repo
		lines := strings.Split(script, "\n")
		foundRawURL := false
		for _, line := range lines {
			if strings.Contains(line, "raw.githubusercontent.com") &&
				strings.Contains(line, "Crisbr10/sequoia") {
				foundRawURL = true
				break
			}
		}
		assert.True(t, foundRawURL, "raw.githubusercontent.com URLs must reference Crisbr10/sequoia")
	})
}

// TestInstallPs1RepoRefs validates that install.ps1 references:
// 1. The correct GitHub repo (Crisbr10/sequoia)
// 2. Download URLs matching goreleaser zip artifact naming
// 3. Checksums URL pointing to checksums.txt
func TestInstallPs1RepoRefs(t *testing.T) {
	content, err := os.ReadFile("scripts/install.ps1")
	require.NoError(t, err, "scripts/install.ps1 must exist")

	script := string(content)

	t.Run("default Repo is Crisbr10/sequoia", func(t *testing.T) {
		assert.Contains(t, script, "Crisbr10/sequoia",
			"default Repo must be Crisbr10/sequoia, not sequoia-ai/sequoia-ai")
		assert.NotContains(t, script, "sequoia-ai/sequoia-ai",
			"install.ps1 must NOT reference old repo sequoia-ai/sequoia-ai")
	})

	t.Run("download URL uses goreleaser naming", func(t *testing.T) {
		// Windows artifact: sequoia_windows_amd64.zip
		assert.Contains(t, script, "sequoia_${OS}_${Arch}.zip",
			"download URL must use goreleaser naming: sequoia_OS_Arch.zip")
		assert.Contains(t, script, "releases/download",
			"must use GitHub releases download URL")
	})

	t.Run("checksum URL references checksums.txt", func(t *testing.T) {
		assert.Contains(t, script, "checksums.txt",
			"must reference the goreleaser-generated checksums.txt")
	})

	t.Run("raw URL examples reference correct repo", func(t *testing.T) {
		lines := strings.Split(script, "\n")
		foundRawURL := false
		for _, line := range lines {
			if strings.Contains(line, "raw.githubusercontent.com") &&
				strings.Contains(line, "Crisbr10/sequoia") {
				foundRawURL = true
				break
			}
		}
		assert.True(t, foundRawURL, "raw.githubusercontent.com URLs must reference Crisbr10/sequoia")
	})
}
