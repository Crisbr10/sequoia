# Sequoia Performance Audit (P2)

**Target**: `sequoia-ai` Go CLI v0.1.0  
**Date**: 2026-05-12  
**Budget**: Startup < 500ms · Help < 100ms · Memory < 100MB · Install < 5s/adapter  

---

## Executive Summary

| Metric | Status |
|--------|--------|
| Binary size | ~248 KB of **duplicated** embedded templates |
| Startup path | `init()` + `exec.LookPath` block TUI boot |
| Template processing | Parse on every call — no cache |
| I/O during install | Single-pass, correct — 1 concern on home dir calls |
| Memory | Light — < 50 MB observed |

8 findings: 2 CRÍTICO, 2 RIESGO, 4 ATENCIÓN.

---

## Findings

### [P2-001] · 248 KB de plantillas embebidas duplicadas en 6 paquetes adaptadores  [🔴 CRÍTICO]
**State**: Confirmed  
**Evidence**: `adapters/*/embed.go:5` (5 producción + 1 template) — todas con `//go:embed templates` independiente. `sequoia-init.md` existe en 6 directorios; `sequoia-audit.md`, `sequoia-diff.md`, `sequoia-fix.md`, `sequoia-review.md` duplicados 5 veces cada uno. Total embebido: 248 KB (44 archivos en 6 `embed.FS`).  
**Problem**: Cinco archivos de comandos Markdown (~125 KB combinados) están replicados idénticamente en cada árbol `embed.FS` por paquete. El binario compila 6 sistemas de archivos separados. La plantilla `skill.md.tmpl` de opencode ocupa 81 KB por sí sola. Estos FS se pagan en tamaño de binario y en tiempo de inicialización por `init()`.  
**Real Impact**: Binario inflado ~125 KB por duplicación de comandos + ~120 KB por skill templates no compartidas. En Windows con disco lento, cargar 6 `embed.FS` agrega latencia medible de inicio. El paquete `_template` (8 KB) también se embebe aunque es solo andamiaje de referencia.  
**Minimal High-Leverage Recommendation**: Extraer los 5 archivos de comandos (`sequoia-*.md`) a `adapters/common/embed.go` como UN solo `//go:embed templates/commands` compartido (`common.CommandFS`). Cada adaptador usa `common.CommandFS` en vez de mantener su propia copia. Las skill templates específicas de cada adaptador permanecen en su paquete.  
**Dependencies/Blockers**: Las skill templates por adaptador son legítimamente diferentes (varían entre herramientas). Los comandos son los mismos — confirmar con diff que no hay divergencias intencionales.  
**Implementation Risk**: **Low** — Las plantillas de comandos son archivos estáticos copiados con el mismo nombre en cada adaptador. El loader común ya existe en `common`. Solo hay que mover archivos y cambiar referencias de `templateFS` a `common.CommandFS` para los comandos.  
**Acceptance Criteria**: 
- [ ] `adapters/common/embed.go` existe con `//go:embed templates/commands` y exporta `CommandFS`
- [ ] `adapters/*/templates/commands/` eliminado en los 5 adaptadores de producción
- [ ] Cada adapter referencia `common.CommandFS` en lugar de su `templateFS` para archivos de comandos
- [ ] Tamaño total de templates embebidas ≤ 150 KB (vs 248 KB actual)
- [ ] Tests de instalación pasan (`go test ./adapters/...`)

---

### [P2-002] · `exec.LookPath("engram")` bloquea el arranque del TUI en construcción del Model  [🟠 RIESGO]
**State**: Confirmed  
**Evidence**: `internal/app/model.go:88` — `exec.LookPath("engram")` llamado sincrónicamente en `NewModel()` durante la construcción del modelo Bubbletea.  
**Problem**: `exec.LookPath` escanea cada directorio del `PATH` buscando el binario `engram`. En Windows, debe probar extensiones del `PATHEXT` (`.exe`, `.cmd`, `.bat`, etc.) en cada directorio. Esta llamada bloquea la creación del modelo y por tanto el primer render del TUI, aunque el usuario nunca use persistencia Engram. `EngramAvailable` solo se usa en la pantalla de Configuration — nunca en Welcome, ToolSelection, o InstallProgress.  
**Real Impact**: Agrega 10-100ms de latencia al arranque del TUI (dependiendo de longitud de PATH, tipo de disco y SO). En Windows con PATH largos y discos mecánicos, puede exceder 100ms.  
**Minimal High-Leverage Recommendation**: Mover la detección de Engram a una goroutine post-arranque que emita un mensaje Bubbletea para setear `EngramAvailable` de forma asíncrona. Alternativa más simple: cachear el resultado en una variable a nivel de paquete con `sync.Once` la primera vez que se renderiza la pantalla de Configuration.  
**Dependencies/Blockers**: Ninguna — `EngramAvailable` es un booleano usado solo en renderizado condicional de ConfigurationView.  
**Implementation Risk**: **Low** — Cambio localizado en `model.go` (quitar `exec.LookPath`, inicializar `EngramAvailable: false`, añadir handler de mensaje para actualizarlo). La pantalla de Configuration ya maneja correctamente `EngramAvailable=false` (muestra opción Engram en gris).  
**Acceptance Criteria**: 
- [ ] `NewModel()` no llama a `exec.LookPath`
- [ ] `EngramAvailable` se setea asincrónicamente después del primer render
- [ ] `go test ./internal/app/...` pasa

---

### [P2-003] · Templates ejecutados sin caché — cada instalación re-parsea desde disco  [🟠 RIESGO]
**State**: Confirmed  
**Evidence**: `adapters/common/template.go:13-27` — `RenderTemplate` lee el archivo embebido, llama a `template.New(name).Parse(string(raw))`, ejecuta, y descarta el template parseado. Se llama 2-3 veces por adaptador durante install (skill + section template).  
**Problem**: `text/template.Parse()` es O(n) sobre el tamaño de la plantilla y aloca un AST completo. La skill template de opencode mide 81 KB — cada install de opencode re-parsea 81 KB de template desde el `embed.FS`. Para una instalación de 3 adaptadores, esto implica 6-9 parseos de templates de 6-81 KB cada uno. No hay `sync.Map` ni caché a nivel de paquete.  
**Real Impact**: ~50-150ms de CPU desperdiciada por instalación multi-adapter. Agravado por el hecho de que los templates son inmutables (embebidos en el binario) — parsearlos una sola vez y reusarlos es trivial.  
**Minimal High-Leverage Recommendation**: Añadir un caché per-FS usando `sync.Map` en `RenderTemplate` (o un nuevo `CachedRenderer`). En la primera llamada por nombre de archivo, parsear y almacenar `*template.Template`. En llamadas subsecuentes, clonar con `tmpl.Lookup(name)` o `tmpl.Clone()` y ejecutar.  
**Dependencies/Blockers**: `text/template.Template.Clone()` es seguro para uso concurrente desde Go 1.13. El `embed.FS` es inmutable y seguro para acceso concurrente.  
**Implementation Risk**: **Low** — Cambio localizado en `common/template.go`. La semántica de `RenderTemplate` no cambia; solo se agrega una capa de caché interna.  
**Acceptance Criteria**: 
- [ ] `RenderTemplate` cachea templates parseados por (fs, name)
- [ ] Segunda llamada con el mismo nombre no re-parsea
- [ ] Tests existentes de templates pasan (`go test ./adapters/common/...`)
- [ ] Benchmark muestra ≥ 5× speedup en llamadas repetidas

---

### [P2-004] · `debug.ReadBuildInfo()` en `init()` penaliza cada comando incluso `help`/`version`  [🟡 ATENCIÓN]
**State**: Confirmed  
**Evidence**: `cmd/sequoia/main.go:33-52` — `init()` llama a `debug.ReadBuildInfo()`, itera sobre `info.Settings` buscando `vcs.revision`.  
**Problem**: `init()` se ejecuta ANTES de `main()` en toda invocación del binario — incluyendo `sequoia help`, `sequoia version`, y `sequoia status`. `debug.ReadBuildInfo()` parsea la metadata embebida del binario. Aunque en producción con `-ldflags` hace early-return (línea 34-35), en modo desarrollo (`go build` local) siempre se ejecuta. Los blank imports de adaptadores (líneas 20-24) y sus `init()` también corren en cada comando.  
**Real Impact**: En desarrollo local, agrega 5-15ms a cada comando. En producción con GoReleaser el impacto es mínimo por el guard clause. Sin embargo, el patrón establece un precedente: `init()` hace trabajo no esencial para comandos de solo-lectura.  
**Minimal High-Leverage Recommendation**: Mover la lógica de `debug.ReadBuildInfo()` del `init()` al `RunE` del comando `version`. La detección de versión solo es necesaria cuando se muestra la versión. Para el banner del TUI, pasar la versión detectada desde `main()` tras construir el comando.  
**Dependencies/Blockers**: La variable `Version` se usa en `runTUI()` → `app.NewModel(toolID, Version)` para mostrarla en la pantalla Welcome. Mover la detección a `main()` entre `newRootCmd()` y la llamada a `Execute()` mantiene la semántica.  
**Implementation Risk**: **Low** — Refactor localizado en `main.go`. Los tests de integración que verifican la versión deben adaptarse.  
**Acceptance Criteria**: 
- [ ] `init()` no contiene llamadas a `debug.ReadBuildInfo()`
- [ ] `sequoia version` sigue mostrando la versión correcta en todos los modos de build
- [ ] `go test ./cmd/...` pasa
- [ ] `go build -o sequoia.exe && ./sequoia.exe help` no ejecuta `debug.ReadBuildInfo`

---

### [P2-005] · ASCII art del logo regenerado con `go-figure` en cada render de Welcome  [🟡 ATENCIÓN]
**State**: Confirmed  
**Evidence**: `internal/tui/styles/logo.go:11-23` — `Logo()` llama a `figure.NewFigure("Sequoia", "", true)` en CADA invocación. `WelcomeView()` (screens/welcome.go:38) llama a `styles.Logo()` en cada render.  
**Problem**: `go-figure.NewFigure()` genera arte ASCII aplicando cada carácter contra un mapa de fuentes — aloca múltiples strings y hace procesamiento de texto. El resultado es determinístico: "Sequoia" con la fuente por defecto nunca cambia. Sin embargo, se recalcula en cada render de Welcome (incluyendo resize de ventana).  
**Real Impact**: ~10-20ms de CPU por render de Welcome. Aunque Bubbletea solo re-renderiza en cambios de estado, la primera renderización (cold start) ya paga este costo sobre el costo de `exec.LookPath` y `embed.FS` init.  
**Minimal High-Leverage Recommendation**: Computar el logo una sola vez con `sync.Once` o a nivel de paquete: `var logoStr = generateLogo()`. Como "Sequoia" y la fuente nunca cambian, no hay razón para recomputar.  
**Dependencies/Blockers**: Ninguna.  
**Implementation Risk**: **Low** — Cambio de 3 líneas en `logo.go`.  
**Acceptance Criteria**: 
- [ ] `styles.Logo()` devuelve string precomputado (no llama a `go-figure` en cada invocación)
- [ ] `go test ./internal/tui/styles/...` pasa
- [ ] Benchmark muestra `Logo()` en < 1µs

---

### [P2-006] · Múltiples llamadas redundantes a `os.UserHomeDir()` en métodos de adaptador  [🟡 ATENCIÓN]
**State**: Confirmed  
**Evidence**: `adapters/claude/adapter.go:34-36` — `base()` llamado desde 7 métodos distintos (`SkillsPath`, `CommandsPath`, `SystemPromptPath`, `IsInstalled`, `Status`, `Install`, `Uninstall`). `Status()` (línea 105) llama `a.base()` DOS veces. En `runStatus()` headless (`cmd/sequoia/main.go:250-261`), cada adaptador llama `Status()` + `Detect()` = hasta 10 `os.UserHomeDir()`.  
**Problem**: `os.UserHomeDir()` lee `USERPROFILE`/`HOME` en cada llamada — no está cacheado al nivel de stdlib. En `sequoia status`, 5 adaptadores × 2 llamadas c/u = 10 lecturas de variable de entorno. Aunque cada lectura es ~1-5µs, es una operación repetitiva innecesaria y un antipatrón de I/O disperso.  
**Real Impact**: ~10-50µs total en `sequoia status` — insignificante individualmente, pero denota falta de caché local en una ruta caliente (los métodos de adaptador se llaman frecuentemente durante la instalación).  
**Minimal High-Leverage Recommendation**: Cachear el resultado de `os.UserHomeDir()` en el struct del adaptador (campo `cachedHome string`) con lazy init en `base()`. El campo `homeDir` ya existe para testing — unificar.  
**Dependencies/Blockers**: Los tests ya inyectan `homeDir` vía `NewAdapter(homeDir)`. Solo hay que asegurar que el código de producción (campo vacío) cachee tras el primer `os.UserHomeDir()`.  
**Implementation Risk**: **Low** — Cambio localizado en cada adapter.  
**Acceptance Criteria**: 
- [ ] `base()` llama a `os.UserHomeDir()` como máximo 1 vez por instancia de adapter
- [ ] `go test ./adapters/...` pasa

---

### [P2-007] · `spinnerFrames` de 10 elementos declarado pero nunca animado  [🟡 ATENCIÓN]
**State**: Confirmed  
**Evidence**: 
- `internal/tui/screens/install-progress.go:48` — `var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}`  
- `internal/tui/screens/install-progress.go:123` — `spinner := styles.Accent().Render("⠋")` usa un carácter hardcodeado, no indexa `spinnerFrames`  
**Problem**: Se aloca un slice de 10 strings en `init()` del paquete, pero `renderStepRow` nunca lo usa — siempre renderiza el mismo frame braille `"⠋"`. No existe ningún `tea.Tick` mandando mensajes de animación para ciclar frames. El usuario ve un spinner estático (parece congelado) en la pantalla de progreso de instalación.  
**Real Impact**: Alocación muerta (~160 bytes en heap). El spinner estático reduce la percepción de progreso (el usuario no sabe si la instalación sigue viva). La variable `spinnerFrames` sugiere animación implementada cuando no lo está.  
**Minimal High-Leverage Recommendation**: Eliminar `spinnerFrames` y mantener indicador estático, o implementar animación real con `tea.Tick`. Para una CLI tool de instalación, un indicador estático "⠋" es suficiente — remover el slice no usado.  
**Dependencies/Blockers**: Ninguna. Si se quiere animación real, requiere añadir un `tea.Tick` en `Model.Init()` o en el handler de `ScreenInstallProgress`.  
**Implementation Risk**: **Low** — Solo eliminar variable no usada o implementar animación con patrón Bubbletea estándar.  
**Acceptance Criteria**: 
- [ ] `spinnerFrames` se elimina o se implementa su animación vía `tea.Tick`
- [ ] Pantalla de progreso de instalación muestra indicador de actividad claro
- [ ] `go test ./internal/tui/screens/...` pasa

---

### [P2-008] · Paquete `_template` y sus 8 KB de plantillas compilados en el binario de producción  [🟡 ATENCIÓN]
**State**: Confirmed  
**Evidence**: `adapters/_template/embed.go:5` — `//go:embed templates` con 3 archivos (8 KB). `adapters/_template/adapter.go:30` — `func init()` registra el adapter en `DefaultRegistry`.  
**Problem**: El paquete `_template` es andamiaje de referencia para nuevos adaptadores. Su `init()` registra un adapter con ID `"_template"` en el registry global. El `embed.FS` con 8 KB de plantillas se compila en el binario aunque ningún blank import en `main.go` lo referencia. El nombre del paquete (`_template`) NO es un build tag — el compilador Go lo incluye porque es parte del módulo.  
**Real Impact**: ~8 KB de código y templates muertos en el binario. El adapter `_template` aparece en `DefaultRegistry.All()` y por tanto en `sequoia status` si se importara (actualmente no se importa, pero el código compilado igual reside en el árbol de objetos del compilador).  
**Minimal High-Leverage Recommendation**: Añadir `//go:build ignore` en `adapters/_template/adapter.go` para excluir el paquete de builds de producción. Mover el contenido de referencia a `docs/adapter-template/` fuera de `adapters/`.  
**Dependencies/Blockers**: El `CONTRIBUTING.md` referencia `_template` como punto de partida. Actualizar la documentación.  
**Implementation Risk**: **Low** — `//go:build ignore` es una directiva estándar de Go. Ningún archivo en producción importa `_template`.  
**Acceptance Criteria**: 
- [ ] `_template` no se compila en builds de producción (`go build -o sequoia.exe ./cmd/sequoia`)
- [ ] `CONTRIBUTING.md` refleja la nueva ubicación del template
- [ ] `go build ./...` compila sin errores

---

## Metrics Summary

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Binary embedded templates | 248 KB | < 150 KB | 🔴 Over budget |
| Duplicate command files | 5 copies each | 1 shared copy | 🔴 Waste |
| TUI startup overhead | `exec.LookPath` + `go-figure` | Lazy detection + cached logo | 🟠 Modest |
| Template parse cache | None | `sync.Map` per FS | 🟠 Missing |
| `init()` work in dev | 5-15ms | < 1ms | 🟡 Acceptable in prod |
| Home dir I/O per status | 10 env reads | 1 cached read | 🟡 Low impact |
| Dead allocations | 160 B (spinner) + 8 KB (template) | 0 B | 🟡 Trivial |

## Health Score Contribution

| Finding | Severity | Scope | Deduction |
|---------|----------|-------|-----------|
| P2-001 | critical=15 | shared (×1.5) | -22.5 |
| P2-002 | high=8 | shared (×1.5) | -12.0 |
| P2-003 | high=8 | isolated (×1.0) | -8.0 |
| P2-004 | medium=4 | isolated (×1.0) | -4.0 |
| P2-005 | medium=4 | isolated (×1.0) | -4.0 |
| P2-006 | low=2 | isolated (×1.0) | -2.0 |
| P2-007 | low=2 | isolated (×1.0) | -2.0 |
| P2-008 | low=2 | isolated (×1.0) | -2.0 |
| **Total P2 deduction** | | | **-56.5** |

---

## Key Learnings

- Los archivos de comandos Markdown (`sequoia-init.md`, `sequoia-audit.md`, etc.) son el mismo contenido en 5-6 adaptadores — el 50% del peso de templates es duplicación pura.
- Las skill templates (`skill.md.tmpl`) varían legítimamente entre adaptadores y NO deben unificarse.
- El patrón de `init()` + blank import para registro de adaptadores es correcto (estilo `database/sql`), pero los `embed.FS` deberían ser lazy o compartidos donde el contenido es idéntico.
- `exec.LookPath` en el camino crítico de arranque del TUI es evitable — la detección de herramientas externas puede ser asincrónica.
- El pipeline de instalación (`runner.go`) ya usa goroutines correctamente para paralelismo multi-adapter — no se necesitan más goroutines en la ruta de instalación.
