# Sequoia — Arquitectura del Sistema

## Visión General

Sequoia opera en tres capas secuenciales. Cada capa consume la salida de la anterior y produce información más refinada. No hay saltos de capa — el contexto alimenta las fases, las fases alimentan la meta-síntesis.

```
┌─────────────────────────────────────────────────────────┐
│                    CAPA 0: ORQUESTADOR                   │
│  Valida entrada · Selecciona flujo · Coordina capas      │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                 CAPA 1: CONTEXTO (C0)                    │
│  Escanea estructura · Identifica stack · Genera mapa     │
│  Output: Project Map                                     │
└────────────────────────┬────────────────────────────────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
    ┌─────▼─────┐ ┌─────▼─────┐ ┌─────▼─────┐
    │  P1: Sec  │ │  P3: Arch │ │  P4: Qual │  ...
    │  P2: Perf │ │  P5: UX   │ │  P6: Ops  │
    └─────┬─────┘ └─────┬─────┘ └─────┬─────┘
          │              │              │
          └──────────────┼──────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│              CAPA 2: FASE AGENTS (P1-P6)                │
│  Análisis especializado en paralelo                      │
│  Input: Project Map · Output: Hallazgos crudos           │
└────────────────────────┬────────────────────────────────┘
                         │
          ┌──────────────┴──────────────┐
          │                             │
    ┌─────▼──────┐              ┌───────▼──────┐
    │ M1: Correl │──────────────│ M2: Reporter │
    └────────────┘              └──────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│              CAPA 3: META AGENTS (M1-M2)                 │
│  Correlación · Scoring · Reporte final                   │
│  Input: Hallazgos crudos · Output: Entregable            │
└─────────────────────────────────────────────────────────┘
```

## Protocolo de Comunicación entre Agentes

### Contrato de datos

Todo agente en Sequoia se comunica mediante estructuras de datos bien definidas. No hay comunicación ad-hoc.

```
Orquestador ──Project Map──► Agente Fase
Agente Fase ──Hallazgos[]──► Correlator (M1)
Correlator  ──Hallazgos Correlacionados──► Reporter (M2)
Reporter    ──Reporte + Health Score──► Orquestador
```

### Project Map (salida de C0)

El Project Map es el contrato central. Todo agente lo consume para calibrar su análisis.

```yaml
project_map:
  # Identidad
  name: string                    # Nombre del proyecto
  type: enum                      # frontend | backend | cli | library | fullstack | mobile | infrastructure
  description: string | null      # Del package.json, README, o inferido

  # Stack tecnológico
  stack:
    languages:                    # Lenguajes detectados con % aproximado
      - name: string
        percentage: number
    frameworks:                   # Frameworks principales
      - name: string
        version: string | null
        role: string              # ej: "frontend", "testing", "orm"
    runtimes:                     # Node, Deno, Bun, JVM, etc.
      - name: string
        version: string | null

  # Estructura
  structure:
    root_files: string[]          # Archivos en raíz
    entry_points:                 # Puntos de entrada principales
      - path: string
        type: enum                # app | api | cli | test | config
    directories:                  # Directorios principales con propósito inferido
      - path: string
        purpose: string           # ej: "source", "tests", "config", "docs"
        file_count: number
    total_files: number
    total_dirs: number

  # Dependencias
  dependencies:
    package_manager: string | null  # npm, yarn, pnpm, pip, cargo, go, etc.
    production_deps: number
    dev_deps: number
    outdated:                      # Solo si se puede determinar rápidamente
      count: number
    vulnerabilities:               # Solo si se detectan con certeza
      count: number

  # Indicadores de madurez
  indicators:
    has_tests: bool
    has_ci: bool                   # .github/workflows, .gitlab-ci, etc.
    has_containerization: bool     # Dockerfile, docker-compose
    has_iac: bool                  # Terraform, Pulumi, CloudFormation
    has_linting: bool              # ESLint, Pylint, etc.
    has_formatting: bool           # Prettier, Black, etc.
    has_type_checking: bool        # TypeScript, mypy, etc.
    has_env_config: bool           # .env, config por ambiente
    has_documentation: bool        # README, docs/

  # Convenciones detectadas
  conventions:
    architecture_pattern: string | null  # MVC, hexagonal, clean, monolith, etc.
    module_system: string | null         # ESM, CJS, etc.
    test_framework: string | null
    style_guide: string | null           # airbnb, standard, google, etc.

  # Tamaño estimado
  estimated_size: enum            # small (<50 archivos) | medium (50-500) | large (>500)
```

### Hallazgo (salida de agentes fase)

Cada agente produce un array de hallazgos. El formato es estricto — el correlator depende de él.

```yaml
finding:
  id: string                      # Formato: [AGENT_ID]-[NNN] ej: P1-001, P3-014
  agent: string                   # ID del agente que lo produjo
  severity: enum                  # critical | high | medium | low | info
  category: string                # Categoría específica del dominio del agente

  title: string                   # ≤80 caracteres, descriptivo y específico

  evidence:
    file: string                  # Path relativo al root del proyecto
    line: number | [start, end]   # Línea o rango de líneas
    code: string                  # Fragmento de código relevante, ≤10 líneas
    explanation: string           # Por qué esto constituye un hallazgo

  impact: string                  # Qué ocurre si no se aborda, consecuencias reales
  effort: enum                    # small (<2h) | medium (2-8h) | large (>8h)

  related_findings: string[]      # IDs de otros hallazgos potencialmente relacionados
  references: string[]            # CWE, docs, estándares, URLs relevantes
```

### Hallazgo Correlacionado (salida de M1)

```yaml
correlated_finding:
  id: string                      # Formato: CX-[NNN] ej: CX-001
  source_findings: string[]       # IDs de hallazgos originales fusionados
  root_cause: string | null       # Causa raíz identificada, null si es independiente

  severity: enum                  # Re-calibrada post-correlación
  title: string                   # Título consolidado
  description: string             # Descripción unificada con todas las perspectivas

  manifestations:                 # Cómo se manifiesta en cada dominio
    - agent: string
      finding_id: string
      perspective: string         # Cómo ve el problema desde su dominio

  dependencies: string[]          # IDs de otros hallazgos correlacionados que deben resolverse primero
  blocks: string[]                # IDs que se desbloquean al resolver este
```

### Health Score (salida de M2)

```yaml
health_score:
  global: number                  # 0-100, promedio ponderado de categorías

  categories:
    security:                     # Siempre presente
      score: number               # 0-100
      findings: number            # Cantidad de hallazgos
      critical: number
      high: number
      medium: number
      low: number

    performance:                  # Presente si P2 se ejecutó
      score: number | null
      findings: number
      critical: number
      high: number
      medium: number
      low: number

    architecture:                 # Siempre presente
      score: number
      findings: number
      critical: number
      high: number
      medium: number
      low: number

    quality:                      # Siempre presente
      score: number
      findings: number
      critical: number
      high: number
      medium: number
      low: number

    experience:                   # Presente si P5 se ejecutó
      score: number | null
      findings: number
      critical: number
      high: number
      medium: number
      low: number

    operations:                   # Presente si P6 se ejecutó
      score: number | null
      findings: number
      critical: number
      high: number
      medium: number
      low: number

  methodology: >
    score = 100 − Σ(severity_weight × scope_multiplier), floored at 0
    severity_weight: critical=15, high=8, medium=4, low=2, info=0
    scope_multiplier: 1.0 (isolated finding) | 1.5 (shared root cause across ≥2 findings)
    Score global = promedio ponderado de categorías aplicables
    Pesos: security×1.3, architecture×1.1, performance×1.0, quality×1.0, experience×0.9, operations×0.9
    Floor: 0 por categoría, 0 global
    See references/scoring-criteria.md for full formula, grade table, and worked example.
```

## Metodología de Health Score

### Cálculo por categoría

```
category_score = score = 100 − Σ(severity_weight × scope_multiplier), floored at 0

severity_weight:
  critical = 15 puntos
  high     = 8 puntos
  medium   = 4 puntos
  low      = 2 puntos
  info     = 0 puntos

scope_multiplier:
  1.0 = isolated finding (independent root cause)
  1.5 = shared root cause (≥2 findings share the same root cause)
```

### Cálculo global

```
global_score = Σ(category_score × weight) / Σ(weight)

weights (según criticidad del dominio para el tipo de proyecto):
  security:     1.3  (siempre)
  architecture: 1.1  (siempre)
  performance:  1.0  (si aplica)
  quality:      1.0  (siempre)
  experience:   0.9  (si aplica)
  operations:   0.9  (si aplica)

Categorías no aplicables (ej: experience en CLI) se excluyen del cálculo.
```

### Rangos de interpretación (Health Grade)

| Score | Grade | Significado |
|-------|-------|-------------|
| 90–100 | A | Proyecto sin deuda significativa. Mantener. |
| 75–89 | B | Deuda manejable. Planificar mejoras a corto plazo. |
| 60–74 | C | Deuda técnica acumulada. Intervención prioritaria. |
| 40–59 | D | Problemas estructurales. Requiere plan de remedición serio. |
| 0–39 | F | Riesgo significativo. Acción inmediata necesaria. |

## Adaptación por Stack

Sequoia no aplica el mismo análisis a todos los proyectos. El Project Map calibra cada agente.

### Frontend (React, Angular, Vue, Svelte, etc.)

**Agentes activos**: P1, P2, P3, P4, P5, (P6 si tiene CI/CD)

**P2 (Performance) enfoca en**:
- Bundle size y tree-shaking efectivo
- Render performance (re-renders innecesarios, memoización)
- Core Web Vitals (LCP, FID, CLS)
- Lazy loading y code splitting
- Image optimization y asset pipeline
- Memory leaks (event listeners, subscriptions sin cleanup)

**P3 (Architecture) enfoca en**:
- Component architecture (presentacional vs contenedor)
- State management (complejidad, derivaciones, sincronización)
- API coupling (acoplamiento frontend↔backend)
- Routing y navigation patterns
- Form handling y validation patterns

**P5 (Experience) enfoca en**:
- Accesibilidad (WCAG 2.1 AA mínimo)
- Semantic HTML y ARIA usage
- Keyboard navigation y focus management
- Responsive design y breakpoints
- Error states y feedback al usuario
- Loading states y perceived performance

### Backend (Node.js, Go, Python, Java, Rust, etc.)

**Agentes activos**: P1, P2, P3, P4, P6, (P5 solo si tiene API consumida por frontend)

**P2 (Performance) enfoca en**:
- Query performance y N+1 patterns
- Connection pooling y resource management
- Caching strategy (invalidación, TTL, granularidad)
- Cold start time (serverless)
- Memory usage y GC pressure
- Serialization overhead

**P3 (Architecture) enfoca en**:
- API design (RESTful compliance, consistency, versioning)
- Service boundaries y coupling
- Error handling strategy (consistencia, propagación)
- Data flow y transformation layers
- Dependency injection y modularidad

**P6 (Operations) enfoca en**:
- Deployment strategy y rollback capability
- Observabilidad (logging, metrics, tracing)
- Configuration management (env vars, secrets)
- Scaling strategy y resource limits
- Health checks y graceful shutdown

### CLI / Library

**Agentes activos**: P1, P2, P3, P4

**P2 (Performance) enfoca en**:
- Startup time (tiempo hasta primera output)
- Memory footprint en operación normal
- Dependency tree size y transitivas
- I/O efficiency (lectura de archivos, network calls)

**P3 (Architecture) enfoca en**:
- API surface design (complejidad, intuitividad)
- Backward compatibility y breaking changes
- Extensibilidad (plugins, hooks, configuración)
- Error reporting y UX del output
- Documentation y examples

**P4 (Quality) enfoca en**:
- Test coverage y tipos (unit, integration, e2e)
- API documentation accuracy
- Changelog y semver compliance
- Cross-platform compatibility (si aplica)

### Fullstack

**Agentes activos**: P1, P2, P3, P4, P5, P6

**Consideraciones adicionales**:
- P3 analiza comunicación entre capas (frontend ↔ backend)
- P6 evalúa consistencia de deployment entre capas
- P2 analiza latencia end-to-end, no solo por capa
- El correlator (M1) puede encontrar causas raíz que cruzan capas

### Infrastructure (Terraform, Pulumi, CloudFormation, Docker, K8s)

**Agentes activos**: P1, P3, P4, P6

**P1 (Security) enfoca en**:
- IAM policies y least privilege
- Secrets management (no hardcodeados)
- Network segmentation y security groups
- Encryption at rest y in transit

**P6 (Operations) toma rol principal**:
- IaC modularidad y reutilización
- Drift detection y state management
- Environment parity (dev/staging/prod)
- Cost optimization y resource rightsizing

## Flujo de Datos Detallado

```
Usuario: /sequoia audit
                │
                ▼
        ┌───────────────┐
        │  Orquestador   │
        │  valida input  │
        └───────┬───────┘
                │
                ▼
        ┌───────────────┐
        │     C0:        │
        │  Context Agent │──── Escanea directorios
        │                │──── Lee package.json / go.mod / etc
        │                │──── Detecta patrones de estructura
        └───────┬───────┘
                │
                │ Project Map
                ▼
        ┌───────────────┐
        │  Orquestador   │
        │  selecciona    │──── Aplica tabla de selección
        │  agentes       │──── Determina paralelismo
        └───────┬───────┘
                │
        ┌───────┼───────┐
        │       │       │
        ▼       ▼       ▼
      ┌───┐   ┌───┐   ┌───┐
      │P1 │   │P3 │   │P4 │   ... (en paralelo)
      │Sec│   │Arc│   │Qal│
      └─┬─┘   └─┬─┘   └─┬─┘
        │       │       │
        └───────┼───────┘
                │
                │ Hallazgos crudos de cada agente
                ▼
        ┌───────────────┐
        │     M1:        │
        │   Correlator   │──── Deduplica
        │                │──── Agrupa causas raíz
        │                │──── Re-calibra severidad
        └───────┬───────┘
                │
                │ Hallazgos correlacionados
                ▼
        ┌───────────────┐
        │     M2:        │
        │   Reporter     │──── Calcula Health Score
        │                │──── Genera plan de acción
        │                │──── Formatea entregable
        └───────┬───────┘
                │
                │ Reporte final
                ▼
        ┌───────────────┐
        │  Orquestador   │
        │  presenta al   │
        │    usuario      │
        └───────────────┘
```

## Garantías del Sistema

1. **Determinismo**: Mismo proyecto, mismo Project Map → misma selección de agentes. No hay decisiones aleatorias.

2. **Completitud**: Si un agente se ejecuta, analiza su dominio completo. No hay muestreo.

3. **Trazabilidad**: Todo hallazgo enlaza a archivo y línea. Todo score se justifica con hallazgos.

4. **Idempotencia**: Ejecutar la misma auditoría dos veces produce el mismo resultado (salvo cambios en el código).

5. **Aislamiento**: Los agentes fase no comparten estado. Un error en P2 no afecta P3.

6. **Fallback**: Si un agente falla, los demás continúan. El reporte indica qué dominio no fue analizado.

---

*Esta arquitectura prioriza la composición sobre la complejidad. Cada capa hace una cosa bien, y la combinación produce un resultado que ninguna capa individual podría lograr.*
