# Sequoia — Referencia de Comandos

> Todos los comandos se ejecutan **dentro de tu herramienta de código** (Claude Code, OpenCode, etc.) usando el prefijo `/sequoia`.
> Antes de usar cualquier comando, Sequoia debe estar instalado vía la TUI de instalación (`sequoia install`).

---

## Índice

| Comando | Propósito | Precondición |
|---------|-----------|-------------|
| [`/sequoia init`](#sequoia-init) | Detectar stack, construir Project Map | Ninguna (primer paso obligatorio) |
| [`/sequoia audit`](#sequoia-audit) | Auditoría técnica completa | `init` ejecutado |
| [`/sequoia review`](#sequoia-review) | Revisión de PR/diff focalizada | `init` ejecutado |
| [`/sequoia fix`](#sequoia-fix) | Generar tareas accionables | Auditoría previa existente |
| [`/sequoia diff`](#sequoia-diff) | Comparar estado vs última auditoría | Auditoría previa existente |

---

## Flujo típico de uso

```
Primera vez:
  /sequoia init     →  mapea el proyecto
  /sequoia audit    →  auditoría completa
  /sequoia fix all  →  genera plan de tareas

Día a día:
  /sequoia review                    →  review del último commit
  /sequoia review --pr=42            →  review de un PR
  /sequoia diff                      →  ver qué mejoró desde la última auditoría

Post-fix:
  /sequoia audit --phase=security    →  re-auditar solo la fase que fixeaste
  /sequoia diff                      →  verificar que los hallazgos se resolvieron
```

---

## `/sequoia init`

**Qué hace**: Inicializa Sequoia en el proyecto. Detecta stack, construye el Project Map, determina qué agentes aplican y persiste todo en Engram.

**Precondición**: Ninguna. Este es SIEMPRE el primer comando.

### Qué ejecuta

| Paso | Acción |
|------|--------|
| 1 | Escanea la estructura del proyecto (manifiestos, configs, frameworks, CI/CD) |
| 2 | Identifica el tech stack con evidencia (lenguaje, framework, runtime, bundler, test runner, package manager) |
| 3 | Determina el paradigma (SPA, API REST, CLI, monolito, microservicios, etc.) |
| 4 | Estima el tamaño (LOC, módulos, dependencias) |
| 5 | Verifica infraestructura existente (tests, CI/CD, linting, docs, Docker) |
| 6 | Evalúa la madurez del proyecto (prototipo / desarrollo / producción) |
| 7 | Determina qué agentes P1-P6 aplican con motivo explícito |
| 8 | Persiste el Project Map en Engram |

### Output: Project Map

```markdown
## Sequoia Project Map — {nombre}

**Madurez**: desarrollo | **Paradigma**: fullstack

### Stack detectado
- Lenguaje: TypeScript
- Framework: Next.js 14
- Runtime: Node 20
- Test runner: Vitest
- Package manager: pnpm

### Agentes aplicables
| Agente | Aplica | Motivo |
|--------|--------|--------|
| P1 Security | ✅ | Todo proyecto |
| P2 Performance | ✅ | SSR + API routes |
| P3 Architecture | ✅ | Siempre |
| P4 Quality | ✅ | Siempre |
| P5 Experience | ✅ | Tiene UI |
| P6 Operations | ✅ | Tiene Docker + CI |
```

### Ejemplos

```bash
# Inicializar (no tiene flags, simplemente se ejecuta)
/sequoia init
```

---

## `/sequoia audit`

**Qué hace**: Ejecuta la auditoría técnica integral. Orquesta agentes de fase en paralelo, luego meta-agentes para correlación y reportes.

**Precondición**: `/sequoia init` ejecutado previamente.

### Flags

| Flag | Valores | Default | Descripción |
|------|---------|---------|-------------|
| `--phase` | `security` `performance` `architecture` `quality` `experience` `operations` | todas | Ejecuta solo una fase específica |
| `--scope` | `changed` `module=<path>` | todo el proyecto | Limita el scope de la auditoría |
| `--mode` | `full` `quick` | `full` | Profundidad del análisis |
| `--output` | `report` `tasks` `both` | `both` | Tipo de entregable |

### `--phase`: Seleccionar qué auditar

Cada fase corresponde a un agente especializado:

| Flag | Agente | Qué audita |
|------|--------|-----------|
| `--phase=security` | P1 | Vulnerabilidades, auth, secrets, XSS/CSRF, superficie de ataque |
| `--phase=performance` | P2 | Bundle, N+1, caching, Core Web Vitals, render path |
| `--phase=architecture` | P3 | Módulos, acoplamiento, API design, god objects |
| `--phase=quality` | P4 | Tests, CVE scanning, riesgo de deps, complejidad |
| `--phase=experience` | P5 | UX flows, accesibilidad WCAG, conversion funnels, SEO |
| `--phase=operations` | P6 | CI/CD, monitoring, migraciones, integridad de datos |
| *(sin flag)* | Todos | Todos los agentes aplicables según el Project Map |

> **Nota**: Con `--phase`, siempre se ejecutan M1 (correlator) y M2 (reporter) para generar el health score.

### `--scope`: Limitar qué archivos auditar

| Valor | Qué hace |
|-------|----------|
| *(sin flag)* | Audita todo el proyecto |
| `changed` | Solo archivos modificados (`git diff --name-only HEAD`) |
| `module=src/auth` | Solo el módulo indicado y sus subdirectorios |

### `--mode`: Profundidad del análisis

| Aspecto | `full` | `quick` |
|---------|--------|---------|
| Hallazgos | Todas las severidades | Solo 🔴 crítico y 🟠 riesgo |
| Tiempo estimado | 15-45 min | 5-15 min |
| Extras | Presupuestos, mapas, matrices | Sin extras |
| Correlación | Completa | Simplificada |

### `--output`: Tipo de entregable

| Valor | Genera |
|-------|--------|
| `report` | `sequoia-master.md` + `sequoia-phases/*.md` + `sequoia-score.md` |
| `tasks` | `sequoia-tasks.md` con plan accionable |
| `both` | Todo lo anterior (default) |

### Orden de ejecución de agentes

```
Paralelo (sin dependencias):
  P1 Security  ─┐
  P2 Performance─┤
  P3 Architecture─┤  →  Después de P3:
  P4 Quality   ─┘     P5 Experience (usa mapa de arch)
                       P6 Operations (usa modelo de arch)

Meta-agentes (siempre secuenciales):
  M1 Correlator → M2 Reporter (scoring + documentos)
```

### Entregables generados

```
docs/sequoia/
├── sequoia-master.md          # Documento maestro
├── sequoia-score.md           # Health scorecard
├── sequoia-tasks.md           # [si --output=tasks|both]
└── sequoia-phases/
    ├── 01-security.md
    ├── 02-performance.md
    ├── 03-architecture.md
    ├── 04-quality.md
    ├── 05-experience.md
    └── 06-operations.md
```

### Ejemplos

```bash
# Auditoría completa
/sequoia audit

# Solo seguridad, modo rápido
/sequoia audit --phase=security --mode=quick

# Solo rendimiento, módulo específico, análisis profundo
/sequoia audit --phase=performance --scope=module=src/api --mode=full

# Solo archivos cambiados, solo reporte
/sequoia audit --scope=changed --output=report

# Solo calidad, generar solo tareas
/sequoia audit --phase=quality --output=tasks

# Quick check de todo (rápido, solo críticos)
/sequoia audit --mode=quick
```

---

## `/sequoia review`

**Qué hace**: Revisión de código tipo PR review. Analiza cambios recientes, selecciona agentes relevantes automáticamente y cruza contra hallazgos previos.

**Precondición**: `/sequoia init` ejecutado.

### Flags

| Flag | Valor | Default | Descripción |
|------|-------|---------|-------------|
| `--diff` | rango git | `HEAD~1..HEAD` | Rango de commits a revisar |
| `--pr` | número de PR | — | Obtiene diff del PR vía `gh` CLI |
| `--strict` | *(flag booleano)* | off | Sin tolerancia para hallazgos medios |

### `--strict` mode

| Aspecto | Sin `--strict` | Con `--strict` |
|---------|---------------|----------------|
| Hallazgos reportados | 🔴 Crítico + 🟠 Riesgo | 🔴 + 🟠 + 🟡 Atención |
| Veredicto con riesgos | `⚠️ WARN` | `🔴 BLOCK` |
| Uso típico | Feedback rápido | Gate de merge en branch protegida |

### Auto-selección de agentes

El review selecciona agentes automáticamente según los archivos cambiados:

| Patrón en archivos | Agentes activados |
|---------------------|-------------------|
| Auth, sesión, tokens | P1 Security |
| Componentes UI, páginas | P2 Performance, P5 Experience |
| Rutas, endpoints | P3 Architecture |
| Modelos, migraciones | P3 Architecture, P6 Operations |
| Tests | P4 Quality |
| CI/CD, Docker | P6 Operations |
| `package.json`, `go.mod` | P4 Quality (deps) |
| Config de build | P2 Performance |
| **Siempre** | P3 Architecture |

### Cruce con hallazgos previos

El review compara contra auditorías anteriores:

- Si los cambios tocan líneas cerca de un hallazgo previo → `📋 HALLAZGO PREVIO AFECTADO`
- Si los cambios resuelven un hallazgo previo → `✅ HALLAZGO RESUELTO`
- Si no tocan hallazgos previos → no se mencionan (reduce ruido)

### Veredicto

| Veredicto | Condición |
|-----------|-----------|
| ✅ `PASS` | Sin hallazgos críticos ni riesgos |
| ⚠️ `WARN` | Hay riesgos 🟠 (sin `--strict`) |
| 🔴 `BLOCK` | Hay críticos 🔴, o riesgos con `--strict` |

### Diferencia con `audit`

| Aspecto | `audit` | `review` |
|---------|---------|----------|
| Scope | Proyecto completo | Solo archivos cambiados |
| Agentes | Todos los aplicables | Solo relevantes al diff |
| Tiempo | 15-45 min | 2-8 min |
| Output | Reportes completos | Hallazgos + veredicto |

### Ejemplos

```bash
# Review del último commit
/sequoia review

# Review de los últimos 3 commits
/sequoia review --diff=HEAD~3..HEAD

# Review de un PR específico
/sequoia review --pr=42

# Review estricto como gate de merge
/sequoia review --pr=42 --strict

# Review de un rango de commits específico
/sequoia review --diff=v2.0.0..HEAD
```

---

## `/sequoia fix`

**Qué hace**: Genera tareas implementables desde los hallazgos de una auditoría. Cada tarea es autosuficiente: un agente implementador puede ejecutarla sin releer la auditoría completa.

**Precondición**: Al menos una auditoría previa ejecutada (`audit` o `review`).

### Uso

```bash
/sequoia fix <fase>           # Tareas de una fase específica
/sequoia fix all              # Tareas de todas las fases
/sequoia fix <fase> --task=ID # Una tarea específica
```

### Fases disponibles

| Argumento | Qué hace |
|-----------|----------|
| `security` | Tareas solo de P1 Security |
| `performance` | Tareas solo de P2 Performance |
| `architecture` | Tareas solo de P3 Architecture |
| `quality` | Tareas solo de P4 Quality |
| `experience` | Tareas solo de P5 Experience |
| `operations` | Tareas solo de P6 Operations |
| `all` | Todas las fases, usando correlación de causas raíz |

### Formato de cada tarea

```markdown
### [TASK-ID] · Título accionable

**Prioridad**: 🔴 Bloqueante | 🟠 Alto leverage | 🟡 Backlog
**Fase origen**: P1-P6 | M1-M2
**Hallazgo(s) origen**: IDs del hallazgo

**Contexto mínimo**:
Qué está mal y por qué importa (3-5 líneas).

**Archivos involucrados**:
- path/al/archivo.ext — qué papel cumple
- path/otro/archivo.ext — qué modificar

**Qué hacer**:
1. Paso concreto 1
2. Paso concreto 2
3. Paso concreto 3

**Criterio de aceptación**:
- [ ] Condición verificable 1
- [ ] Condición verificable 2
```

### Orden de implementación

Las tareas se ordenan por:

1. 🔴 **Bloqueantes de producción** → primero (sin excepción)
2. **Causas raíz** → antes que sus síntomas (del correlator)
3. **Dependencias técnicas** → B requiere que A esté hecha
4. **Alto leverage** → máximo impacto con mínimo cambio
5. **Riesgo bajo** → quick wins primero

### Ejemplos

```bash
# Generar tareas de seguridad
/sequoia fix security

# Generar todas las tareas (cross-fase, con correlación)
/sequoia fix all

# Generar una tarea específica
/sequoia fix security --task=P1-003

# Generar tareas de rendimiento
/sequoia fix performance
```

---

## `/sequoia diff`

**Qué hace**: Compara el estado actual del proyecto contra la última auditoría registrada. Muestra evolución: qué mejoró, qué empeoró, qué apareció nuevo.

**Precondición**: Al menos una auditoría previa en Engram.

### Categorías de evolución

| Icono | Categoría | Significado |
|-------|-----------|-------------|
| ✅ | Resuelto | El hallazgo anterior ya no se reproduce |
| 🔸 | Parcialmente resuelto | Se mejoró pero no cumple criterio de aceptación |
| ⏸️ | Sin cambio | El hallazgo anterior sigue igual |
| 🔻 | Empeorado | El hallazgo sigue y ha empeorado |
| 🆕 | Nuevo | Problema que no existía en la auditoría anterior |

### Qué ejecuta

1. Recupera la última auditoría de Engram
2. Verifica cambios en la estructura del proyecto
3. Re-verifica cada hallazgo anterior (¿el problema sigue?)
4. Detecta hallazgos nuevos (scan rápido, solo 🔴 y 🟠)
5. Genera reporte de evolución

### Output

```markdown
## Sequoia Diff — [Proyecto]

**Auditoría anterior**: 2025-01-10
**Comparación actual**: 2025-01-24
**Tiempo transcurrido**: 14 días

### Resumen de evolución
| Categoría | Cantidad | Porcentaje |
|-----------|----------|------------|
| ✅ Resueltos | 5 | 33% |
| 🔸 Parciales | 2 | 13% |
| ⏸️ Sin cambio | 6 | 40% |
| 🔻 Empeorados | 1 | 7% |
| 🆕 Nuevos | 1 | 7% |

### Health Score comparativo
| Fase | Antes | Ahora | Tendencia |
|------|-------|-------|-----------|
| Security | 🟠 55 | 🟢 72 | ↗️ Mejorando |
| Performance | 🟡 68 | 🟡 70 | → Estable |

### Tendencia global
📈 Mejorando | ➡️ Estable | 📉 Degradando
```

### Cuándo usar `diff` vs `audit` nuevo

| Situación | Usar |
|-----------|------|
| Implementaste fixes y querés verificar | `diff` |
| Pasó 1-2 semanas y querés tracking | `diff` |
| Post-merge de feature grande | `diff` primero, `audit` si hay sorpresas |
| Cambios grandes en el proyecto | `audit` (nueva auditoría) |
| Pasó más de un mes | `audit` (nueva auditoría) |
| Nuevo miembro en el equipo | `audit` (nueva auditoría) |

> Si la última auditoría tiene más de 30 días, `diff` muestra una advertencia y recomienda ejecutar `audit`.

### Ejemplos

```bash
# Ver evolución desde la última auditoría
/sequoia diff
```

---

## Resumen rápido de flags

### `audit`

```
/sequoia audit [--phase=security|performance|architecture|quality|experience|operations]
               [--scope=changed|module=<path>]
               [--mode=full|quick]
               [--output=report|tasks|both]
```

### `review`

```
/sequoia review [--diff=<git-range>]
                [--pr=<number>]
                [--strict]
```

### `fix`

```
/sequoia fix <security|performance|architecture|quality|experience|operations|all>
             [--task=<ID>]
```

### `init` / `diff`

```
/sequoia init     # Sin flags
/sequoia diff     # Sin flags
```
