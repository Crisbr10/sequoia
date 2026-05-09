# Sequoia — Framework de Auditoría y Revisión de Código

> "Un árbol sequoia no crece con prisas. Crece con raíces profundas."
> Sequoia no audita para sonar inteligente. Audita para que el proyecto sea intervenable.

---

## Visión

Sequoia es un framework de auditoría técnica integral diseñado como plugin de Claude Code. Funciona como un equipo de arquitectos especializados que inspeccionan un proyecto desde múltiples ángulos en paralelo o en fases, sin asumir tecnología específica, sin relleno enterprise, y con evidencia concreta del repositorio como única fuente de verdad.

**Principio fundamental**: cada hallazgo debe trazarse a un archivo real, una línea real, o una ausencia documentable. Lo que no se puede verificar, se declara explícitamente como no verificable.

---

## Filosofía de Diseño

| Principio | Descripción |
|-----------|-------------|
| **Evidencia sobre opinión** | Ningún hallazgo sin cita al repo. Ni una. |
| **Contexto sobre dogma** | Sequoia detecta el stack y adapta el análisis. No aplica reglas de React a un proyecto Go. |
| **Causa raíz sobre síntoma** | Distingue entre lo que se ve y lo que lo genera. |
| **Accionabilidad absoluta** | Cada recomendación tiene un responsable, un criterio de aceptación y un riesgo estimado. |
| **Separación de agentes** | Cada dominio tiene su propio agente. No hay un agente que sepa todo medianamente; hay agentes que saben un dominio profundamente. |
| **Deuda priorizable** | No toda deuda es igual. Sequoia clasifica: bloqueante, alto leverage, backlog, aceptable. |

---

## Arquitectura del Sistema

```
┌─────────────────────────────────────────────────────────────┐
│                    SEQUOIA ORCHESTRATOR                      │
│         Detecta contexto · Coordina agentes · Sintetiza      │
└────────────────────────┬────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ Context Agent│ │ Phase Agents │ │ Meta Agents  │
│  (pre-flight)│ │  (1-10)      │ │  (post-run)  │
└──────────────┘ └──────────────┘ └──────────────┘
```

### Capa 0 — Agente de Contexto (Pre-flight)

Antes de cualquier auditoría, `sequoia-context` corre automáticamente y construye el **Mapa de Proyecto**:

- Stack detectado (lenguaje, framework, runtime, bundler)
- Paradigma dominante (SPA, SSR, API, CLI, monolito, microservicio, etc.)
- Tamaño del proyecto (LOC, módulos, dependencias)
- Presencia de tests, CI/CD, documentación
- Estado de salud de dependencias
- Madurez estimada del proyecto (prototipo / en desarrollo / producción)

Este mapa es pasado como contexto a TODOS los agentes subsiguientes. Cada agente adapta sus criterios según el mapa.

### Capa 1 — Agentes de Fase

Diez agentes especializados, cada uno dueño de un dominio:

| ID | Agente | Dominio |
|----|--------|---------|
| A1 | `sequoia-security` | Seguridad, autenticación, superficie de ataque |
| A2 | `sequoia-performance` | Rendimiento, bundles, métricas, cargas |
| A3 | `sequoia-architecture` | Arquitectura, escalabilidad, patrones, límites |
| A4 | `sequoia-ux` | Experiencia de usuario, accesibilidad, flujos |
| A5 | `sequoia-quality` | Testing, cobertura, calidad de código, contratos |
| A6 | `sequoia-product` | Conversión, SEO, funnel, coherencia producto-código |
| A7 | `sequoia-devops` | CI/CD, observabilidad, operaciones, releases |
| A8 | `sequoia-api` | Diseño de API, contratos, versionado, documentación |
| A9 | `sequoia-data` | Modelos de datos, esquemas, migraciones, integridad |
| A10 | `sequoia-deps` | Dependencias, CVEs, licencias, salud del ecosistema |

### Capa 2 — Agentes Meta (Post-run)

| Agente | Rol |
|--------|-----|
| `sequoia-correlator` | Cruza hallazgos entre fases. Detecta causas raíz que generan problemas en múltiples agentes. |
| `sequoia-reporter` | Genera el documento maestro y los markdowns por fase con plantilla uniforme. |
| `sequoia-scorecard` | Calcula el Health Score por fase y global. Produce el semáforo de estado del proyecto. |

---

## Comandos

### Comandos Principales

```bash
/sequoia init
```
Inicializa Sequoia en el proyecto actual. Corre `sequoia-context`, construye el mapa del proyecto, detecta qué agentes son relevantes (no todos aplican a todos los proyectos), y persiste el contexto en Engram.

---

```bash
/sequoia audit
```
Auditoría completa. Corre los 10 agentes de fase en paralelo (donde no hay dependencias) y luego los 3 meta-agentes. Genera todos los entregables.

**Flags**:
- `--phase=security|performance|architecture|ux|quality|product|devops|api|data|deps` — Audita solo una fase
- `--scope=changed` — Solo audita archivos modificados desde el último commit/PR
- `--scope=module=<path>` — Audita un módulo específico
- `--mode=full|quick` — Full: análisis profundo. Quick: solo bloqueantes y hallazgos críticos
- `--output=report|tasks|both` — Qué tipo de entregable generar

---

```bash
/sequoia review
```
Modo revisión de código. Diseñado para PR review o revisión de un diff específico. Corre un subconjunto de agentes relevantes según los archivos cambiados. Más rápido que audit, más profundo que un linter.

**Flags**:
- `--diff=HEAD~1..HEAD` — Rango de commits a revisar
- `--pr=<number>` — Revisar un PR específico (requiere gh CLI)
- `--strict` — Sin tolerancia para hallazgos medios

---

```bash
/sequoia score
```
Genera el Health Scorecard del proyecto. Requiere haber corrido al menos una auditoría previa. Muestra evolución si hay historial.

---

```bash
/sequoia report
```
Regenera los documentos desde hallazgos cacheados. No re-ejecuta agentes.

---

```bash
/sequoia fix <phase> [--task=<id>]
```
Genera plan de tareas accionables desde los hallazgos de una fase, optimizado para que otro agente lo implemente. Incluye: contexto mínimo necesario, archivos candidatos, criterio de aceptación, riesgo.

---

```bash
/sequoia diff
```
Compara el estado actual del proyecto contra la última auditoría registrada. Muestra qué mejoró, qué empeoró, qué aparece nuevo.

---

## Agentes en Detalle

---

### A0 · sequoia-context (Pre-flight)

**Propósito**: Construir el mapa de proyecto que informará a todos los demás agentes.

**Inputs**:
- Estructura de directorios
- Archivos de configuración (cualquier formato detectado)
- Manifiestos de dependencias
- README y documentación interna
- Workflows CI/CD presentes

**Outputs: Mapa de Proyecto**

```markdown
## Project Map
- Stack: [detectado]
- Runtime: [detectado]
- Paradigma: [SPA / SSR / API REST / API GraphQL / CLI / Library / Fullstack / ...]
- Bundler/Build: [detectado o ausente]
- Test infra: [presente / parcial / ausente]
- CI/CD: [presente / parcial / ausente]
- Madurez estimada: [prototipo / desarrollo activo / producción]
- Módulos principales: [lista]
- Dependencias de riesgo: [lista preliminar]
- Agentes aplicables: [lista de A1-A10 que aplican al contexto]
- Agentes no aplicables: [lista con motivo — ej: A9 no aplica, no hay capa de datos]
```

**Mejora clave sobre el original**: El prompt original asume frontend (menciona package.json, vite, hooks, stores). Sequoia-context elimina esa suposición y hace que cada agente adapte sus preguntas al stack real.

---

### A1 · sequoia-security

**Dominio**: Seguridad, autenticación, autorización, superficie de ataque, manejo de secretos.

**Adaptación por contexto**:
- Frontend SPA: foco en tokens, XSS, CSRF, storage inseguro, redirects
- API/Backend: foco en autenticación de endpoints, injection, rate limiting, RBAC
- Fullstack: ambos
- CLI tool: foco en manejo de credenciales, permisos de sistema

**Inspecciones específicas**:
- Manejo de tokens: dónde se guardan, cómo expiran, cómo se rotan
- Persistencia de PII: qué datos se loguean, cachean o exponen en errores
- Logout real vs cosmético (solo UI vs invalidación real de sesión/token)
- Redirects: ¿pueden ser manipulados externamente?
- Superficie XSS: interpolación sin sanitizar, dangerouslySetInnerHTML equivalentes
- Secretos en código: keys, tokens, passwords hardcodeados o en archivos trackeados
- Headers de seguridad si hay servidor propio
- Contratos front/back necesarios para hardening real
- CORS, CSP, cookies (httpOnly, Secure, SameSite)
- Dependencias con CVEs conocidos (coordinado con A10)

**Entregable adicional nuevo**: Matriz de superficie de ataque — tabla de vectores de entrada × estado de mitigación actual.

---

### A2 · sequoia-performance

**Dominio**: Rendimiento, tiempos de carga, render, memoria, bundles, assets.

**Adaptación por contexto**:
- Frontend: Core Web Vitals, bundle splitting, lazy loading, renders innecesarios
- Backend/API: latencia de endpoints, N+1 queries, caché, concurrencia
- Fullstack: ambos
- CLI: tiempo de arranque, uso de memoria

**Inspecciones específicas**:
- Peso de dependencias: ¿qué entra en el bundle que no debería?
- Imports innecesarios o completos cuando solo se usa una función
- Eager vs lazy: ¿qué se carga al inicio que podría diferirse?
- Assets sobredimensionados: imágenes, fuentes, JSON estáticos
- Trabajo de render evitable: cómputos en el render path que podrían cachearse
- Cargas percibidas vs reales: skeletons, optimistic UI, streaming
- Consultas costosas sin índices
- Operaciones bloqueantes en el critical path

**Entregable adicional nuevo**: Presupuesto de performance — tabla con métricas objetivo medibles y forma de verificación:

| Métrica | Objetivo | Cómo medir | Estado actual |
|---------|----------|------------|---------------|
| LCP | < 2.5s | Lighthouse / WebVitals | [verificado] |
| TTI | < 3.5s | Lighthouse | [verificado] |
| Bundle JS | < 300kb gzip | build output | [verificado] |

---

### A3 · sequoia-architecture

**Dominio**: Diseño de sistema, escalabilidad, patrones, límites de módulos, deuda estructural.

**Adaptación por contexto**:
- SPA: stores, routing, state management, service layer
- Backend: capas de dominio, repositorios, servicios, controllers
- Fullstack: integración entre capas, contratos internos
- Library: API pública, encapsulamiento, composabilidad

**Inspecciones específicas**:
- Duplicación funcional: misma lógica en múltiples lugares
- Límites de módulos: ¿cada módulo tiene responsabilidad clara?
- Acoplamiento innecesario: ¿quién sabe demasiado de quién?
- Validación runtime ausente: ¿se confía ciegamente en tipos estáticos en runtime?
- Convenciones inconsistentes: ¿el mismo problema se resuelve de 3 formas distintas?
- Superficie pública innecesaria: APIs internas expuestas sin necesidad
- Puntos de escalado peligroso: ¿qué se rompe primero si el proyecto crece 10x?
- God objects/modules: componentes o módulos que saben demasiado
- Antipatrones específicos del stack detectado

**Entregable adicional nuevo**: Mapa de dependencias de módulos — diagrama textual de quién depende de quién, con flechas de acoplamiento identificadas.

---

### A4 · sequoia-ux

**Dominio**: Experiencia de usuario, accesibilidad, flujos, responsive, onboarding, interacciones.

**Adaptación por contexto**:
- Solo aplica a proyectos con interfaz de usuario (web, mobile, desktop, CLI interactivo)
- Para APIs puras: aplica a la experiencia del desarrollador (DX) como "UX de la API"

**Inspecciones específicas**:
- Bloqueos de flujo: pasos donde el usuario puede quedar atascado sin salida
- Errores sin mensaje de recuperación o sin acción posible
- Estados de carga pobres: spinners eternos, ausencia de feedback
- Accesibilidad real: roles ARIA, contraste, navegación por teclado, screen reader
- Responsive/mobile: ¿el layout se rompe en viewports reales?
- Fricción en onboarding: ¿cuántos pasos hasta el primer valor percibido?
- Modales, tablas, formularios, menús, tabs: ¿cada uno tiene estado de error, vacío y carga?
- Formularios: ¿validación en tiempo real? ¿mensajes claros? ¿recovery de errores?

**Nueva inspección**: Developer Experience (DX) para proyectos tipo biblioteca o API — ¿qué tan difícil es empezar a usar este código?

---

### A5 · sequoia-quality

**Dominio**: Testing, cobertura, calidad de código, deuda técnica medible.

**Adaptación por contexto**:
- Detecta el framework de testing presente (o su ausencia)
- Adapta las recomendaciones según el tipo de proyecto y su madurez

**Inspecciones específicas**:
- Estado real de tests: ¿cuántos existen? ¿pasan? ¿cubren qué?
- Infraestructura ausente: ¿hay test runner, fixtures, mocks, factories?
- Módulos de mayor riesgo sin cobertura (identificados junto con A3)
- Calidad de tests existentes: ¿testean comportamiento o implementación?
- Tests de contrato: ¿hay validación de que el contrato API se cumple?
- Smoke tests mínimos: ¿hay algo que verifique que el proyecto arranca?
- Deuda de código: complejidad ciclomática alta, funciones largas, duplicación
- Lint/format: ¿está configurado? ¿corre en CI? ¿tiene reglas serias?

**Nueva inspección**: Mutation testing readiness — ¿los tests detectarían si se cambia un operador lógico? No pedir implementación; evaluar si los tests tienen esa capacidad.

**Estrategia incremental obligatoria**: No proponer "llegar al 80% de cobertura". Proponer el camino mínimo viable: primero smoke, luego módulos críticos, luego integración.

---

### A6 · sequoia-product

**Dominio**: Conversión, SEO, coherencia producto-código, funnel, metadatos, rutas legales.

**Adaptación por contexto**:
- Solo aplica a proyectos con presencia pública (web, SaaS, e-commerce, landing)
- Para apps internas o APIs: aplica como coherencia entre documentación y funcionalidad real

**Inspecciones específicas**:
- Coherencia entre lo que promete el producto y lo que el código implementa
- CTAs: ¿funcionan? ¿llevan a donde dicen llevar? ¿tienen tracking?
- Funnel real: ¿se puede rastrear el camino desde "llega" hasta "convierte"?
- Callejones sin salida: páginas o estados desde los que el usuario no puede avanzar ni volver
- SEO técnico: meta tags, Open Graph, robots.txt, sitemap, canonical
- Rutas legales: privacidad, términos, cookies — ¿existen? ¿son alcanzables?
- Performance de landing: separado de la app (el primero es conversión, el segundo es uso)
- Instrumentación mínima: ¿hay algún mecanismo para saber si alguien convierte?

**Nueva inspección**: Content-code drift — ¿el copy en el código coincide con la propuesta de valor real del producto? Detectar promesas que el producto no cumple.

---

### A7 · sequoia-devops

**Dominio**: CI/CD, operaciones, releases, monitoring, observabilidad, guardrails del repo.

**Adaptación por contexto**:
- Para proyectos pequeños/personales: foco en lo mínimo viable (lint en CI, env contract, básico de monitoring)
- Para producción: foco en staging, previews, rollback, alertas, postmortem capability

**Inspecciones específicas**:
- Scripts faltantes: start, build, test, lint, typecheck — ¿todos existen y funcionan?
- Env contract: ¿hay `.env.example` o similar? ¿están documentadas todas las variables requeridas?
- `.env` trackeado: ¿hay archivos de entorno con secretos reales en el repo?
- Hooks de Git: pre-commit, pre-push — ¿existen? ¿tienen guardrails serios?
- CI/CD: ¿existe? ¿corre tests? ¿bloquea el merge si falla?
- Staging/previews: ¿hay entorno de validación antes de producción?
- Monitoring básico: ¿hay uptime monitoring? ¿se sabe cuándo cae?
- Logger: ¿hay logging estructurado? ¿los errores son observables?
- Observabilidad: ¿hay error tracking (Sentry equivalente)?
- Documentación operativa: ¿hay runbook mínimo? ¿cómo se despliega? ¿cómo se rollbackea?
- Política de releases: ¿hay semver? ¿changelog? ¿tags?
- Salud de dependencias: ¿se actualiza regularmente? ¿hay scanning de CVEs?

**Entregable adicional nuevo**: Ownership mínimo — tabla de quién es responsible de qué en operaciones (aunque sea una sola persona).

---

### A8 · sequoia-api (NUEVO)

**Dominio**: Diseño de API, contratos, versionado, documentación de interfaz, DX para consumidores.

**Cuándo aplica**: Proyectos con APIs REST, GraphQL, RPC, SDKs, o cualquier interfaz programática pública o entre servicios.

**Inspecciones específicas**:
- Consistencia de naming: ¿snake_case, camelCase, kebab-case? ¿es consistente?
- Verbos HTTP correctos: ¿se usa POST donde debería ser PUT/PATCH?
- Estructura de respuestas: ¿errores y éxitos tienen estructura predecible?
- Manejo de errores: ¿los errores incluyen código, mensaje, y contexto útil?
- Versionado: ¿hay estrategia? ¿hay breaking changes sin versión?
- Documentación: ¿hay OpenAPI/Swagger, GraphQL schema, o equivalente?
- Autenticación de endpoints: ¿todos los endpoints privados están protegidos?
- Rate limiting y throttling: ¿hay protección contra abuso?
- Paginación: ¿los endpoints que devuelven listas tienen paginación?
- Contratos internos: ¿servicios que se llaman entre sí tienen contrato documentado?

**Entregable nuevo**: Breaking Change Risk Map — lista de endpoints o contratos que, si cambian, romperían consumidores conocidos.

---

### A9 · sequoia-data (NUEVO)

**Dominio**: Modelos de datos, esquemas, migraciones, integridad, privacidad de datos.

**Cuándo aplica**: Proyectos con base de datos, esquemas definidos, ORMs, migraciones, o modelos de dominio explícitos.

**Inspecciones específicas**:
- Normalización: ¿hay duplicación de datos que podría generar inconsistencias?
- Índices: ¿las consultas frecuentes tienen índices apropiados?
- Migraciones: ¿el proceso de migración es reversible? ¿hay riesgo de pérdida de datos?
- Validación de integridad: ¿hay constraints a nivel de base de datos, no solo en aplicación?
- Soft delete vs hard delete: ¿la estrategia es explícita y consistente?
- PII y datos sensibles: ¿se guardan datos personales? ¿con qué nivel de protección?
- Backups: ¿hay estrategia documentada?
- Schema drift: ¿el esquema en código coincide con el esquema real en producción?
- Queries N+1: ¿hay carga de relaciones sin optimizar?

---

### A10 · sequoia-deps (NUEVO)

**Dominio**: Dependencias, seguridad del ecosistema, licencias, salud a largo plazo.

**Cuándo aplica**: Siempre. Todo proyecto con dependencias externas.

**Inspecciones específicas**:
- CVEs conocidos: dependencias con vulnerabilidades públicas
- Versiones desactualizadas: major versions atrás con cambios de seguridad relevantes
- Dependencias abandonadas: sin mantenimiento en 2+ años
- Licencias incompatibles: ¿hay dependencias GPL en un proyecto comercial privado?
- Dependencias no utilizadas: código muerto en el manifiesto
- Lock file presente: ¿hay lockfile? ¿está commiteado?
- Dependencias duplicadas: misma funcionalidad resuelta con dos librerías
- Dependencias de un solo contribuyente: riesgo de bus factor externo

**Entregable nuevo**: Risk Score de dependencias — tabla con columnas: nombre, versión actual, versión latest, CVEs conocidos, mantenimiento, acción recomendada.

---

### M1 · sequoia-correlator (Meta)

**Propósito**: Encontrar causas raíz transversales. Muchos síntomas en fases distintas tienen la misma causa.

**Ejemplo real**:
- A3 detecta: "no hay validación runtime"
- A1 detecta: "datos del usuario no se sanitizan antes de usar"
- A5 detecta: "no hay tests de los módulos de entrada de datos"
- **Correlación**: la causa raíz es una sola — ausencia de una capa de validación de inputs

**Output**: Lista de causas raíz con sus síntomas en múltiples fases, priorizada por impacto agregado.

---

### M2 · sequoia-reporter (Meta)

**Propósito**: Generar los entregables finales con plantilla uniforme.

**Entregables**:
- `sequoia-master.md` — Documento maestro con resumen ejecutivo, severidades por fase, roadmap
- `sequoia-phases/01-security.md` ... `10-deps.md` — Un markdown por fase
- `sequoia-score.md` — Health scorecard

---

### M3 · sequoia-scorecard (Meta)

**Propósito**: Calcular métricas de salud del proyecto por fase y globalmente.

**Health Score por fase**:

```
🔴 CRÍTICO    — Bloqueantes de producción o seguridad
🟠 RIESGO     — Problemas serios sin solución activa
🟡 ATENCIÓN   — Deuda técnica priorizable
🟢 SALUDABLE  — Bajo nivel de hallazgos serios
⚪ N/A         — Fase no aplica al proyecto
```

**Health Score global**: promedio ponderado (seguridad y devops pesan más).

---

## Reglas Innegociables (heredadas y extendidas)

Estas reglas aplican a TODOS los agentes sin excepción:

1. **No asumir**. Si algo no está en el repo, declararlo como ausente, no inventarlo.
2. **No confirmar claims sin verificar**. Ni los del usuario, ni los de documentación interna.
3. **Si no es verificable, decirlo explícitamente** con el label `[NO VERIFICABLE]`.
4. **Citar archivos reales**. Si mencionás un archivo, tiene que existir con ese path.
5. **No teoría genérica**. Evidencia del repo o silencio.
6. **No checklists decorativos**. Cada ítem debe ser verificable y accionable.
7. **No soluciones mágicas**. Cada recomendación incluye impacto, dependencias y riesgo.
8. **Si repo y documentación previa se contradicen, prevalece el repo**.
9. **No cambios destructivos salvo instrucción explícita**.
10. **Adaptar al contexto real del proyecto**. No aplicar criterios enterprise a un prototipo.

**Nueva regla 11**: Si un agente no puede verificar algo porque requiere acceso externo (infra, DB en producción, logs de Sentry), lo declara como `[REQUIERE ACCESO EXTERNO]` y describe qué verificar y cómo.

**Nueva regla 12**: Si una recomendación solo aplica si el proyecto crece (futura escala), marcarlo con `[SOLO SI ESCALA]`. No mezclarlo con recomendaciones para el estado actual.

---

## Formato Estándar de Hallazgo

Todos los agentes usan exactamente esta estructura:

```markdown
### [FASE-ID] · [Título del hallazgo]  [🔴 CRÍTICO | 🟠 RIESGO | 🟡 ATENCIÓN]

**Estado**: Confirmado | Parcial | No verificable | Desactualizado

**Evidencia**:
- `path/real/al/archivo.ext:línea` — descripción de lo observado
- Comportamiento o ausencia detectada

**Problema**:
Qué está mal y por qué técnicamente importa. Sin generalidades.

**Impacto real**:
Qué puede pasar en producción si esto sigue así.

**Recomendación mínima de alto leverage**:
Qué cambio concreto conviene hacer primero y por qué ese específicamente.

**Dependencias / bloqueos**:
Backend, infra, contrato de API, otros módulos, equipo externo, etc.

**Riesgo de implementación**: Bajo | Medio | Alto
Motivo del riesgo estimado.

**Criterio de aceptación**:
Cómo verificar que el hallazgo fue resuelto.
```

---

## Plantilla Obligatoria por Fase

Cada documento de fase generado por `sequoia-reporter` usa exactamente esta estructura:

```markdown
# Fase [N] — [Nombre]

**Agente**: sequoia-[nombre]
**Proyecto**: [nombre]
**Fecha de auditoría**: [fecha]
**Stack detectado**: [del mapa de proyecto]

## 1. Objetivo de la fase

## 2. Scope de inspección
Qué archivos, directorios y configuraciones fueron revisados.
Qué quedó fuera del scope y por qué.

## 3. Estado actual verificado
### Qué dice la documentación interna (si existe)
### Qué se confirmó en código
### Qué quedó desactualizado, ambiguo o no verificable

## 4. Hallazgos consolidados
Ordenados por severidad. Sin duplicados. Sin relleno.

## 5. Faltantes de alto leverage
Solo mejoras justificadas técnicamente. Con impacto esperado.

## 6. Plan de tareas
Cada tarea con: contexto, archivos, impacto, dependencias, riesgo, criterio de aceptación, prioridad.

## 7. Orden de implementación recomendado
Secuencia que minimiza riesgo y maximiza impacto.

## 8. Riesgos y bloqueos de la fase

## 9. Checklist de cierre de fase
Lista verificable de qué debe ser verdad cuando esta fase esté "done".
```

---

## Entregables Finales

### Documento Maestro: `sequoia-master.md`

```markdown
# Sequoia Audit Report — [Proyecto]

**Fecha**: [fecha]
**Modo**: Full | Quick | Phase | Review
**Stack**: [detectado]
**Madurez estimada**: [del mapa de contexto]

## Resumen Ejecutivo
Estado general del proyecto en 5-10 líneas. Sin exageración.

## Health Scorecard
| Fase | Agente | Score | Bloqueantes | Alto Leverage | Backlog |
|------|--------|-------|-------------|---------------|---------|
| Seguridad | A1 | 🟠 | 2 | 3 | 1 |
| ... | | | | | |

## Top 10 Hallazgos Globales
Los más críticos, sin importar la fase. Priorizados por impacto × urgencia.

## Causas Raíz Transversales
Output del sequoia-correlator.

## Roadmap Sugerido
Ordenado por: bloquea producción → alto leverage → backlog → aceptable.

## Fases No Aplicables
Lista con motivo (del mapa de contexto).
```

---

## Flujos de Trabajo

### Flujo 1: Auditoría Completa Nueva

```
/sequoia init
  └─► sequoia-context → Project Map

/sequoia audit
  ├─► A1-A10 en paralelo (donde no hay dependencias)
  │     Cada agente recibe: Project Map + instrucción de su dominio
  ├─► M1 (sequoia-correlator) — después de todos los agentes
  ├─► M3 (sequoia-scorecard)
  └─► M2 (sequoia-reporter) — genera todos los documentos
```

### Flujo 2: Revisión de PR / Diff

```
/sequoia review --diff=HEAD~1..HEAD
  └─► sequoia-context (solo archivos cambiados)
        └─► Selección automática de agentes relevantes
              (basado en qué tipos de archivos cambiaron)
        └─► Hallazgos solo sobre el diff
        └─► Flag si el cambio toca áreas con hallazgos previos abiertos
```

### Flujo 3: Auditoría Incremental (Re-run)

```
/sequoia diff
  └─► Compara estado actual vs última auditoría en Engram
        └─► Muestra: resuelto / nuevo / empeorado / sin cambio
```

### Flujo 4: Generar Tareas para Agente Implementador

```
/sequoia fix security
  └─► sequoia-reporter genera plan de tareas de A1
        Formato: cada tarea autosuficiente, con contexto mínimo para otro agente
        Sin necesidad de releer toda la auditoría
```

---

## Configuración: `sequoia.config.json`

Archivo opcional en la raíz del proyecto:

```json
{
  "project": "nombre-del-proyecto",
  "maturity": "prototype | development | production",
  "agents": {
    "disabled": ["A6", "A9"],
    "focus": ["A1", "A7"]
  },
  "thresholds": {
    "security": "strict",
    "performance": "standard",
    "quality": "relaxed"
  },
  "outputs": {
    "dir": "docs/sequoia",
    "master": true,
    "phases": true,
    "scorecard": true
  },
  "context": {
    "stack": "auto",
    "entryPoints": ["src/main.ts", "app/index.tsx"],
    "excludeDirs": ["node_modules", "dist", ".git"]
  }
}
```

**Niveles de umbral por fase**:
- `strict` — tolerancia cero para hallazgos medios
- `standard` — solo reporta medios+ (default)
- `relaxed` — solo reporta críticos y riesgos altos

---

## Integración con Claude Code

### Hooks Sugeridos

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{
          "type": "command",
          "command": "echo 'Sequoia: operación de escritura detectada'"
        }]
      }
    ]
  }
}
```

### Slash Commands (Skills)

| Comando | Skill | Descripción |
|---------|-------|-------------|
| `/sequoia` | `sequoia:orchestrator` | Entry point principal |
| `/sequoia-security` | `sequoia:security` | Auditoría de seguridad |
| `/sequoia-review` | `sequoia:review` | Revisión de PR/diff |
| `/sequoia-score` | `sequoia:scorecard` | Health scorecard |
| `/sequoia-fix` | `sequoia:fix` | Genera plan de tareas |

### Integración con Engram (Memoria Persistente)

Sequoia persiste en Engram:
- El Project Map de cada proyecto
- Los hallazgos de cada auditoría con timestamp
- El historial de Health Scores para ver evolución
- Las tareas generadas y su estado

Esto permite:
- `sequoia diff`: comparar estado actual vs auditoría anterior
- Que el agente implementador recuerde hallazgos de sesiones anteriores
- Evolución de la salud del proyecto a lo largo del tiempo

---

## Mejoras sobre el Prompt Original

| Aspecto | Original | Sequoia |
|---------|----------|---------|
| **Stack** | Asume frontend (package.json, vite, hooks) | Auto-detect: cualquier stack |
| **Fases** | 7 fases | 10 agentes de fase + 3 meta |
| **Nuevas áreas** | — | API Design, Data/Schema, Dependency Health |
| **Modo revisión** | Solo auditoría completa | Audit + Review (PR/diff) + Incremental |
| **Correlación** | No existe | sequoia-correlator: causas raíz transversales |
| **Scoring** | No existe | Health Scorecard por fase y global |
| **Persistencia** | No existe | Engram: historial, diff, evolución |
| **Configuración** | No existe | sequoia.config.json: umbrales, agentes, outputs |
| **DX de API** | No existe | A8 incluye DX para consumidores de API |
| **PII/Data** | Mencionado en seguridad | A9 dedicado a datos, integridad, privacidad |
| **Deps** | Mencionado en DevOps | A10 dedicado con CVEs, licencias, risk score |
| **Tareas para agente** | Formato estático | /sequoia fix: output optimizado para implementador |
| **Madurez del proyecto** | Ignora contexto | Adapta criterios según madurez (prototipo vs producción) |
| **Hallazgos futuros** | Mezclados con actuales | `[SOLO SI ESCALA]` los separa explícitamente |
| **Acceso externo** | No contemplado | `[REQUIERE ACCESO EXTERNO]` declara límites verificables |

---

## Principio de Cierre

Sequoia no es una checklist. Es un equipo de agentes que razonan, correlacionan y priorizan.

Un proyecto auditado por Sequoia debe quedar con:
1. Un mapa de estado real — no aspiracional
2. Una lista de hallazgos trazables a evidencia real
3. Un roadmap que cualquier arquitecto firmaría
4. Tareas que cualquier agente implementador puede ejecutar sin ambigüedad
5. Un score que evoluciona con el proyecto

El objetivo no es encontrar problemas. El objetivo es dejar el proyecto entendido, priorizado y listo para intervenir.
