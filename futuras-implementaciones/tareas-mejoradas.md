# Formato de Tareas Mejorado

> Comparación del formato actual de `sequoia-tasks.md` vs el formato propuesto.
> Objetivo: que un dev pueda ejecutar la tarea sin volver a leer el reporte de auditoría.

---

## Diagnóstico del formato actual

El formato actual de `docs/sequoia/sequoia-tasks.md` es un **checklist funcional pero minimalista**:

```markdown
- [ ] **TASK-001**: Create `adapters/common/base.go` — `BaseResolver(...)` 
  to eliminate 5 identical `xxxBase()` functions
```

### Lo que le falta a un dev para ejecutarla sin fricción

| Falta | Problema que genera |
|-------|--------------------|
| **Contexto del problema** | El dev tiene que abrir el reporte de auditoría para entender por qué existe esta tarea |
| **Archivos específicos a editar** | "Eliminate 5 functions" — ¿en qué archivos? ¿Qué adapters? |
| **Before/After del código** | Sin ver el diff esperado, el dev puede implementar algo distinto a lo que el auditor imaginó |
| **Criterio de aceptación** | ¿Cómo verifico que la tarea está completa? ¿Qué tests deben pasar? |
| **Dependencias** | ¿Esta tarea bloquea a otras? ¿Debo hacerla antes o después de TASK-003? |
| **Labels** | Sin tags de severidad/categoría/esfuerzo/módulo, no se puede filtrar ni priorizar en un gestor de proyectos |
| **Tareas grandes sin dividir** | RC1 tiene 11 subtareas pero no están organizadas en milestones funcionales — el dev no sabe qué puede mergear primero |

---

## Comparación A/B: TASK-001

### ❌ Formato actual

```markdown
- [ ] **TASK-001**: Create `adapters/common/base.go` — `BaseResolver(homeDir, relativeDir string)` 
  to eliminate 5 identical `xxxBase()` functions
```

### ✅ Formato propuesto

```markdown
### TASK-001 · Extract BaseResolver to `adapters/common/`

**Severity**: 🔴 high &nbsp;|&nbsp; **Category**: architecture &nbsp;|&nbsp; **Effort**: 30min &nbsp;|&nbsp; **Module**: adapters/common  
**Fixes**: P3-002 &nbsp;|&nbsp; **Blocks**: TASK-003, TASK-007

---

**Why**  
Los 5 adapters (claude, opencode, cursor, gemini, codex) copian la misma lógica de 25 líneas
para resolver el home directory con fallback de symlinks. Solo cambia el path final (`.claude`,
`.config/opencode`, `.cursor/rules`, `.gemini`, `.codex`). Si `os.UserHomeDir()` cambia su
comportamiento en Go 2.0, hay que editar 5 archivos idénticos.

**Files**

| Acción | Archivo |
|--------|---------|
| ✨ NEW | `adapters/common/base.go` |
| ✏️ EDIT | `adapters/claude/paths.go:12-26` |
| ✏️ EDIT | `adapters/opencode/paths.go:12-26` |
| ✏️ EDIT | `adapters/cursor/paths.go:12-26` |
| ✏️ EDIT | `adapters/gemini/paths.go:12-25` |
| ✏️ EDIT | `adapters/codex/paths.go:12-26` |
| ✏️ EDIT | `adapters/_template/paths.go` (comentarios TODO) |

**Before → After**

```go
// BEFORE (copied 5 times — adapters/claude/paths.go:12-26)
func claudeBase(homeDir string) (string, error) {
    if homeDir == "" {
        var err error
        homeDir, err = os.UserHomeDir()
        if err != nil { return "", err }
    }
    resolved, err := filepath.EvalSymlinks(homeDir)
    if err != nil { resolved = homeDir }
    return filepath.Join(resolved, ".claude"), nil
}

// AFTER (once in common — adapters/common/base.go)
func BaseResolver(homeDir, relativeDir string) (string, error) {
    if homeDir == "" {
        var err error
        homeDir, err = os.UserHomeDir()
        if err != nil { return "", err }
    }
    resolved, err := filepath.EvalSymlinks(homeDir)
    if err != nil { resolved = homeDir }
    return filepath.Join(resolved, relativeDir), nil
}

// Usage in claude/paths.go:
func claudeBase(homeDir string) (string, error) {
    return common.BaseResolver(homeDir, ".claude")
}
```

**Acceptance Criteria**
- [ ] `BaseResolver("", ".claude")` resuelve a `~/.claude/` correctamente
- [ ] Comportamiento de fallback de symlinks preservado (si EvalSymlinks falla, usa path sin resolver)
- [ ] Los 5 adapters compilan y todos sus tests existentes pasan sin cambios
- [ ] El template `_template/paths.go` usa `BaseResolver` en vez de copiar la lógica
- [ ] `go vet ./...` limpio

**Dependencies**
```
TASK-001 (esta)
  ├── bloquea → TASK-003 (mover templateData a common)
  └── bloquea → TASK-007 (refactorizar 5 Install() con helpers comunes)
```
```

---

## Comparación A/B: Tarea grande dividida (RC1)

RC1 "Extract shared adapter boilerplate" tiene 11 subtareas (~5h). Sin división, no se puede mergear incrementalmente.

### ❌ Formato actual

```markdown
### 🔴 RC1: Extract shared adapter boilerplate into `adapters/common/`
**Effort**: 4–6h | **Fixes**: P3-001, P3-002, P3-004, P2-001, P2-002, P4-002, P4-003, P4-007

- [ ] **TASK-001**: Create `adapters/common/base.go` ...
- [ ] **TASK-002**: Create `adapters/common/adapter_installer.go` ...
- [ ] **TASK-003**: Move `templateData` struct to `adapters/common/` ...
- [ ] **TASK-004**: Centralize command template embedding ...
- [ ] **TASK-005**: Move `InjectSection`/`RemoveSection` ...
- [ ] **TASK-006**: Add `"Version\n"` trailing newline ...
- [ ] **TASK-007**: Refactor all 5 adapters to use the new common helpers
- [ ] **TASK-008**: Remove duplicated `embed.go` files ...
- [ ] **TASK-009**: Remove duplicated template directories ...
- [ ] **TASK-010**: Update `_template/adapter.go` ...
- [ ] **TASK-011**: Run full test suite ...
```

### ✅ Formato propuesto (3 milestones independientes)

```markdown
## Milestone A: Extract Path Resolution (1h, mergeable solo)

📦 **PR**: `refactor/extract-base-resolver`
✅ **Verificación**: todos los tests de paths.go pasan sin cambios

### TASK-A1 · Extract BaseResolver (30min)
[Formato completo como TASK-001 arriba]

### TASK-A2 · Refactor adapters to use BaseResolver (30min)
**Why**: Una vez que BaseResolver existe, eliminar las 5 copias.
**Files**: EDIT `adapters/{claude,opencode,cursor,gemini,codex}/paths.go` — cada `xxxBase()` pasa de 15 líneas a 3
**Acceptance**: tests pasan, `go vet` limpio, sin cambios de comportamiento


## Milestone B: Centralize Templates + Installer (2.5h, mergeable solo)

📦 **PR**: `refactor/centralize-templates-installer`
✅ **Verificación**: install en 5 adapters produce los mismos archivos que antes del refactor
🔗 **Depende de**: Milestone A (usa BaseResolver)

### TASK-B1 · Move templateData to common (20min)
**Why**: 4 de 5 adapters definen un struct idéntico. Solo Codex agrega campos extra.  
**Files**: NEW `adapters/common/template_data.go` | EDIT `adapters/{claude,opencode,cursor,gemini}/install.go`  
**Acceptance**: compila, Codex sigue funcionando con su struct extendido

### TASK-B2 · Centralize command templates in common/embed.go (1h)
**Why**: 5 archivos de comando × 5 adapters = 25 copias idénticas. 885 líneas duplicadas.  
**Files**: NEW `adapters/common/embed.go` con `//go:embed templates/commands` | DELETE `adapters/*/templates/commands/` | EDIT `adapters/*/adapter.go` para usar `common.CommandTemplates()`  
**Acceptance**: `sequoia install --no-tui` en cada adapter produce los mismos archivos de comando

### TASK-B3 · Move InjectSection/RemoveSection to common (45min)
**Why**: Claude y Gemini tienen 72 líneas idénticas de lógica de inyección de marcadores.  
**Files**: NEW `adapters/common/sections.go` | EDIT `adapters/claude/installer.go` | EDIT `adapters/gemini/installer.go` | DELETE `adapters/gemini/installer.go` entero si ya no tiene más lógica propia  
**Acceptance**: `sequoia install` en Claude y Gemini produce CLAUDE.md y GEMINI.md con los mismos marcadores


## Milestone C: Unified Installer Helper (1.5h, mergeable solo)

📦 **PR**: `refactor/unified-install-helper`
✅ **Verificación**: diff de archivos instalados antes/después es vacío para los 5 adapters
🔗 **Depende de**: Milestones A y B

### TASK-C1 · Create InstallSkills helper (1h)
**Why**: Los 5 métodos Install() comparten ~80 líneas idénticas. Solo la inyección de system prompt varía.  
**Files**: NEW `adapters/common/adapter_installer.go` — función `InstallSkills(cfg InstallConfig, injectSystemPrompt func() error) error` | EDIT `adapters/*/adapter.go` — cada Install() pasa de ~80 líneas a ~15  
**Acceptance**: filesystem diff vacío para install en los 5 adapters

### TASK-C2 · Fix cursor version newline (5min)
**Why**: 4 adapters escriben `Version + "\n"`, cursor omite el newline.  
**Files**: EDIT `adapters/cursor/adapter.go:197`  
**Acceptance**: `strings.TrimSpace()` lee la versión correctamente (ya lo hace, esto es preventivo)

### TASK-C3 · Cleanup duplicated embed.go and template dirs (15min)
**Why**: Después de C1 y B2, los `embed.go` y `templates/` de cada adapter son residuos.  
**Files**: DELETE `adapters/{claude,opencode,cursor,gemini,codex}/embed.go` | DELETE `adapters/{claude,opencode,cursor,gemini,codex}/templates/`  
**Acceptance**: compila, tests pasan, `go vet` limpio
```

---

## Template reutilizable para cualquier tarea

```markdown
### TASK-XXX · [título accionable en imperativo]

**Severity**: 🔴 critical | 🔴 high | 🟡 medium | 🟢 low | ℹ️ info
&nbsp;|&nbsp; **Category**: [security|performance|architecture|quality|i18n|...]
&nbsp;|&nbsp; **Effort**: [15min|30min|1h|2h|4h|8h|>8h]
&nbsp;|&nbsp; **Module**: [paquete o área del proyecto]
**Fixes**: [finding ID(s)] &nbsp;|&nbsp; **Blocks**: [task IDs] &nbsp;|&nbsp; **Blocked by**: [task IDs]

---

**Why**  
[2-4 oraciones explicando el problema y por qué esta tarea existe. Contexto suficiente 
para que un dev nuevo entienda sin leer el reporte de auditoría.]

**Files**

| Acción | Archivo |
|--------|---------|
| ✨ NEW | `path/to/new/file.go` |
| ✏️ EDIT | `path/to/existing.go:42-58` |
| 🗑️ DELETE | `path/to/removed.go` |

**Before → After** *(opcional — incluir si el cambio no es obvio)*

```lang
// BEFORE (qué hay hoy)
[código actual con el problema]

// AFTER (qué debería quedar)
[código esperado después del cambio]
```

**Acceptance Criteria**
- [ ] [Criterio verificable 1 — ej: "todos los tests existentes pasan sin cambios"]
- [ ] [Criterio verificable 2 — ej: "el nuevo adapter _template usa el helper sin copy-paste"]
- [ ] [Criterio verificable 3 — ej: "go vet ./... y golangci-lint run limpios"]
- [ ] [Criterio verificable 4 — ej: "diff de archivos instalados antes/después es vacío"]

**Dependencies**
```
TASK-XXX (esta)
  ├── depende de → TASK-YYY (porque necesita BaseResolver)
  └── bloquea → TASK-ZZZ (porque el installer unificado usa esto)
```
```

---

## Reglas de división de tareas

| Si la tarea... | Entonces... |
|---------------|-------------|
| Tiene **>4h** de esfuerzo estimado | Dividir en milestones funcionales de ≤2h cada uno |
| Toca **>5 archivos** | Evaluar si se puede partir en: (a) extraer lo nuevo, (b) migrar consumidores |
| Tiene **>3 Acceptance Criteria** | Probablemente son 2 o más tareas independientes |
| Bloquea **>3 tareas** posteriores | Priorizarla como milestone independiente para desbloquear rápido |
| Tiene **dependencia circular** con otra | Red flag — repensar el orden o fusionarlas |

### Ejemplo de división aplicado

| Tarea original | Cómo se dividió | Resultado |
|---------------|----------------|-----------|
| RC1 (11 subtareas, 5h) | Milestone A (paths, 1h) + B (templates, 2.5h) + C (installer, 1.5h) | 3 PRs mergeables independientemente |
| TASK-007 "refactor all 5 adapters" | Absorbido en C1 (helper) + A2 (paths) + B2 (templates) | Cada milestone migra lo suyo |

---

## Referencias

- [Sequoia Tasks actual](../docs/sequoia/sequoia-tasks.md) — 49 tareas en formato actual
- [Roadmap v1.0](./roadmap-v1.0.md) — nuevos dominios y agentes planificados
- [Audit Master Report](../docs/sequoia/sequoia-master.md) — hallazgos completos
