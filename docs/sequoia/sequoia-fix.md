# Plan de Implementación — sequoia-ai

**Generado desde**: Auditoría del 2026-05-12 · Health Score 28/100 (F)
**Total tareas**: 19 (6 causas raíz fusionadas + 13 hallazgos aislados)
**🔴 Bloqueantes**: 6 | **🟠 Alto leverage**: 8 | **🟡 Backlog**: 5

---

## Orden de implementación

```
FIX-001 ──► FIX-002 ──► FIX-003 ──► FIX-004 ──► FIX-005 ──► FIX-006
  (indep)    (indep)    (indep)     (indep)    (usa FIX-003) (usa FIX-003)

FIX-007 ──► FIX-008 ──► FIX-009 ──► FIX-010 ──► FIX-011 ──► FIX-012
  (indep)    (indep)    (indep)     (indep)    (indep)      (indep)

FIX-013 ──► FIX-014 ──► FIX-015 ──► FIX-016 ──► FIX-017 ──► FIX-018 ──► FIX-019
  (indep)    (usa 015)  (indep)     (indep)    (indep)      (indep)     (indep)
```

---

## 🔴 Bloqueantes

---

### [FIX-001] · Excluir `_template` de builds de producción

**Prioridad**: 🔴 Bloqueante
**Fase origen**: P4 Quality + P2 Performance
**Hallazgo(s) origen**: P4-002, P2-008, P4-011 (causa raíz R6)

**Contexto mínimo**:
El directorio `adapters/_template/` es un paquete Go válido cuyo `init()` registra un adapter `ID="template"` en `DefaultRegistry`. Aunque ningún archivo lo importa explícitamente, es compilable y cualquier `import _ "..."` accidental lo metería en producción, contaminando `sequoia status` y la TUI con "Template Tool". Además, sus 8 KB de plantillas embebidas y 20 TODOs se compilan innecesariamente.

**Archivos involucrados**:
- `adapters/_template/adapter.go` — `init()` que registra en `DefaultRegistry`
- `adapters/_template/embed.go` — `//go:embed templates` con 3 archivos (8 KB)
- `adapters/_template/paths.go` — funciones de paths
- `adapters/_template/installer.go` — funciones de estrategia duplicadas
- `adapters/_template/install.go` — lógica de install

**Qué hacer**:
1. Agregar `//go:build ignore` como primera línea en cada archivo `.go` dentro de `adapters/_template/` (5 archivos: `adapter.go`, `embed.go`, `paths.go`, `installer.go`, `install.go`)
2. Verificar que `go build ./...` ya no compila el paquete:
   ```
   go build ./...
   ```
3. Confirmar que el adapter "Template Tool" NO aparece en `DefaultRegistry` al iniciar Sequoia

**Impacto esperado**:
- Binario de producción 8 KB más chico
- Cero riesgo de que "Template Tool" aparezca en `sequoia status`
- 20 TODOs eliminados del scope de linting en producción

**Dependencias**: Ninguna

**Riesgo de implementación**: Bajo
`//go:build ignore` es una directiva estándar de Go. Ningún código en producción importa `_template`. Si el `CONTRIBUTING.md` referencia el directorio, actualizar la ruta.

**Criterio de aceptación**:
- [ ] `go build ./...` compila sin incluir `adapters/_template/`
- [ ] `go build -o sequoia.exe ./cmd/sequoia` produce un binario sin el adapter template
- [ ] `CONTRIBUTING.md` sigue siendo correcto (actualizar si referenciaba `_template/`)

**Verificación**:
```powershell
go build ./...
# Debe compilar sin errores. Verificar que no hay output de _template:
go list ./adapters/...
# No debe listar adapters/_template
```

---

### [FIX-002] · Overhaul de CI: version, linting, coverage, dependencias

**Prioridad**: 🔴 Bloqueante
**Fase origen**: P6 Operations + P4 Quality
**Hallazgo(s) origen**: P6-001, P6-002, P6-003, P6-004, P6-006, P6-007, P4-004, P4-008 (causa raíz R2)

**Contexto mínimo**:
El CI tiene 4 problemas encadenados: (1) `ci.yml` usa Go 1.22 pero `go.mod` declara 1.24.2 → toolchain auto-download oculto en cada run, (2) `.golangci.yaml` con 10 linters nunca se ejecuta en CI, (3) la cobertura de tests nunca se mide — el claim "90%+ coverage" del README es estático y no verificable, (4) no hay Dependabot → las dependencias envejecen sin alertas. Además, `go.mod` declara `go 1.24.2` pero el código solo usa features de Go 1.19+ y el README dice "mín Go 1.22".

**Archivos involucrados**:
- `.github/workflows/ci.yml:25` — `go-version: '1.22'` (fuera de sync con go.mod)
- `action.yml:55` — `go-version: '1.24'` (hardcodeado, debería usar go-version-file)
- `go.mod:3` — `go 1.24.2` (innecesariamente restrictivo)
- `.golangci.yaml` — 10 linters configurados, nunca ejecutados
- `README.md:13` — claim estático "90%+ code coverage"
- `.github/dependabot.yml` — no existe

**Qué hacer**:
1. **Fix go.mod version**: Cambiar `go 1.24.2` a `go 1.22.0` en `go.mod:3`. Ejecutar `go mod tidy -go=1.22`.
2. **Fix CI Go version**: En `.github/workflows/ci.yml:25`, reemplazar `go-version: '1.22'` por `go-version-file: go.mod`. Hacer lo mismo en `action.yml:55`.
3. **Agregar golangci-lint**: En `ci.yml`, después del step `go vet`, agregar:
   ```yaml
   - name: Lint
     if: runner.os == 'Linux'
     uses: golangci/golangci-lint-action@v6
     with:
       version: v2
   ```
4. **Agregar coverage**: Cambiar el step `go test` en `ci.yml` de `go test ./... -count=1 -timeout 120s` a `go test -coverprofile=coverage.out -covermode=atomic ./...`. Agregar step de upload:
   ```yaml
   - name: Upload coverage
     uses: actions/upload-artifact@v4
     with:
       name: coverage-${{ matrix.os }}
       path: coverage.out
   ```
5. **Crear Dependabot**: Crear `.github/dependabot.yml`:
   ```yaml
   version: 2
   updates:
     - package-ecosystem: gomod
       directory: "/"
       schedule: { interval: weekly }
       open-pull-requests-limit: 5
       commit-message: { prefix: "chore(deps)" }
     - package-ecosystem: github-actions
       directory: "/"
       schedule: { interval: weekly }
       open-pull-requests-limit: 3
       commit-message: { prefix: "chore(ci)" }
   ```
6. **Fix CI badge**: En `README.md`, reemplazar el badge estático por:
   ```
   ![CI](https://github.com/Crisbr10/sequoia/actions/workflows/ci.yml/badge.svg)
   ```
7. Ejecutar `golangci-lint run` localmente, fixear cualquier violación pre-existente, commitear los fixes.

**Impacto esperado**:
- CI ~30-60s más rápido por trabajo (sin toolchain download oculto)
- Linting automatizado → no más `gofmt`/`unused`/`misspell` en PRs
- Coverage medible → el claim "90%+" se vuelve verificable
- Dependabot PRs automáticos → CVEs y updates surfaced sin esfuerzo manual
- `go.mod` compatible con Go 1.22+ → más usuarios pueden buildear

**Dependencias**: Ninguna. FIX-002 es independiente.

**Riesgo de implementación**: Medio
`golangci-lint` puede revelar violaciones pre-existentes. Ejecutar primero localmente, fixear todo, commitear fixes en un PR separado antes de agregar el step de CI. Dependabot puede generar varios PRs iniciales — configurar `open-pull-requests-limit: 3` mitiga esto.

**Criterio de aceptación**:
- [ ] `go.mod` declara `go 1.22` y `go mod tidy` no introduce cambios
- [ ] `ci.yml` y `action.yml` usan `go-version-file: go.mod`
- [ ] CI incluye step de `golangci-lint` (ubuntu) y pasa
- [ ] CI incluye `-coverprofile` y upload del artifact
- [ ] `.github/dependabot.yml` existe y está enabled en el repo
- [ ] README muestra badge de CI dinámico (verde/rojo según estado real)
- [ ] `go test ./...` pasa en Go 1.22, 1.23, y 1.24

**Verificación**:
```powershell
# Verificar go.mod
go mod tidy
# Verificar CI localmente (si tenés act instalado)
act push
# O revisar en GitHub Actions después del push
```

---

### [FIX-003] · Extraer BaseAdapter + strategies compartidas + comandos unificados

**Prioridad**: 🔴 Bloqueante
**Fase origen**: P3 Architecture + P4 Quality + P2 Performance
**Hallazgo(s) origen**: P3-001, P3-002, P4-001, P2-001, P4-005 (causa raíz R1)

**Contexto mínimo**:
5 adaptadores (claude, opencode, cursor, gemini, codex) duplican 70-85% de su código (~700 líneas de boilerplate). Los métodos `Install()`, `Uninstall()`, y `Status()` siguen el mismo patrón de 8 pasos en todos. Las funciones `InjectSection`/`RemoveSection` son byte-idénticas en claude+gemini; `GenerateRulesMD`/`RemoveRulesMD` son byte-idénticas en opencode+cursor+template. Además, 5 archivos de comandos Markdown (`sequoia-init.md`, `sequoia-audit.md`, etc.) están duplicados en 6 `embed.FS` (~125 KB). También `runInstallSteps` y `runUninstallSteps` en `pipeline/runner.go` son 95% idénticas (85 líneas duplicadas).

**Archivos involucrados**:
- `adapters/{claude,opencode,cursor,gemini,codex}/adapter.go` — métodos `Install()`, `Uninstall()`, `Status()` duplicados
- `adapters/{claude,gemini}/installer.go` — `InjectSection`/`RemoveSection` byte-idénticas
- `adapters/{opencode,cursor}/installer.go` — `GenerateRulesMD`/`RemoveRulesMD` byte-idénticas
- `adapters/{claude,opencode,cursor,gemini,codex}/embed.go` — `//go:embed templates` con comandos duplicados
- `adapters/*/templates/commands/` — 5 archivos duplicados 5-6 veces
- `adapters/common/` — destino para el nuevo código compartido
- `internal/pipeline/runner.go:75-130, 161-216` — `runInstallSteps`/`runUninstallSteps` duplicadas

**Qué hacer**:

**Parte A — Mover strategies a common (2h)**:
1. Crear `adapters/common/strategy.go` con 4 funciones exportadas:
   - `InjectMarkdownSection(path, content string) error` — copiar cuerpo de `claude/installer.go:InjectSection`
   - `RemoveMarkdownSection(path string) error` — copiar cuerpo de `claude/installer.go:RemoveSection`
   - `ReplaceFile(path, content string) error` — copiar cuerpo de `opencode/installer.go:GenerateAgentsMD`
   - `RestoreOrRemoveFile(path string) error` — copiar cuerpo de `opencode/installer.go:RemoveAgentsMD`
   - Mover constantes `markerStart`, `markerEnd`, `sequoiaMarker` a `strategy.go`
2. Crear `adapters/common/strategy_test.go` con los tests existentes de los archivos `installer_test.go` movidos y adaptados.
3. En cada adapter, reemplazar las llamadas a funciones locales por llamadas a `common.InjectMarkdownSection(...)`, etc.
4. Eliminar `adapters/{claude,gemini,opencode,cursor,codex}/installer.go`.
5. Ejecutar `go test ./adapters/...` — todos los tests deben pasar.

**Parte B — Extraer BaseAdapter (4h)**:
1. Crear `adapters/common/base_adapter.go` con un struct `BaseAdapter`:
   ```go
   type BaseAdapter struct {
       ID             string
       Name           string
       homeDir        string
       promptStrategy adapters.PromptStrategy
   }
   ```
2. Implementar métodos compartidos en `BaseAdapter`:
   - `Install(opts adapters.InstallOpts) error` — el flujo de 8 pasos común
   - `Uninstall(opts adapters.InstallOpts) error` — el flujo de eliminación común
   - `Status() (adapters.AdapterStatus, error)` — el Status común (100% idéntico hoy)
   - `SkillsPath()`, `CommandsPath()`, `SystemPromptPath()`, `IsInstalled()` — delegar a métodos que cada adapter concreto overridea
3. Definir una interfaz `AdapterCustomizer` con los métodos que cada adapter debe implementar:
   ```go
   type AdapterCustomizer interface {
       WriteSystemPrompt(base string) error
       SkillsDir(base string) string
       CommandsDir(base string) string
       // etc.
   }
   ```
4. Refactorizar cada adapter concreto para embeber `BaseAdapter`:
   ```go
   type Adapter struct {
       common.BaseAdapter
   }
   ```
   Cada adapter solo implementa `ID()`, `Name()`, `Detect()`, `PromptStrategy()`, y los métodos de `AdapterCustomizer`.
5. Verificar que `ToolAdapter` interface sigue satisfecha. Cada adapter concreto debe reducirse a ≤80 líneas (de ~200-250 actuales).
6. Ejecutar `go test ./adapters/...` — todos los tests existentes deben pasar sin modificación.

**Parte C — Unificar comandos embebidos (2h)**:
1. Crear `adapters/common/templates/commands/` y copiar los 5 archivos de comando (`sequoia-init.md`, `sequoia-audit.md`, `sequoia-diff.md`, `sequoia-fix.md`, `sequoia-review.md`) desde cualquier adapter.
2. Crear `adapters/common/embed.go`:
   ```go
   package common
   import "embed"
   //go:embed templates/commands
   var CommandFS embed.FS
   ```
3. En cada adapter, eliminar `adapters/{tool}/templates/commands/` y sus `embed.go` respectivos (o modificarlos para que solo embeban las skill templates específicas del adapter).
4. Actualizar todas las referencias de `templateFS` a `common.CommandFS` para los archivos de comandos.
5. Verificar que `go test ./adapters/...` pasa.

**Parte D — DRY pipeline runner (1h)**:
1. En `internal/pipeline/runner.go`, extraer una función `runSteps(ctx, t, ch, lang string, fn func(adapters.ToolAdapter, adapters.InstallOpts) error)`.
2. `runInstallSteps` y `runUninstallSteps` llaman a `runSteps` pasando `adapter.Install` o `adapter.Uninstall` como `fn`.
3. Verificar que `go test ./internal/pipeline/...` pasa.

**Impacto esperado**:
- ~700 líneas de código duplicado eliminadas
- Binario ~125 KB más chico (comandos unificados)
- Nuevo adapter requiere ≤50 líneas de código único (vs ~200 actuales)
- Bug fix en lógica de install se aplica en 1 lugar, no en 5

**Dependencias**: FIX-001 (`_template` excluido) corre antes para no distraer. Las partes A→B→C→D pueden hacerse en secuencia. FIX-005 y FIX-006 dependen de este (tocan los mismos archivos).

**Riesgo de implementación**: Medio
Refactor grande que toca 5 paquetes de producción. El riesgo se mitiga haciendo las partes en orden (A primero — es la de menor riesgo porque solo mueve código byte-idéntico) y corriendo tests después de cada parte. La interfaz `ToolAdapter` no cambia, así que los consumidores (pipeline, TUI) no se tocan.

**Criterio de aceptación**:
- [ ] `adapters/common/strategy.go` existe con 4 funciones + tests
- [ ] `adapters/{claude,gemini,opencode,cursor,codex}/installer.go` eliminados
- [ ] `adapters/common/base_adapter.go` existe con `Install()`, `Uninstall()`, `Status()`
- [ ] Cada adapter concreto ≤80 líneas
- [ ] `adapters/common/embed.go` exporta `CommandFS` con 5 archivos
- [ ] `adapters/*/templates/commands/` eliminado en los 5 adapters
- [ ] `internal/pipeline/runner.go` tiene una sola función `runSteps`
- [ ] `go test ./...` pasa completo
- [ ] `go build -o sequoia.exe ./cmd/sequoia` produce binario funcional

**Verificación**:
```powershell
go test ./adapters/...
go test ./internal/pipeline/...
go build -o sequoia.exe ./cmd/sequoia
./sequoia.exe status  # Debe mostrar los 5 adaptadores reales, sin regresión
```

---

### [FIX-004] · Agregar manejo de señales del SO (SIGTERM/SIGINT)

**Prioridad**: 🔴 Bloqueante
**Fase origen**: P1 Security
**Hallazgo(s) origen**: P1-002

**Contexto mínimo**:
Cuando el usuario presiona Ctrl+C o el proceso recibe SIGTERM durante una instalación, el pipeline no ejecuta rollback. Los archivos quedan a medio instalar en los directorios de configuración de los AI assistants, y el marker de versión puede estar ausente → `sequoia status` muestra "not installed" pero hay artifacts presentes. El pipeline ya soporta context cancellation (`runner.go:43-49`), pero nunca se conecta a señales del SO.

**Archivos involucrados**:
- `cmd/sequoia/main.go:56` — `root.Execute()` sin `signal.NotifyContext`
- `internal/pipeline/runner.go:43-49` — el contexto se cancela pero nunca por señal del SO

**Qué hacer**:
1. En `cmd/sequoia/main.go`, en la función `main()` o donde se llame a `root.Execute()`, envolver con signal handling:
   ```go
   ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
   defer stop()
   // Pasar ctx al pipeline (requiere modificar la firma de RunInstall/RunUninstall o usar context.Background con el ctx)
   ```
2. En `internal/pipeline/runner.go`, modificar `RunInstall` y `RunUninstall` para aceptar un `context.Context` externo (o crear el contexto interno a partir del externo).
3. En los métodos `Install()` de cada adapter, verificar `ctx.Err()` entre pasos para abortar limpiamente.
4. Agregar test: simular envío de señal durante install, verificar que se ejecuta rollback:
   ```go
   func TestInstall_RollbackOnCancel(t *testing.T) { ... }
   ```

**Impacto esperado**:
- Ctrl+C durante install → rollback automático, directorios limpios
- Sin artifacts residuales en `~/.claude/`, `~/.config/opencode/`, etc.

**Dependencias**: Ninguna técnica, pero conviene hacerlo después de FIX-003 (BaseAdapter) para no tener que implementar el chequeo de `ctx.Err()` en 5 archivos separados.

**Riesgo de implementación**: Bajo
Signal handling en Go es straightforward. En Windows, `syscall.SIGTERM` no existe — usar build tags (`//go:build !windows`) para SIGTERM y solo `os.Interrupt` en Windows.

**Criterio de aceptación**:
- [ ] Ctrl+C durante `sequoia install` ejecuta rollback
- [ ] SIGTERM ejecuta rollback (Unix)
- [ ] No quedan archivos residuales después de cancelación
- [ ] `go test ./cmd/...` incluye test de signal handling

**Verificación**:
```powershell
# Iniciar install y matar con Ctrl+C en ~2 segundos
./sequoia.exe install
# Verificar que no hay archivos de Sequoia en los directorios de config
./sequoia.exe status
```

---

### [FIX-005] · Fix colisión de nombres en archivos de backup

**Prioridad**: 🔴 Bloqueante
**Fase origen**: P1 Security
**Hallazgo(s) origen**: P1-001, P1-007

**Contexto mínimo**:
Tres adaptadores (OpenCode, Cursor, Codex) crean backups con `path + ".sequoia-backup"`. Si el usuario ya tiene un archivo llamado exactamente así, se sobrescribe silenciosamente. En uninstall, el backup restaurado reemplaza el archivo original del usuario con datos que nunca fueron suyos. Además, el directorio `.sequoia-backup` es predecible — un atacante local puede pre-crearlo con permisos restrictivos para causar denial-of-service. Claude y Gemini usan un directorio `.sequoia-backup/` que es mejor pero también predecible.

**Archivos involucrados**:
- `adapters/opencode/installer.go:36` — `backup := path + ".sequoia-backup"`
- `adapters/cursor/installer.go:36` — `backup := path + ".sequoia-backup"`
- `adapters/codex/installer.go:22` — `backupPath := path + ".sequoia-backup"`
- `adapters/common/installer.go:84` — `os.MkdirAll(backupPath, 0o755)` con path predecible

**Qué hacer**:
1. Para los adaptadores que usan `path + ".sequoia-backup"` (OpenCode, Cursor, Codex): cambiar a `path + ".sequoia-backup-" + timestamp` donde timestamp es `time.Now().UnixMilli()`.
2. Para el `common.Installer` (usado por Claude, Gemini): cambiar `backupPath()` para usar `filepath.Join(base, ".sequoia-backup-"+randomSuffix)` donde randomSuffix es `strconv.FormatInt(time.Now().UnixNano(), 36)`.
3. En `Uninstall()`, solo restaurar backups cuyo timestamp coincida con la sesión de install actual (guardar el timestamp en el marker de versión o en un archivo `.sequoia-session`).
4. Agregar test: pre-crear un archivo `.sequoia-backup`, ejecutar install, verificar que el archivo pre-existente no se sobrescribe.

**Impacto esperado**:
- Cero riesgo de colisión con archivos de usuario
- Cero riesgo de DoS por directorio pre-creado
- Backups trazables a la sesión de install que los creó

**Dependencias**: FIX-003 (BaseAdapter) — es más eficiente implementar esto en `BaseAdapter.Install()` una sola vez que en 5 archivos separados.

**Riesgo de implementación**: Bajo
Cambio localizado en la lógica de backup. Si FIX-003 ya está hecho, se modifica solo `BaseAdapter` y `common/Installer`.

**Criterio de aceptación**:
- [ ] Nombres de archivo/directorio de backup incluyen componente único (timestamp)
- [ ] Archivo pre-existente `*.sequoia-backup` no se sobrescribe
- [ ] Uninstall solo restaura backups de la sesión actual
- [ ] Test: pre-crear `.sequoia-backup`, instalar, verificar preservación

**Verificación**:
```powershell
go test ./adapters/... -run TestBackup -v
```

---

### [FIX-006] · Colectar y reportar errores de uninstall + sentinel errors

**Prioridad**: 🔴 Bloqueante
**Fase origen**: P1 Security + P3 Architecture
**Hallazgo(s) origen**: P1-004, P3-010 (causa raíz R4 parcial)

**Contexto mínimo**:
Los 5 adaptadores usan `_ = os.Remove(...)` en `Uninstall()`, descartando silenciosamente cualquier error. Si un archivo está lockeado (Windows) o tiene permisos insuficientes, el uninstall reporta éxito pero deja archivos. El usuario ve "Done" pero `sequoia status` sigue mostrando "installed". Además, todo el código usa un solo sentinel error (`ErrUnknownAdapter`) — no hay forma de distinguir un error de permisos de un error de template.

**Archivos involucrados**:
- `adapters/{claude,opencode,cursor,gemini,codex}/adapter.go` — métodos `Uninstall()` con `_ = os.Remove(...)`
- `adapters/errors.go` — solo `ErrUnknownAdapter`
- `internal/pipeline/runner.go` — manejo de errores genérico
- `cmd/sequoia/main.go` — display de errores al usuario

**Qué hacer**:
1. En cada `Uninstall()`, reemplazar `_ = os.Remove(...)` por:
   ```go
   var errs []error
   if err := os.Remove(path); err != nil {
       errs = append(errs, fmt.Errorf("remove %s: %w", path, err))
   }
   // al final:
   return errors.Join(errs...)
   ```
   Si FIX-003 (BaseAdapter) ya está implementado, hacer esto una sola vez en `BaseAdapter.Uninstall()`.
2. Agregar sentinel errors en `adapters/errors.go`:
   ```go
   var (
       ErrUnknownAdapter  = errors.New("unknown adapter")
       ErrInstallFailed   = errors.New("install failed")
       ErrUninstallFailed = errors.New("uninstall failed")
       ErrNotDetected     = errors.New("adapter not detected")
   )
   ```
3. Wrappear errores en `Install()` y `Uninstall()`:
   ```go
   return fmt.Errorf("%w: %w", ErrInstallFailed, err)
   ```
4. En `cmd/sequoia/main.go`, usar `errors.Is()` para decidir exit codes y mensajes.
5. En la TUI, mostrar warning cuando `Uninstall()` retorna error parcial: "Uninstall completed with warnings — 2 files could not be removed".

**Impacto esperado**:
- Uninstall reporta exactamente qué archivos no se pudieron eliminar
- CLI puede distinguir tipos de error y dar mensajes apropiados
- TUI muestra advertencias en vez de éxito falso

**Dependencias**: FIX-003 (BaseAdapter) reduce el trabajo de 5 archivos a 1.

**Riesgo de implementación**: Bajo
Cambio puramente aditivo en el manejo de errores. `errors.Join` está disponible desde Go 1.20 (el proyecto ya usa ≥1.22).

**Criterio de aceptación**:
- [ ] `Uninstall()` retorna errores agregados, no los descarta
- [ ] `adapters/errors.go` contiene al menos 3 sentinel errors
- [ ] TUI muestra warnings de archivos no eliminados
- [ ] `go test ./adapters/...` incluye test con archivo read-only

**Verificación**:
```powershell
go test ./adapters/... -run TestUninstall -v
```

---

## 🟠 Alto Leverage

---

### [FIX-007] · Hacer mandatory la verificación de checksum en install scripts

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P1 Security
**Hallazgo(s) origen**: P1-003

**Contexto mínimo**:
`install.sh` y `install.ps1` intentan bajar `checksums.txt` para verificar el binario. Si la descarga falla (red, rate limiting, 404), el script muestra un warning y continúa sin verificación. Un atacante que controle la red puede bloquear `checksums.txt` (más fácil que modificar el binario) y servir un binario malicioso que pasa todas las verificaciones porque fueron saltadas.

**Archivos involucrados**:
- `scripts/install.sh:220-248` — fallback silencioso con `|| true`
- `scripts/install.ps1:165` — `Write-Warn "Skipping checksum verification"`

**Qué hacer**:
1. En `install.sh`, reemplazar `|| true` en la descarga de checksums por lógica de abort:
   ```bash
   if ! curl -sSL --retry 3 "$CHECKSUMS_URL" -o "$CHECKSUMS_FILE"; then
       log_error "Could not download checksums. Aborting."
       exit 2  # EXIT_CHECKSUM
   fi
   ```
2. Agregar flag `--skip-checksums` que permite continuar sin verificación (para entornos air-gapped).
3. En `install.ps1`, mismo cambio: abortar en vez de warn, agregar `-SkipChecksum` switch ya existente pero documentarlo como opt-in consciente.
4. Aplicar el mismo retry logic (`--retry 3` o `-Retry 3`) que ya se usa para la descarga del binario.

**Impacto esperado**:
- Instalación aborta si no se puede verificar el binario (en vez de instalar sin verificar)
- Usuarios en entornos sin internet pueden usar `--skip-checksums` explícitamente

**Dependencias**: Ninguna

**Riesgo de implementación**: Bajo
Cambio de política en scripts de shell. El retry logic ya existe para el binario — solo hay que aplicarlo a checksums también.

**Criterio de aceptación**:
- [ ] Fallo en descarga de checksums.txt → aborta con exit code 2
- [ ] Flag `--skip-checksums` permite continuar
- [ ] Ambos scripts (`install.sh`, `install.ps1`) actualizados
- [ ] Retry logic (3 intentos) aplicado a descarga de checksums

**Verificación**:
```powershell
# Simular fallo de red
./scripts/install.sh --checksums-url=http://localhost:1/fake
# Debe abortar con exit code 2
```

---

### [FIX-008] · Mover detección de Engram a async en startup del TUI

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P2 Performance
**Hallazgo(s) origen**: P2-002

**Contexto mínimo**:
`NewModel()` en `internal/app/model.go:88` llama a `exec.LookPath("engram")` sincrónicamente durante la construcción del Model. Esto bloquea el primer render del TUI. `EngramAvailable` solo se usa en la pantalla de Configuration — nunca en Welcome, ToolSelection, o InstallProgress. En Windows con PATH largos, `exec.LookPath` puede tomar >100ms.

**Archivos involucrados**:
- `internal/app/model.go:88` — `exec.LookPath("engram")` en `NewModel()`

**Qué hacer**:
1. En `NewModel()`, eliminar la llamada a `exec.LookPath("engram")`. Inicializar `EngramAvailable: false`.
2. Agregar un nuevo mensaje Bubbletea `type EngramDetectedMsg bool`.
3. En `Model.Init()`, agregar un comando que lance una goroutine:
   ```go
   func detectEngram() tea.Msg {
       _, err := exec.LookPath("engram")
       return EngramDetectedMsg(err == nil)
   }
   ```
4. En `Update()`, agregar un case para `EngramDetectedMsg` que actualice `m.EngramAvailable`.
5. La pantalla de Configuration ya maneja correctamente `EngramAvailable=false` (muestra la opción en gris).

**Impacto esperado**:
- Arranque del TUI ~10-100ms más rápido
- Usuario ve la pantalla Welcome inmediatamente, sin bloqueo

**Dependencias**: Ninguna

**Riesgo de implementación**: Bajo
Cambio localizado en `model.go`. El mensaje asincrónico es un patrón estándar de Bubbletea.

**Criterio de aceptación**:
- [ ] `NewModel()` no llama a `exec.LookPath`
- [ ] `EngramAvailable` se actualiza asincrónicamente después del primer render
- [ ] `go test ./internal/app/...` pasa

**Verificación**:
```powershell
go test ./internal/app/... -v
```

---

### [FIX-009] · Cachear parseo de templates

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P2 Performance
**Hallazgo(s) origen**: P2-003

**Contexto mínimo**:
`RenderTemplate` en `adapters/common/template.go` lee el archivo embebido, lo parsea con `text/template`, lo ejecuta, y descarta el template parseado. La skill template de OpenCode mide 81 KB — cada install de OpenCode re-parsea 81 KB. Para 3 adaptadores, son 6-9 parseos. Los templates son inmutables (embebidos en el binario) — parsearlos una vez y cachearlos es trivial.

**Archivos involucrados**:
- `adapters/common/template.go:13-27` — `RenderTemplate` sin caché

**Qué hacer**:
1. Agregar un `sync.Map` a nivel de paquete en `template.go`:
   ```go
   var templateCache sync.Map
   ```
2. En `RenderTemplate`, antes de parsear, buscar en el caché por `(fs, name)`:
   ```go
   key := fmt.Sprintf("%p:%s", fs, name)  // %p da la dirección del FS como identificador único
   if cached, ok := templateCache.Load(key); ok {
       tmpl := cached.(*template.Template)
       return tmpl.Execute(buf, data)
   }
   ```
3. Si no está en caché, parsear, guardar, y ejecutar.
4. Agregar benchmark test:
   ```go
   func BenchmarkRenderTemplate(b *testing.B) { ... }
   ```

**Impacto esperado**:
- Segunda llamada a `RenderTemplate` con el mismo archivo es ≥5× más rápida
- ~50-150ms ahorrados en instalación multi-adapter

**Dependencias**: Ninguna

**Riesgo de implementación**: Bajo
`text/template.Template` es seguro para uso concurrente. `embed.FS` es inmutable. El caché es interno y no cambia la semántica de `RenderTemplate`.

**Criterio de aceptación**:
- [ ] `RenderTemplate` cachea templates parseados
- [ ] Segunda llamada no re-parsea
- [ ] Tests existentes de templates pasan
- [ ] Benchmark muestra speedup ≥5×

**Verificación**:
```powershell
go test ./adapters/common/... -bench=BenchmarkRenderTemplate -benchtime=1s
```

---

### [FIX-010] · Romper dependencia internal/model → adapters

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P3 Architecture
**Hallazgo(s) origen**: P3-003

**Contexto mínimo**:
`internal/model/types.go` importa `adapters` para usar `ToolAdapter` en `ToolState.Adapter`. Esto viola la convención Go de que `internal/` no debe depender de paquetes no-internal. Si `ToolAdapter` cambia su interfaz, `internal/model` debe recompilarse aunque `ToolState` solo use la interfaz estructuralmente.

**Archivos involucrados**:
- `internal/model/types.go:6` — `import "...sequoia/adapters"`
- `internal/model/types.go:36` — `ToolState.Adapter adapters.ToolAdapter`
- `internal/app/model.go` — construye `ToolState` con adapters del registry

**Qué hacer**:
1. Definir una interfaz `ToolInfo` en `internal/model/types.go` con solo los métodos que `ToolState` necesita:
   ```go
   type ToolInfo interface {
       ID() string
       Name() string
       IsInstalled() bool
       Status() (adapters.AdapterStatus, error)
       Detect() bool
   }
   ```
   Nota: `AdapterStatus` también está en `adapters`. Para romper la dependencia completamente, definir `ToolStatus` en `internal/model/` o mover `AdapterStatus` a `adapters/common/`.
2. Cambiar `ToolState.Adapter` de `adapters.ToolAdapter` a `ToolInfo`.
3. En `internal/app/model.go`, al construir `ToolState`, hacer type assertion o wrapping: `ToolState{Adapter: adapter}` (si `adapter` ya satisface `ToolInfo`).
4. Verificar que `internal/model/types.go` ya no tiene imports de `adapters`.

**Impacto esperado**:
- `internal/model` es verdaderamente interno, sin dependencias externas
- Cambios en `ToolAdapter` no fuerzan recompilación de `internal/model`

**Dependencias**: Ninguna, pero FIX-003 (BaseAdapter) no interfiere.

**Riesgo de implementación**: Bajo
Interfaz subset — `ToolAdapter` ya satisface `ToolInfo`. Solo cambia el tipo del campo.

**Criterio de aceptación**:
- [ ] `internal/model/types.go` no importa `adapters`
- [ ] `ToolInfo` interfaz definida con métodos necesarios
- [ ] `go build ./...` compila sin errores
- [ ] `go test ./internal/model/...` pasa

**Verificación**:
```powershell
go test ./internal/model/...
go test ./internal/app/...
```

---

### [FIX-011] · Simplificar pipeline a 1 paso real (eliminar pasos cosméticos)

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P3 Architecture + P2 Performance
**Hallazgo(s) origen**: P2-007, P3-006, P3-012 (causa raíz R5)

**Contexto mínimo**:
El pipeline envía 6 mensajes de progreso por adapter (3 "running" + 3 "done") para los pasos "Skills", "Commands", "System Prompt", pero `adapter.Install()` es una sola llamada monolítica. Los 3 pasos siempre completan atómicamente — el usuario ve todos flickear de pending→running→done instantáneamente. Es progreso cosmético que no refleja la realidad. Además, los nombres de paso están duplicados en `runner.go:18` y `update.go:302`. Y `spinnerFrames` es un slice de 10 strings nunca usado (el spinner es estático).

**Archivos involucrados**:
- `internal/pipeline/runner.go:18` — `defaultStepNames` (3 pasos)
- `internal/pipeline/runner.go:75-130` — `runInstallSteps` con 3 pasos
- `internal/app/update.go:302,327` — `buildProgressTools` duplica los step names
- `internal/tui/screens/install-progress.go:48` — `spinnerFrames` nunca animado

**Qué hacer**:
1. Reducir `defaultStepNames` a `[]string{"Installing"}`. Un solo paso por adapter.
2. Simplificar `runInstallSteps` para enviar 1 mensaje "running" antes de `adapter.Install()` y 1 mensaje "done"/"error" después.
3. Eliminar la duplicación de step names en `update.go:302,327` — referenciar `pipeline.InstallSteps` (exportado) en vez de redefinir.
4. Eliminar `spinnerFrames` (slice de 10 strings no usado) de `install-progress.go:48`.
5. Si se quiere animación real, implementar con `tea.Tick` en `Model.Init()`. Si no, mantener indicador estático "⠋" (suficiente para una CLI tool).

**Impacto esperado**:
- UX honesta: 1 paso real = 1 indicador de progreso
- ~30 líneas de código eliminadas (spinner muerto + step names duplicados)
- Progress screen más simple de mantener

**Dependencias**: FIX-003 (BaseAdapter + pipeline DRY) ya reduce la duplicación en el pipeline.

**Riesgo de implementación**: Bajo
Simplificación — se elimina código, no se agrega. Los tests de progreso deben actualizarse para esperar 2 mensajes en vez de 6.

**Criterio de aceptación**:
- [ ] Pipeline envía 2 mensajes por tool (running + done/error), no 6
- [ ] `defaultStepNames` reducido a 1 elemento o eliminado
- [ ] `update.go` no duplica los nombres de paso
- [ ] `spinnerFrames` eliminado o implementado con animación real
- [ ] `go test ./internal/pipeline/...` pasa
- [ ] `go test ./internal/tui/screens/...` pasa

**Verificación**:
```powershell
go test ./internal/pipeline/... -v
go test ./internal/app/... -v
```

---

### [FIX-012] · Resolver dead code del ScreenRouter

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P3 Architecture
**Hallazgo(s) origen**: P3-004

**Contexto mínimo**:
`internal/tui/router.go` define 151 líneas de infraestructura de ruteo (`TransitionMap`, `ScreenRouter` interface, `router` implementation, `IsValidTransition`, `NextScreen`, `NewRouter`) que no son usadas por ninguna parte del código de producción. `app.Model` maneja transiciones inline en `update.go` con switch-case por screen. Esto crea dos fuentes de verdad para las transiciones de pantalla.

**Archivos involucrados**:
- `internal/tui/router.go` — 151 líneas (interface + implementation + transition map)
- `internal/app/update.go:51-67` — transiciones inline que duplican la lógica del router

**Qué hacer**:
**Opción A (recomendada — eliminar)**:
1. Eliminar `TransitionMap`, `ScreenRouter`, `IsValidTransition`, `NextScreen`, `NewRouter` de `router.go`.
2. Conservar `NavigateMsg` (sí se usa en `update.go` y `view.go`).
3. Conservar `CurrentScreenName()` si se usa en tests.

**Opción B (wirear)**:
1. Agregar un `ScreenRouter` al `Model`.
2. Reemplazar las transiciones inline en `update.go` por llamadas a `m.router.NavigateTo(...)`.
3. Esto requiere refactorizar 389 líneas de `update.go` — mayor riesgo.

**Impacto esperado**:
- 151 líneas de dead code eliminadas (Opción A)
- Single source of truth para navegación
- Menos confusión para nuevos contributors

**Dependencias**: Ninguna

**Riesgo de implementación**: Bajo (Opción A) / Medio (Opción B)
Opción A es pura eliminación de código no referenciado. Opción B toca la lógica de navegación.

**Criterio de aceptación**:
- [ ] No hay lógica de transición duplicada entre `router.go` y `update.go`
- [ ] `NavigateMsg` se conserva y funciona igual
- [ ] `go test ./internal/tui/...` pasa
- [ ] `go test ./internal/app/...` pasa

**Verificación**:
```powershell
go test ./internal/tui/... -v
go test ./internal/app/... -v
```

---

### [FIX-013] · Agregar firma de artifacts con cosign

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P6 Operations
**Hallazgo(s) origen**: P6-005

**Contexto mínimo**:
Los usuarios descargan binarios de GitHub Releases y verifican integridad con SHA-256, pero no pueden verificar autenticidad (quién buildéo). Un atacante que comprometa el repo podría reemplazar binario + checksums. Cosign keyless signing usa OIDC de GitHub Actions para firmar sin manejar claves.

**Archivos involucrados**:
- `.goreleaser.yaml` — sin sección `signs`
- `.github/workflows/release.yml` — sin `id-token: write`

**Qué hacer**:
1. Agregar sección `signs` a `.goreleaser.yaml`:
   ```yaml
   signs:
     - cmd: cosign
       args:
         - sign-blob
         - --output-signature=${signature}
         - --yes
         - ${artifact}
       artifacts: all
   ```
2. En `release.yml`, agregar permisos:
   ```yaml
   permissions:
     id-token: write
     contents: write
   ```
3. Agregar nota en README con comando de verificación:
   ```bash
   cosign verify-blob --signature sequoia.sig --certificate-identity ... sequoia
   ```

**Impacto esperado**:
- Usuarios pueden verificar que el binario fue buildedo por CI oficial
- Cumplimiento con SLSA Level 2+

**Dependencias**: Ninguna

**Riesgo de implementación**: Medio
Requiere permisos OIDC y entender el flow de cosign. GoReleaser tiene docs claros para esto.

**Criterio de aceptación**:
- [ ] `.goreleaser.yaml` incluye `signs` con cosign
- [ ] Release assets incluyen `.sig` o bundle de Sigstore
- [ ] README incluye comando de verificación cosign

**Verificación**:
```powershell
# Hacer un release de prueba (pre-release tag) y verificar los assets
```

---

### [FIX-014] · Wirear language selector o esconderlo hasta que i18n esté listo

**Prioridad**: 🟠 Alto leverage
**Fase origen**: P7 i18n + P4 Quality + P1 Security
**Hallazgo(s) origen**: P1-006, P4-003, P7-001, P7-002 (causa raíz R3 parcial — solo la parte de wiring)

**Contexto mínimo**:
La TUI tiene un selector de idioma (EN/ES) en la pantalla de Configuration, y `InstallOpts.Language` se pasa por todo el pipeline. Pero los 5 adaptadores lo descartan con `_ = opts.Language`. Los usuarios que seleccionan "Español" ven todo en inglés. Hay dos caminos: wirear el selector a traducciones reales (requiere FIX-015 primero), o esconder el selector hasta que i18n esté implementado.

**Archivos involucrados**:
- `internal/tui/screens/configuration.go:14-19` — selector EN/ES
- `adapters/{claude,opencode,cursor,gemini,codex}/adapter.go` — `_ = opts.Language`
- `internal/pipeline/runner.go:78` — `opts := adapters.InstallOpts{Language: lang}`
- `internal/model/types.go` — `TUIConfig.Language`

**Qué hacer**:
**Opción recomendada (corto plazo)**: Esconder el selector agregando un feature flag o comentándolo en la UI, con un mensaje "i18n coming soon". Mantener `InstallOpts.Language` en la interfaz (ya está cableado, solo falta la implementación).

1. En `configuration.go`, wrappear el selector de idioma con:
   ```go
   // TODO(i18n): uncomment when translation catalog is implemented (FIX-015)
   // Language selector here
   ```
2. Agregar un comentario `// TODO(i18n): implement opts.Language` en cada adapter donde está `_ = opts.Language`.
3. Si se quiere wirear (requiere FIX-015 completado primero): en cada adapter, usar `opts.Language` para seleccionar `templates/{lang}/skill.md.tmpl` en vez de `templates/skill.md.tmpl`.

**Impacto esperado**:
- Usuarios no son engañados por un selector que no funciona
- Infraestructura de `Language` se conserva para cuando i18n esté lista

**Dependencias**: Para wirear de verdad, depende de FIX-015 (i18n catalog). Para esconder, no tiene dependencias.

**Riesgo de implementación**: Bajo (esconder) / Medio (wirear)
Esconder es comentar código. Wirear requiere traducciones reales.

**Criterio de aceptación**:
- [ ] Selector de idioma no se muestra si no hay traducciones implementadas
- [ ] `InstallOpts.Language` se conserva en la interfaz
- [ ] Cada adapter tiene `// TODO(i18n)` donde descarta `opts.Language`

**Verificación**:
```powershell
go build -o sequoia.exe ./cmd/sequoia
./sequoia.exe  # Navegar a Configuration, verificar que el selector no aparece o está deshabilitado
```

---

## 🟡 Backlog

---

### [FIX-015] · Agregar librería i18n y catálogo de mensajes

**Prioridad**: 🟡 Backlog
**Fase origen**: P7 i18n
**Hallazgo(s) origen**: P7-004, P7-005, P7-006, P7-003 (causa raíz R3 parcial)

**Contexto mínimo**:
El proyecto tiene cero infraestructura de i18n: sin librería, sin archivos de traducción, sin mecanismo de lookup. Las 51 strings de la TUI están hardcodeadas en inglés en 8 archivos. Los agentes instalados (SKILL.md) están en español, mientras que las secciones de system prompt están en inglés — una mezcla confusa. Para wirear el language selector (FIX-014), se necesita este catálogo primero.

**Archivos involucrados**:
- `go.mod` — agregar `github.com/nicksnyder/go-i18n/v2`
- `internal/i18n/` (nuevo directorio) — bundle + traducciones
- `internal/tui/screens/*.go` — 51 strings hardcodeadas
- `adapters/*/templates/` — crear `templates/en/` y mover `templates/` actual a `templates/es/`

**Qué hacer**:
1. Agregar `go-i18n/v2` a `go.mod`:
   ```
   go get github.com/nicksnyder/go-i18n/v2
   ```
2. Crear `internal/i18n/bundle.go` con inicialización del bundle y función `T(key, lang string) string`.
3. Crear `internal/i18n/translations/en.json` con las 51+ strings de la TUI.
4. Crear `internal/i18n/translations/es.json` con traducciones al español.
5. Reemplazar strings hardcodeadas en `screens/*.go` por llamadas a `i18n.T("welcome.menu.install", lang)`.
6. Crear `adapters/claude/templates/en/skill.md.tmpl` con traducción al inglés del prompt del orquestador (actualmente en español). Mover el actual a `templates/es/skill.md.tmpl`.
7. Repetir para cada adapter.
8. Modificar `RenderTemplate` para que acepte un parámetro `lang` y seleccione `templates/{lang}/` o haga fallback a `templates/en/`.

**Impacto esperado**:
- TUI cambia de idioma al seleccionar EN/ES
- Templates instalados en el idioma elegido
- Infraestructura lista para agregar más idiomas

**Dependencias**: FIX-014 (wirear selector) consume este catálogo.

**Riesgo de implementación**: Medio
~51 strings para traducir + templates de 176 líneas. Las traducciones deben ser nativas (no Google Translate). Requiere coordinar con alguien que hable español nativo para las traducciones inversas (los templates ya están en español, las strings de TUI necesitan traducción al español).

**Criterio de aceptación**:
- [ ] `go-i18n/v2` en `go.mod`
- [ ] `internal/i18n/translations/en.json` con todas las strings
- [ ] `internal/i18n/translations/es.json` con todas las strings
- [ ] `T("key", "es")` retorna string en español
- [ ] Templates en inglés disponibles como default

**Verificación**:
```powershell
go test ./internal/i18n/...
go build -o sequoia.exe ./cmd/sequoia
./sequoia.exe  # Seleccionar Español en Configuration, verificar TUI en español
```

---

### [FIX-016] · Agregar warnings para symlink resolution fallback

**Prioridad**: 🟡 Backlog
**Fase origen**: P1 Security
**Hallazgo(s) origen**: P1-005

**Contexto mínimo**:
Los 5 adaptadores usan `filepath.EvalSymlinks` con fallback silencioso al path sin resolver. En macOS con directorios syncados (iCloud/Dropbox), si el symlink target está temporalmente offline, Sequoia instala en el path del symlink (local), que luego es sobrescrito cuando el volumen se monta.

**Archivos involucrados**: `adapters/{claude,opencode,cursor,gemini,codex}/paths.go:20-24`

**Qué hacer**:
1. Usar `os.Lstat` para verificar si el path es symlink antes de resolver.
2. Si `EvalSymlinks` falla y el path ES un symlink, loguear warning en stderr/TUI.
3. Si no es symlink, continuar sin warning (fallback legítimo).

**Criterio de aceptación**: Warning visible en TUI status cuando se usa fallback.

**Verificación**: `go test ./adapters/... -run Symlink -v`

---

### [FIX-017] · Restringir permisos de archivos de backup

**Prioridad**: 🟡 Backlog
**Fase origen**: P1 Security
**Hallazgo(s) origen**: P1-008

**Contexto mínimo**:
Todos los archivos instalados usan `0o644` (world-readable). En sistemas Unix multi-usuario, otros usuarios pueden leer el contenido de los backups que incluyen configuraciones personalizadas del usuario.

**Archivos involucrados**: `adapters/common/files.go:15`, `adapters/*/installer.go` (todos los `os.WriteFile`)

**Qué hacer**:
1. Cambiar `0o644` a `0o600` para backups.
2. Cambiar `0o755` a `0o700` para directorios de backup.
3. Preservar permisos originales al restaurar (no hardcodear).

**Criterio de aceptación**: `ls -la ~/.claude/.sequoia-backup-*/` muestra `-rw-------`.

---

### [FIX-018] · Usar atomic writes (temp-then-rename) para backup+replace

**Prioridad**: 🟡 Backlog
**Fase origen**: P1 Security
**Hallazgo(s) origen**: P1-010

**Contexto mínimo**:
En Windows, `os.WriteFile` trunca y escribe en lugar — un crash durante escritura deja el archivo truncado. La solución: escribir a archivo temporal, luego `os.Rename` (atómico en el mismo volumen).

**Archivos involucrados**: `adapters/opencode/installer.go:41-44`, `adapters/cursor/installer.go:41-44`, `adapters/codex/installer.go:22-25`

**Qué hacer**:
1. Reemplazar `os.WriteFile(path, content, perm)` por:
   ```go
   tmp := path + ".tmp"
   os.WriteFile(tmp, content, perm)
   os.Rename(tmp, path)
   ```

**Criterio de aceptación**: Crash durante escritura no deja archivo truncado.

---

### [FIX-019] · Varios: quick wins acumulados (cache logo, cache home dir, mover debug.ReadBuildInfo)

**Prioridad**: 🟡 Backlog
**Fase origen**: P2 Performance + P4 Quality
**Hallazgo(s) origen**: P2-004, P2-005, P2-006, P4-006, P4-007, P4-009, P4-010, P4-012, P6-008

**Contexto mínimo**:
Varios hallazgos de bajo esfuerzo que pueden empaquetarse juntos:

| Sub-tarea | Hallazgo | Qué hacer | Esfuerzo |
|-----------|----------|-----------|----------|
| Cache `go-figure` logo | P2-005 | `sync.Once` en `logo.go` — el logo nunca cambia | 15 min |
| Cache `os.UserHomeDir()` | P2-006 | Campo `cachedHome` en adapter struct, lazy init | 30 min |
| Mover `debug.ReadBuildInfo` | P2-004 | De `init()` a `RunE` del comando `version` | 15 min |
| Shared mock adapter | P4-006 | `adapters/testutil/mock_adapter.go` con 4 test packages compartiendo | 1h |
| Fix recursion limit en tests | P4-007 | Loop explícito con `t.Fatalf` si se excede el límite | 30 min |
| Fix dead code guard | P4-009 | Eliminar `if len(defaultStepNames) > 0` o loguear error | 15 min |
| Mover `renderUninstallConfirm` | P4-010 | A `screens/uninstall.go`, exportar como `RenderConfirmPrompt` | 30 min |
| Eliminar test sin valor | P4-012 | `TestWaitForProgress_ContextCancellationIgnored` no testea nada real | 5 min |
| CHANGELOG vs GoReleaser | P6-008 | Elegir CHANGELOG.md como fuente canónica, cambiar `changelog.use: github` | 1h |

**Criterio de aceptación**: Todas las sub-tareas pasan sus tests específicos.

**Verificación**:
```powershell
go test ./...  # Todo verde
```

---

## Dependencias entre tareas

```
Bloqueantes (secuenciales donde se indica):
FIX-001 (_template exclude)     ← independiente, hacelo YA (5 min)
FIX-002 (CI overhaul)           ← independiente
FIX-003 (BaseAdapter)           ← independiente, pero FIX-005 y FIX-006 se benefician
  ├── FIX-005 (backup collision) ← más fácil después de FIX-003
  └── FIX-006 (uninstall errors) ← más fácil después de FIX-003
FIX-004 (signal handling)       ← independiente

Alto Leverage (todos independientes entre sí):
FIX-007 (checksum mandatory)    ← independiente
FIX-008 (async engram)          ← independiente
FIX-009 (template cache)        ← independiente
FIX-010 (model→adapters dep)    ← independiente
FIX-011 (pipeline simplify)     ← independiente (pero FIX-003 ayuda)
FIX-012 (ScreenRouter dead)     ← independiente
FIX-013 (artifact signing)      ← independiente
FIX-014 (language selector)     ← para wirear, depende de FIX-015

Backlog:
FIX-015 (i18n catalog)          ← habilita FIX-014
FIX-016 (symlink warnings)      ← independiente
FIX-017 (backup permissions)    ← independiente
FIX-018 (atomic writes)         ← independiente
FIX-019 (quick wins)            ← 9 micro-tareas independientes
```

---

## Estimación de riesgo global

| Fase | Tareas | Esfuerzo | Riesgo |
|------|--------|----------|--------|
| Bloqueantes (6) | FIX-001 a FIX-006 | ~14h | Medio — FIX-003 es el más grande pero se parte en A/B/C/D |
| Alto Leverage (8) | FIX-007 a FIX-014 | ~12h | Bajo — cambios localizados, bien acotados |
| Backlog (5) | FIX-015 a FIX-019 | ~12h | Bajo-Medio — FIX-015 requiere traducciones de calidad |

**Total**: 19 tareas · ~38h · Riesgo global: **Medio**

El riesgo principal está en FIX-003 (BaseAdapter) por tocar 5 paquetes simultáneamente. Se mitiga haciendo las partes A→B→C→D en secuencia con tests después de cada una. FIX-001 y FIX-002 son quick wins de muy bajo riesgo que se pueden mergear inmediatamente.

---

*Plan de implementación generado por /sequoia fix · Sequoia v0.1.0 · 2026-05-12*
