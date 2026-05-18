# Operations Tasks — Sequoia Audit

## Context
Sequoia is distributed via GitHub Releases, Homebrew, Scoop, install scripts (curl/irm), and a GitHub Composite Action. The CI/CD pipeline uses GitHub Actions with 3 workflows (ci.yml, release.yml, test-action.yml) and GoReleaser v2 for cross-platform builds signed with cosign keyless. The audit found gaps in supply chain security, release quality gates, and CI coverage. The audit identified 10 operations findings: 3 MEDIUM, 6 LOW, 1 INFO.

**Key files**: `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `action.yml`, `.goreleaser.yaml`, `.github/dependabot.yml`, `scripts/install.sh`, `scripts/install.ps1`

## Priority Tiers

### Tier 1 — Immediate (HIGH via RC-003)

| ID | Task | Effort | Blocks |
|----|------|--------|--------|
| RC-003 | Add CodeQL, pre-release testing, vuln scanning to CI/CD | medium | P6-001..P6-009 |

### Tier 2 — Short Term (MEDIUM)

| ID | Task | Effort |
|----|------|--------|
| P6-001 | Enable CodeQL and secret scanning | small |
| P6-002 | Add pre-release test gate to release.yml | small |
| P6-003 | Fix GitHub Action to download signed binary for "latest" | small |
| P6-007 | Remove continue-on-error from CI smoke tests | small |

### Tier 3 — Long Term (LOW + INFO)

| ID | Task | Effort |
|----|------|--------|
| P6-004 | Add ARM64 to CI test matrix | medium |
| P6-005 | Use separate fine-grained token for Scoop tap | small |
| P6-006 | Add input validation for fail-on severity | small |
| P6-008 | Re-enable -race on Windows in CI | small |
| P6-009 | Add Dependabot grouping, reviewers, labels | small |
| P6-010 | Add --verbose/--debug flag for CLI troubleshooting | small |

---

## Detailed Tasks

### RC-003 — Add security and quality gates to CI/CD pipeline
- **Severity**: HIGH
- **Evidence**: `.github/workflows/ci.yml` — runs go test/lint/vet but lacks: CodeQL, secret scanning, dependency vuln check, pre-release testing, SBOM generation, ARM64 coverage, Windows race detection
- **Problem**: The CI/CD pipeline was built for build+ship, not verify+defend. Vulnerabilities and supply-chain issues enter production undetected. No automated assurance that releases match source.
- **Fix**:
  1. **CodeQL**: Create `.github/workflows/codeql.yml` with `go` language analysis, scheduled weekly + on PR to main
  2. **Secret scanning**: Enable via GitHub repo settings (no code change needed, free for public repos)
  3. **Dependency vuln check**: Add `go run golang.org/x/vuln/cmd/govulncheck ./...` step to ci.yml
  4. **Pre-release testing**: Add `go test -race -count=1 ./...` job before GoReleaser in release.yml
  5. **SBOM**: Add `go version -m sequoia > sbom.txt` step in release.yml
  6. See child tasks below for individual CI fixes
- **Verification**: CI passes with all new gates. Release pipeline blocks on test failure. CodeQL produces zero alerts (or known-accepted alerts).
- **References**: SLSA Level 3; GitHub CodeQL documentation; Go supply chain security

### P6-001 — Enable CodeQL and secret scanning
- **Severity**: MEDIUM (absorbed by RC-003)
- **Evidence**: `.github/` — no `codeql/` directory, no codeql-analysis.yml workflow, no secret scanning config
- **Problem**: With ~19,200 LOC and 120 Go files, manual review cannot catch all vulnerabilities. CodeQL would catch common Go issues (SQL injection patterns, path traversal, crypto weaknesses). Secret scanning prevents accidental credential commits.
- **Fix**:
  1. Create `.github/workflows/codeql.yml`:
     ```yaml
     name: CodeQL
     on: [push, pull_request, schedule: {cron: '0 6 * * 1'}]
     jobs:
       analyze:
         runs-on: ubuntu-latest
         steps:
           - uses: actions/checkout@v4
           - uses: github/codeql-action/init@v3
             with: { languages: go }
           - uses: github/codeql-action/analyze@v3
     ```
  2. Enable secret scanning in GitHub repo Settings → Security → Secret scanning
- **Verification**: CodeQL workflow runs on PR. No critical alerts (or documented exceptions).
- **References**: https://docs.github.com/en/code-security/code-scanning

### P6-002 — Add pre-release test gate to release workflow
- **Severity**: MEDIUM (absorbed by RC-003)
- **Evidence**: `.github/workflows/release.yml:12-39` — GoReleaser job runs directly with no testing
- **Problem**: CI only triggers on push/PR to main, not on tags. A tag push on a branch that hasn't passed CI can trigger a release. The go test suite is never invoked before building and publishing release artifacts.
- **Fix**:
  1. Add job before goreleaser in release.yml:
     ```yaml
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v6
           with: { go-version-file: go.mod, cache: true }
         - run: go test -race -count=1 ./...
     ```
  2. Make goreleaser job `needs: test`
- **Verification**: Push a test tag — test job must pass before goreleaser runs. Test failure blocks release.
- **References**: SLSA Level 3 non-falsifiable provenance

### P6-003 — Fix GitHub Action "latest" to download signed binary
- **Severity**: MEDIUM (absorbed by RC-003)
- **Evidence**: `action.yml:84-89` — `if [ "$VERSION" = "latest" ]; then go build -o sequoia ...`
- **Problem**: The default path (`sequoia-version: latest`) builds from source via `go build`, bypassing cosign-signed releases entirely. Every default usage gets an unsigned binary. The download-with-curl logic (lines 91-94) is unreachable for the default case.
- **Fix**:
  1. Query GitHub API for latest release tag: `curl -s https://api.github.com/repos/Crisbr10/sequoia/releases/latest | jq -r .tag_name`
  2. Use the returned tag to construct download URL: `https://github.com/Crisbr10/sequoia/releases/download/${TAG}/sequoia_${OS}_${ARCH}${EXT}`
  3. Download checksums.txt and verify SHA-256 before executing
  4. Keep `go build` as fallback only if download fails
- **Verification**: Default action run downloads signed binary. SHA-256 matches checksums.txt. Cosign signature verifiable.
- **References**: SLSA supply chain integrity; cosign keyless verification

### P6-007 — Remove continue-on-error from CI smoke tests
- **Severity**: MEDIUM (recalibrated from LOW, absorbed by RC-004)
- **Evidence**: `.github/workflows/ci.yml:56-65` — Install Smoke and Uninstall Smoke both have `continue-on-error: true`
- **Problem**: The smoke tests validate the tool's core functionality (installing into AI tools). With `continue-on-error`, regressions in the install/uninstall flow go completely undetected — CI stays green while the tool is broken. This is the last line of defense for P3-001 (backup data loss) and P2-001 (wrong template content).
- **Fix**:
  1. Remove `continue-on-error: true` from Install Smoke step
  2. Remove `continue-on-error: true` from Uninstall Smoke step
  3. If environment-specific issues exist (e.g., missing AI tools on CI runner), use conditional: `if: runner.os == 'Linux'` to test only where tools exist
  4. Add explicit `|| echo "SKIP: no tools detected" && exit 0` if no tools found
- **Verification**: CI fails red when install smoke fails. CI passes green only when install succeeds or is legitimately skipped.
- **References**: CI best practice: smoke tests must gate deployment

### P6-004 — Add ARM64 to CI test matrix
- **Severity**: LOW
- **Evidence**: `.github/workflows/ci.yml:15-16` — matrix varies only by OS (ubuntu/macos/windows), all x86_64
- **Problem**: `.goreleaser.yaml` builds `linux/arm64` and `darwin/arm64` artifacts that ship to users untested. Apple Silicon Macs and ARM cloud instances receive binaries never validated in CI.
- **Fix**:
  1. Add `macos-14` (Apple Silicon) runner to matrix: `os: [ubuntu-latest, macos-latest, macos-14, windows-latest]`
  2. Alternatively: use `architecture: [amd64, arm64]` dimension with runner selection
  3. Note: ARM64 runners cost more minutes; evaluate cost vs coverage
  4. At minimum: add `macos-14` for darwin/arm64 coverage
- **Verification**: CI runs on ARM64. Binary built and tested on Apple Silicon.
- **References**: GitHub larger runners documentation

### P6-005 — Use separate fine-grained token for Scoop tap
- **Severity**: LOW
- **Evidence**: `.goreleaser.yaml:149` — `token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"` in Scoop config
- **Problem**: Scoop configuration uses the Homebrew token. If repos require different scopes or token has excessive permissions, this violates least-privilege. Variable name is misleading.
- **Fix**:
  1. Create a dedicated fine-grained PAT with write access only to `Crisbr10/scoop-sequoia`
  2. Store as `SCOOP_TAP_TOKEN` in repo secrets
  3. Reference: `token: "{{ .Env.SCOOP_TAP_TOKEN }}"`
  4. Document both tokens in CONTRIBUTING.md
- **Verification**: Release workflow pushes to Scoop using dedicated token. Homebrew token has no Scoop access.
- **References**: GitHub Actions security hardening; least-privilege principle

### P6-006 — Add input validation for fail-on severity
- **Severity**: LOW (absorbed by RC-003)
- **Evidence**: `action.yml:20-26` — `fail-on` input accepts any string without validation
- **Problem**: Users can pass invalid severity strings like "none" or "urgent". When real audit logic is added (v0.2.0), unvalidated severity could fail silently or produce incorrect exit codes.
- **Fix**:
  1. Add validation step in action shell script:
     ```bash
     VALID_SEVERITIES="critical high medium low never"
     if ! echo "$VALID_SEVERITIES" | grep -qw "$FAIL_ON"; then
       echo "Error: fail-on must be one of: $VALID_SEVERITIES"
       exit 1
     fi
     ```
- **Verification**: Invalid severity input causes action to fail with clear error message.
- **References**: GitHub Actions metadata syntax; input validation best practices

### P6-008 — Re-enable -race on Windows in CI
- **Severity**: LOW (absorbed by RC-003)
- **Evidence**: `.github/workflows/ci.yml:38-43` — `-race` flag excluded on Windows with platform conditional
- **Problem**: Go 1.24 has significantly improved Windows race detector support. The project uses Go 1.24 but continues to exclude -race on Windows with no tracking issue. Windows-specific data races go undetected.
- **Fix**:
  1. Test `go test -race ./...` on Windows locally (`sequoia.exe` exists in repo)
  2. If clean: remove the Windows conditional, use `-race` on all platforms
  3. If issues: file issue tracking specific failures, add `TODO(#N): re-enable -race on Windows when Go fixes #XXXXX`
- **Verification**: CI runs -race on Windows. No false positives. Data races caught on all platforms.
- **References**: Go 1.24 release notes; race detector improvements on Windows

### P6-009 — Add Dependabot grouping, reviewers, labels
- **Severity**: INFO (absorbed by RC-003)
- **Evidence**: `.github/dependabot.yml` — configured but lacks grouping, reviewers, labels
- **Problem**: Each dependency bump gets its own PR generating noise and CI cost. No reviewers assigned means PRs sit unreviewed.
- **Fix**:
  1. Add groups for Go modules:
     ```yaml
     groups:
       go-deps:
         patterns: ["*"]
         update-types: ["minor", "patch"]
     ```
  2. Add `reviewers: ["Crisbr10"]` and `labels: ["dependencies"]`
  3. Increase `open-pull-requests-limit` slightly (to 10) since groups reduce PR count
- **Verification**: Dependabot opens grouped PRs with labels and reviewers. CI runs once per group.
- **References**: Dependabot configuration options

### P6-010 — Add debug/verbose mode for CLI troubleshooting
- **Severity**: LOW
- **Evidence**: `cmd/sequoia/main.go:81-101` — root command has `SilenceUsage: true`, no `--debug` or `--verbose` flag
- **Problem**: When install fails, error messages are generic. No way to increase verbosity for debugging adapter-level details, backup paths, or file operation traces. Support teams can't ask users to "run with --debug".
- **Fix**:
  1. Add persistent flag: `root.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")`
  2. Add persistent flag: `root.PersistentFlags().Bool("debug", false, "Enable debug logging (adapter paths, file operations)")`
  3. Create `internal/log` package with Debug/Info/Warn/Error levels controlled by flags
  4. Add debug logs: adapter detection paths, backup directory paths, file copy operations, template rendering inputs
- **Verification**: `sequoia install --debug` outputs file paths and adapter details. `sequoia status --verbose` shows detection logic.
- **References**: 12-factor app logs; Cobra persistent flags pattern
