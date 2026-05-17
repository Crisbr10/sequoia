# Delta Spec: test-infrastructure — ADDED

## Action: ADDED

| # | Requirement | Scenario |
|---|-------------|----------|
| N5 | Golden files SHALL be regenerated for English-only output | GIVEN `testdata/golden/` → WHEN `go test -race -count=1 ./...` runs → THEN all tests pass; golden files match English-only strings |
| N6 | Tests verifying i18n key completeness or Spanish translation SHALL be removed | GIVEN test suite → WHEN i18n tests deleted → THEN `go test ./...` passes with no i18n-key or translation assertions |
| N7 | Full test suite SHALL pass | GIVEN all changes applied → WHEN `go test -race -count=1 ./...` → THEN zero failures; no race conditions |
