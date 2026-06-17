# Tasks: Fix Install Script Source-Only Release Detection

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~146 (ps1: ~28, sh: ~68, test: ~50) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low

## Phase 1: RED — Failing Tests (Strict TDD)

- [x] 1.1 **Add failing test for install.ps1 source-only-release detection**
  Add `TestInstallPs1SourceOnlyReleaseDetection` to `scripts/install_scripts_test.go`. Assert the script contains: (a) the marker `source-only release`, (b) the GitHub API call pattern `releases/tags/$ResolvedVersion` (or `releases/tags/${VERSION}`), (c) remediation text containing `previous version` and `https://github.com/Crisbr10/sequoia/issues`. Test must FAIL at this point — the script does not yet contain the new code.
  Files: `scripts/install_scripts_test.go`
  Verification: `go test ./scripts/... -run TestInstallPs1SourceOnlyReleaseDetection -count=1` — must FAIL.

- [x] 1.2 **Add failing test for install.sh source-only-release detection**
  Add `TestInstallShSourceOnlyReleaseDetection` to `scripts/install_scripts_test.go`. Assert: (a) marker `source-only release`, (b) GitHub API URL appears ≥2 times (both curl and wget branches), (c) remediation text with `previous version` and `https://github.com/Crisbr10/sequoia/issues`. Test must FAIL at this point.
  Files: `scripts/install_scripts_test.go`
  Verification: `go test ./scripts/... -run TestInstallShSourceOnlyReleaseDetection -count=1` — must FAIL.

## Phase 2: GREEN — Implementation

- [x] 2.1 **Implement source-only-release detection in install.ps1**
  In `scripts/install.ps1`, modify the binary-download `catch` block (near lines 198–207). Detect HTTP 404 via `$_.Exception.Response.StatusCode.value__ -eq 404` OR string-match on `404|Not Found`. On 404, call `Invoke-WebRequest -Uri "https://api.github.com/repos/$Repo/releases/tags/$ResolvedVersion" -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop`. If API returns HTTP 200, emit `Write-Err` with the source-only-release message containing the marker `source-only release`, the resolved tag, `previous version` hint, and `https://github.com/Crisbr10/sequoia/issues`. Exit `$EXIT_NETWORK`. On API failure or non-404, fall through to existing generic `Download failed. Please check:` error and exit `$EXIT_NETWORK`.
  Files: `scripts/install.ps1`
  Verification: `go test ./scripts/... -run TestInstallPs1SourceOnlyReleaseDetection -count=1` — must PASS. All other tests must still pass.

- [x] 2.2 **Implement source-only-release detection in install.sh**
  In `scripts/install.sh`, extend both the `curl` (line ~274) and `wget` (line ~282) error branches. Detect 404 via `curl -fsSL ... -w '%{http_code}'` exit code 22 or wget exit code 8. On detection, call `curl -fsSL --max-time 10 "https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"`. If API succeeds (exit 0), call `log_error` with the source-only-release message containing `source-only release`, `$VERSION`, `previous version`, and `https://github.com/Crisbr10/sequoia/issues`. Exit `$EXIT_NETWORK`. On API failure fall through to existing generic `Download failed. Please check:` error and exit `$EXIT_NETWORK`. Both branches must be covered.
  Files: `scripts/install.sh`
  Verification: `go test ./scripts/... -run TestInstallShSourceOnlyReleaseDetection -count=1` — must PASS. All other tests must still pass.

## Phase 3: Verification — Regression Suite

- [x] 3.1 **Run full scripts test suite and confirm no regressions**
  Run `go test ./scripts/... -count=1`. All existing tests plus the 2 new tests must pass. No existing assertion substring must be removed from either script.
  Files: (no edits)
  Verification: `go test ./scripts/... -count=1` — all tests in the package report PASS.

## Phase 4: Cleanup — Lint and Format

- [x] 4.1 **Run linter and formatter on test file**
  Run `gofmt -l scripts/install_scripts_test.go` — must produce no output. Run `go vet ./scripts/...` — must report no issues. Verify bash script syntax with `bash -n scripts/install.sh` — must report no errors.
  Files: `scripts/install_scripts_test.go` (format only)
  Verification: `gofmt -l scripts/install_scripts_test.go` returns empty. `go vet ./scripts/...` returns no issues. `bash -n scripts/install.sh` returns no syntax errors.