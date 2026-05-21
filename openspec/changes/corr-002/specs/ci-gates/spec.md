# ci-gates Specification

## Purpose

Defines the CI pipeline gates for the sequoia-ai Go project: blocking vulnerability checks, coverage enforcement, stale artifact removal, and a phased job structure that prevents masking failures.

## Requirements

### Requirement: Blocking Vulnerability Check

The CI pipeline and release workflow MUST fail when `govulncheck` discovers known CVEs in Go dependencies. The release pipeline SHALL run vulncheck before publishing artifacts.

#### Scenario: PR vulncheck blocks merge on CVE

- GIVEN a pull request against `main` introduces a dependency with a known CVE
- WHEN CI runs the vulncheck step
- THEN the vulncheck exits non-zero
- AND the CI job fails, preventing merge

#### Scenario: Release vulncheck blocks publish on CVE

- GIVEN a version tag triggers the release workflow
- WHEN the release pipeline runs vulncheck before GoReleaser
- THEN vulncheck exits non-zero on CVE discovery
- AND GoReleaser does not execute

#### Scenario: Clean dependency tree passes vulncheck

- GIVEN a pull request with no known CVEs in its dependency tree
- WHEN CI runs the vulncheck step
- THEN vulncheck exits zero
- AND the pipeline proceeds to subsequent jobs

---

### Requirement: Coverage Collection and Enforcement

The CI pipeline MUST collect Go test coverage via `-coverprofile` and SHALL fail if total coverage drops below 70%. Windows runners SHOULD omit `-covermode=atomic`.

#### Scenario: Coverage above threshold passes

- GIVEN the test suite produces coverage ≥70%
- WHEN the coverage enforcement step parses `go tool cover -func` total
- THEN the step exits zero
- AND the pipeline proceeds

#### Scenario: Coverage below threshold fails CI

- GIVEN the test suite produces coverage <70%
- WHEN the coverage enforcement step parses the total percentage
- THEN the step exits non-zero
- AND the CI job fails with a message indicating the threshold

#### Scenario: Coverage threshold enforced on all platforms

- GIVEN the test matrix runs on ubuntu, macos, and windows runners
- WHEN each runner executes `go test -coverprofile`
- THEN coverage is collected on all platforms
- AND the threshold is enforced on all platforms

#### Scenario: Windows runner excludes atomic mode

- GIVEN the test runs on a `windows-latest` runner
- WHEN the test step executes
- THEN `-coverprofile` is used without `-covermode=atomic`

---

### Requirement: Repository Cleanup

Stale coverage artifacts committed to the repository MUST be removed from git tracking. The `.gitignore` SHALL prevent re-commit of coverage files.

#### Scenario: Stale coverage files untracked

- GIVEN `coverage` and `coverage_rc002` exist in git tracking
- WHEN `git rm --cached` is applied to both files
- THEN the files are no longer tracked by git
- AND the files may remain on disk but are ignored

#### Scenario: Gitignore prevents re-commit

- GIVEN the `.gitignore` contains patterns `coverage`, `coverage_*`, and `coverage*.out`
- WHEN a developer runs `go test -coverprofile=coverage.out` locally
- THEN the generated `coverage.out` is not staged by `git add`
- AND `git status` shows the file as untracked

#### Scenario: Verify cleanup completeness

- GIVEN the cleanup has been applied
- WHEN `git ls-files coverage coverage_rc002` is executed
- THEN no tracked files are listed

---

### Requirement: Phased CI Job Structure

The CI pipeline MUST split into independent, dependent jobs: lint, test (matrix), build, and smoke. Each downstream job SHALL depend on the success of its upstream predecessor.

#### Scenario: Lint runs independently first

- GIVEN a push or PR triggers CI
- WHEN the workflow starts
- THEN a `lint` job runs `golangci-lint` and `go vet`
- AND the `lint` job has no upstream dependencies

#### Scenario: Test matrix gates build

- GIVEN the `lint` job has completed (regardless of outcome)
- WHEN the `test` matrix job runs across all platform runners
- THEN each matrix entry runs `go test` with coverage
- AND the `build` job only starts if all `test` matrix entries succeed

#### Scenario: Build produces and shares artifact

- GIVEN the `test` matrix has succeeded
- WHEN the `build` job compiles the binary
- THEN the binary is uploaded as a workflow artifact
- AND the `smoke` job downloads and uses the shared artifact

#### Scenario: Smoke tests the built binary

- GIVEN the `build` job has succeeded and shared the binary
- WHEN the `smoke` job runs on the downloaded binary
- THEN install, status, and uninstall smoke checks execute
- AND the pipeline fails if any smoke check fails

#### Scenario: Build does not run after test failure

- GIVEN any entry in the `test` matrix fails
- WHEN the workflow evaluates the `build` job
- THEN `build` is skipped via the `needs` dependency chain
- AND `smoke` is also skipped

#### Scenario: Monolithic job split preserves platform coverage

- GIVEN the original CI job tested on 5 platforms (ubuntu, macos, macos-14, ubuntu-arm, windows)
- WHEN the pipeline is restructured into phased jobs
- THEN the `test` matrix SHALL cover all 5 platforms
- AND the `build` job SHALL also build for all 5 platforms
