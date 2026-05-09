# Proposal: One-Line Installer Scripts

## Intent

Provide zero-dependency installation via `curl | bash` (Unix) and `irm | iex` (Windows). Users can install Sequoia without Go toolchain or manual binary management.

## Scope

### In Scope
- `scripts/install.sh` — bash: detect OS/arch, download binary, verify SHA-256, run `sequoia install`
- `scripts/install.ps1` — PowerShell: same flow for Windows
- Graceful handling when Sequoia is already installed (idempotent)
- Smoke tests validating detection, download URL construction, and hash verification on each target platform

### Out of Scope
- GoReleaser config (T-033)
- CI pipeline matrix (T-023)
- Binary signing / notarization
- Homebrew formula / Scoop manifest (T-033)
- `sequoia update` command (future)
- Telemetry or analytics

## Capabilities

### New Capabilities
- `installer-scripts`: Platform-native one-line installers that download, verify, and invoke the Sequoia CLI

### Modified Capabilities
None — no existing specs to modify.

## Approach

Two scripts, same contract:

1. **Detection**: `uname -s` / `uname -m` → normalized OS (`darwin|linux`) + arch (`amd64|arm64`)
2. **Download**: Construct URL from `REPO` variable (default `/sequoia-ai/sequoia-ai`, configurable) → `$REPO/releases/download/$VERSION/sequoia_$OS_$ARCH.tar.gz`
3. **Verify**: SHA-256 checksum against `$REPO/releases/download/$VERSION/checksums.txt`
4. **Install**: Extract binary, make executable, run `./sequoia install`
5. **Idempotent**: Check if `sequoia` binary already exists at target; skip download if version matches

`install.ps1` mirrors the flow using `Invoke-WebRequest`, `Get-FileHash`, and `Expand-Archive`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `scripts/install.sh` | New | Bash one-line installer |
| `scripts/install.ps1` | New | PowerShell one-line installer |
| `scripts/` | New | Directory (currently does not exist) |

No Go code changes. Scripts are standalone distributables.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| SHA-256 mismatch on download | Low | Script aborts with clear error; user re-runs |
| Architecture string normalization edge cases (e.g., `aarch64` vs `arm64`) | Med | Map explicitly in a case statement; fail on unknown |
| `curl | bash` security concerns | Low | Document `REPO` override; recommend inspecting script first |
| Network failure mid-download | Low | `curl -f` / `Invoke-WebRequest` with error handling; retry guidance |

## Rollback Plan

Delete `scripts/` directory. No Go code is touched. Users who installed via script retain the installed binary (standard `sequoia uninstall` path).

## Dependencies

- T-019 (CLI base) — **DONE**. Script invokes `sequoia install` once binary is downloaded.
- GitHub repo with releases — not yet created. Scripts default `REPO` to a placeholder; must be configurable via env var.

## Success Criteria

- [ ] `curl -sSL https://raw.githubusercontent.com/$REPO/main/scripts/install.sh | bash` installs Sequoia on macOS (amd64, arm64) and Linux (amd64, arm64)
- [ ] `irm https://raw.githubusercontent.com/$REPO/main/scripts/install.ps1 | iex` installs Sequoia on Windows (amd64)
- [ ] Re-running either script when already installed exits gracefully with "already installed" message
- [ ] Smoke test validates SHA-256 mismatch triggers error, not partial install
