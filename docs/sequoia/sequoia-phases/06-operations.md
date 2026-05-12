# P6 · Operations Audit Report

## 1. Objective

Analyze the Sequoia CLI (`github.com/Crisbr10/sequoia`) for CI/CD readiness, release automation quality, cross-platform build health, install script robustness, version management, and dependency hygiene. Sequoia is a Go 1.24.2 CLI shipped via GoReleaser to 3 OS × 2 architectures with Homebrew and Scoop distribution.

## 2. Scope

| Dimension | Coverage |
|-----------|----------|
| **CI workflows** | `.github/workflows/ci.yml`, `release.yml`, `test-action.yml` |
| **GoReleaser config** | `.goreleaser.yaml` (v2) |
| **Install scripts** | `scripts/install.sh`, `scripts/install.ps1` |
| **GitHub Action** | `action.yml` (composite action) |
| **Version embedding** | `cmd/sequoia/main.go` — ldflags + debug.ReadBuildInfo() |
| **Release docs** | `CHANGELOG.md`, `docs/release-notes/v0.1.0.md` |
| **Dependency mgmt** | `go.mod`, `go.sum`, Dependabot presence |
| **Build artifacts** | `.gitignore` coverage of dist/, binaries, coverage |
| **Excluded** | Docker, Kubernetes, cloud monitoring (CLI tool — not applicable) |

## 3. Verified State

- **GoReleaser v2 config**: Builds for darwin/linux/windows × amd64/arm64 (5 active targets). Windows arm64 explicitly excluded (known limitation — `scoops` section handles this). Archives as tar.gz (Unix) and zip (Windows). Checksums (SHA-256) generated. Homebrew formula and Scoop manifest auto-published. Release set as draft. Changelog groups by conventional commit type.
- **Install scripts**: Both `install.sh` (303 lines) and `install.ps1` (275 lines) are production-quality. Structured exit codes (0/1/2/3 for ok/general/checksum/network). Idempotent install with version comparison. SHA-256 checksum verification with graceful skip-on-failure. Retry logic (curl `--retry 3`). Cleanup traps. PATH guidance.
- **Version embedding**: Multi-layered fallback: ldflags (GoReleaser) → `debug.ReadBuildInfo()` (`go install @version`) → VCS revision → `"0.1.0-dev"` default. Correct.
- **Cross-platform CI**: Matrix strategy on ubuntu, macos, windows. Fail-fast disabled. `go vet` + `go test` + `go build` + smoke tests (install/status/uninstall) on all 3 OSes. Race detector runs on Unix, skipped on Windows (expected — race detector has known Windows limitations).
- **Release workflow**: Triggers on `v[0-9]+.[0-9]+.[0-9]+` tags. Uses `goreleaser/goreleaser-action@v6`. Sets `fetch-depth: 0` for changelog generation. Passes `GITHUB_TOKEN` and `HOMEBREW_TAP_TOKEN`.
- **Test-action workflow**: Validates the composite GitHub Action end-to-end. Verifies outputs (health-score, findings-count) and report artifact upload. Only triggers on relevant file changes (`action.yml`, `test-action.yml`, `cmd/sequoia/**`).
- **`.gitignore`**: Covers binaries (`*.exe`, `*.dll`, `*.so`, `*.dylib`), coverage artifacts, test binaries, `dist/`, `vendor/`, IDE files, OS files, temp files, logs, `.env`. Comprehensive.
- **`.golangci.yaml`**: v2 config with 10 linters enabled (errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, goimports, misspell, revive). Exported function check enabled. Test files excluded from errcheck.

## 4. Consolidated Findings

---

### [P6-001] · CI Go version pinned to 1.22 while go.mod declares go 1.24.2  [🔴 CRÍTICO]

**State**: Confirmed

**Evidence**: `.github/workflows/ci.yml:25` — `go-version: '1.22'` vs `go.mod:3` — `go 1.24.2`

**Problem**:
The CI workflow installs Go 1.22 via `setup-go`, but the module declares `go 1.24.2`. Go 1.22's toolchain auto-download mechanism (`GOTOOLCHAIN=auto`, default) detects the mismatch and downloads Go 1.24.2 on every CI run — bypassing `setup-go`'s cache entirely. This means:
1. Every CI matrix job (3 OS × every push) incurs a hidden ~30–60s Go toolchain download.
2. `setup-go`'s `cache: true` is wasted — the cached Go 1.22 toolchain is never used for compilation.
3. If the Go download mirror is unreachable or rate-limited, CI fails with a cryptic toolchain error.
4. `go vet` and `go test` behavior may differ between the CI's Go 1.22 and the actual Go 1.24.2 runtime (unlikely but possible with standard library changes).

**Real Impact**:
Silently slower CI (wasted runner minutes), fragile builds if Go download infrastructure is unavailable, and misleading cache configuration. Over 100 CI runs, this wastes approximately 50–100 minutes of runner time.

**Minimal High-Leverage Recommendation**:
Change `go-version: '1.22'` to `go-version: '1.24'` in `ci.yml:25` (and `go-version-file: go.mod` would be even better — auto-syncs). Same fix applies to `action.yml:55` (currently `'1.24'`, which is closer but should use `go-version-file` for auto-sync).

**Dependencies/Blockers**: None

**Implementation Risk**: Low — changing the Go version spec is a one-line YAML change. All 3 OS runners in the matrix support Go 1.24.

**Acceptance Criteria**:
- [ ] `ci.yml:25` matches the Go version in `go.mod` (1.24) or uses `go-version-file: go.mod`
- [ ] CI matrix jobs complete within the same time or faster (no hidden toolchain download)
- [ ] `setup-go` cache hit is visible in CI logs for all 3 OSes

---

### [P6-002] · No linting step in CI despite configured .golangci.yaml  [🟠 RIESGO]

**State**: Confirmed

**Evidence**:
- `.golangci.yaml` — 10 linters enabled, well-configured
- `.github/workflows/ci.yml:28-29` — only `go vet` runs; `golangci-lint` is never invoked
- Grep for `golangci` in `.github/workflows/` — zero matches

**Problem**:
The project has a thorough `.golangci.yaml` (v2) with 10 linters (errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, goimports, misspell, revive), but it is never executed in CI. Linting is only enforced locally via developer discipline. A PR can merge with `gofmt` violations, unused imports, misspellings, or static analysis warnings — and CI will happily pass.

**Real Impact**:
Code quality drift accumulates silently. A contributor who doesn't run `golangci-lint` locally can introduce lint violations that are only discovered later (or never). The `go vet` step in CI is a subset of what golangci-lint provides — it catches fewer issues.

**Minimal High-Leverage Recommendation**:
Add a `golangci-lint` step to `ci.yml` using `golangci/golangci-lint-action@v6`. Run it only on one OS (ubuntu-latest) to avoid 3× overhead. Use `version: v2` to match the `.golangci.yaml` v2 format.

**Dependencies/Blockers**: None

**Implementation Risk**: Low — golangci-lint-action is a well-maintained official action. Initial run may surface pre-existing violations that need fixing first.

**Acceptance Criteria**:
- [ ] `golangci-lint` runs on every push/PR (ubuntu job only)
- [ ] Violations fail the CI job (not just warnings)
- [ ] No pre-existing lint violations in the codebase (or explicitly excluded)

---

### [P6-003] · No coverage collection or upload in CI  [🟠 RIESGO]

**State**: Confirmed

**Evidence**:
- `.github/workflows/ci.yml:31-37` — `go test ./... -count=1 -timeout 120s` without `-cover` flag
- `.gitignore:12` — `coverage.out` is ignored
- No Codecov, Coveralls, or artifact upload for coverage data
- `README.md:13` — claims "90%+ code coverage" but this is a static, unverified claim

**Problem**:
Coverage is never collected in CI. The project claims "90%+ code coverage" in its README and release notes, but there is no automated enforcement or even measurement of this metric. Coverage can silently regress with no CI indication. The `coverage.out` file in `.gitignore` prevents any coverage artifact from being committed or uploaded.

**Real Impact**:
Coverage claims are trust-based, not evidence-based. A PR that deletes tests and drops coverage to 60% will pass CI. Over time, coverage regressions compound and test discipline erodes. This undermines the project's "Strict TDD" positioning.

**Minimal High-Leverage Recommendation**:
Add `-coverprofile=coverage.out` to `go test` in CI and upload the artifact via `actions/upload-artifact@v4`. For enforcement, add a `-cover` threshold check (e.g., fail if < 85%). Optionally integrate Codecov for historical trend visualization.

**Dependencies/Blockers**: None

**Implementation Risk**: Low — coverage collection adds negligible time to `go test`. The threshold check should start at a realistic current value to avoid breaking CI immediately.

**Acceptance Criteria**:
- [ ] `go test` in CI includes `-coverprofile=coverage.out`
- [ ] Coverage artifact is uploaded and retained for 7 days
- [ ] A minimum coverage threshold is enforced (e.g., 85%)
- [ ] Coverage badge in README is dynamic (optional, could be follow-up)

---

### [P6-004] · No Dependabot or Renovate for automated dependency updates  [🟠 RIESGO]

**State**: Confirmed

**Evidence**:
- `grep dependabot .github/` — no `.github/dependabot.yml` in project root
- `docs/Ejemplos de Skills/engram/.github/dependabot.yml` — exists but is example/docs content, not the project's own Dependabot config (references `Gentleman-Programming` reviewer, not this repo)
- `go.mod` — 7 direct dependencies, 19 transitive. No automated update mechanism.

**Problem**:
The project has no automated dependency update mechanism. Security patches for transitive dependencies (e.g., `golang.org/x/sys`, `golang.org/x/text`) will not be surfaced. GitHub Actions versions (`actions/checkout@v4`, `setup-go@v5`, `goreleaser-action@v6`) will age without notice. A critical CVE in a transitive dependency could go unnoticed for weeks or months.

**Real Impact**:
Supply-chain risk accumulates silently. When a CVE is eventually discovered in a dependency, the fix requires manual discovery + manual PR — no automated nudge. For a security-auditing tool, this is ironic and undermines credibility.

**Minimal High-Leverage Recommendation**:
Create `.github/dependabot.yml` with two update groups: `gomod` (weekly, max 5 PRs) and `github-actions` (weekly, max 3 PRs). Use `commit-message.prefix: "chore(deps)"` to match the GoReleaser changelog chore group.

**Dependencies/Blockers**: None — Dependabot is free for public repos and enabled by default on GitHub.

**Implementation Risk**: Low — Dependabot PRs are automatically created. Initial burst of PRs may be high if deps are stale; reduce `open-pull-requests-limit` initially.

**Acceptance Criteria**:
- [ ] `.github/dependabot.yml` exists in project root with `gomod` and `github-actions` ecosystems
- [ ] Dependabot is enabled in repo settings (Settings > Code security > Dependabot)
- [ ] First Dependabot PR is received within 1 week
- [ ] Commit message prefix matches conventional commit conventions (`chore(deps)`)

---

### [P6-005] · No release artifact signing — authenticity unverifiable  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `.goreleaser.yaml` — no `signs` section (no cosign, GPG, or keyless signing)
- `.goreleaser.yaml:49-51` — checksum generation is SHA-256 only, no signature over the checksum file
- GitHub release assets include only the binary archives and checksums.txt — no `.sig`, `.pem`, or Sigstore bundle

**Problem**:
Users downloading `sequoia_0.1.0_darwin_amd64.tar.gz` can verify the file's integrity (SHA-256) but cannot verify its authenticity (who built it). An attacker who compromises the GitHub release could replace both the binary and the checksums file, and the user's `sha256sum` verification would pass. Release signing (cosign keyless with OIDC, or GPG) would allow users to verify that the binary was built by the authorized CI pipeline, not a third party.

**Real Impact**:
Supply-chain attack surface: users who pipe `curl | bash` (the documented install method) trust the GitHub release endpoint implicitly. For a tool that audits code for security vulnerabilities, lacking artifact signing is a credibility gap.

**Minimal High-Leverage Recommendation**:
Add a `signs` section to `.goreleaser.yaml` using cosign keyless signing (GitHub Actions OIDC). This requires no key management — cosign uses the GitHub Actions identity token. Add a verification note to the install scripts.

**Dependencies/Blockers**: Requires the GitHub Actions workflow to have `id-token: write` permission (for OIDC). Must add `--timeout 5m` per GoReleaser docs.

**Implementation Risk**: Medium — cosign keyless signing requires OIDC permissions and `id-token: write` in the release workflow. First-time setup has a learning curve. However, GoReleaser's cosign integration is well-documented.

**Acceptance Criteria**:
- [ ] `.goreleaser.yaml` includes `signs` with cosign keyless configuration
- [ ] `release.yml` includes `id-token: write` permission
- [ ] Release assets include Sigstore bundle or `.sig` file
- [ ] Install scripts or README include cosign verification command

---

### [P6-006] · action.yml Go version diverges from go.mod — future drift risk  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `action.yml:55` — `go-version: '1.24'` (string literal)
- `go.mod:3` — `go 1.24.2` (module directive)
- `.github/workflows/ci.yml:25` — `go-version: '1.22'` (string literal, already flagged in P6-001)

**Problem**:
Three different files specify Go versions independently:
- `go.mod` → `go 1.24.2` (source of truth)
- `action.yml` → `'1.24'` (close but hardcoded)
- `ci.yml` → `'1.22'` (severely outdated, P6-001)

When `go.mod` is bumped to `go 1.25`, both `action.yml` and `ci.yml` must be manually updated. If either is missed, CI breaks or the composite action uses a wrong toolchain. Currently `action.yml` uses `'1.24'` which happens to resolve to the correct major version, but the pattern is fragile.

**Real Impact**:
Manually synchronized version pins across 3 files are a classic source of drift in Go projects. A future contributor who bumps `go.mod` may not know about the other two files. The composite action (`action.yml`) is the most dangerous — if it uses a Go version too old for the module, the action fails for every downstream user.

**Minimal High-Leverage Recommendation**:
Replace all hardcoded `go-version` strings with `go-version-file: go.mod` in both `ci.yml` and `action.yml`. The `setup-go` action reads the `go` directive from `go.mod` and installs the correct version automatically. This eliminates version drift permanently.

**Dependencies/Blockers**: `setup-go@v5` (already in use) supports `go-version-file`. No upgrade needed.

**Implementation Risk**: Low — `go-version-file` is a drop-in replacement for `go-version`. The action tests in `test-action.yml` will validate the change.

**Acceptance Criteria**:
- [ ] `ci.yml` uses `go-version-file: go.mod` instead of `go-version: '1.22'`
- [ ] `action.yml` uses `go-version-file: go.mod` instead of `go-version: '1.24'`
- [ ] `test-action.yml` CI passes with the updated action
- [ ] Bumping `go.mod` to a new Go version does not require changes to any workflow file

---

### [P6-007] · README CI badge is static — no live pipeline status  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `README.md:15` — `<img src="https://img.shields.io/badge/CI-GitHub_Actions-2088FF?style=flat&logo=githubactions" alt="CI">`
- This is a generic shields.io badge, not linked to any GitHub Actions workflow
- Compare to a dynamic badge: `![CI](https://github.com/Crisbr10/sequoia/actions/workflows/ci.yml/badge.svg)`

**Problem**:
The README displays a static "CI" badge that always shows the same blue color and text. It conveys zero information about whether CI is currently passing, failing, or was never run. Visitors, potential contributors, and users evaluating the project cannot see build health at a glance.

**Real Impact**:
Minor — the badge is cosmetic, not functional. But for a project that positions itself as a quality/audit tool, a static CI badge undermines the "evidence over opinion" philosophy. It also means the README doesn't benefit from the matrix CI results that are already running on every push.

**Minimal High-Leverage Recommendation**:
Replace the static badge with a dynamic GitHub Actions badge: `https://github.com/Crisbr10/sequoia/actions/workflows/ci.yml/badge.svg`. Also add badges for Go version (dynamic from go.mod) and optionally coverage (once P6-003 is addressed).

**Dependencies/Blockers**: None — GitHub generates badge SVGs automatically for any public workflow.

**Implementation Risk**: Low — one-line URL change. The badge will show "no status" until the next CI run completes, then update automatically.

**Acceptance Criteria**:
- [ ] README CI badge links to the actual `ci.yml` workflow status
- [ ] Badge shows green (passing) or red (failing) based on latest `main` run
- [ ] Badge is clickable and links to the Actions tab

---

### [P6-008] · CHANGELOG.md manually maintained alongside GoReleaser auto-changelog  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `CHANGELOG.md:1-28` — manually written, one entry (v0.1.0), Keep a Changelog format
- `.goreleaser.yaml:71-101` — `changelog.use: github-native` generates release notes from conventional commits, grouped by type (feat, fix, docs, refactor, build/ci, test, chore)
- `.goreleaser.yaml:97-101` — filters exclude `docs:`, `test:`, merge commits from release notes

**Problem**:
Two sources of truth for release notes exist:
1. `CHANGELOG.md` — manually maintained, Keep a Changelog format, human-curated
2. GoReleaser-generated release body — auto-generated from commit messages, grouped by type

Over time, these will diverge. The GoReleaser release body provides detailed per-commit breakdowns; the CHANGELOG.md provides curated highlights. If a maintainer updates one but not the other, users see inconsistent information. Additionally, the GoReleaser filters exclude `docs:` and `test:` commits from release notes, but `CHANGELOG.md` might include documentation changes — causing further divergence.

**Real Impact**:
Contributors and users consulting `CHANGELOG.md` may see different information than what appears on the GitHub Releases page. For a project that follows Keep a Changelog, the manual file would ideally be the source of truth — but GoReleaser's auto-generation makes the release page potentially more complete for per-commit details.

**Minimal High-Leverage Recommendation**:
Choose one strategy and commit to it:
- **Option A**: Keep `CHANGELOG.md` as the single source of truth. Set GoReleaser `changelog.use: github` (not `github-native`) to link to the CHANGELOG.md file. Update CHANGELOG.md as part of the release PR.
- **Option B**: Let GoReleaser generate the release notes. Delete `CHANGELOG.md` (or reduce it to a pointer to GitHub Releases). The GoReleaser notes are always accurate because they're commit-derived.

Option A is recommended since Keep a Changelog is already adopted and human-curated notes are more user-friendly.

**Dependencies/Blockers**: Requires discipline to update CHANGELOG.md on every release.

**Implementation Risk**: Low — process change, not code change.

**Acceptance Criteria**:
- [ ] Decision documented: CHANGELOG.md is the canonical release notes source
- [ ] GoReleaser `changelog.use` changed from `github-native` to `github` (or kept but with note that CHANGELOG.md is authoritative)
- [ ] Release checklist includes "Update CHANGELOG.md" step
- [ ] `CHANGELOG.md` and GitHub Release body are consistent for v0.2.0

---

## 5. High-Leverage Missing Items

These are infrastructure capabilities not present in the project that would significantly improve operations readiness beyond the specific findings above.

| # | Missing Item | Why It Matters | Effort | Priority |
|---|-------------|----------------|--------|----------|
| 1 | **`go mod verify` in CI** | Validates that `go.sum` matches downloaded modules — detects tampered or corrupted dependency downloads. Adds 2 seconds to CI. | 5 min | Immediate |
| 2 | **govulncheck in CI** | Scans dependencies for known CVEs. Essential for a security-auditing tool. `govulncheck ./...` runs in < 30s. | 30 min | High |
| 3 | **SLSA provenance generation** | Generates build provenance (SLSA Level 3) via GoReleaser, proving the binary was built by the official CI pipeline. Complements artifact signing (P6-005). | 1 h | Medium |
| 4 | **SBOM generation** | Software Bill of Materials via GoReleaser's `sboms` section. Users can audit what's inside the binary. | 30 min | Medium |
| 5 | **Release checklist / runbook** | Documented steps for cutting a release (tag, verify CI, review draft, publish). Prevents forgotten steps. | 1 h | Medium |
| 6 | **Install script smoke tests in CI** | `scripts_test.go` validates script content but doesn't execute the scripts. A CI job that actually runs `install.sh` and `install.ps1` on their respective platforms would catch regressions. | 2 h | Low |
| 7 | **Matrix test for go.mod minimum Go version** | CI tests with the minimum Go version declared in `go.mod` (not just latest). Ensures backward compatibility. | 30 min | Low |
| 8 | **Nightly CI run** | Scheduled CI on `main` to detect bit-rot from external dependency changes. Currently CI only runs on push/PR. | 15 min | Low |

## 6. Task Plan

### Immediate (this sprint — 0–3 days)

| Task | Finding | Effort | Priority |
|------|---------|--------|----------|
| Fix CI Go version to match go.mod | P6-001 | 5 min | P0 |
| Add `go-version-file: go.mod` to ci.yml and action.yml | P6-006 | 10 min | P0 |
| Add `golangci-lint` step to CI (ubuntu only) | P6-002 | 30 min | P1 |
| Add `-coverprofile` and coverage upload to CI | P6-003 | 30 min | P1 |
| Create `.github/dependabot.yml` | P6-004 | 15 min | P1 |
| Add dynamic CI badge to README | P6-007 | 5 min | P2 |
| Add `go mod verify` step to CI | Missing #1 | 5 min | P1 |

### Short-term (next sprint — 1–2 weeks)

| Task | Finding | Effort | Priority |
|------|---------|--------|----------|
| Add cosign keyless signing to GoReleaser | P6-005 | 2 h | P2 |
| Add `govulncheck` step to CI | Missing #2 | 30 min | P2 |
| Resolve CHANGELOG.md vs GoReleaser dual source | P6-008 | 1 h | P2 |
| Add coverage threshold enforcement | P6-003 | 30 min | P2 |
| Document release checklist | Missing #5 | 1 h | P3 |

### Long-term (backlog)

| Task | Finding | Effort | Priority |
|------|---------|--------|----------|
| Add SLSA provenance to GoReleaser | Missing #3 | 1 h | P3 |
| Add SBOM generation to GoReleaser | Missing #4 | 30 min | P3 |
| Add install script execution tests to CI | Missing #6 | 2 h | P4 |
| Add minimum-Go-version test to matrix | Missing #7 | 30 min | P4 |
| Add nightly CI schedule | Missing #8 | 15 min | P4 |

## 7. Implementation Order

```
P6-001 (Go version fix)
├──► P6-006 (go-version-file: eliminate all hardcoded versions)
│    └──► P6-002 (lint in CI) — independent, can run in parallel
│         └──► P6-003 (coverage) — independent, can run in parallel
│              └──► P6-004 (Dependabot) — independent
│                   └──► P6-007 (CI badge) — cosmetic, last
│
Missing #1 (go mod verify) — add alongside P6-001
Missing #2 (govulncheck) — add after P6-002 (same CI step style)

P6-005 (artifact signing) — requires release workflow change, test with pre-release
└──► Missing #3 (SLSA) — builds on signing infrastructure

P6-008 (changelog) — process decision, no dependencies
```

## 8. Risks/Blockers

| Risk | Probability | Mitigation |
|------|-----------|------------|
| `golangci-lint` reveals many pre-existing violations | High | Run once locally, fix all violations in a preparatory PR before adding CI enforcement |
| Coverage threshold too aggressive | Medium | Measure current coverage first; set threshold at current value minus 2% to allow small fluctuations |
| cosign OIDC setup requires `id-token: write` permission | Low | Documented in GoReleaser docs; add to `release.yml` permissions block |
| Dependabot PR flood on first enablement | Medium | Set `open-pull-requests-limit: 3` initially; increase after backlog clears |
| `go-version-file` breaks if go.mod is moved | Very Low | go.mod is in repo root and unlikely to move |
| Windows race detector skip | Known (not a risk) | Go race detector does not fully support Windows; skip is expected and documented |

## 9. Phase Close Checklist

- [x] 8 findings documented with evidence, impact, and recommendations
- [x] Severity calibrated to CLI tool context (no server surface, focus on build/release pipeline)
- [x] No fabricated findings — all traceable to specific file:line references
- [x] Findings ordered by severity (CRÍTICO → RIESGO → ATENCIÓN)
- [x] Section 5 identifies 8 missing operations practices with effort estimates
- [x] Task plan with 7 immediate, 5 short-term, 5 long-term tasks
- [x] Implementation order diagram showing dependency chain
- [x] Risks/blockers table covering cross-platform and toolchain concerns
- [x] Install scripts rated: production-quality (both install.sh and install.ps1)
- [x] Version embedding rated: robust (multi-layered fallback)
- [x] GoReleaser config rated: comprehensive (builds, archives, checksums, brews, scoops, changelog)

---

*Report generated by P6 sequoia-operations · Sequoia v0.1.0*
