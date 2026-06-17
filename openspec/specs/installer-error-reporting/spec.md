# Delta for installer-error-reporting

> Capability: `installer-error-reporting` (NEW). Diagnostics-only fix for `scripts/install.ps1` and `scripts/install.sh`. Post-archive main spec: `openspec/specs/installer-error-reporting/spec.md`.

## ADDED Requirements

## REQ-IER-01 — install.ps1 source-only release detection

On HTTP 404 from the binary download URL, `scripts/install.ps1` MUST distinguish a "source-only release" (tag exists but no precompiled asset was uploaded) from a missing tag. The script MUST query `https://api.github.com/repos/{owner}/{repo}/releases/tags/{tag}`. On HTTP 200 it MUST emit an error containing the resolved tag and the marker `source-only release`. On API failure (timeout, DNS, or HTTP 403/429) it MUST fall back to the generic network-error message. In both cases the script MUST exit with `$EXIT_NETWORK`.

#### Scenario: 404 with reachable API emits source-only error

- GIVEN `Invoke-WebRequestWithRetry` on `$DownloadUrl` throws HTTP 404
- AND `Invoke-RestMethod` on the GitHub Releases API returns a release object (HTTP 200)
- WHEN the install.ps1 download catch block runs
- THEN `Write-Err` text contains the resolved tag and the substring `source-only release`
- AND the script exits with `$EXIT_NETWORK`

#### Scenario: 404 with unreachable API falls back to generic error

- GIVEN the binary download throws HTTP 404
- AND the GitHub Releases API call fails (timeout, DNS, or HTTP 403/429)
- WHEN the install.ps1 download catch block runs
- THEN the generic `Download failed. Please check:` message is emitted
- AND `source-only release` is NOT present
- AND the script exits with `$EXIT_NETWORK`

## REQ-IER-02 — install.sh source-only release detection

`scripts/install.sh` MUST mirror REQ-IER-01. A 404 on `$DOWNLOAD_URL` MUST trigger a GitHub Releases API lookup. On HTTP 200 the script MUST emit an error containing `$VERSION` and `source-only release` in both `curl` and `wget` branches. On API failure the script MUST fall back to the generic message and exit `$EXIT_NETWORK`.

#### Scenario: source-only error emitted in curl and wget branches

- GIVEN `$DOWNLOAD_URL` returns HTTP 404 in the `curl` branch
- AND the GitHub Releases API lookup succeeds (HTTP 200)
- WHEN the curl download-error block runs
- THEN `log_error` text contains `$VERSION` and `source-only release`
- AND the script exits with `$EXIT_NETWORK`

- GIVEN `$DOWNLOAD_URL` returns HTTP 404 in the `wget` branch
- AND the GitHub Releases API lookup succeeds (HTTP 200)
- WHEN the wget download-error block runs
- THEN `log_error` text contains `$VERSION` and `source-only release`
- AND the script exits with `$EXIT_NETWORK`

#### Scenario: API unreachable falls back to generic error

- GIVEN the binary download fails with HTTP 404 in either branch
- AND the GitHub Releases API call fails
- WHEN the download-error block runs
- THEN the generic network-error message is emitted
- AND the script exits with `$EXIT_NETWORK`

## REQ-IER-03 — Actionable remediation guidance

Both installers MUST include actionable remediation text in the source-only-release error. It MUST contain (a) a suggestion to install a previous version, and (b) a link to `https://github.com/Crisbr10/sequoia/issues`.

#### Scenario: remediation text is present in both installers

- GIVEN the source-only-release error path is taken in install.ps1 or install.sh
- WHEN the error message is inspected
- THEN it contains `previous version` (or equivalent)
- AND it contains `https://github.com/Crisbr10/sequoia/issues`

## REQ-IER-04 — Backward compatibility

All existing assertions in `scripts/install_scripts_test.go` MUST pass without modification. The source-only-release branch is reached only on HTTP 404. Preserved invariants: `$EXIT_NETWORK` exit code; generic `Download failed. Please check:` text on non-404 errors; mandatory checksum verification with `SKIP_CHECKSUMS` / `-SkipChecksum` opt-in; path validation; retry/backoff (2/4/8 delays, 3 attempts); repo refs.

#### Scenario: full existing test suite passes unchanged

- GIVEN the fix is applied to install.ps1 and install.sh
- WHEN `go test ./scripts/... -count=1` runs
- THEN every existing test in `install_scripts_test.go` passes
- AND no existing assertion substring is removed from either script

#### Scenario: generic network-error path remains reachable

- GIVEN a non-404 download failure (DNS, connection refused, HTTP 5xx)
- WHEN the download catch block runs in install.ps1 or install.sh
- THEN the generic `Download failed. Please check:` message is emitted verbatim
- AND `source-only release` is NOT present