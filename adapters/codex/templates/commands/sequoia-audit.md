---
description: "Ejecuta auditoría técnica completa. Corre agentes de fase en paralelo, luego meta-agentes para correlación y reportes. Soporta flags: --phase, --scope, --mode, --output."
argument-hint: "[--phase=security|performance|architecture|quality|experience|operations] [--scope=changed|module=<path>] [--mode=full|quick] [--output=report|tasks|both]"
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia audit

Ejecuta la auditoría técnica integral. Orquesta agentes de fase, meta-agentes de correlación y genera entregables completos.

## Precondición

Debe haberse ejecutado `/sequoia init` previamente. Si no hay Project Map en Engram, solicitar al usuario que ejecute init primero.

## Flujo de ejecución

```
/sequoia audit
  │
  ├─ 1. Recuperar Project Map de Engram
  │     └─ Si no existe → ERROR: "Ejecutá /sequoia init primero"
  │
  ├─ 2. Seleccionar agentes según flags + Project Map
  │     ├─ Sin --phase → todos los agentes aplicables
  │     ├─ Con --phase → solo ese agente + meta-agentes
  │     └─ Agentes no aplicables → se saltan con motivo
  │
  ├─ 3. Ejecutar agentes de fase
  │     ├─ Paralelo: P1, P2, P3, P4 (sin dependencias entre sí)
  │     ├─ Después: P5, P6 (usan hallazgos de P3)
  │     └─ Todos los aplicables según Project Map
  │
  ├─ 4. Ejecutar meta-agentes
  │     ├─ M1 sequoia-correlator (cruza hallazgos entre fases)
  │     └─ M2 sequoia-reporter (calcula health scores + genera documentos)
  │
  ├─ 5. Generar entregables
  │     ├─ sequoia-master.md (documento maestro)
  │     ├─ sequoia-phases/01-security.md ... 06-operations.md
  │     ├─ sequoia-score.md (health scorecard)
  │     └─ [si --output=tasks|both] sequoia-tasks.md
  │
  └─ 6. Persistir en Engram
        ├─ Hallazgos con timestamp
        ├─ Health scores
        └─ Snapshot de estado para futuro diff
```

## Referencia de flags

| Flag | Valores | Default | Descripción |
|------|---------|---------|-------------|
| `--phase` | `security` `performance` `architecture` `quality` `experience` `operations` | todas | Ejecuta solo una fase específica |
| `--scope` | `changed` `module=<path>` | todo el proyecto | Limita el scope de la auditoría |
| `--mode` | `full` `quick` | `full` | Profundidad del análisis |
| `--output` | `report` `tasks` `both` | `both` | Tipo de entregable a generar |

## `--mode` diferencias

### `full` (default)
- Todos los agentes aplicables
- Análisis profundo por agente
- Hallazgos de todas las severidades
- Presupuestos de performance completos
- Mapa de dependencias de módulos
- Matriz de superficie de ataque completa
- Tiempo estimado: 15-45 min según tamaño

### `quick`
- Solo hallazgos CRÍTICO y RIESGO
- Sin entregables adicionales (presupuestos, mapas, matrices)
- Agentes reducidos a inspecciones de mayor impacto
- Sin correlación profunda (correlator simplificado)
- Tiempo estimado: 5-15 min según tamaño

## `--scope` opciones

| Valor | Qué hace |
|-------|----------|
| *(sin flag)* | Audita todo el proyecto |
| `changed` | Solo archivos modificados vs último commit. Usa `git diff --name-only HEAD` |
| `module=src/auth` | Solo el módulo indicado y sus subdirectorios |

Con `--scope=changed`, cada agente solo inspecciona los archivos del diff. Los meta-agentes correlacionan solo contra esos hallazgos.

## `--output` opciones

| Valor | Genera |
|-------|--------|
| `report` | `sequoia-master.md` + `sequoia-phases/*.md` + `sequoia-score.md` |
| `tasks` | `sequoia-tasks.md` con plan accionable por fase |
| `both` | Todo lo anterior (default) |

## Lógica de paralelismo

### Pueden correr en paralelo (sin dependencias):
- P1 Security, P2 Performance, P3 Architecture, P4 Quality

### Deben correr después de P3:
- P5 Experience (usa mapa de arquitectura)
- P6 Operations (usa modelo de arquitectura)

### Meta-agentes (siempre secuenciales):
- M1 Correlator → M2 Reporter (en ese orden; scoring es parte de M2)

## Delegación del orquestador

El orquestador delega a cada agente proporcionándole:
1. El **Project Map** completo
2. El **scope** aplicable (todo, changed files, o módulo)
3. El **modo** (full o quick)
4. La **plantilla de hallazgo** estándar (de `references/finding-format.md`)

Cada agente retorna sus hallazgos en el formato estándar. El orquestador no interpreta hallazgos, solo los enruta.

## Entregables generados

Todos se crean en el directorio configurado (default: `docs/sequoia/`):

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

## Ejemplos de uso

```bash
# Auditoría completa
/sequoia audit

# Solo seguridad, modo rápido
/sequoia audit --phase=security --mode=quick

# Solo archivos cambiados, solo reporte
/sequoia audit --scope=changed --output=report

# Auditoría profunda de un módulo
/sequoia audit --scope=module=src/auth --mode=full

# Generar solo tareas de calidad
/sequoia audit --phase=quality --output=tasks
```
