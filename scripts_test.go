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
		// Script uses: sequoia_${VERSION}_${OS}_${ARCH}.tar.gz
		assert.Regexp(t, `sequoia_.*\$\{OS\}_\$\{ARCH\}\.tar\.gz`, script,
			"download URL must use goreleaser naming: sequoia_VERSION_OS_ARCH.tar.gz")
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

// TestInstallShChecksumMandatory validates FIX-007: checksum verification is
// mandatory in install.sh. When checksums.txt download fails, the script MUST
// abort with exit code 2. The --skip-checksums flag is opt-in to bypass.
func TestInstallShChecksumMandatory(t *testing.T) {
	content, err := os.ReadFile("scripts/install.sh")
	require.NoError(t, err, "scripts/install.sh must exist")

	script := string(content)

	t.Run("has SKIP_CHECKSUMS or skip-checksums flag", func(t *testing.T) {
		// The script MUST document the --skip-checksums opt-in flag.
		hasFlag := strings.Contains(script, "skip-checksums") ||
			strings.Contains(script, "SKIP_CHECKSUMS")
		assert.True(t, hasFlag,
			"install.sh must expose --skip-checksums / SKIP_CHECKSUMS flag for air-gapped environments")
	})

	t.Run("aborts on checksum download failure", func(t *testing.T) {
		// The script MUST NOT silently skip checksum verification when download fails.
		// It must exit with code 2 (EXIT_CHECKSUM) or call log_error and exit.
		// Verify the || true fallback is REMOVED from checksum download commands.
		lines := strings.Split(script, "\n")
		checksumOrTrue := false
		for _, line := range lines {
			// Only check lines that are actual download commands (curl/wget) AND
			// mention checksum — not diagnostic lines that list files.
			isDownload := strings.Contains(line, "curl") || strings.Contains(line, "wget")
			isChecksum := strings.Contains(line, "checksum") ||
				strings.Contains(line, "CHECKSUM") ||
				strings.Contains(line, "checksums")
			if isDownload && isChecksum && strings.Contains(line, "|| true") {
				checksumOrTrue = true
				break
			}
		}
		assert.False(t, checksumOrTrue,
			"checksum download must NOT use || true — must abort on failure")
	})

	t.Run("checksum failure exits with code 2", func(t *testing.T) {
		// When checksum download fails, the script must exit with EXIT_CHECKSUM (2)
		// unless SKIP_CHECKSUMS is set.
		// Verify the script has the mandatory-checksum message AND the abort code.
		hasMandatory := strings.Contains(script, "mandatory")
		hasExitChecksum := strings.Contains(script, "exit $EXIT_CHECKSUM")
		assert.True(t, hasMandatory,
			"install.sh must state checksum verification is mandatory")
		assert.True(t, hasExitChecksum,
			"install.sh must exit with EXIT_CHECKSUM on checksum failure")

		// Verify that exit $EXIT_CHECKSUM appears AFTER the "Could not download"
		// error message (not just in hash mismatch path). Both exist, but we
		// also verify no || true on download lines (tested separately).
		dlFailIdx := strings.Index(script, "Could not download checksums")
		exitIdx := strings.LastIndex(script, "exit $EXIT_CHECKSUM")
		if dlFailIdx >= 0 && exitIdx >= 0 {
			assert.True(t, exitIdx > dlFailIdx,
				"exit $EXIT_CHECKSUM must appear after 'Could not download checksums' error")
		}
	})

	t.Run("skip-checksums documented in help", func(t *testing.T) {
		// The --help flag must document the --skip-checksums option
		// and mention air-gapped environments.
		hasAirGapped := strings.Contains(script, "air-gapped")
		assert.True(t, hasAirGapped,
			"install.sh help must mention air-gapped use case for --skip-checksums")

		// The help text must list SKIP_CHECKSUMS as an environment variable
		hasSkipEnv := strings.Contains(script, "SKIP_CHECKSUMS")
		assert.True(t, hasSkipEnv,
			"install.sh must document SKIP_CHECKSUMS environment variable")
	})

	t.Run("checksum verification is default-on", func(t *testing.T) {
		// SKIP_CHECKSUMS defaults to "false" — verification is ON by default.
		// This ensures the opt-in flag is truly opt-in.
		assert.Contains(t, script, `SKIP_CHECKSUMS="${SKIP_CHECKSUMS:-false}"`,
			"SKIP_CHECKSUMS must default to false (verification mandatory by default)")
	})
}

// TestInstallPs1ChecksumMandatory validates FIX-007: checksum verification is
// mandatory in install.ps1. When checksums.txt download fails, the script MUST
// abort with exit code 2 unless -SkipChecksum is explicitly set.
func TestInstallPs1ChecksumMandatory(t *testing.T) {
	content, err := os.ReadFile("scripts/install.ps1")
	require.NoError(t, err, "scripts/install.ps1 must exist")

	script := string(content)

	t.Run("SkipChecksum switch is documented", func(t *testing.T) {
		// The SkipChecksum parameter already exists — verify it is documented.
		assert.Contains(t, script, "SkipChecksum",
			"install.ps1 must have -SkipChecksum switch parameter documented")
	})

	t.Run("checksum download aborts on failure", func(t *testing.T) {
		// When checksum download fails and -SkipChecksum is NOT set,
		// the script must abort with exit $EXIT_CHECKSUM (2).
		// The old behavior was: catch → Write-Warn "Skipping" → continue.
		// New behavior: catch → if -SkipChecksum → warn; else → error + exit 2.
		hasSkippingWarning := strings.Contains(script,
			"Skipping verification")
		assert.True(t, hasSkippingWarning,
			"warning about skipping checksums should still exist (for -SkipChecksum case)")

		// Verify the mandatory verification message and abort code exist
		hasMandatory := strings.Contains(script, "mandatory")
		assert.True(t, hasMandatory,
			"install.ps1 must state checksum verification is mandatory")

		// Verify exit $EXIT_CHECKSUM appears after "Could not download checksums"
		dlFailIdx := strings.Index(script, "Could not download checksums")
		exitIdx := strings.LastIndex(script, "exit $EXIT_CHECKSUM")
		if dlFailIdx >= 0 && exitIdx >= 0 {
			assert.True(t, exitIdx > dlFailIdx,
				"exit $EXIT_CHECKSUM must appear after 'Could not download checksums' error")
		}
	})

	t.Run("checksum download uses retry", func(t *testing.T) {
		// The checksum download should use retry logic similar to binary download.
		// Invoke-WebRequest supports -RetryIntervalSec and -MaximumRetryCount in PS 6+,
		// but for PS 5.1 compatibility, look for retry pattern or at least a try/catch.
		// As a minimum, the download should be wrapped in try/catch.
		hasTryCatchInChecksum := false
		lines := strings.Split(script, "\n")
		inChecksumBlock := false
		tryDepth := 0
		for _, line := range lines {
			if strings.Contains(line, "checksum") || strings.Contains(line, "Checksum") {
				inChecksumBlock = true
			}
			if inChecksumBlock && strings.Contains(line, "try {") {
				tryDepth++
			}
			if inChecksumBlock && tryDepth > 0 && strings.Contains(line, "Invoke-WebRequest") &&
				(strings.Contains(line, "Checksum") || strings.Contains(line, "checksums")) {
				hasTryCatchInChecksum = true
				break
			}
		}
		assert.True(t, hasTryCatchInChecksum,
			"checksum download should be in try/catch block for error handling")
	})

	t.Run("SkipChecksum documented as opt-in", func(t *testing.T) {
		// The -SkipChecksum flag must be documented as opt-in for air-gapped envs.
		hasAirGapped := strings.Contains(script, "air-gapped")
		assert.True(t, hasAirGapped,
			"install.ps1 help must mention air-gapped use case for -SkipChecksum")
	})

	t.Run("checksum verification is default-on", func(t *testing.T) {
		// Verification must be ON by default. The SkipChecksum switch is opt-in.
		// Verify the default value is not $true.
		hasParamSkip := strings.Contains(script, "[switch]$SkipChecksum")
		assert.True(t, hasParamSkip,
			"install.ps1 must have -SkipChecksum as a switch parameter (default: off)")
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
		// Windows artifact: sequoia_${ResolvedVersion}_${OS}_${Arch}.zip
		assert.Regexp(t, `sequoia_.*\$\{OS\}_\$\{Arch\}\.zip`, script,
			"download URL must use goreleaser naming: sequoia_VERSION_OS_Arch.zip")
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
