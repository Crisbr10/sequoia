# Design: RC-003 — Add CI/CD Security Gates

## Technical Approach

Insert 5 independent security gates into the CI/release pipeline: (1) `govulncheck` in ci.yml on every push, (2) `gosec` in `.golangci.yaml` for code-pattern scanning, (3) ARM64 runners in the CI matrix, (4) SBOM generation in `release.yml` attached to GitHub Releases, (5) cosign signature verification in installer scripts with graceful fallback. Plus Dependabot grouping. All changes are additive single-file modifications. No application code changes. Each component independently revertable.

## Architecture Decisions

| Decision | Options | Choice | Rationale |
|----------|---------|--------|-----------|
| SBOM tool | anchore/sbom-action@v0 vs cyclonedx-gomod | **anchore/sbom-action@v0** | Runs natively on GHA (no Docker). `upload-release-assets: true` attaches SPDX+CycloneDX directly to the release GoReleaser creates. Already adopted by 2k+ repos. cyclonedx-gomod requires Docker or extra plugin setup for release attachment. |
| govulncheck timing | Every push vs schedule-only vs release-only | **Every push/PR** | ~30s CI overhead is negligible. Schedule-only leaves a detection gap between runs. Release-only is too late — the code has already shipped. Running only on Linux avoids triple-execution across the matrix. |
| ARM64 runners | macos-14 only vs +ubuntu-24.04-arm | **Both** | `macos-14` covers Apple Silicon (primary target). `ubuntu-24.04-arm` covers Linux ARM servers (growing market for CLI tools). `fail-fast: false` already mitigates runner flakiness. Both are free for public repos. |
| gosec severity threshold | Fail on medium+ vs high+ only | **High+ only** | `govulncheck` already catches known CVEs. gosec adds code-pattern scanning (G101 hardcoded secrets, G110 decompression bombs). High-severity rules have low false-positive rates. Medium adds noise (G104, G304) that exclusion rules alone may not fully suppress. |
| Cosign fallback behavior | Silent skip vs warning vs error exit | **Warning + continue** | Mirrors existing `SKIP_CHECKSUMS` pattern: inform the user, recommend cosign installation, continue with SHA-256. Blocking the installer because cosign is absent would break non-technical users. SHA-256 is the primary integrity check; cosign adds cryptographic provenance on top. |

## Data Flow

### CI Pipeline (after)

```
Checkout → Setup Go → govulncheck (Linux only) → Vet → Lint (Linux) → Test (all OS/arch) → Upload coverage → Build → Smoke tests
```

`govulncheck` inserts between Setup Go and Vet — it only needs the Go toolchain and source. Running on Linux only avoids triple-execution; ARM64 runners still get the `go vet` + `go test` coverage.

### Release Pipeline (after)

```
Tag push
  ├─ test job (ubuntu-latest)           ← NEW pre-release gate
  │    ├─ Checkout → Setup Go → go test ./... -count=1 -race
  │    └─ PASS ✓
  │
  └─ goreleaser job (needs: test)       ← dependency ensures tests pass first
       ├─ Checkout → Setup Go → Install Cosign → GoReleaser (builds + signs + releases)
       └─ Generate SBOM (anchore/sbom-action)  ← NEW, attaches .spdx.json + .cdx.json to release
```

### Installer Signature Flow (install.sh / install.ps1)

```
Download tarball → SHA-256 checksum verification → [NEW] cosign verify-blob attempt:
  ├─ cosign on PATH? → Download .sig + .cert → verify-blob → PASS: log_info / FAIL: log_warn
  └─ cosign absent? → log_warn "Install cosign for cryptographic verification"
→ Extract → Install
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `.github/workflows/ci.yml` | Modify | Add `macos-14`, `ubuntu-24.04-arm` to `os` matrix. Insert `govulncheck` step after Setup Go (Linux-only, `if: runner.os == 'Linux'`) |
| `.github/workflows/release.yml` | Modify | Add `test` job (checkout + setup-go + `go test`) with `needs: test` on goreleaser. Add SBOM step after GoReleaser using `anchore/sbom-action@v0` |
| `.github/dependabot.yml` | Modify | Add `groups` (gomod minor/patch grouped, actions grouped), `reviewers: ["Crisbr10"]`, `labels: ["dependencies"]` |
| `.golangci.yaml` | Modify | Add `gosec` to `enable` list. Add `exclude-rules` entries for G104 (unchecked err in main.go, matches existing errcheck exclusion) and G304 (filepath.Join patterns already sanitized) |
| `scripts/install.sh` | Modify | Add cosign verification block (~35 lines) between checksum verification and extraction. Detect cosign, download .sig/.cert, attempt `verify-blob`, warn on failure/absence |
| `scripts/install.ps1` | Modify | Add cosign verification block (~40 lines) between checksum verification and extraction. PowerShell-native equivalent of install.sh logic |

## Key Implementation Details

### ci.yml — govulncheck step

```yaml
- name: Vulncheck
  if: runner.os == 'Linux'
  run: go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Position: after Setup Go (needs go binary + go.sum), before Vet (fast-fail if CVEs exist). Linux-only: avoids redundant scans on macOS/Windows. Uses `@latest` to always scan with the freshest vulnerability DB.

### release.yml — pre-release test job

```yaml
test:
  name: Pre-release tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v6
      with: { go-version: "1.24", cache: true }
    - run: go test -race -count=1 ./...

goreleaser:
  needs: test   # ← added dependency
  # ... rest unchanged
```

### release.yml — SBOM step (inside goreleaser job, after GoReleaser)

```yaml
- name: Generate SBOM
  uses: anchore/sbom-action@v0
  with:
    path: ./
    format: spdx-json,cyclonedx-json
    upload-release-assets: true
```

Permissions: `contents: write` already present — SBOM attaches as release asset to the draft release GoReleaser created.

### .golangci.yaml — gosec with exclusions

Exclusions mirror the existing `errcheck` exclusion for `cmd/sequoia/main.go`:

```yaml
- path: cmd/sequoia/main\.go
  linters: [gosec]
  text: "G104:"   # unchecked errors in audit entrypoint
- path: adapters/
  linters: [gosec]
  text: "G304:"   # filepath.Join paths are sanitized before use
```

### install.sh / install.ps1 — cosign verification

Pattern mirrors existing checksum verification: detect tool → attempt download → verify → warn on failure. No new exit codes. The `SKIP_CHECKSUMS` flag does NOT skip cosign (cosign is always attempted when available — it's additive, not blocking).

Key variables (install.sh):
- `SIG_URL`: `https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}.sig`
- `CERT_URL`: `https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}.cert`

Command: `cosign verify-blob --signature "$SIG_FILE" --certificate "$CERT_FILE" "${TMPDIR}/${TARBALL}"`

The `.sig` and `.cert` files are produced by GoReleaser's `signs` block (already configured with keyless cosign signing via GitHub OIDC).

PowerShell equivalent uses `Get-Command cosign` for detection, `Invoke-WebRequest` for downloads, and `& cosign verify-blob` for verification.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| CI config | Matrix expansion, govulncheck step | Push to branch → verify CI runs on all 5 OS entries, govulncheck runs only on Linux |
| Release config | Pre-release gate blocks GoReleaser on test failure | Push a failing test → verify GoReleaser skipped. Fix test → verify GoReleaser runs |
| Dependabot | Grouped PRs | Verify next Dependabot PR groups gomod updates under one PR |
| gosec | No false-positive failures | Run `golangci-lint run` locally → verify zero blocking issues |
| install.sh | Cosign present/absent, sig available/missing | 4 scenarios: cosign+files OK, cosign+sig missing, cosign absent, sig OK but verify fails — all must NOT exit error, only warn |
| install.ps1 | Same 4 scenarios | PowerShell-specific implementation, same expected behavior |

## Rollback Plan

Each component is a single-file revert, independent of others:

| Component | Rollback |
|-----------|----------|
| govulncheck in ci.yml | Remove the `Vulncheck` step block |
| ARM64 runners | Remove `macos-14` and `ubuntu-24.04-arm` from matrix list |
| Pre-release test gate | Remove `test` job + remove `needs: test` from goreleaser |
| SBOM generation | Remove `Generate SBOM` step from release.yml |
| Dependabot groups | Remove `groups`, `reviewers`, `labels` blocks |
| gosec linter | Remove `gosec` from enable list + remove gosec exclude-rules |
| Cosign in installers | Remove the cosign verification block (between checksum and extraction) |

No migration required. No data schema changes. No breaking changes to existing workflows or user experience.

## Open Questions

- [ ] Verify `anchore/sbom-action@v0` `upload-release-assets` works with draft releases (GoReleaser creates drafts). If it requires published releases, switch to `upload-artifact-retention` + manual upload step.
- [ ] Confirm `ubuntu-24.04-arm` runner is GA for this repo (it was announced GA in Jan 2025 for public repos, but org-level enablement may be needed).
