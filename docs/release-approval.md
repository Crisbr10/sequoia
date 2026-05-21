# Release Approval Gate

**Every Sequoia release requires a human maintainer to explicitly approve the GoReleaser
publish step before binaries are built, signed, and shipped.** This document explains how
the gate works, who can approve, and what to expect when you push a tag.

## Quick Path

A release flow from tag to publish:

1. Push a semver tag → CI starts automatically
2. Pre-release tests run (vulncheck + cross-platform matrix) — no approval needed
3. GoReleaser job pauses at `environment: release` → waits for approval
4. A maintainer approves the deployment in the GitHub Actions UI
5. GoReleaser builds, signs (Cosign), and publishes binaries
6. Post-deploy smoke test downloads and verifies the published binary

## Gate Mechanism

| Component | What it does |
|-----------|-------------|
| `environment: release` (`.github/workflows/release.yml` line 63) | References the protected GitHub Environment on the `goreleaser` job |
| `release` GitHub Environment | Holds the deployment protection rules — required reviewers, wait timer |
| Required reviewers | Maintainers listed in the environment who must approve before the job executes |
| Wait timer (optional) | Additional delay after approval to catch mistakes before publish |

When a tag push or manual dispatch triggers the workflow, the `vulncheck` and `test` jobs
run immediately (they are NOT behind the approval gate). Only after both pass does the
`goreleaser` job queue — and **that** is where the gate applies.

## Who Can Approve

Approval is granted to **repository maintainers with write access**. The current
maintainers are listed in the `release` Environment under
**Settings → Environments → release → Deployment protection rules**.

Code ownership for release-critical files also requires `@Crisbr10` review for any
changes to `.github/workflows/release.yml`, `scripts/install.ps1`, or
`.github/CODEOWNERS` (see `docs/CONTRIBUTING.md`).

## How to Approve

When a release is waiting for approval:

1. Go to **Actions** tab in the repository
2. Select the running **Release** workflow
3. Click **Review deployments** on the `goreleaser` job
4. Select the `release` environment
5. Click **Approve and deploy**

The GoReleaser job will proceed immediately after approval (subject to any configured
wait timer).

## How to Request Approval

If you are not a maintainer but need a release:

1. Push the tag and note the workflow run URL
2. Open a discussion or Slack/message a maintainer with the run URL
3. The maintainer approves through the Actions UI (see above)
4. Monitor the workflow run for completion

## Step-by-Step: Full Release

### Prerequisites (one-time setup)

- [ ] `release` GitHub Environment created (Settings → Environments → New Environment)
- [ ] Required reviewers added (at least 1)
- [ ] Optional wait timer configured
- [ ] `HOMEBREW_TAP_TOKEN` secret set (for Homebrew tap publish)
- [ ] CODEOWNERS covers release-critical files

### Performing a Release

**Option A — Tag push (automatic)**:
```bash
git tag v0.2.0
git push origin v0.2.0
```

**Option B — Manual dispatch** (Actions UI → Release → Run workflow):
- Enter a version tag (e.g., `v0.2.0`) or leave empty for latest
- Optionally check **Skip GoReleaser publish** for dry-run mode

### What Happens After Trigger

| Phase | Job | Gate? | Duration (approx.) |
|-------|-----|-------|--------------------|
| 1. Security | `vulncheck` | None | ~30s |
| 2. Tests | `test` (3 OS matrix) | None | ~3min |
| 3. **Approval gate** | `goreleaser` | **Yes — wait for maintainer** | Variable |
| 4. Build + Sign | `goreleaser` steps | None (after approval) | ~3min |
| 5. Smoke test | `smoke` | None | ~30s |

The approval gate sits between phase 2 (tests pass) and phase 4 (build/sign/publish).
The job will appear as **"Waiting for deployment approval"** in the Actions UI until a
maintainer acts.

## Configuring the Release Environment

For new maintainers or repository setup:

### Step 1: Create the Environment

1. Go to **Settings → Environments** (left sidebar)
2. Click **New Environment**
3. Name it exactly `release`
4. Click **Configure environment**

### Step 2: Add Deployment Protection Rules

Under **Deployment protection rules**:

- [ ] **Required reviewers**: Add at least 1 reviewer. Check **Prevent self-review** to
  require a second pair of eyes.
- [ ] **Wait timer** (optional): Set to e.g., 5 minutes. This adds a delay after
  approval, giving time to cancel if something was missed.

### Step 3: Save

The workflow already references `environment: release` — no workflow changes needed.

## Dry-Run Mode (No Publish)

For testing the pipeline without publishing:

1. Trigger via manual dispatch (`workflow_dispatch`)
2. Check **Skip GoReleaser publish**
3. The pipeline will build, sign, and verify — but skip the final `goreleaser release --clean` step

This still requires approval (the gate applies to the entire `goreleaser` job). The dry
run is useful for verifying the signing and build steps before a real release.

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| Workflow doesn't start on tag push | Tag doesn't match `v[0-9]+.[0-9]+.[0-9]+` pattern | Use semver: `v1.2.3` |
| `goreleaser` job stuck "waiting" | No maintainer has approved yet | Notify a maintainer with the run URL |
| Approval button not visible | You are not in the required reviewers list | Ask a listed maintainer to approve |
| GoReleaser fails after approval | Secret missing or goreleaser config wrong | Check `HOMEBREW_TAP_TOKEN` and `.goreleaser.yaml` |
| Smoke test fails | Published binary corrupted or wrong version | Re-run; if persistent, check the release assets manually |

## Related Documents

- `docs/CONTRIBUTING.md` — Release process overview and action pinning policy
- `.github/workflows/release.yml` — The workflow definition
- `.github/CODEOWNERS` — Release-critical file ownership
