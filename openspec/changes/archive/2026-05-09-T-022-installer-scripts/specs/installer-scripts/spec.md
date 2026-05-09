# Installer Scripts Specification

## Purpose

Platform-native one-line installers (bash and PowerShell) that download, verify,
and invoke the Sequoia CLI without requiring Go toolchain or manual binary
management.

## Requirements

### Requirement: OS and Architecture Detection

The install scripts MUST detect the host OS and CPU architecture and normalize
them to a canonical `OS-ARCH` pair used in binary download URLs.

| OS | Canonical | Architectures |
|----|-----------|---------------|
| Darwin (macOS) | `darwin` | `amd64`, `arm64` |
| Linux | `linux` | `amd64`, `arm64` |
| Windows | `windows` | `amd64` |

Supported pairs: `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`,
`windows-amd64`.

#### Scenario: macOS Intel detection

- GIVEN the script runs on macOS with an Intel (x86_64) CPU
- WHEN OS/arch detection executes
- THEN `OS` MUST be set to `darwin` and `ARCH` to `amd64`

#### Scenario: Apple Silicon detection with aarch64 normalization

- GIVEN the script runs on Apple Silicon where `uname -m` reports `aarch64` or `arm64`
- WHEN OS/arch detection executes
- THEN `ARCH` MUST be normalized to `arm64`

#### Scenario: Unsupported platform

- GIVEN the script runs on an unsupported OS/arch (e.g., FreeBSD, linux/s390x)
- WHEN OS/arch detection executes
- THEN the script MUST exit with code 1 and print an error listing supported
  platforms

### Requirement: Binary Download

The install scripts MUST download the correct release binary tarball from a
configurable GitHub repository.

The download URL SHALL be constructed as:

```
$REPO/releases/download/$VERSION/sequoia_$OS_$ARCH.tar.gz
```

Where `REPO` and `VERSION` are configurable via environment variables with
sensible defaults.

#### Scenario: Successful download with curl

- GIVEN curl is available and `REPO`/`VERSION` env vars are set to valid values
- WHEN the download step executes
- THEN the script MUST fetch the release asset via `curl -fsSL`
- AND store it in a temporary directory

#### Scenario: Successful download with wget fallback

- GIVEN curl is NOT available but wget is
- WHEN the download step executes
- THEN the script MUST use wget with equivalent error handling

#### Scenario: Network error during download

- GIVEN the network is unavailable or the release URL returns HTTP 4xx/5xx
- WHEN the download step executes
- THEN the script MUST exit with a network-specific exit code
- AND print a message recommending the user verify connectivity and
  `REPO`/`VERSION` values

### Requirement: SHA-256 Checksum Verification

The install scripts MUST verify the downloaded binary's SHA-256 checksum against
the published `checksums.txt` from the release.

#### Scenario: Checksum match

- GIVEN the release asset is downloaded successfully
- WHEN SHA-256 verification executes
- THEN the script MUST download `checksums.txt` from the same release
- AND compare the computed hash against the published entry
- AND proceed to installation ONLY if they match

#### Scenario: Checksum mismatch

- GIVEN the computed SHA-256 does NOT match the published checksum
- WHEN verification completes
- THEN the script MUST exit with a checksum-specific exit code
- AND print expected vs. actual hashes
- AND MUST NOT leave the downloaded binary in an executable state

### Requirement: Binary Installation

The install scripts MUST extract the downloaded binary, set correct permissions,
and invoke the Sequoia installer.

#### Scenario: Unix installation

- GIVEN a verified tarball is available for macOS or Linux
- WHEN the install step executes
- THEN the script MUST extract the `sequoia` binary
- AND make it executable (`chmod +x`)
- AND run `./sequoia install`

#### Scenario: Windows installation

- GIVEN a verified ZIP archive is available for Windows
- WHEN the install step executes
- THEN the script MUST extract `sequoia.exe` using `Expand-Archive`
- AND run `.\sequoia.exe install`

### Requirement: Idempotent Installation

The install scripts MUST detect when Sequoia is already installed and skip the
download/install process with an informational message.

#### Scenario: Already installed — same version

- GIVEN the `sequoia` binary exists at the expected location AND its version
  matches the resolved version
- WHEN the script runs
- THEN it MUST print "Sequoia {version} is already installed"
- AND exit with code 0 without downloading

#### Scenario: Already installed — different version

- GIVEN the `sequoia` binary exists BUT its version differs from the resolved version
- WHEN the script runs
- THEN it MUST proceed with download and install (upgrading the binary)

### Requirement: Missing Tool Detection

The install scripts MUST detect missing required tools (curl, wget, tar, unzip,
shasum) and exit with a clear diagnostic before proceeding.

#### Scenario: No download tool available

- GIVEN neither curl nor wget is found on the system
- WHEN the script runs
- THEN it MUST exit with code 1
- AND print "Error: curl or wget is required to download the binary"

#### Scenario: Missing hash utility

- GIVEN sha256sum and shasum are both unavailable (Unix)
- WHEN the checksum verification step is reached
- THEN the script MUST exit with code 1
- AND print which tool it was looking for
