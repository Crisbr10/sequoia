# Sequoia Action Plan — Tasks

**Proyecto**: sequoia-ai
**Auditoría**: 2026-05-12 · Health Score 28/100 (F)
**Formato**: Tareas por fase, priorizadas, con criterios de aceptación

---

## 🔴 Bloqueantes (Sprint Actual — ~10h)

### TASK-SEC.001 · Fix backup file collision

- **Contexto**: 3 adaptadores (OpenCode, Cursor, Codex) usan `path + ".sequoia-backup"` como nombre de backup. Si el usuario ya tiene un archivo con ese nombre, se sobrescribe silenciosamente.
- **Archivos**: `adapters/opencode/installer.go:36`, `adapters/cursor/installer.go:36`, `adapters/codex/installer.go:22`
- **Fix**: Agregar timestamp o random suffix al nombre de backup.
- **Prioridad**: 🔴 Bloqueante · **Esfuerzo**: 1h · **Hallazgo**: P1-001

### TASK-SEC.002 · Add OS signal handling

- **Contexto**: Ctrl+C durante install no ejecuta rollback. El pipeline ya soporta context cancellation pero no está wired a OS signals.
- **Archivos**: `cmd/sequoia/main.go:56`
- **Fix**: `signal.NotifyContext` con `os.Interrupt` y `syscall.SIGTERM`; pasar contexto al pipeline.
- **Prioridad**: 🔴 Bloqueante · **Esfuerzo**: 2h · **Hallazgo**: P1-002

### TASK-ARC.001 · Extract BaseAdapter to common

- **Contexto**: `Install()`/`Uninstall()`/`Status()` duplicados 85%+ en 5 adaptadores (~700 líneas). Cada nuevo adapter copia 150 líneas.
- **Archivos**: `adapters/{claude,opencode,cursor,gemini,codex}/adapter.go`
- **Fix**: Crear `BaseAdapter` struct en `adapters/common/` con la lógica compartida. Cada adapter embebe `BaseAdapter` y solo override `ID()`, `Name()`, `Detect()`, `PromptStrategy()`, `WriteSystemPrompt()`.
- **Prioridad**: 🔴 Bloqueante · **Esfuerzo**: 4h · **Hallazgo**: P3-001, P4-001

### TASK-ARC.002 · Move strategies to common/strategy.go

- **Contexto**: `InjectSection`/`RemoveSection` byte-idénticas en claude+gemini. `GenerateRulesMD`/`RemoveRulesMD` byte-idénticas en opencode+cursor+template.
- **Archivos**: `adapters/{claude,gemini,opencode,cursor,codex}/installer.go`
- **Fix**: Mover las 4 funciones a `adapters/common/strategy.go`. Eliminar archivos `installer.go` por adapter.
- **Prioridad**: 🔴 Bloqueante · **Esfuerzo**: 2h · **Hallazgo**: P3-002

### TASK-QUA.001 · Exclude _template from production builds

- **Contexto**: `_template` tiene `init()` que registra un adapter en `DefaultRegistry`. Si se importa accidentalmente, aparece en producción.
- **Archivos**: `adapters/_template/*.go`
- **Fix**: Agregar `//go:build ignore` en cada archivo `.go` del directorio `_template/`.
- **Prioridad**: 🔴 Bloqueante · **Esfuerzo**: 5 min · **Hallazgo**: P4-002

### TASK-OPS.001 · Fix CI Go version

- **Contexto**: `ci.yml` usa Go 1.22 pero `go.mod` declara 1.24.2. Toolchain auto-download en cada CI run.
- **Archivos**: `.github/workflows/ci.yml:25`, `action.yml:55`
- **Fix**: Cambiar a `go-version-file: go.mod` en ambos archivos.
- **Prioridad**: 🔴 Bloqueante · **Esfuerzo**: 10 min · **Hallazgo**: P6-001, P6-006

### TASK-QUA.002 · Fix go.mod version to match reality

- **Contexto**: `go 1.24.2` en go.mod pero el código usa solo features de Go 1.19+. README dice "mín Go 1.22".
- **Archivos**: `go.mod:3`
- **Fix**: Bajar a `go 1.22.0` y correr `go mod tidy -go=1.22`.
- **Prioridad**: 🔴 Bloqueante · **Esfuerzo**: 5 min · **Hallazgo**: P4-008

---

## 🟠 Alto Leverage (Próximo Sprint — ~12h)

### TASK-PER.001 · Extract shared command templates

- **Contexto**: 5 archivos de comandos duplicados en 6 `embed.FS` (~125 KB). Binario inflado y 6 FS cargados en `init()`.
- **Archivos**: `adapters/*/embed.go`, `adapters/*/templates/commands/`
- **Fix**: Crear `adapters/common/embed.go` con `//go:embed templates/commands` y `var CommandFS embed.FS`. Referenciar desde cada adapter.
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 2h · **Hallazgo**: P2-001

### TASK-SEC.003 · Make checksum verification mandatory

- **Contexto**: `install.sh` y `install.ps1` saltan checksum verification si falla la descarga de `checksums.txt`.
- **Archivos**: `scripts/install.sh:220-248`, `scripts/install.ps1:165`
- **Fix**: Abortar con exit code 2 si checksums.txt no se puede descargar. Agregar `--skip-checksums` flag para opt-out explícito.
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 1h · **Hallazgo**: P1-003

### TASK-SEC.004 · Collect and report uninstall errors

- **Contexto**: Los 5 adaptadores usan `_ = os.Remove(...)` en `Uninstall()`, descartando errores. Archivos locked dejan estado inconsistente.
- **Archivos**: `adapters/{claude,opencode,cursor,gemini,codex}/adapter.go` (métodos `Uninstall`)
- **Fix**: Colectar errores con `errors.Join` y reportar en TUI y CLI.
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 1.5h · **Hallazgo**: P1-004

### TASK-PER.002 · Cache template parsing

- **Contexto**: `RenderTemplate` parsea templates de 81 KB en cada install. Sin caché.
- **Archivos**: `adapters/common/template.go:13-27`
- **Fix**: `sync.Map` para cachear `*template.Template` por (fs, nombre). Usar `.Clone()` en llamadas subsecuentes.
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 1h · **Hallazgo**: P2-003

### TASK-PER.003 · Async engram detection at TUI startup

- **Contexto**: `exec.LookPath("engram")` bloquea la construcción del Model en TUI startup.
- **Archivos**: `internal/app/model.go:88`
- **Fix**: Mover a goroutine post-arranque con mensaje Bubbletea. Inicializar `EngramAvailable: false`.
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 1h · **Hallazgo**: P2-002

### TASK-OPS.002 · Add golangci-lint to CI

- **Contexto**: `.golangci.yaml` con 10 linters configurados pero nunca ejecutados en CI.
- **Archivos**: `.github/workflows/ci.yml`
- **Fix**: Agregar step con `golangci/golangci-lint-action@v6` (solo ubuntu).
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 30 min · **Hallazgo**: P6-002

### TASK-OPS.003 · Add coverage to CI

- **Contexto**: "90%+ coverage" claim es estática y no verificable. Coverage nunca se mide en CI.
- **Archivos**: `.github/workflows/ci.yml`
- **Fix**: `go test -coverprofile=coverage.out ./...` + `actions/upload-artifact@v4`.
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 30 min · **Hallazgo**: P6-003, P4-004

### TASK-OPS.004 · Add Dependabot

- **Contexto**: Sin actualización automática de dependencias. CVEs en deps transitivas pasarían desapercibidos.
- **Archivos**: `.github/dependabot.yml` (nuevo)
- **Fix**: Crear config con `gomod` (semanal, max 5 PRs) y `github-actions` (semanal, max 3 PRs).
- **Prioridad**: 🟠 Alto leverage · **Esfuerzo**: 15 min · **Hallazgo**: P6-004

---

## 🟡 Backlog Priorizado (~25h)

### Fase 1: Seguridad (3h)

| Task | Descripción | Hallazgo | Esfuerzo |
|------|-------------|----------|----------|
| TASK-SEC.005 | Atomic writes (temp-then-rename) para backup+replace | P1-010 | 2h |
| TASK-SEC.006 | Restrict backup permissions (0o600 backups, 0o700 dirs) | P1-008 | 1h |

### Fase 2: Performance (1.5h)

| Task | Descripción | Hallazgo | Esfuerzo |
|------|-------------|----------|----------|
| TASK-PER.004 | Cache `go-figure` logo (precomputar con `sync.Once`) | P2-005 | 15 min |
| TASK-PER.005 | Cache `os.UserHomeDir()` en adapter struct | P2-006 | 30 min |
| TASK-PER.006 | Remove dead `spinnerFrames` or implement animation | P2-007 | 30 min |
| TASK-PER.007 | Move `debug.ReadBuildInfo` from `init()` to `RunE` | P2-004 | 15 min |

### Fase 3: Arquitectura (5h)

| Task | Descripción | Hallazgo | Esfuerzo |
|------|-------------|----------|----------|
| TASK-ARC.003 | Fix `internal/model` → `adapters` dependency | P3-003 | 1h |
| TASK-ARC.004 | Delete or wire `ScreenRouter`/`TransitionMap` dead code | P3-004 | 1h |
| TASK-ARC.005 | Split `app.Model` into sub-structs (PipelineState + NavState) | P3-005 | 2h |
| TASK-ARC.006 | Single-step pipeline (honest UX) | P3-006, P3-012 | 1h |

### Fase 4: Calidad (4h)

| Task | Descripción | Hallazgo | Esfuerzo |
|------|-------------|----------|----------|
| TASK-QUA.003 | DRY `runInstallSteps`/`runUninstallSteps` | P4-005 | 1h |
| TASK-QUA.004 | Shared mock adapter for tests | P4-006 | 1h |
| TASK-QUA.005 | Fix fragile recursion limit in test helper | P4-007 | 30 min |
| TASK-QUA.006 | Move `renderUninstallConfirm` to screens package | P4-010 | 30 min |
| TASK-QUA.007 | Remove meaningless test `TestWaitForProgress_ContextCancellation...` | P4-012 | 15 min |
| TASK-QUA.008 | Fix `len(defaultStepNames) > 0` dead code guard | P4-009 | 15 min |

### Fase 5: Operaciones (3h)

| Task | Descripción | Hallazgo | Esfuerzo |
|------|-------------|----------|----------|
| TASK-OPS.005 | Cosign keyless artifact signing | P6-005 | 2h |
| TASK-OPS.006 | Dynamic CI badge in README | P6-007 | 5 min |
| TASK-OPS.007 | Resolve CHANGELOG.md vs GoReleaser dual source | P6-008 | 1h |

### Fase 6: i18n (8h)

| Task | Descripción | Hallazgo | Esfuerzo |
|------|-------------|----------|----------|
| TASK-I18N.001 | Add `go-i18n/v2` + English message catalog | P7-004 | 2h |
| TASK-I18N.002 | Externalize 51 TUI strings to catalog | P7-002 | 3h |
| TASK-I18N.003 | Wire language selector to catalog lookups | P7-001 | 2h |
| TASK-I18N.004 | Translate skill.md.tmpl to English (default) | P7-003 | 1h |

---

## Dependencias entre Tareas

```
TASK-ARC.001 (BaseAdapter)
├── TASK-ARC.002 (common/strategy.go) — pueden hacerse juntas
│   └── TASK-PER.001 (shared command templates) — usa la misma capa common
│
TASK-QUA.001 (_template build ignore) — independiente, 5 min
│
TASK-OPS.001 (CI Go version) — independiente
├── TASK-OPS.002 (golangci-lint)
├── TASK-OPS.003 (coverage)
└── TASK-OPS.004 (Dependabot)
│
TASK-SEC.002 (signal handling)
└── TASK-SEC.001 (backup collision) — ambos afectan el lifecycle de install
    └── TASK-SEC.004 (uninstall errors) — hace el rollback observable
│
TASK-I18N.001 (i18n library)
└── TASK-I18N.002 (externalize strings)
    └── TASK-I18N.003 (wire language selector)
```

---

## Estimación Total

| Sprint | Tareas | Esfuerzo |
|--------|--------|----------|
| Actual (bloqueantes) | 7 | ~10h |
| Próximo (alto leverage) | 8 | ~12h |
| Backlog | 19 | ~25h |
| **Total** | **34** | **~47h** |

---

*Plan de acción generado por Sequoia v0.1.0 · M2 Reporter · 2026-05-12*
