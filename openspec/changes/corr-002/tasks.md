# Tasks: CI Gates Hardening (CORR-002)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~130 (ci.yml: ~80 add + ~45 del, release.yml: ~6 add) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Verify Existing Cleanup (pre-flight checks)

- [x] 1.1 Run `git ls-files coverage coverage_rc002 coverage*.out` → confirm empty output (files already untracked per design)
- [x] 1.2 Confirm `.gitignore` lines 8-14 cover `coverage`, `coverage_*`, `coverage*.out`, `*.cover` → no changes needed

## Phase 2: Release Pipeline Vulncheck (independent, deployable alone)

- [x] 2.1 Add `govulncheck` step to release.yml `test` job after `Setup Go` step (zero-cost: reuses checkout+setup)
- [x] 2.2 Validate release.yml YAML syntax: `yamllint .github/workflows/release.yml` or GH Actions VS Code extension
- [x] 2.3 Commit: `ci(release): add blocking vulncheck before goreleaser` with work-unit message

## Phase 3: CI Pipeline Rewrite (monolithic → 5 phased jobs)

- [x] 3.1 Define `vulncheck` job (ubuntu, checkout→setup-go→`govulncheck`, no `continue-on-error`)
- [x] 3.2 Define `lint` job (ubuntu, independent/no `needs`, checkout→setup-go→`go vet`→`golangci-lint`)
- [x] 3.3 Define `test` job matrix (5 OS, `needs: [vulncheck]`, add `-coverprofile=coverage.out`, `-covermode=atomic` on !windows)
- [x] 3.4 Add coverage threshold step in `test`: parse `go tool cover -func coverage.out | tail -1`, fail if <70%
- [x] 3.5 Define `build` job matrix (5 OS, `needs: [test]`, build binary, `upload-artifact@v4` as `sequoia-${{ matrix.os }}`)
- [x] 3.6 Define `smoke` job (ubuntu, `needs: [build]`, `download-artifact@v4`, install→status→uninstall smoke checks)
- [x] 3.7 Validate ci.yml YAML syntax and job graph (lint→parallel, vulncheck→test→build→smoke dependency chain)
- [x] 3.8 Commit: `ci: split into phased jobs with coverage gate and blocking vulncheck`

## Phase 4: Integration Verification

- [ ] 4.1 Push to test branch → observe GitHub Actions workflow run: verify all 5 jobs execute in correct dependency order
- [ ] 4.2 Confirm `vulncheck` failure blocks `test`/`build`/`smoke` (temporarily introduce known-CVE dep)
- [ ] 4.3 Confirm `lint` failure does NOT block `test`/`build`/`smoke` (introduce lint error, observe pipeline)
- [ ] 4.4 Verify artifact flow: build artifact URL exists → smoke job downloads and runs binary successfully
- [ ] 4.5 Verify coverage gate: current ~84% baseline passes; deliberately lower threshold to 100% to confirm fail path
- [ ] 4.6 Merge to main after all gates verified green
