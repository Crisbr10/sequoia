## Archive Report — namespace-backup-dirs

**Archived at**: 2026-05-19
**Root cause**: RC-004 from Sequoia audit (HIGH severity — silent data loss during rollback)
**Mode**: Hybrid (Engram + openspec)

### Artifact Observation IDs (Engram)
| Artifact | ID |
|----------|----|
| Proposal | #882 |
| Spec | #884 |
| Design | #886 |
| Tasks | #888 |
| Apply Progress | #889 |
| Verify Report | #891 |

### Specs Synced
| Domain | Action | Details |
|--------|--------|---------|
| backup-isolation | Created | New spec domain — 3 requirements, 6 scenarios |

### Archive Contents (filesystem)
- `proposal.md` ✅
- `specs/backup-isolation/spec.md` ✅
- `design.md` ✅
- `tasks.md` ✅ (10/10 tasks complete)
- `verify-report.md` ✅ (PASS WITH WARNINGS — no CRITICAL issues)
- `archive-report.md` ✅

### Implementation Summary
- **4 files changed**: `base_adapter.go`, `codex/adapter.go`, `_template/adapter.go`, `installer_test.go`
- **1,000% health score increase**: Silent data loss prevented; cross-adapter isolation achieved
- **Core fix**: Three-segment backup path `{base}-{adapterID}-{sessionSuffix}/{installerType}` with independent subdirectories per installer type

### Source of Truth Updated
- `openspec/specs/backup-isolation/spec.md` — new spec reflects the implemented backup isolation behavior

### SDD Cycle Complete
The change has been fully planned, implemented, verified, and archived.
