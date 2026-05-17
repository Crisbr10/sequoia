# Sequoia Audit — Master Report

**Project**: sequoia-ai (Sequoia CLI v0.1.0)
**Stack**: Go 1.24.2, Cobra CLI, Bubbletea TUI, Lipgloss
**Type**: CLI Tool + Plugin Framework
**Date**: 2026-05-12
**Mode**: Full · 6 phases · 59 findings
**Health Score**: **28 / 100 (F)**

---

## Executive Summary

Sequoia CLI is a Go-based tool that installs AI audit framework skills into multiple AI coding assistants (Claude Code, OpenCode, Cursor, Gemini, OpenAI Codex). The codebase is ~34K LOC with 64 test files, robust install scripts, GoReleaser cross-platform builds, and a well-structured adapter pattern.

**The audit reveals a project with strong foundations but severe code duplication and incomplete CI automation.** The single biggest problem: 5 adapter packages duplicate 70-85% of each other's code (~800 lines of near-identical boilerplate), tripling the maintenance surface and making bug fixes error-prone. Combined with a CI pipeline that doesn't enforce linting or coverage, and critical metrics like "90%+ code coverage" being unverifiable, the project's technical debt is concentrated in two areas: adapter architecture and CI maturity.

**Immediate wins**: Extracting a `BaseAdapter` and moving shared strategies to `adapters/common/` eliminates ~700 lines of duplication (solves 4 CRITICAL findings at once). Adding golangci-lint, coverage, and Dependabot to CI automates quality gates that are currently trust-based. Together, these two actions would raise the Health Score from 28 to ~55.

---

## Health Scorecard

| Categoría | Score | Grade |
|-----------|-------|-------|
| Security | 30 | F |
| Architecture | 1 | F |
| Performance | 46 | D |
| Quality | 0 | F |
| Operations | 39 | F |
| i18n | 58 | D |
| **Global** | **28** | **F** |

> *Full score breakdown in `sequoia-score.md`*

---

## Critical Findings (6)

### 🔴 P3-001 / P4-001 — Massive Code Duplication in 5 Adapters
**Root cause**: R1 — No common adapter layer
**Impact**: ~700 duplicated lines. Each bug fix requires touching 5 files. Each new adapter copies 150 lines of boilerplate.
**Fix**: Extract `BaseAdapter` in `adapters/common/` and move `InjectSection`/`GenerateRulesMD`/`RemoveRulesMD` to `adapters/common/strategy.go`.
**Effort**: 4-6h · **Priority**: P0

### 🔴 P3-002 — Byte-by-Byte Duplicated Strategy Functions
**Root cause**: R1 — No common adapter layer
**Impact**: `InjectSection`/`RemoveSection` duplicated in 2 packages; `GenerateRulesMD`/`RemoveRulesMD` duplicated in 3 packages. Fix in backup logic requires 3 changes.
**Fix**: Move the 4 functions to `adapters/common/strategy.go`. Delete `adapters/{claude,gemini,cursor,opencode,codex}/installer.go`.
**Effort**: 2h · **Priority**: P0

### 🔴 P4-002 — `_template` Compiles and Auto-Registers in Production
**Root cause**: R6 — `_template` package in production
**Impact**: The "Template Tool" adapter appears in `DefaultRegistry`. If anyone imports it, it contaminates `sequoia status` and the TUI.
**Fix**: Add `//go:build ignore` to all files in `adapters/_template/`.
**Effort**: 5 min · **Priority**: P0

### 🔴 P2-001 — 248 KB de Plantillas Embebidas Duplicadas
**Causa raíz**: R1 (parcial) — Comandos duplicados en 6 `embed.FS`
**Impacto**: Binario inflado ~125 KB por duplicación de comandos. 6 `embed.FS` cargados en `init()`. 5 archivos de comandos (`sequoia-*.md`) copiados idénticamente en 5-6 adaptadores.
**Fix**: Extraer comandos a `adapters/common/embed.go` con un solo `//go:embed templates/commands`. Cada adapter referencia `common.CommandFS`.
**Esfuerzo**: 2h · **Prioridad**: P1

### 🔴 P6-001 — CI Usa Go 1.22 pero go.mod Declara 1.24.2
**Causa raíz**: R2 — CI sin automatización
**Impacto**: Descarga oculta de toolchain en cada CI run (~30-60s desperdiciados). Cache de `setup-go` nunca usado.
**Fix**: Cambiar `go-version: '1.22'` a `go-version-file: go.mod` en `ci.yml`.
**Esfuerzo**: 5 min · **Prioridad**: P0

---

## High Findings (15) — Summary

| ID | Título | Causa Raíz |
|----|--------|------------|
| P1-001 | Backup file collision — archivos de usuario sobrescritos | — |
| P1-002 | Sin signal handling — Ctrl+C deja estado inconsistente | — |
| P1-003 | Checksum verification saltada silenciosamente en install scripts | — |
| P1-004 | Errores de uninstall descartados con `_ = os.Remove(...)` | R4 |
| P2-002 | `exec.LookPath("engram")` bloquea arranque del TUI | — |
| P2-003 | Templates sin caché — re-parseo en cada install | — |
| P3-003 | `internal/model` importa `adapters` — viola encapsulación | — |
| P3-004 | `ScreenRouter`/`TransitionMap` son dead code (151 líneas) | — |
| P4-003 | `opts.Language` plumbed pero descartado en 6 lugares | R3 |
| P4-004 | Cobertura 0% en todos los paquetes no-main | R2 |
| P6-002 | `.golangci.yaml` configurado pero nunca ejecutado en CI | R2 |
| P6-003 | Sin recolección/upload de cobertura en CI | R2 |
| P6-004 | Sin Dependabot/Renovate — dependencias sin alertas | R2 |
| P7-001 | Language selector cosmético — no wired a traducciones | R3 |
| P7-002 | 51 strings de TUI hardcodeadas en inglés | R3 |

---

## Root Causes (6)

| ID | Cause | Findings | Priority Fix |
|----|-------|-----------|-----------------|
| R1 | No common adapter layer | P3-001, P3-002, P4-001, P2-001, P4-005 | `BaseAdapter` + `common/strategy.go` |
| R2 | Minimal CI without automation | P6-001, P6-002, P6-003, P6-004, P6-005, P6-006, P6-007, P4-004 | Lint + coverage + Dependabot |
| R3 | Dead i18n (plumbing without implementation) | P1-006, P4-003, P7-001, P7-002, P7-005, P7-006 | Wire or hide language selector |
| R4 | No error taxonomy | P1-004, P3-010, P7-006 | `ErrInstallFailed`, `ErrUninstallFailed` |
| R5 | Over-designed pipeline (3 steps vs 1 real) | P2-007, P3-006, P3-012 | Simplify to 1-step or wire real |
| R6 | `_template` compilable in production | P2-008, P4-002, P4-011 | `//go:build ignore` |

---

## Prioritized Action Plan

### Immediate (Current sprint — 1-3 days)

| Order | Action | Findings | Effort |
|-------|--------|-----------|----------|
| 1 | `//go:build ignore` en `_template/` | P4-002, P2-008, P4-011 | 5 min |
| 2 | Fix CI: `go-version-file: go.mod` | P6-001, P6-006 | 10 min |
| 3 | Fix CI: `go-version: 1.22` en go.mod a `1.22` | P4-008 | 5 min |
| 4 | Mover `InjectSection`/`GenerateRulesMD` a `common/strategy.go` | P3-002 | 2h |
| 5 | Extraer `BaseAdapter` en `common/` | P3-001, P4-001 | 4h |
| 6 | Extraer comandos a `common/embed.go` | P2-001 | 2h |
| 7 | Agregar `golangci-lint` a CI | P6-002 | 30 min |
| 8 | Agregar `-coverprofile` y upload a CI | P6-003, P4-004 | 30 min |
| 9 | Crear `.github/dependabot.yml` | P6-004 | 15 min |

### Short Term (Next sprint — 1-2 weeks)

| Orden | Acción | Hallazgos | Esfuerzo |
|-------|--------|-----------|----------|
| 10 | Template cache con `sync.Map` | P2-003 | 1h |
| 11 | Signal handling (SIGTERM/SIGINT) | P1-002 | 2h |
| 12 | Backup file collision fix (unique suffix) | P1-001 | 1h |
| 13 | Checksum verification mandatory | P1-003 | 1h |
| 14 | Collect uninstall errors | P1-004 | 1.5h |
| 15 | `exec.LookPath` async en TUI startup | P2-002 | 1h |
| 16 | Cache `go-figure` logo | P2-005 | 15 min |
| 17 | Fix `internal/model` → `adapters` dependency | P3-003 | 1h |
| 18 | Dead `ScreenRouter` code — delete or wire | P3-004 | 1h |
| 19 | CI badge dinámico en README | P6-007 | 5 min |
| 20 | Cosign artifact signing | P6-005 | 2h |

### Long Term (Backlog)

| Acción | Hallazgos |
|--------|-----------|
| Wire language selector o removerlo | P7-001, P7-002, P1-006, P4-003 |
| i18n infrastructure (`go-i18n/v2` + catalog) | P7-004 |
| English template variants | P7-003 |
| Structured error types | P3-010, P7-006 |
| Single-step pipeline (honest UX) | P3-006 |
| Split `app.Model` (19 fields → sub-structs) | P3-005 |
| Shared mock adapter for tests | P4-006 |
| Atomic writes (temp-then-rename) | P1-010 |
| Permission hardening (0o600 backups) | P1-008 |
| `govulncheck` in CI | P6-Missing#2 |
| SBOM + SLSA provenance | P6-Missing#3, #4 |
| CHANGELOG.md vs GoReleaser reconciliation | P6-008 |

---

## Positives Detected

- ✅ **Zero hardcoded secrets** — no API keys, tokens, o credenciales en el código
- ✅ **Sin command injection** — `exec.LookPath` solo para detección, sin `exec.Command` con input de usuario
- ✅ **Sin path traversal** — paths construidos desde `UserHomeDir` + subdirectorios conocidos
- ✅ **Install scripts producción-quality** — exit codes estructurados, retry logic, checksum verification
- ✅ **GoReleaser config comprehensivo** — 5 targets, checksums, Homebrew, Scoop
- ✅ **Registry pattern testeable** — `DefaultRegistry` con mutex, restore en tests
- ✅ **Pipeline con rollback** — Prepare→Apply→Verify→Rollback bien implementado
- ✅ **48 dependencias saludables** — sin paquetes abandonados, sin CVEs conocidos
- ✅ **Error wrapping consistente** — `fmt.Errorf("context: %w", err)` en todo el código
- ✅ **Sin goroutine leaks** — `sync.WaitGroup` correcto, channels siempre cerrados
- ✅ **Go doc comments exhaustivos** — todo símbolo exportado tiene documentación

---

## Report Phases

| Fase | Archivo | Hallazgos |
|------|---------|-----------|
| P1 Security | `sequoia-phases/01-security.md` | 12 |
| P2 Performance | `sequoia-phases/02-performance.md` | 8 |
| P3 Architecture | `sequoia-phases/03-architecture.md` | 12 |
| P4 Quality | `sequoia-phases/04-quality.md` | 12 |
| P6 Operations | `sequoia-phases/06-operations.md` | 8 |
| P7 i18n | `sequoia-phases/07-i18n.md` | 7 |

---

*Audit generated by Sequoia v0.1.0 · Orchestrator C0 · 2026-05-12*
