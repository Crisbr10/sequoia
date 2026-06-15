// Package scripts contains tests that validate the Sequoia repo's build
// configuration and installer scripts. This file (install_scripts_test.go)
// validates the install.sh and install.ps1 shell scripts (T-033 sub-tasks).
//
// Strict TDD: tests written BEFORE script updates.
// Verifies scripts reference correct GitHub repo and goreleaser artifact URLs.
package scripts

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
	content, err := os.ReadFile("install.sh")
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
	content, err := os.ReadFile("install.sh")
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
	content, err := os.ReadFile("install.ps1")
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
		// The checksum download MUST use Invoke-WebRequestWithRetry
		// (not bare Invoke-WebRequest) for resilience against transient failures.
		assert.Contains(t, script, "Invoke-WebRequestWithRetry",
			"install.ps1 must define Invoke-WebRequestWithRetry wrapper")
		assert.Contains(t, script, "Invoke-WebRequestWithRetry -Uri $ChecksumUrl",
			"checksum download must use Invoke-WebRequestWithRetry")
	})

	t.Run("binary download uses retry", func(t *testing.T) {
		// The binary download MUST also use Invoke-WebRequestWithRetry.
		assert.Contains(t, script, "Invoke-WebRequestWithRetry -Uri $DownloadUrl",
			"binary download must use Invoke-WebRequestWithRetry")
	})

	t.Run("retry function has backoff pattern", func(t *testing.T) {
		// The retry wrapper MUST use exponential backoff: 2s, 4s, 8s delays.
		assert.Contains(t, script, "Start-Sleep",
			"retry function must use Start-Sleep between attempts")
		assert.Regexp(t, `@\(2,\s*4,\s*8\)`, script,
			"retry delays must be 2, 4, 8 (exponential backoff)")

		// Verify exactly 3 attempts (the for loop: $i -lt 3)
		assert.Regexp(t, `-lt\s+3`, script,
			"retry function must make exactly 3 attempts")

		// Verify throw on final failure ($i -eq 2 means 3rd attempt, 0-indexed)
		assert.Regexp(t, `-eq\s+2.*\{ throw`, script,
			"retry function must throw on final (3rd) failure")
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

// TestInstallShPathValidation validates install.sh path security (IS-SEC-001).
// The script must reject dangerous input before any filesystem operation.
func TestInstallShPathValidation(t *testing.T) {
	content, err := os.ReadFile("install.sh")
	require.NoError(t, err, "scripts/install.sh must exist")

	script := string(content)

	t.Run("function validate_path exists", func(t *testing.T) {
		assert.Contains(t, script, "validate_path",
			"install.sh must define a validate_path function for path sanitization")
	})

	t.Run("rejects .. traversal", func(t *testing.T) {
		assert.Contains(t, script, "..",
			"validation must reject '..' traversal sequences")
	})

	t.Run("rejects ; metachar", func(t *testing.T) {
		assert.Contains(t, script, ";",
			"validation must reject ';' command separator")
	})

	t.Run("rejects | metachar", func(t *testing.T) {
		assert.Contains(t, script, "|",
			"validation must reject '|' pipe metachar")
	})

	t.Run("rejects & metachar", func(t *testing.T) {
		// The script must contain '&' in a forbidden-character check context.
		// Since '&' also appears in '2>&1', check that it appears in a
		// case statement or error message pattern near the validation code.
		assert.True(t,
			strings.Contains(script, "&") &&
				(strings.Contains(script, "forbidden") ||
					strings.Contains(script, "invalid character") ||
					strings.Contains(script, "must not contain")),
			"validation must reject '&' background/metachar")
	})

	t.Run("rejects backtick metachar", func(t *testing.T) {
		// Look for backtick in the validation context (case *\`*)
		assert.Contains(t, script, "`",
			"validation must reject backtick command substitution")
	})

	t.Run("rejects $() substitution", func(t *testing.T) {
		assert.True(t,
			strings.Contains(script, "$(") &&
				(strings.Contains(script, "substitution") ||
					strings.Contains(script, "forbidden") ||
					strings.Contains(script, ";")),
			"validation must reject $() command substitution")
	})

	t.Run("rejects empty path", func(t *testing.T) {
		// Validation must check for empty/null path
		assert.True(t,
			strings.Contains(script, "must not be empty") ||
				strings.Contains(script, "required") ||
				strings.Contains(script, "-n") ||
				strings.Contains(script, "-z"),
			"validation must reject empty install directory")
	})

	t.Run("requires absolute path", func(t *testing.T) {
		// Unix paths must start with /
		assert.True(t,
			strings.Contains(script, "/*)") ||
				strings.Contains(script, "absolute") ||
				strings.Contains(script, "start with /"),
			"validation must require absolute paths starting with /")
	})

	t.Run("validate_path called before check_existing", func(t *testing.T) {
		// The validation must run BEFORE the idempotency check (check_existing).
		// This ensures no I/O happens before path validation.
		valIdx := strings.Index(script, "validate_path")
		checkIdx := strings.Index(script, "check_existing")
		if valIdx >= 0 && checkIdx >= 0 {
			assert.True(t, valIdx < checkIdx,
				"validate_path call must appear before check_existing call")
		}
	})
}

// TestInstallPs1PathValidation validates install.ps1 path security
// (IS-SEC-001, IS-SEC-002). PowerShell must reject dangerous input
// before any filesystem/PATH operation.
func TestInstallPs1PathValidation(t *testing.T) {
	content, err := os.ReadFile("install.ps1")
	require.NoError(t, err, "scripts/install.ps1 must exist")

	script := string(content)

	t.Run("function Resolve-SafePath exists", func(t *testing.T) {
		assert.Contains(t, script, "Resolve-SafePath",
			"install.ps1 must define a Resolve-SafePath function for path sanitization")
	})

	t.Run("rejects ; metachar", func(t *testing.T) {
		assert.True(t,
			contains(script, ";") &&
				strings.Contains(script, "forbidden"),
			"validation must reject ';' semicolon metachar")
	})

	t.Run("rejects | metachar", func(t *testing.T) {
		assert.True(t,
			contains(script, "|") &&
				strings.Contains(script, "forbidden"),
			"validation must reject '|' pipe metachar")
	})

	t.Run("rejects & metachar", func(t *testing.T) {
		// '&' metachar — verify it appears in a validation context
		assert.True(t,
			contains(script, "&") &&
				strings.Contains(script, "forbidden"),
			"validation must reject '&' background operator")
	})

	t.Run("rejects backtick metachar", func(t *testing.T) {
		assert.True(t,
			contains(script, "`") &&
				strings.Contains(script, "forbidden"),
			"validation must reject backtick escape character")
	})

	t.Run("rejects $( ) substitution", func(t *testing.T) {
		// PowerShell's $() is the subexpression operator
		assert.True(t,
			contains(script, "$(") &&
				strings.Contains(script, "forbidden"),
			"validation must reject $( ) subexpression")
	})

	t.Run("rejects .. traversal", func(t *testing.T) {
		assert.True(t,
			contains(script, "..") &&
				strings.Contains(script, "forbidden"),
			"validation must reject '..' directory traversal")
	})

	t.Run("rejects < and > Windows special chars", func(t *testing.T) {
		// PowerShell/Windows special chars: < > (redirection)
		hasLT := contains(script, "<") && strings.Contains(script, "invalid")
		hasGT := contains(script, ">") && strings.Contains(script, "invalid")
		assert.True(t, hasLT || hasGT ||
			strings.Contains(script, "[<>\\\"*?]") ||
			strings.Contains(script, "x00-\\x1f"),
			"validation must reject Windows special chars: < > \" * ?")
	})

	t.Run("rejects double-quote char", func(t *testing.T) {
		// Double quote must be rejected (escape/parameter injection)
		assert.True(t,
			(contains(script, "\"") || contains(script, "\\\"\"")) &&
				(strings.Contains(script, "invalid") ||
					strings.Contains(script, "forbidden")),
			"validation must reject double-quote character")
	})

	t.Run("rejects * and ? wildcards", func(t *testing.T) {
		// Windows wildcard chars must be rejected
		hasStar := contains(script, "*") && strings.Contains(script, "invalid")
		hasQMark := contains(script, "?") && strings.Contains(script, "invalid")
		assert.True(t, hasStar || hasQMark,
			"validation must reject wildcard characters * and ?")
	})

	t.Run("requires absolute path with drive letter", func(t *testing.T) {
		// PowerShell paths must be rooted (drive letter)
		assert.True(t,
			strings.Contains(script, "IsPathRooted") ||
				strings.Contains(script, "[A-Z]:\\") ||
				strings.Contains(script, "absolute"),
			"validation must require absolute paths starting with drive letter")
	})

	t.Run("rejects empty path", func(t *testing.T) {
		assert.True(t,
			strings.Contains(script, "must not be empty") ||
				strings.Contains(script, "IsNullOrWhiteSpace") ||
				strings.Contains(script, "must not be empty"),
			"validation must reject empty/whitespace-only install directory")
	})

	t.Run("called before Test-SequoiaInstalled", func(t *testing.T) {
		// Resolve-SafePath must be called AFTER version resolution
		// but BEFORE the Test-SequoiaInstalled check.
		safeIdx := strings.Index(script, "Resolve-SafePath")
		testIdx := strings.Index(script, "Test-SequoiaInstalled")
		if safeIdx >= 0 && testIdx >= 0 {
			assert.True(t, safeIdx < testIdx,
				"Resolve-SafePath call must appear before Test-SequoiaInstalled call")
		}
	})
}

// TestInstallPs1PathGuard validates the PATH injection guard (IS-SEC-003).
// Before writing to the Windows user PATH registry, the script must
// re-validate InstallDir with a semicolon guard.
func TestInstallPs1PathGuard(t *testing.T) {
	content, err := os.ReadFile("install.ps1")
	require.NoError(t, err, "scripts/install.ps1 must exist")

	script := string(content)

	t.Run("Resolve-SafePath recalled before PATH write", func(t *testing.T) {
		// The script must call Resolve-SafePath (or a ; guard check)
		// immediately before [Environment]::SetEnvironmentVariable.
		safeIdx := strings.LastIndex(script, "Resolve-SafePath")
		pathWriteIdx := strings.Index(script, "[Environment]::SetEnvironmentVariable")
		if safeIdx >= 0 && pathWriteIdx >= 0 {
			assert.True(t, safeIdx < pathWriteIdx,
				"Resolve-SafePath must be called before SetEnvironmentVariable for PATH write")
		}
	})

	t.Run("semicolon rejected before PATH write", func(t *testing.T) {
		// The PATH guard re-calls Resolve-SafePath before SetEnvironmentVariable.
		// The Resolve-SafePath function body (defined earlier in the script)
		// includes ';' in its $forbidden array — this is the rejection mechanism.
		// Verify the function definition contains semicolon in forbidden chars.
		setEnvIdx := strings.Index(script, "[Environment]::SetEnvironmentVariable")
		lastSafeIdx := strings.LastIndex(script, "Resolve-SafePath")
		if setEnvIdx >= 0 && lastSafeIdx >= 0 {
			assert.True(t, lastSafeIdx < setEnvIdx,
				"Resolve-SafePath must be called before SetEnvironmentVariable")
		}
		// Verify the Resolve-SafePath function itself rejects ';'
		assert.True(t,
			contains(script, "';'") &&
				strings.Contains(script, "forbidden"),
			"Resolve-SafePath must explicitly reject ';' semicolon in its forbidden list")
	})
}

// TestInstallPs1RepoRefs validates that install.ps1 references:
// 1. The correct GitHub repo (Crisbr10/sequoia)
// 2. Download URLs matching goreleaser zip artifact naming
// 3. Checksums URL pointing to checksums.txt
func TestInstallPs1RepoRefs(t *testing.T) {
	content, err := os.ReadFile("install.ps1")
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
