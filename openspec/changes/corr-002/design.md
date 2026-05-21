# Design: CI Gates Hardening (CORR-002)

## Technical Approach

Split the monolithic CI job into 5 phased jobs with `needs`-based gating. Vulncheck becomes blocking (removing `continue-on-error: true`) and is added to the release pipeline. Coverage is collected via `-coverprofile` with a 70% threshold enforced by post-processing `go tool cover -func`. Artifacts flow via `upload-artifact@v4`/`download-artifact@v4`.

## Architecture Decisions

| Decision | Choice | Tradeoff | Rationale |
|----------|--------|----------|-----------|
| Vulncheck placement | Separate `vulncheck` job (ubuntu only) | +1 job in graph vs. cleaner separation | Must block pipeline. Putting it in `lint` means lint failure wouldn't gate test (spec says test runs regardless of lint outcome). Separate job with `test` depending on it ensures CVE blocks everything. |
| Test depends on vulncheck | `needs: [vulncheck]` (default: must succeed) | vulncheck failure blocks test/build/smoke | Matches spec: "CI pipeline MUST fail when govulncheck discovers CVEs" |
| Lint independence | No `needs`, runs parallel to vulncheck+test | Lint failures visible but don't block test | Spec: "lint job has no upstream dependencies" and "test runs regardless of lint outcome" |
| Build depends on test | `needs: [test]` (must succeed all matrix entries) | All 5 platforms must pass before any build | Prevents shipping a binary that failed tests on any platform |
| Smoke platform | Ubuntu only (single job after build) | Loses cross-platform smoke, saves 4 runners | Binary behavior is platform-independent; install/uninstall pattern is same |
| Coverage threshold parser | `go tool cover -func \| tail -1 \| grep -oP '\d+\.\d+'` | Requires bash; not Windows-native | CI test runs on all platforms; enforcement runs as post-test bash step |
| Windows coverage flags | `-coverprofile` only (no `-covermode=atomic`) | No race-safe coverage on Windows | Spec requirement; `-covermode=atomic` is not supported on Windows |

## Job Dependency Graph

```
vulncheck (ubuntu) ──┐
                     ├──> test (matrix: ubuntu, macos, macos-14, ubuntu-arm, windows)
lint (ubuntu) ───────┘         │
                               ├──> build (matrix: same 5 OS)
                               │         │
                               │         └──> smoke (ubuntu)
                               └──> (coverage check per matrix entry)
```

**Key**: `vulncheck` failure → blocks `test` → blocks `build` → blocks `smoke`. `lint` runs independently — its failure marks the check red but does not stop the pipeline.

## Data / Artifact Flow

```
test (each OS)
  │  go test -coverprofile=coverage.out
  │  go tool cover -func → parse total → fail if <70%
  │  upload: coverage.out (optional, for inspection)
  └──> build (each OS)
         │  go build -o sequoia[.exe] ./cmd/sequoia/
         │  upload-artifact@v4: sequoia-${{ matrix.os }}
         └──> smoke (ubuntu)
                download-artifact@v4: sequoia-ubuntu-latest
                ./sequoia install --no-tui → status → uninstall --all --yes
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `.github/workflows/ci.yml` | Rewrite | Split into 5 jobs: `vulncheck`, `lint`, `test`, `build`, `smoke`. Add coverage flags + threshold check. Remove `continue-on-error: true`. |
| `.github/workflows/release.yml` | Modify | Add `vulncheck` job before `goreleaser` (with `needs: [test, vulncheck]` on goreleaser, OR add vulncheck step inside `test` job). |
| `.gitignore` | Verify-only | Lines 8-14 already cover `coverage`, `coverage_*`, `coverage*.out`. No change needed. |
| `coverage`, `coverage_rc002` | No action | Already untracked (`git ls-files` returns empty). Files exist on disk (94KB / 12 bytes) but are ignored by `.gitignore`. |

## Release Pipeline Vulncheck

Current: `test` → `goreleaser` (2 jobs). After: add vulncheck step inside the existing `test` job (avoids a 3rd job with another checkout+setup). The `test` job already runs `go test`; adding `go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...` before or after the test step is zero-cost.

Alternatively, a separate `vulncheck` job with `goreleaser` depending on both `test` and `vulncheck`. Simpler to modify the existing `test` job since it already has Go setup and checkout.

## Testing Strategy

| Layer | What | How |
|-------|------|-----|
| YAML validation | ci.yml, release.yml syntax | `yamllint` or GitHub Actions VS Code extension |
| Workflow validation | Job graph correctness | Manual push to test branch, observe GitHub Actions UI |
| Coverage threshold | 70% gate logic | Run locally: `go tool cover -func \| tail -1` → verify ≥84.1% (current baseline) passes |
| Artifact flow | upload/download chain | Push to test branch, verify build artifact URL and smoke download |
| Vulncheck blocking | CVE detection gates pipeline | Temporarily introduce known-vulnerable dep, verify CI fails |

## Migration / Rollout

No migration required. CI config changes apply immediately on push to `main`. Rollback: revert `ci.yml` and `release.yml` to current state in a single commit. No database or infra dependencies.

## Open Questions

None. All design decisions resolved.
