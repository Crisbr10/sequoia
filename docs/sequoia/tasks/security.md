# Security Tasks — Sequoia Audit

## Context

Sequoia is a Go CLI installer that deploys AI tool skills into user home directories (`~/.claude/`, `~/.gemini/`, etc.). Security concerns center on:
1. **Supply chain integrity** — install scripts run without verification
2. **Input validation** — user-provided paths and tool detection logic are unsanitized
3. **CI/CD security** — pipeline lacks SAST, vuln scanning, SBOM generation

Security findings include 11 items: 4 medium, 4 low, 3 info. The root cause RC-002 (no install script integrity) drives the medium findings P1-001 and P1-002. RC-003 (CI/CD lacks gates) drives P6-001 and P6-003.

## Priority Tiers

### Tier 1 — Immediate (critical + high)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| RC-002 | Add checksum verification and script signing to install pipeline | medium | P1-001, P1-002 |
| RC-003 | Add SAST, vuln scanning, SBOM generation to CI/CD | medium | P6-001, P6-003 |

### Tier 2 — Short Term (medium)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| P1-001 | Sanitize user-provided paths in install/uninstall commands | small | — |
| P1-002 | Guard adapter detection against PATH injection | small | — |
| P6-001 | Generate SBOM with `go version -m` in CI | small | — |
| P6-003 | Add `govulncheck` and `gosec` to CI workflow | small | — |

### Tier 3 — Long Term (low + info)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| P1-004 | Create SECURITY.md with vulnerability disclosure policy | small | — |
| P1-005 | Pin and verify go.sum entries with `go mod verify` in CI | small | — |
| P1-006 | Redact filesystem paths from debug output | small | — |
| P1-007 | Add `t.Parallel()` annotations to independent security tests | small | — |
| P1-008 | Add rate-limiting to install retry logic | small | — |
| P1-009 | Randomize staging directory prefix | small | — |
| P1-010 | Add structured security event logging | medium | — |

## Detailed Tasks

### RC-002 — Install Scripts Lack Integrity-by-Design
- **Severity**: HIGH (drives 6 findings)
- **Evidence**: `install.ps1` (root), shell install scripts — no checksum verification, no GPG/minisign signature, no mechanism for users to verify script authenticity.
- **Problem**: A compromised GitHub release, MitM proxy, or supply-chain attack silently installs malicious content. Users have zero cryptographic assurance that the script they downloaded matches what was authored. This affects 100% of installations.
- **Fix**:
  1. Generate SHA256 checksums for all release artifacts (`.exe`, `.ps1`, `.sh`) using `goreleaser` checksum config
  2. Sign checksums file with GPG or minisign
  3. Add `--verify` flag to install scripts that checks the downloaded binary against published checksums
  4. Document the verification procedure in README.md getting-started section
  5. Add CI step that validates checksum file integrity before publishing release
- **Verification**: Run `install.ps1 --verify` on a clean Windows VM — must fail if checksums don't match. Modify one byte of the binary and confirm verification catches it. Test with valid checksums to confirm success path.
- **References**: SLSA Level 2+, CWE-494 (Download of Code Without Integrity Check), NIST SP 800-218 (SSDF)

### RC-003 — CI/CD Pipeline Lacks Security Gates
- **Severity**: HIGH (drives 8 findings)
- **Evidence**: `.github/workflows/ci.yml` — runs only `go test`, `golangci-lint`, `go vet`. Missing: SAST, dependency scanning, fuzz testing, SBOM generation.
- **Problem**: Security regressions enter production undetected. Dependencies with known CVEs are not flagged. No automated proof that the binary was built from audited source.
- **Fix**:
  1. Add `gosec` scan step to `ci.yml` with minimum severity threshold
  2. Add `govulncheck ./...` step to scan for known CVEs in Go modules
  3. Add `go test -fuzz=. -fuzztime=30s` step for fuzz-tested functions
  4. Add SBOM generation: `go version -m sequoia.exe > sbom.spdx.json` and attach to release
  5. Enable Dependabot for Go modules (currently only GH Actions in `dependabot.yml`)
  6. Add reproducible build check: build on ubuntu+windows+macos and compare binary hashes
- **Verification**: Open a PR that introduces `os.Setenv("KEY", "hardcoded-secret")` — gosec must fail. Add an old dependency version with a known CVE — govulncheck must fail. Check that SBOM is attached to release artifacts.
- **References**: OWASP Top 10 CI/CD, SLSA Level 3, CWE-1104 (Use of Unmaintained Third-Party Components)

### P1-001 — Install Paths Lack Input Sanitization
- **Severity**: MEDIUM
- **Evidence**: Adapter `Install()` and `Uninstall()` methods accept paths from configuration without validating for path traversal or injection characters.
- **Problem**: A malicious or misconfigured path value could write files outside intended directories (path traversal) or execute unintended operations.
- **Fix**:
  1. Add `filepath.Clean()` normalization to all user-provided paths before use
  2. Validate that resolved paths are within expected base directories (prefix check)
  3. Reject paths containing null bytes, shell metacharacters on Unix
  4. Add unit tests with traversal attempts (`../../etc/passwd`, `..\Windows\System32`)
- **Verification**: Run tests with path traversal inputs. Confirm they are rejected with a clear error message. Test on both Windows and Unix paths.
- **References**: CWE-22 (Path Traversal), OWASP Input Validation Cheat Sheet

### P1-002 — PATH Injection Risk in Adapter Detection
- **Severity**: MEDIUM
- **Evidence**: `Detect()` methods search for tool executables on the system PATH. No validation that the found binary is from a trusted location.
- **Problem**: An attacker with write access to a user-writable directory early in PATH (e.g., `.local/bin`) can place a malicious binary that Sequoia's Detect() finds before the real tool. This could lead to code execution during adapter operations.
- **Fix**:
  1. Verify detected binary paths are in standard installation locations (e.g., `/usr/local/bin`, `C:\Program Files\`)
  2. Check file ownership/permissions on detected binaries
  3. Emit a warning when binary is found in a non-standard location
  4. Add `--tool-path` flag to let users explicitly specify binary location
- **Verification**: Create a fake binary in a temp directory early in PATH. Confirm Detect() either rejects it or logs a clear warning.
- **References**: CWE-426 (Untrusted Search Path), MITRE ATT&CK T1574.007 (Path Interception by PATH Environment Variable)

### P6-001 — No SBOM Generation in Release Pipeline
- **Severity**: MEDIUM
- **Evidence**: `.github/workflows/release.yml` — goreleaser publishes binaries with checksums but no SBOM. No `go version -m` or CycloneDX/SPDX output.
- **Problem**: Users and auditors cannot determine which dependencies and versions are embedded in a given release. Compliance with Executive Order 14028 (US) or EU Cyber Resilience Act becomes impossible.
- **Fix**:
  1. Add SBOM generation step in `ci.yml` and `release.yml`: `go version -m sequoia.exe > sbom.spdx.json`
  2. Attach SBOM to GitHub Release artifacts
  3. Add SBOM validation step: verify that SBOM lists match `go.sum`
- **Verification**: After next release, download the SBOM and verify all go.mod dependencies are listed with correct versions.
- **References**: NTIA SBOM Minimum Elements, SPDX 2.3, CycloneDX 1.5

### P6-003 — No Dependency Vulnerability Scanning in CI
- **Severity**: MEDIUM
- **Evidence**: `.github/workflows/ci.yml` — no `govulncheck` or `nancy` step. Dependabot config (`.github/dependabot.yml`) only monitors GitHub Actions, not Go modules.
- **Problem**: Dependencies with known CVEs are not flagged during PRs or releases. A `go get -u` could pull in a vulnerable version.
- **Fix**:
  1. Add `govulncheck ./...` step to `ci.yml`
  2. Update `.github/dependabot.yml` to include `go_modules` ecosystem
  3. Configure Dependabot to open PRs automatically for security updates
  4. Add CI check that fails on critical/high CVEs
- **Verification**: Trigger a Dependabot PR by pinning a known-vulnerable dependency version. Confirm govulncheck fails the CI run.
- **References**: CWE-1104, GitHub Advisory Database, Go Vulnerability Database (govulncheck)
