---
description: "Develop a task using SDD with configured strategies. Auto-detects simple tasks for fast-forward mode."
argument-hint: "<task-id> [--ff | --full]"
allowed-tools: Read, Glob, Grep, Bash, mem_search, mem_get_observation, mem_save
---

# /sequoia-dev

> **SDD task executor** — lee una tarea, computa complejidad, y lanza el flujo SDD apropiado. Configurado vía `~/__SEQUOIA_BASE__/.sequoia-dev.yaml`.

## 🔴 BLOCKING SEQUENCE — ejecutá estos pasos EN ORDEN. No saltees, no reordenes, no "optimices".

---

### STEP 0 — IDENTIFICÁ LA TAREA (primera acción, sin excepciones)

El mensaje del usuario contiene el task ID después de `/sequoia-dev`. Extraélo AHORA.
El formato es `{PHASE}-{NNN}` (ej: `P1-003`, `M2-001`).
También verificá si hay flags: `--ff` (forzar fast-forward) o `--full` (forzar ciclo completo).

**Antes de hacer NADA más**, anunciá obligatoriamente:

```
🔧 /sequoia-dev {task-id}
```

Si ambos `--ff` y `--full` están presentes, error inmediato: "Cannot use both --ff and --full. Choose one." No continúes.

---

### STEP 1 — CARGÁ LA CONFIGURACIÓN (BLOQUEANTE — no avances sin completar este paso)

**Este paso es OBLIGATORIO. No lo omitas bajo ninguna circunstancia. La configuración gobierna TODO lo que sigue.**

1. **Resolvé la ruta absoluta** al archivo de configuración:
   - Windows: `$env:USERPROFILE\__SEQUOIA_BASE__\.sequoia-dev.yaml`
   - Linux/macOS: `$HOME/__SEQUOIA_BASE__/.sequoia-dev.yaml`
2. **Leé el archivo** con la herramienta `read` usando la ruta absoluta. NO uses `glob` ni `grep` — estas herramientas solo buscan dentro del workspace y fallarán silenciosamente.
3. **Parseá el YAML**. Mergeá con defaults (valores del usuario sobreescriben defaults, claves faltantes usan su default, claves desconocidas se ignoran).

| Key | Default | Valores válidos |
|-----|---------|-----------------|
| `sdd.execution_mode` | `auto` | `auto` \| `interactive` |
| `sdd.tdd` | `strict` | `strict` \| `standard` |
| `sdd.delivery` | `ask-on-risk` | `ask-on-risk` \| `auto-chain` \| `single-pr` \| `exception-ok` |
| `sdd.chain` | `stacked-to-main` | `stacked-to-main` \| `feature-branch-chain` |
| `sdd.auto_ff` | `true` | `true` \| `false` |
| `sdd.complexity.ff_max_score` | `2` | 0–9 |
| `paths.tasks_dir` | `.sequoia/tasks/` | ruta relativa al proyecto |

**Manejo de errores de configuración:**
- **Archivo no existe** → usá defaults en silencio (primer uso, todavía no hay config).
- **Error de sintaxis YAML** → advertí al usuario: "⚠️ .sequoia-dev.yaml tiene un error de sintaxis: {error}. Usando defaults." y continuá con defaults.
- **Valor inválido** (ej: `tdd: banana`) → advertí al usuario y usá el default para ESA clave solamente.
- **Claves desconocidas** → ignorá en silencio (forward compatibility).

**Anunciá la configuración activa** antes de continuar:

```
⚙️  Configuración activa:
   execution_mode: {valor}
   tdd: {valor}
   delivery: {valor}
   chain: {valor}
   auto_ff: {valor}
   ff_max_score: {valor}
   tasks_dir: {valor}
```

**Estos valores GOBIERNAN todo el flujo que sigue. No son sugerencias, son reglas.**

---

### STEP 2 — PRECONDICIÓN: SDD Init

Verificá que SDD esté inicializado. Buscá en Engram: `mem_search(query: "sdd-init/{project}")`.

Si no se encuentra → error bloqueante:
```
SDD has not been initialized for this project. Run `/sdd-init` first to set up SDD workspace, testing capabilities, and registry.
```
No continúes sin SDD init.

---

### STEP 3 — LOCALIZÁ LA TAREA

Usá `paths.tasks_dir` de la configuración (STEP 1) como directorio base.

1. Mapeá el prefijo de fase a archivo de área:
   - `P1` → `security.md`
   - `P2` → `performance.md`
   - `P3` → `architecture.md`
   - `P4` → `quality.md`
   - `P5` → `experience.md`
   - `P6` → `operations.md`
   - `M1`, `M2` → `index.md`
2. Buscá en `{paths.tasks_dir}{area}.md` el encabezado `### {TASK-ID} ·`.
3. Parseá los metadatos de la tarea:

| Metadato | Fuente | Puntuación |
|----------|--------|-----------|
| Implementation risk | `**Implementation risk**: Low / Medium / High` | Low=0, Medium=1, High=3 |
| Effort | `**Effort**: small / medium / large` | small=0, medium=1, large=3 |
| Files involved | Contar bullets `- \`path\`` en **Files involved** | 1-2=0, 3-4=1, 5+=2 |
| Dependencies | ¿Tiene contenido más allá de "None"? | none=0, has_deps=1 |

Si la tarea no se encuentra en el archivo esperado, buscá en TODOS los `*.md` de `tasks_dir`. Si aparece en otra área, sugerí: "Task {id} not found in {area}.md. Did you mean {id} in {other_area}.md?". Si no aparece en ningún lado, error: "Task {id} not found in {tasks_dir}. Run `/sequoia fix` to regenerate tasks."

Si faltan metadatos, usá defaults seguros: `risk=Medium (1)`, `effort=medium (1)`, `files=3`, `deps=none (0)`. Advertí: "⚠️ Task {id} is missing some metadata. Using safe defaults."

---

### STEP 4 — COMPUTÁ COMPLEJIDAD

```
score = implementation_risk + effort + files_score + dependencies
```

| Factor | Valor | Score |
|--------|-------|-------|
| Risk | Low | 0 |
| Risk | Medium | 1 |
| Risk | High | 3 |
| Effort | small | 0 |
| Effort | medium | 1 |
| Effort | large | 3 |
| Files | 1-2 | 0 |
| Files | 3-4 | 1 |
| Files | 5+ | 2 |
| Dependencies | none | 0 |
| Dependencies | has deps | 1 |

**Decisión** (las flags `--ff` / `--full` tienen prioridad absoluta):
- `--ff` → Forzar fast-forward, ignorar score
- `--full` → Forzar ciclo completo, ignorar score
- `score ≤ sdd.complexity.ff_max_score` AND `sdd.auto_ff = true` → Fast-forward (`sdd-ff`)
- `score > sdd.complexity.ff_max_score` OR `sdd.auto_ff = false` → Ciclo SDD completo

---

### STEP 5 — MOSTRAR FEEDBACK (obligatorio, incluso en modo auto)

Antes de lanzar CUALQUIER subagente, mostrá este resumen:

```
🔧 /sequoia-dev {task-id}
   Mode: {fast-forward | full SDD cycle}
   TDD: {strict | standard}
   Delivery: {ask-on-risk | auto-chain | single-pr | exception-ok}
   Chain: {stacked-to-main | feature-branch-chain}
   Complexity score: {score} (risk={r} + effort={e} + files={f} + deps={d})
   Config: ~/__SEQUOIA_BASE__/.sequoia-dev.yaml

{Breve descripción de lo que va a pasar}
```

---

### STEP 6 — LANZAR SDD

Usá EXCLUSIVAMENTE los valores de configuración cargados en STEP 1.

#### Fast-Forward (`sdd-ff`)

Cuando score ≤ `ff_max_score` o `--ff`:
1. Leé spec scenarios y design decisions de Engram (si existen)
2. Leé los acceptance criteria y verification steps de la tarea
3. Implementá la tarea directamente (apply)
4. Verificá la implementación (verify)

Pasá **explícitamente** a cada subagente: `execution_mode={valor de STEP 1}`, `tdd={valor de STEP 1}`, `delivery={valor de STEP 1}`, `chain={valor de STEP 1}`.

#### Full SDD Cycle

Cuando score > `ff_max_score` o `--full`:
1. `sdd-propose` → Definir scope y approach
2. `sdd-spec` → Escribir requirements y scenarios
3. `sdd-design` → Crear diseño técnico
4. `sdd-tasks` → Desglosar en tasks de implementación
5. `sdd-apply` → Implementar cada task
6. `sdd-verify` → Verificar contra specs
7. `sdd-archive` → Sincronizar delta specs

Pasá **explícitamente** a cada subagente: `execution_mode={valor de STEP 1}`, `tdd={valor de STEP 1}`, `delivery={valor de STEP 1}`, `chain={valor de STEP 1}`.

---

## Configuración de referencia

Tu `~/__SEQUOIA_BASE__/.sequoia-dev.yaml`:

```yaml
sdd:
  execution_mode: auto       # auto | interactive
  tdd: strict                # strict | standard
  delivery: ask-on-risk      # ask-on-risk | auto-chain | single-pr | exception-ok
  chain: stacked-to-main     # stacked-to-main | feature-branch-chain
  auto_ff: true              # true | false
  complexity:
    ff_max_score: 2          # 0-9

paths:
  tasks_dir: .sequoia/tasks/
```

Este archivo se crea automáticamente con `sequoia install`. Podés editarlo cuando quieras — `/sequoia-dev` lo lee fresco en cada invocación.
