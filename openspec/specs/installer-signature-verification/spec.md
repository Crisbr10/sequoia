# installer-signature-verification Specification

## Purpose

Cosign signature verification in install.sh and install.ps1, layered on top of existing SHA-256 checksum verification, with graceful fallback when cosign is absent.

## Requirements

### REQ-IV-001: Cosign Verification — Unix (install.sh)

After SHA-256 checksum verification passes, install.sh MUST attempt `cosign verify-blob` using `.sig` and `.cert` files from the release. It MUST verify against the Fulcio/Rekor public keyless trust root.

#### Scenario: Signature valid — cosign available

- GIVEN cosign is on PATH and the release provides valid `.sig` + `.cert` files
- WHEN install.sh runs signature verification after checksum check
- THEN verification SHALL pass and installation SHALL proceed

#### Scenario: Signature invalid — cosign available

- GIVEN cosign is on PATH but `.sig` is tampered or mismatched
- WHEN install.sh runs signature verification
- THEN the script MUST exit with an error and MUST NOT install the binary

#### Scenario: cosign not installed

- GIVEN cosign is not found on PATH
- WHEN install.sh reaches signature verification
- THEN the script SHALL print a warning recommending cosign installation
- AND SHALL skip signature verification and proceed with SHA-256 only

#### Scenario: Signature files missing from release

- GIVEN cosign is on PATH but `.sig` or `.cert` files fail to download
- WHEN install.sh attempts signature verification
- THEN the script SHALL print a warning and proceed with SHA-256 only

### REQ-IV-002: Cosign Verification — Windows (install.ps1)

install.ps1 MUST mirror the cosign verification behavior of install.sh, with identical logic adapted for PowerShell.

#### Scenario: Signature valid — cosign available (Windows)

- GIVEN cosign.exe is on PATH and `.sig` + `.cert` are valid
- WHEN install.ps1 runs signature verification after checksum check
- THEN verification SHALL pass and installation SHALL proceed

#### Scenario: cosign not installed (Windows)

- GIVEN cosign.exe is not found
- WHEN install.ps1 reaches signature verification
- THEN the script SHALL print a warning and proceed with SHA-256 only

#### Scenario: Signature invalid (Windows)

- GIVEN cosign.exe is on PATH but signature verification fails
- WHEN install.ps1 runs verification
- THEN the script MUST exit with error and MUST NOT install the binary

### REQ-IV-003: Graceful Fallback Contract

Both install scripts MUST treat cosign absence as non-fatal. The scripts SHALL:

- Detect cosign availability via `command -v cosign` (Unix) or `Get-Command cosign` (Windows)
- Print a clear warning when cosign is missing, including installation instructions
- NEVER fail or block installation due to cosign unavailability
- Log the fallback reason (absent binary, missing sig files, download failure)

#### Scenario: All fallback paths

- GIVEN any condition where signature verification cannot complete (absent tool, missing files, network error downloading sig/cert)
- WHEN the installer reaches signature verification
- THEN the installer MUST NOT fail and SHALL proceed with SHA-256 verification only
