# Tasks: RC-003 — Add CI/CD Security Gates

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~132 across 6 files |
| 400-line budget risk | Low |
| Chained PRs recommended | Yes |
| Suggested split | PR1 → PR2 → PR3 → PR4 |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | PR | Notes |
|------|------|-----|-------|
| 1 | Foundation — gosec + govulncheck + Dependabot | PR 1 | ~35 lines, 3 files, no pipeline disruption |
| 2 | CI Matrix — ARM64 runners | PR 2 | ~10 lines, 1 file, stacks on PR 1 |
| 3 | Release hardening — test gate + SBOM | PR 3 | ~22 lines, 1 file, independent |
| 4 | Installer signatures — cosign verify-blob | PR 4 | ~75 lines, 2 files, independent |

## Phase 1: Foundation (PR 1 — gosec, govulncheck, Dependabot)

- [x] T-001: `.golangci.yaml` — add `gosec` to `linters.enable`. Add exclude-rules for G104 (`cmd/sequoia/main\.go`) and G304 (`adapters/`). **Verify**: `golangci-lint run` passes with zero blocking issues.
- [x] T-002: `.github/workflows/ci.yml` — insert `govulncheck` step after `Setup Go`: `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` with `if: runner.os == 'Linux'`. **Verify**: push branch, govulncheck runs only on Linux.
- [x] T-003: `.github/dependabot.yml` — add `groups` (gomod minor+patch, github-actions), `reviewers: ["Crisbr10"]`, `labels: ["dependencies"]` to both entries. **Verify**: next Dependabot PR groups updates + assigns reviewer.

## Phase 2: CI Matrix (PR 2 — ARM64 runners)

- [x] T-004: `.github/workflows/ci.yml` — add `macos-14`, `ubuntu-24.04-arm` to `strategy.matrix.os`. Confirm `fail-fast: false` remains. **Verify**: push branch, both ARM jobs pass in CI.

## Phase 3: Release Hardening (PR 3 — test gate + SBOM)

- [x] T-005: `.github/workflows/release.yml` — add `test` job (ubuntu-latest: checkout + setup-go@v6 go-1.24 + `go test -race -count=1 ./...`). **Verify**: push tag with failing test → GoReleaser blocked.
- [x] T-006: `.github/workflows/release.yml` — add `needs: test` to `goreleaser`. Insert SBOM step after GoReleaser: `anchore/sbom-action@v0` with `path: ./`, `format: spdx-json,cyclonedx-json`, `upload-release-assets: true`. **Verify**: passing release → .spdx.json + .cdx.json attached.

## Phase 4: Installer Signatures (PR 4 — cosign verify-blob)

- [x] T-007: `scripts/install.sh` — insert cosign block (~47 lines) between checksum and extraction. Detect cosign via `command -v`, download `.sig`/`.cert` from release, run `cosign verify-blob`, warn on failure/absence. Cosign absence NEVER exits error. **Verify**: 4 scenarios — cosign+sig OK, sig missing, cosign absent, invalid sig (only last exits error with EXIT_CHECKSUM=2).
- [x] T-008: `scripts/install.ps1` — insert cosign block (~58 lines) between checksum and extraction. PowerShell: `Get-Command cosign`, `Invoke-WebRequest` downloads, `& cosign verify-blob`. Same fallback contract. Exit code: EXIT_CHECKSUM=2 on sig failure. **Verify**: same 4 scenarios in PowerShell. Syntax confirmed valid via PSParser.
