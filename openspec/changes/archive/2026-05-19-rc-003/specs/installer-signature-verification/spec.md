# installer-signature-verification Specification (Delta)

## Purpose

Cosign signature verification in install.sh and install.ps1, layered after SHA-256 checksum verification, with graceful fallback.

## Requirements

### REQ-IV-001: Cosign Verification — Unix (install.sh)

After checksum check, MUST attempt cosign verify-blob using .sig + .cert from release, verifying against Fulcio/Rekor trust root.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Sig valid | cosign on PATH, valid sig+cert | verify-blob runs | Passes, install proceeds |
| Sig invalid | cosign on PATH, tampered sig | verify-blob runs | Script exits error, binary NOT installed |
| cosign absent | cosign not on PATH | reaches verification | Warning printed, skipped, SHA-256 only |
| Sig files missing | cosign on PATH, .sig/.cert missing | download fails | Warning printed, SHA-256 only |

### REQ-IV-002: Cosign Verification — Windows (install.ps1)

Mirrors install.sh behavior in PowerShell.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Sig valid (Win) | cosign.exe on PATH, valid sig+cert | verify-blob runs | Passes, install proceeds |
| cosign absent (Win) | cosign.exe not found | reaches verification | Warning, skipped, SHA-256 only |
| Sig invalid (Win) | cosign.exe on PATH, verification fails | verify-blob runs | Exit error, binary NOT installed |

### REQ-IV-003: Graceful Fallback Contract

Both scripts MUST treat cosign absence as non-fatal:
- Detect via command -v (Unix) / Get-Command (Windows)
- Print clear warning with install instructions
- NEVER block installation due to cosign unavailability
- Log fallback reason

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Any fallback | Any condition blocks sig verification | Reaches sig check | MUST NOT fail, proceed SHA-256 only |
