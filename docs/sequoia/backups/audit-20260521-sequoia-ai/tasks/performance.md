# Performance Tasks — sequoia-ai

**Score**: 0/100 (Critical) | **Findings**: 12

---

## 🔴 P2-001 (high): Cache Detect() results for process lifetime

**Problem**: Every adapter's `Detect()` (`adapters/claude/adapter.go:64` and siblings) calls `exec.LookPath` — runs on every CLI command and TUI screen with zero caching. `sequoia status` makes up to 3 PATH scans.

**Acceptance Criteria**:
- [ ] Add `detectedOnce sync.Once` and `detectedResult bool` fields to `BaseAdapter` (or `Detector`)
- [ ] `Detect()` calls `sync.Once.Do()` wrapping the `exec.LookPath` + `isInstalled` check
- [ ] Subsequent calls return cached result immediately
- [ ] Test verifies `exec.LookPath` is called exactly once per adapter per process

**Effort**: small (<2h) | **Risk**: low | **Blocks**: P2-008, P2-011

---

## 🔴 P2-002 (high): Cache IsInstalled() results for process lifetime

**Problem**: `BaseAdapter.IsInstalled()` (`adapters/common/base_adapter.go:225-230`) calls `Detector.IsInstalled()` which reads the system prompt file via `os.ReadFile`. Called in TUI `View()` at 60fps. UninstallView calls it up to 10 times per frame.

**Acceptance Criteria**:
- [ ] Add `installedOnce sync.Once` and `installedResult bool` to `BaseAdapter`
- [ ] `IsInstalled()` calls `sync.Once.Do()` wrapping the file read
- [ ] Cache is process-lifetime (results don't change during a session)
- [ ] Test verifies `os.ReadFile` is called exactly once per adapter

**Effort**: small (<2h) | **Risk**: low | **Blocks**: P2-006, P2-012

---

## P2-003 (medium): Cache Status() version file reads

**Problem**: `Status()` reads the version file via `os.ReadFile` on every invocation without caching.

**Acceptance Criteria**:
- [ ] Cache version string after first successful read in `BaseAdapter`
- [ ] Return cached version on subsequent `Status()` calls
- [ ] Cache respects process lifetime (version file doesn't change during session)

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

## P2-004 (medium): Pre-build Lipgloss styles as package-level vars

**Problem**: All 8 style functions in `internal/tui/styles/styles.go` call `lipgloss.NewStyle()` on every invocation, creating ~200 bytes of heap per call at 60fps.

**Acceptance Criteria**:
- [ ] Convert `Title()`, `Subtitle()`, `Body()`, `Accent()`, `Error()`, `Success()`, `Muted()`, `Highlight()` from functions to package-level `var` declarations
- [ ] Verify no behavioral change (styles are immutable once built)
- [ ] Run TUI golden tests to confirm visual output unchanged

**Effort**: small (<1h) | **Risk**: low | **Blocks**: none

---

## P2-005 (medium): Pre-render Logo at init time

**Problem**: `Logo()` in `internal/tui/styles/logo.go:19-29` splits the logo string, creates a new style, and renders through lipgloss on every frame.

**Acceptance Criteria**:
- [ ] Pre-render logo string once in `init()` or as package-level `var`
- [ ] `Logo()` function returns the pre-rendered string directly
- [ ] Golden test confirms logo output unchanged

**Effort**: small (<30m) | **Risk**: low | **Blocks**: none

---

## P2-006 (medium): Compute installed count in Update(), not View()

**Problem**: `UninstallView` loops through all tools calling `IsInstalled()` twice per frame (once for count, once for rendering).

**Acceptance Criteria**:
- [ ] Add `installedCount` field to `UninstallModel`
- [ ] Compute `installedCount` once in `UninstallUpdate` when model is created
- [ ] `UninstallView` reads pre-computed count from model
- [ ] With P2-002 caching, this becomes a single O(1) field read

**Effort**: small (<1h) | **Risk**: low | **Blocks**: P2-002

---

## P2-007~P2-012 (low/info): Cleanup

- **P2-007**: Remove unused `RegisterIn(adapters.DefaultRegistry)` from adapter `init()` functions (5 adapters + template)
- **P2-008**: Remove redundant `Detect()` loop in `runStatus` (already known from `reg.All()`)
- **P2-009**: Replace `err` shadowing in `copyFile` with explicit named return (minor)
- **P2-010**: Skip map allocation in `MergeSection` when `existingTOML == ""` (minor)
- **P2-011**: Remove redundant `Detect()` call in `runInstall` for specific tool
- **P2-012**: Pre-compute "next installed tool" offsets in UninstallUpdate instead of while-loop scanning

**Effort**: small (1-2h total) | **Risk**: low | **Blocks**: none

---

## Quick Win: Fix P2-001 + P2-002 + P2-004 + P2-005 together (~3h)
These four changes eliminate ~95% of TUI jank and CLI latency in one session.
