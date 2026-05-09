# Flujo: Auditoría Completa

Flujo de trabajo para auditorías integrales en proyectos medianos y grandes.

## Precondiciones

- `/sequoia init` ejecutado y Project Map disponible en Engram
- Si el init tiene más de 7 días, refrescar con un re-init rápido

## Diagrama de flujo

```
┌─────────────────────────────────────────────────────────────────┐
│                    /sequoia audit                                │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │ 1. REFRESH  │ Quick re-scan del Project Map
                    │   CONTEXT   │ (verificar que sigue vigente)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ 2. SELECT   │ Agentes aplicables según
                    │   AGENTS    │ Project Map + flags
                    └──────┬──────┘
                           │
               ┌───────────┼───────────┐
               │           │           │
        ┌──────▼──────┐    │    ┌──────▼──────┐
        │ 3a. BATCH 1 │    │    │ 3a. BATCH 1 │
        │ P1 Security │    │    │ P4 Quality  │
        │ P2 Perform. │    │    └──────┬──────┘
        │ P3 Archit.  │    │           │
        └──────┬──────┘    │           │
               │           │           │
               └─────┬─────┘           │
                     │    (paralelo)   │
                     └────────┬────────┘
                              │
                     ┌────────▼────────┐
                     │ 3b. BATCH 2     │
                     │ P5 Experience   │
                     │ P6 Operations   │
                     │ (usan P3 output)│
                     └────────┬────────┘
                              │
               ┌──────────────┼──────────────┐
               │                             │
        ┌──────▼──────┐              ┌───────▼──────┐
        │ 4a. M1      │─────────────│ 4b. M2       │
        │ CORRELATOR  │             │ REPORTER     │
        │ cruza fases │             │ scoring+docs │
        └─────────────┘             └──────────────┘
                                            │
                                   ┌────────▼────────┐
                                   │ 5. DELIVERABLES │
                                   │ master.md       │
                                   │ phases/*.md     │
                                   │ score.md        │
                                   │ tasks.md        │
                                   └────────┬────────┘
                                            │
                                   ┌────────▼────────┐
                                   │ 6. ENGRAM       │
                                   │ Persistir:      │
                                   │ hallazgos+score │
                                   │ + snapshot      │
                                   └─────────────────┘
```

## Detalle por paso

### Paso 1 — Context Refresh (~1-2 min)

Quick re-scan para verificar que el Project Map sigue vigente:
- ¿Se agregaron dependencias nuevas?
- ¿Cambió la estructura de directorios?
- ¿Hay archivos nuevos relevantes?

Si detecta cambios significativos → re-ejecutar paso de init correspondiente.

### Paso 2 — Agent Selection (~instantáneo)

Determinar agentes a ejecutar:
- Sin `--phase` → todos los marcados como "aplica" en el Project Map
- Con `--phase` → solo ese agente
- Con `--scope` → todos los aplicables, pero cada uno limita su scope

### Paso 3 — Agentes de fase (~10-25 min total)

**Batch 1 (paralelo)** — sin dependencias entre sí:
| Agente | Tiempo estimado | Produce |
|--------|----------------|---------|
| P1 Security | 3-8 min | Hallazgos de seguridad + matriz de ataque |
| P2 Performance | 3-8 min | Hallazgos de performance + presupuesto |
| P3 Architecture | 5-10 min | Hallazgos de arquitectura + API design + mapa deps |
| P4 Quality | 3-6 min | Hallazgos de calidad + deps + estrategia testing |

**Batch 2 (después de P3)** — usan output de architecture:
| Agente | Tiempo estimado | Produce |
|--------|----------------|---------|
| P5 Experience | 3-6 min | Hallazgos de UX + producto |
| P6 Operations | 3-6 min | Hallazgos de DevOps + data + infra |

### Paso 4 — Meta-agentes (~3-5 min total)

Siempre secuenciales en este orden:

1. **M1 Correlator** (~1-2 min): Cruza hallazgos entre fases, detecta causas raíz
2. **M2 Reporter** (~1-2 min): Calcula health score por fase y global + genera todos los documentos

### Paso 5 — Deliverables (~1 min)

Generación de archivos markdown en `docs/sequoia/`.

### Paso 6 — Engram (~instantáneo)

Persistir:
- Hallazgos con timestamp y hash del commit actual
- Health scores para histórico
- Snapshot del estado para futuro `/sequoia diff`

## Decisiones ante edge cases

### Nuevas dependencias detectadas durante la auditoría
Si un agente descubre deps no mapeadas en el init:
1. Anotarlas como hallazgo (P4 Quality)
2. No detener la auditoría
3. Sugerir re-init al final del reporte

### Stack ambiguo o mixto (monorepo)
1. Ejecutar init por cada sub-proyecto si son independientes
2. Si comparten código, auditar el shared como módulo transversal
3. Reporter separa hallazgos por sub-proyecto

### Agente que no puede verificar algo
El agente marca el hallazgo como `[NO VERIFICABLE]` o `[REQUIERE ACCESO EXTERNO]`.
El correlator NO correlaciona hallazgos no verificables. El reporter los incluye en sección separada.

### Proyecto sin tests y sin CI
No es un error. P4 y P6 reportan la ausencia como hallazgos.
El reporter marca esas fases según el estado real, no aspiracional.

## Estimación de tiempo total

| Tamaño proyecto | full | quick |
|----------------|------|-------|
| Pequeño (< 50 archivos) | 10-15 min | 5-8 min |
| Mediano (50-200 archivos) | 15-30 min | 8-15 min |
| Grande (> 200 archivos) | 30-45 min | 12-20 min |

*Con `--scope=module`, restar ~60% del tiempo estimado.*
