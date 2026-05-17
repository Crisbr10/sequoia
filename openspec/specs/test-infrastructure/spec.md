# test-infrastructure Specification

## Purpose

Define the test infrastructure requirements established by the remove-i18n-english-only change: golden file regeneration for English-only output, removal of i18n-specific test assertions, and full test suite compliance.

## Requirements

### Requirement: Golden Files English-Only

Golden test files in `testdata/golden/` SHALL be regenerated to match English-only TUI output. No golden file SHALL contain Spanish text.

#### Scenario: Golden files pass English-only assertions
- GIVEN regenerated golden files
- WHEN `go test -race -count=1 ./...` runs
- THEN all golden tests pass; output matches English-only strings

### Requirement: I18n Test Removal

Tests verifying i18n key completeness or Spanish translation coverage SHALL be removed. The test suite SHALL pass without any i18n-key or translation assertions.

#### Scenario: No i18n tests remain
- GIVEN the full test suite
- WHEN `go test ./...` runs
- THEN zero i18n-specific tests exist; no `bundle_test.go` or `keys_test.go`

### Requirement: Full Test Suite Passes

The complete test suite SHALL pass with zero failures attributable to the i18n removal.

#### Scenario: Test suite green
- GIVEN all i18n removal changes applied
- WHEN `go test -race -count=1 ./...` runs
- THEN zero failures from this change; core packages (internal, cmd, codex, cursor) pass
