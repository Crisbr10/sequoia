# Proposal: RC-001 — Decompose BaseAdapter

## Intent

`adapters/common/base_adapter.go` is a 482-line god-object with 22 function fields across 7 concerns. It blocks 8 Sequoia findings (P3-002 through P4-002). Every bug fix risks all 6 adapters. We extract 3 cohesive structs, segregate the 12-method monolith interface, eliminate the `DefaultRegistry` global, and add error-path + mock FS tests. This unlocks RC-004's BackupManager work (P3-006 depends on stable adapter boundaries).

## Scope

### In Scope
- Extract `PathResolver` struct (~6 fields, pure path mapping)
- Extract `Detector` struct (~2 fields, detection-only)
- Extract `PromptManager` struct (~4 fields, strategy + write/remove + rollback flag)
- Extract `BackupPathBuilder` struct (~2 fields, session-suffix backup path generation)
- Slim `BaseAdapter` to orchestration + template config + state (warnings, backup dir)
- Segregate `ToolAdapter` into 4 role interfaces: `Identifier`, `Detector`, `Installer`, `AdapterPaths`
- Replace `DefaultRegistry` global with constructor DI (`NewModel(registry)`, `RunApp(registry)`)
- Remove `NewAdapter()` factory — consumers use injected registry directly
- Add 20 error-path tests (context cancellation, nil fns, partial failure, template missing)
- Add 15 mock FS tests proving per-adapter template content

### Out of Scope
- BackupManager extraction → RC-004 (we preserve `BackupDirGetter` as-is, deferring BackupManager)
- Performance improvements (symlink caching, template cache key) → RC-006
- Security hardening (path sanitization) → RC-002
- Changing install/uninstall algorithms — behavior-preserving refactor only
- Codex TOML merge logic rewrite
- Gemini sequoia/ subdirectory removal rewrite

## Approach

Six-phase chained PR delivery:
1. **PathResolver extraction** — isolate path logic (~80 LOC new file, 13 tests)
2. **Detector extraction** — isolate detection with baseFn injection (~40 LOC new file, 7 tests)
3. **PromptManager + BackupPathBuilder + Interface segregation** — isolate prompt lifecycle and backup path
4. **DI refactor** — NewRegistry(), RegisterIn(), factory.go deletion, wire through main + model
5. **Error-path + mock FS tests** — 35 new tests across 4 files
6. **Final verification** — go test, go vet, go build, spec compliance check

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `PathResolver` extraction breaks `Base()` caching (sync.Once) | Low | `Base()` moves as-is; sync.Once is self-contained |
| Interface segregation breaks Codex/Gemini custom overrides | Medium | Keep `ToolAdapter` composite for registry; narrow interfaces are opt-in for consumers |
| DefaultRegistry DI breaks 20+ model tests | Medium | Batch-update tests: constructor `NewModel(reg)` vs global swap |
| Codex Installs silently break after PromptManager extraction | Low | Codex sets strategy nil, nil — extracted struct handles nil fns gracefully |

## Rollback Plan
`git revert` any phase independently. Phase 1 (struct extraction) is behavior-preserving — zero downstream breakage. Phase 4 (DI) is the only disruptive change.

## Success Criteria

- [x] All 10 existing `base_adapter_test.go` tests pass unchanged
- [x] All 5 concrete adapter tests pass (claude, codex, cursor, gemini, opencode)
- [x] 35 new error-path + mock FS tests pass
- [x] `go build ./...` succeeds — no compile errors
- [x] `adapters/common/...` coverage reaches 84.6% (exceeds 80% threshold)
- [x] `factory.go` deleted — consumers use `registry.Get(id)` directly
- [x] `DefaultRegistry` deprecated as compat shim; `NewRegistry()` constructor added
- [x] `ToolAdapter` segregated into 4 role interfaces with composite backward compat
