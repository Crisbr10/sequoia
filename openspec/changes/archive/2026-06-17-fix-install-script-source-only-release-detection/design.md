# Design: Fix Install Script Source-Only Release Detection

## Technical Approach

Surgical diagnostics-only change. On HTTP 404 from the binary download, the installer queries `https://api.github.com/repos/$Repo/releases/tags/$ResolvedVersion`. API 200 → emit `source-only release` error with remediation. API failure → fall back to existing generic `Download failed. Please check:` message. Happy path is unchanged. The marker `source-only release` is the regression-test anchor (spec REQ-IER-01 / REQ-IER-02).

## Architecture Decisions

| Decision | Choice | Trade-off | Rationale |
|----------|--------|-----------|-----------|
| Trigger condition | HTTP 404 on binary download only | Bare `curl -fsSL` conflates 404 with network errors | Source-only releases are 404 by definition; others keep generic per REQ-IER-04 |
| API to query | `releases/tags/$ResolvedVersion` (not `latest`) | Extra endpoint vs `latest` | Exact tag user requested is authoritative |
| On API failure | Fall back to generic error | User may miss hint on rate-limit | REQ-IER-01 mandates fall-back |
| Retry on API lookup | None | User already waited 14s for binary retries | Avoid adding 14 more seconds to a failed install |
| Timeouts | 10s (`-TimeoutSec` / `--max-time`) | Long for slow nets, short not to hang | "Fail fast on error path" |
| Marker string | `source-only release` (literal) | User-visible text is test-coupled | Spec REQ-IER-01 / REQ-IER-02 mandates this exact substring |

## Data Flow

```
HTTP 404 from download ──► GET /releases/tags/$VERSION
                              ├─► 200 ──► "source-only release" error + exit 3
                              └─► fail ──► generic error + exit 3
non-404 from download ──► generic error + exit 3 (unchanged)
success ──► continue (happy path unchanged)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `scripts/install.ps1` | Modify | Replace catch block at lines 198–207 with 404-aware variant |
| `scripts/install.sh` | Modify | Extend both `curl` (line 274) and `wget` (line 282) error branches |
| `scripts/install_scripts_test.go` | Modify | Add 2 new test functions |

No changes to `adapters/`, `internal/`, `cmd/`, `plugin/`, workflows, goreleaser, or any other Go file.

## Implementation Sketch

User-visible error (locked by REQ-IER-03):

> `Release $ResolvedVersion exists on GitHub but no precompiled binary asset was found. This usually means the release was published without GoReleaser. Try installing a previous version or report this at https://github.com/Crisbr10/sequoia/issues.`

PowerShell (PS 5.1 `$_.Exception.Response` may be null → string-match fallback):

```powershell
} catch {
    $is404 = ($_.Exception.Response.StatusCode.value__ -eq 404) -or ($_ -match '404|Not Found')
    if ($is404) {
        $api = Invoke-WebRequest -Uri "https://api.github.com/repos/$Repo/releases/tags/$ResolvedVersion" -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
        if ($api.StatusCode -eq 200) { Write-Err "Release $ResolvedVersion exists on GitHub but no precompiled binary asset was found."; ... ; exit $EXIT_NETWORK }
    }
    Write-Err "Download failed. Please check:"; ...   # unchanged
    exit $EXIT_NETWORK
}
```

Bash (both branches; curl 22 = 404, wget 8 = server-error):

```bash
if [ $EXIT -eq 22 ] || [ $EXIT -eq 8 ]; then
    if curl -fsSL --max-time 10 "https://api.github.com/repos/${REPO}/releases/tags/${VERSION}" >/dev/null 2>&1; then
        log_error "Release ${VERSION} exists on GitHub but no precompiled binary asset was found."; ... ; exit $EXIT_NETWORK
    fi
fi
log_error "Download failed. Please check:"; ...   # unchanged
exit $EXIT_NETWORK
```

## Testing Strategy

Strict TDD: tests written before code. Same pattern as `TestInstallPs1ChecksumMandatory` (grep-based).

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `install.ps1` has source-only-release branch | Grep marker substrings |
| Unit | `install.sh` has branch in **both** curl + wget | Grep; verify API URL appears ≥2 times |
| Regression | All existing tests pass | `go test ./scripts/... -count=1` |

New test functions:

- `TestInstallPs1SourceOnlyReleaseDetection` — asserts `source-only release`, `releases/tags/$ResolvedVersion` (or equivalent), `previous version`, `https://github.com/Crisbr10/sequoia/issues`.
- `TestInstallShSourceOnlyReleaseDetection` — same four assertions; asserts API URL appears ≥2 times (both branches).

Uses existing `os.ReadFile`, `assert.Contains`, `assert.Regexp`. The `contains` helper from `goreleaser_config_test.go` (same package) is already available.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| GitHub API rate-limit (60/h unauthenticated) on error path | Low | Fall back to generic per spec |
| Extra round-trip on failure path | Low | ≤10s timeout; happy path unchanged |
| PS 5.1 `$_.Exception.Response` may be null | Medium | String-match `404`/`Not Found` as secondary signal |
| Test grep brittleness to whitespace | Low | Assert substring presence, not exact line |

## Review Workload Forecast

| File | Changed lines |
|------|---------------|
| `scripts/install.ps1` | ~28 |
| `scripts/install.sh` | ~68 |
| `scripts/install_scripts_test.go` | ~50 |
| **Total** | **~146** |

- `fits_400_budget`: **true** (37% of budget)
- `chained_prs_recommended`: **false** — single PR
- `Decision needed before apply`: **No**
- `400-line budget risk`: **Low**

## Open Questions

None. All blockers resolved in spec REQ-IER-01..04.