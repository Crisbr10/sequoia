# adapter-registration Specification

## Purpose

Explicit dependency-injected adapter registration with no global mutable state.
Adapters register via `RegisterIn()` called on an explicit `*Registry`, not via
`init()` side effects into a shared `DefaultRegistry`.

## Requirements

### Requirement: Explicit Registry Creation

The system MUST provide `NewRegistry()` as the sole constructor for `*Registry`.
There MUST NOT be a `DefaultRegistry` global variable.

#### Scenario: Registry created on demand

- GIVEN a new registry is needed
- WHEN `NewRegistry()` is called
- THEN a fully initialized, empty `*Registry` is returned
- AND `r.Get("any")` returns `ErrUnknownAdapter`
- AND `r.All()` returns an empty slice

#### Scenario: No global registry variable

- GIVEN the `adapters` package is compiled
- WHEN `grepping` for `DefaultRegistry` in production code
- THEN zero references exist

---

### Requirement: Explicit Adapter Registration via DI

Adapters MUST be registered by calling `RegisterIn(reg *adapters.Registry)` on
an explicit registry. Importing an adapter package SHALL NOT auto-register it
into any registry.

#### Scenario: Main registers adapters explicitly

- GIVEN `cmd/sequoia/main.go`
- WHEN it calls `NewRegistry()` followed by each adapter's `RegisterIn(reg)`
- THEN all adapters are available via `reg.All()`
- AND `reg.Get("claude-code")` returns the Claude adapter

#### Scenario: No import-time auto-registration

- GIVEN a test imports `adapters/claude` but calls no `RegisterIn()`
- WHEN it creates `adapters.NewRegistry()` and calls `reg.All()`
- THEN the registry is empty

#### Scenario: Fallback factory preserves lazy construction

- GIVEN an adapter uses `RegisterFactory(id, factory)` inside `RegisterIn(reg)`
- WHEN `RegisterIn(reg)` is called but `reg.Get(id)` is NOT yet called
- THEN the factory is stored but NOT invoked
- AND `reg.All()` triggers construction on first demand

---

### Requirement: Immutable Command File List

The ordered list of command template filenames MUST be returned by a function
that returns a defensive copy. The export SHALL NOT be a mutable slice variable.

#### Scenario: CommandFiles() returns a defensive copy

- GIVEN `common.CommandFiles()` is called twice
- WHEN the first caller mutates the returned slice (e.g., `result[0] = "evil"`)
- THEN the second call returns the original unmodified list

#### Scenario: Compile-time enforcement

- GIVEN the `adapters/common` package compiles
- WHEN inspecting the `CommandFiles` symbol
- THEN it is a function signature `func() []string`
- AND no `var CommandFiles = []string{...}` exists

#### Scenario: Call sites use function syntax

- GIVEN any `.go` file referencing command files
- WHEN it ranges over `common.CommandFiles` (bare, no parens)
- THEN compilation fails — call sites MUST use `common.CommandFiles()`

---

### Requirement: Test Isolation

Each test MUST create its own `*Registry` via `NewRegistry()`. Tests MUST NOT
share registry state through a global variable or import-time side effects.

#### Scenario: Parallel tests use independent registries

- GIVEN two parallel subtests
- WHEN each calls `adapters.NewRegistry()` and registers different adapters
- THEN neither registry contains the other's adapters

#### Scenario: Go build passes without DefaultRegistry

- GIVEN `DefaultRegistry` is removed from `adapters/registry.go`
- WHEN `go build ./...` runs on the full project
- THEN zero compilation errors occur
- AND `go vet ./...` reports no issues

#### Scenario: Full test suite passes after migration

- GIVEN all `adapters.DefaultRegistry` references are replaced with `NewRegistry()`
- WHEN `go test -race -count=1 ./...` runs
- THEN all packages pass with no data races
