# Contributing to Sequoia

## Action Pinning Policy

All GitHub Actions in `.github/workflows/` MUST reference full, immutable commit SHAs.
Floating refs (`@vX`, `@main`, `@master`) are **prohibited**.

### Why

Floating tags (`@v4`, `@v6`) can be retagged by their maintainers to point to new commits
at any time. A supply-chain attacker who compromises an action's repository can retag a
major version tag to a malicious commit, executing arbitrary code in your CI pipeline.

Commit SHAs are immutable — once published, they cannot be changed.

### Rules

- Every `uses:` directive MUST use `owner/repo@<40-char-sha>` format
- Include a `# vX` comment to document which major version the SHA resolves to
- Example: `uses: actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd  # v6`
- The `action-pinning` CI check scans all workflow files and fails on floating refs

### Updating Actions

When Dependabot or manual review identifies an outdated action:

1. Look up the latest commit SHA for the major version tag
   via `https://api.github.com/repos/{owner}/{repo}/git/ref/tags/v{major}`
2. For annotated tags, resolve the underlying commit SHA
   via `https://api.github.com/repos/{owner}/{repo}/git/tags/{tag-sha}`
3. Update the SHA in the workflow file, keeping the `# vX` comment
4. Push the change through a PR; the `action-pinning` CI check verifies correctness

## Release Process

### Prerequisites

The `release` GitHub Environment must exist with required reviewers configured.
To create it:

1. Go to **Settings → Environments** in the repository
2. Click **New Environment**, name it `release`
3. Under **Deployment protection rules**, add required reviewers (at least one)
4. Optionally configure a wait timer
5. Save

### Triggering a Release

Releases are triggered by pushing a semver tag or via manual dispatch:

**Tag push (automatic)**:
```bash
git tag v0.2.0
git push origin v0.2.0
```

**Manual dispatch** (from Actions UI):
1. Go to **Actions → Release**
2. Click **Run workflow**
3. Enter a version tag (e.g., `v0.2.0`), or leave empty to use the latest tag
4. Optionally check **Skip GoReleaser publish** for dry-run (build, verify, sign only)
5. Click **Run workflow**

### Release Pipeline

1. **Pre-release Tests** — Cross-platform matrix: `ubuntu-latest`, `macos-latest`, `windows-latest`.
   Runs `go test -race -count=1 ./...` on each OS. Any failure aborts.

2. **GoReleaser** (requires manual approval via `environment: release`)
   - GoReleaser builds binaries (snapshot, no publish)
   - SHA-256 checksums are verified against `checksums.txt`
   - Cosign keyless signing (OIDC) for all binaries
   - GoReleaser publishes to GitHub Releases, Homebrew tap, and Scoop bucket

3. **Post-Deploy Smoke** — Downloads the published Linux AMD64 binary from the
   GitHub Release, verifies SHA-256 against the published `checksums.txt`, and
   runs `sequoia version` + `sequoia status`.

### Code Ownership

Release-critical paths are protected by CODEOWNERS:
- `.github/workflows/release.yml`
- `scripts/install.ps1`
- `.github/CODEOWNERS` (self-referencing — prevents tampering)

All changes to these files require review from `@Crisbr10`.

### Installer Resilience

The Windows installer (`scripts/install.ps1`) uses exponential backoff retry
for all downloads:

- 3 attempts maximum
- Delays: 2 seconds, 4 seconds, 8 seconds
- Final failure exits with `$EXIT_NETWORK` (code 3)

This protects against transient network failures (timeouts, 5xx errors) during
binary and checksum downloads.
