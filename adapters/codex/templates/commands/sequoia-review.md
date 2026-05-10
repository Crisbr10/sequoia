---
description: "Revisión de código enfocada en PR/diff. Analiza archivos cambiados, selecciona agentes relevantes automáticamente, detecta impacto en hallazgos previos. Más rápido que audit, más profundo que un linter."
argument-hint: "[--diff=HEAD~1..HEAD] [--pr=<number>] [--strict]"
allowed-tools: Read, Glob, Grep, Bash
---

# /sequoia review

Revisión de código tipo PR review. Analiza cambios recientes, selecciona agentes relevantes automáticamente y cruza contra hallazgos previos.

## Cuándo usar

- Antes de mergear un PR
- Después de un batch de cambios
- Como gate de calidad pre-commit (con `--strict`)
- Cuando querés feedback rápido sin auditoría completa

## Flujo de ejecución

```
/sequoia review
  │
  ├─ 1. Obtener archivos cambiados
  │     ├─ --diff=HEAD~3..HEAD → git diff en ese rango
  │     ├─ --pr=42 → gh pr diff 42
  │     └─ sin flags → git diff HEAD~1..HEAD (último commit)
  │
  ├─ 2. Clasificar archivos cambiados por tipo
  │     ├── {auth,session,token,login} → P1 Security
  │     ├── {component,page,view,jsx,tsx,vue,svelte} → P5 Experience
  │     ├── {api,route,endpoint,controller,handler} → P3 Architecture
  │     ├── {model,schema,migration,entity,repository} → P3 Architecture, P6 Operations
  │     ├── {test,spec,__tests__} → P4 Quality
  │     ├── {Dockerfile,workflow,deploy,.github} → P6 Operations
  │     ├── {package.json,go.mod,Cargo.toml} → P4 Quality (deps)
  │     ├── {config,bundle,build,webpack,vite} → P2 Performance
  │     └── Todos los archivos → P3 Architecture (siempre)
  │
  ├─ 3. Recuperar hallazgos previos de Engram
  │     └─ Flag si los cambios tocan áreas con hallazgos abiertos
  │
  ├─ 4. Ejecutar agentes seleccionados (solo sobre archivos del diff)
  │
  ├─ 5. Generar output focalizado
  │     ├─ Hallazgos nuevos del diff
  │     ├─ Hallazgos previos afectados por los cambios
  │     └─ Veredicto: ✅ PASS | ⚠️ WARN | 🔴 BLOCK
  │
  └─ 6. Persistir hallazgos en Engram
```

## Referencia de flags

| Flag | Valor | Default | Descripción |
|------|-------|---------|-------------|
| `--diff` | rango git | `HEAD~1..HEAD` | Rango de commits a revisar |
| `--pr` | número de PR | — | Obtiene diff del PR vía `gh` CLI |
| `--strict` | *(flag booleano)* | off | Sin tolerancia para hallazgos medios |

## `--strict` mode

Con `--strict`:
- Todos los hallazgos ATENCIÓN se reportan (normalmente se omiten en review)
- Hallazgos que normalmente serían "deuda aceptable" se marcan como `WARN`
- Si hay cualquier RIESGO, el veredicto es `BLOCK`
- Útil como gate de merge en branches protegidas

Sin `--strict`:
- Solo se reportan CRÍTICO y RIESGO
- El veredicto es `WARN` si hay riesgos, `PASS` si solo hay atención

## Tabla de auto-selección de agentes

| Patrón en archivos cambiados | Agentes activados |
|------------------------------|-------------------|
| Archivos de auth/seguridad | P1, P4 |
| Componentes UI / páginas | P2, P3, P5 |
| Rutas / endpoints | P3 |
| Modelos / migraciones | P3, P6 |
| Tests | P4 |
| CI/CD / Docker / deploy | P6 |
| Manifiestos de deps | P4 |
| Configuración de build | P2, P3 |
| **Siempre** | P3 (architecture) |

## Cruce con hallazgos previos

Para cada archivo cambiado:
1. Buscar en Engram si ese archivo tenía hallazgos abiertos
2. Si los cambios modifican líneas cerca de un hallazgo previo → marcar `HALLAZGO PREVIO AFECTADO`
3. Si los cambios resuelven un hallazgo previo → marcar `HALLAZGO RESUELTO`
4. Si los cambios no tocan el hallazgo previo → no mencionar (reducir ruido)

## Formato de salida

```markdown
## Sequoia Review — [rango o PR]

**Archivos revisados**: {N}
**Agentes ejecutados**: {lista}

### Bloqueantes
{lista de hallazgos críticos, si hay}

### Riesgos
{lista de hallazgos de riesgo}

### Atención {solo con --strict}
{lista de hallazgos de atención}

### Hallazgos previos afectados
{si los cambios tocan áreas con hallazgos abiertos}

### Hallazgos resueltos
{si los cambios corrigen hallazgos previos}

---
**Veredicto**: {PASS | WARN | BLOCK}
```

## Diferencia con `/sequoia audit`

| Aspecto | audit | review |
|---------|-------|--------|
| Scope | proyecto completo | solo archivos cambiados |
| Agentes | todos los aplicables | solo los relevantes al diff |
| Tiempo | 15-45 min | 2-8 min |
| Profundidad | completa | focalizada en cambios |
| Correlación | completa entre fases | solo entre hallazgos del diff |
| Meta-agentes | todos | solo correlator simplificado |
| Output | reportes completos | hallazgos focalizados + veredicto |

## Ejemplos

```bash
# Review del último commit
/sequoia review

# Review de los últimos 3 commits
/sequoia review --diff=HEAD~3..HEAD

# Review de un PR específico
/sequoia review --pr=42

# Review estricto como gate de merge
/sequoia review --pr=42 --strict
```
