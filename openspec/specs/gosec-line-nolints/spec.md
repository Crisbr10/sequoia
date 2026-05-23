# gosec-line-nolints

> Source: CORR-004 — Common Package Split (archived 2026-05-22)
> Domain: MODIFIED — P4-007 blanket gosec exclusion removal

File-level gosec exclusions SHALL be removed from `.golangci.yaml`. Each gosec warning in previously excluded files SHALL be addressed with a line-level `//nolint:gosec` comment containing a justification.

## Requirements

| # | Requirement | Strength |
|---|------------|----------|
| R16 | `.golangci.yaml` has zero file-level gosec exclusions for `adapters/`, `cmd/sequoia/main.go`, test files | MUST |
| R17 | Every `//nolint:gosec` includes a justification comment on the same or preceding line | MUST |
| R18 | CI gosec step passes with line-level nolints only | MUST |

## Scenarios

### New code in previously excluded file is scanned
- GIVEN a new security-sensitive line added to `adapters/common/strategy.go`
- WHEN CI runs gosec
- THEN the new line is scanned
- AND a gosec warning is raised if the line is insecure

### Justified nolint suppresses known false positive
- GIVEN `os.WriteFile` with mode `0o644` triggers gosec G306
- WHEN `//nolint:gosec // G306: file is meant to be world-readable config` precedes the line
- THEN gosec does NOT flag this line
- AND CI passes
