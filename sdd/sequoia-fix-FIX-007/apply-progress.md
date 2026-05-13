## Apply Progress — FIX-007: Mandatory Checksum Verification

**Status**: ✅ Complete (all 3 files modified)
**Mode**: Strict TDD

### Completed Tasks
- [x] Fix `scripts/install.sh` — mandatory checksum + `--skip-checksums` flag
- [x] Fix `scripts/install.ps1` — mandatory checksum + `-SkipChecksum` switch (fix catch block)
- [x] Update `scripts_test.go` — add tests for mandatory checksum behavior

### Files Changed
| File | Action | What Was Done |
|------|--------|---------------|
| `scripts/install.sh` | Modified | Removed `|| true` from checksum download; added abort (exit 2) on download failure; added `SKIP_CHECKSUMS` env var and `--skip-checksums` flag; updated header docs |
| `scripts/install.ps1` | Modified | Fixed catch block to abort (exit 2) on checksum download failure unless `-SkipChecksum`; updated `-SkipChecksum` parameter help text to mention mandatory default and air-gapped use case |
| `scripts_test.go` | Modified | Added `TestInstallShChecksumMandatory` (5 subtests) and `TestInstallPs1ChecksumMandatory` (5 subtests) |

### TDD Cycle Evidence
| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| install.sh checksum | `scripts_test.go` | Unit (content) | ✅ 4/4 | ✅ Written | ✅ Passed | ✅ 5 tests | ➖ None needed |
| install.ps1 checksum | `scripts_test.go` | Unit (content) | ✅ 4/4 | ✅ Written | ✅ Passed | ✅ 5 tests | ➖ None needed |

### Test Summary
- **Total tests written**: 10 (5 for install.sh + 5 for install.ps1)
- **Total tests passing**: 14/14 (including 4 existing repo refs tests)
- **Layers used**: Unit (content matching) — shell scripts tested via Go string assertions
- **Approval tests**: None — modifying scripts, not Go code
- **Pure functions created**: 0 (shell scripts)

### Full Test Suite
All 17 packages pass: `go test ./...` — ✅
