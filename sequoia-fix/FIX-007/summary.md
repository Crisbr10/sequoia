# FIX-007 — Mandatory Checksum Verification

**Status**: ✅ Complete
**Date**: 2026-05-12

## Problem
`install.sh` and `install.ps1` downloaded `checksums.txt` to verify the binary, but if the download failed (network, rate limiting, 404), the scripts silently skipped verification and continued. An attacker controlling the network could block `checksums.txt` and serve a malicious binary.

## Solution

### `scripts/install.sh` (Bash)
- **Removed** `|| true` fallback from checksum download — download failure now exits with code 2
- **Added** `SKIP_CHECKSUMS` environment variable (default: `false`)
- **Added** `--skip-checksums` flag for direct script invocation
- **Added** `--help` flag documenting options
- **Default behavior**: mandatory verification — script aborts if checksums.txt cannot be downloaded
- **Opt-in bypass**: `SKIP_CHECKSUMS=true` or `--skip-checksums` allows continuation without verification (air-gapped environments)

### `scripts/install.ps1` (PowerShell)
- **Fixed** catch block: when checksum download fails and `-SkipChecksum` is NOT set → abort with exit code 2
- **When** `-SkipChecksum` IS set → warn and continue
- **Updated** parameter help text to clarify mandatory default
- `-SkipChecksum` switch already existed but was only used to skip the block entirely; now controls whether failure is fatal

### Tests (`scripts_test.go`)
- Added `TestInstallShChecksumMandatory` — 5 subtests:
  1. Has `SKIP_CHECKSUMS` or `skip-checksums` flag
  2. `|| true` removed from checksum download lines
  3. Checksum download failure exits with code 2
  4. `--skip-checksums` documented in help (air-gapped)
  5. `SKIP_CHECKSUMS` defaults to `false` (verification ON by default)
- Added `TestInstallPs1ChecksumMandatory` — 5 subtests:
  1. `-SkipChecksum` switch is documented
  2. Checksum download failure aborts with exit 2
  3. Checksum download uses try/catch
  4. `-SkipChecksum` documented as opt-in (air-gapped)
  5. `-SkipChecksum` is a `[switch]` parameter (default: off)

### Exit Codes
| Code | Name | Meaning |
|------|------|---------|
| 0 | EXIT_OK | Success |
| 1 | EXIT_GENERAL | General error |
| 2 | EXIT_CHECKSUM | Checksum verification failed (download or mismatch) |
| 3 | EXIT_NETWORK | Network/download error |
