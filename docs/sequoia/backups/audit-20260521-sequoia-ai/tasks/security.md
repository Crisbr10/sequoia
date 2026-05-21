# Security Tasks — sequoia-ai

**Score**: 70/100 (Fair) | **Findings**: 4

---

## P1-001 (medium): Explicit file permissions on installed files

**Problem**: `copyFile` in `adapters/common/installer.go:200` uses `os.Create(dst)` which creates files with umask-dependent permissions. Installed skill/command files could be world-readable on misconfigured systems.

**Acceptance Criteria**:
- [ ] `copyFile` calls `os.Chmod(dst, 0o644)` after `os.Create` and before `io.Copy` (or uses `os.OpenFile` with explicit mode)
- [ ] `StageFile` in `installer.go:109` calls `os.Chmod` if it writes a new file (currently uses `os.WriteFile` with 0o644, which is already correct — verify consistency)
- [ ] Tests verify that installed files have 0o644 permissions (not world-writable)

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

## P1-002 (medium): Explicit permissions on probe file

**Problem**: `Prepare()` creates `.sequoia-probe` with default umask permissions at `adapters/common/installer.go:64`. Race window where empty file exists on disk with uncontrolled permissions.

**Acceptance Criteria**:
- [ ] `os.Create(probe)` replaced with `os.OpenFile(probe, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)`
- [ ] File is cleaned up in `defer` immediately after creation check passes
- [ ] Test verifies probe file is created with restricted permissions

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

## P1-003 (low): Sanitize home directory path in status output

**Problem**: `runStatus` in `cmd/sequoia/main.go:321` prints the user's full home directory path (including resolved symlink target) to stdout, which may leak OS usernames in CI logs.

**Acceptance Criteria**:
- [ ] Home directory path display uses `filepath.Base(home)` or `~` shorthand instead of full absolute path
- [ ] Resolved symlink path is shown relative to home or as `~` equivalent
- [ ] Output still clearly indicates symlink presence without exposing full paths

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

## P1-004 (low): Validate working-directory input in GitHub Action

**Problem**: `action.yml` uses `${{ inputs.working-directory }}` unsanitized in shell expressions and JavaScript interpolation. Path traversal possible if workflow author passes `..` segments.

**Acceptance Criteria**:
- [ ] Add validation step: check that `inputs.working-directory` doesn't contain `..` segments
- [ ] Add validation: path must resolve within `${{ github.workspace }}`
- [ ] Fail action early with clear error message if validation fails
- [ ] Document in action.yml description that path must be within repository

**Effort**: small (<2h) | **Risk**: low | **Blocks**: none

---

## No Critical/High findings in Security. All tasks are backlog priority.
