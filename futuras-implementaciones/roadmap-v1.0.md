# Roadmap v1.0 — Nuevos Dominios de Auditoría

> Basado en el audit del 2026-05-11 (Score 78/100). Agentes actuales: P1–P6.

---

## Dominios nuevos evaluados

Se identificaron 10 dominios adicionales que harían la auditoría más completa:

| # | Dominio | ¿Quién lo cubre hoy? | Conflicto |
|---|---------|----------------------|-----------|
| 1 | API Design | P3 Architecture (parcial) | 🔴 Overlap |
| 2 | Data Layer | P2 (queries) + P3 (data flow) | 🟡 Overlap parcial |
| 3 | Dependencies / Supply Chain | P4 Quality (superficial) + P1 (CVEs) | 🟡 Overlap parcial |
| 4 | Testing Strategy | P4 Quality (coverage, no strategy) | 🟡 Overlap parcial |
| 5 | Documentation Health | P4 Quality (evalúa docs) | 🔴 Overlap |
| 6 | Observability | P6 Operations | 🔴 Overlap |
| 7 | Configuration Management | P1 Security + P6 Operations | 🟡 Overlap parcial |
| 8 | **i18n / Internationalization** | **Ninguno** | 🟢 Sin conflicto |
| 9 | Accessibility (a11y) | P5 Experience | 🔴 Overlap |
| 10 | Resilience | P3 Architecture (patrones) | 🟡 Overlap parcial |

---

## Mapa de conflictos: qué choca con qué

### 🔴 Choca directamente (scope ya cubierto)

| Dominio nuevo | Agente actual | Scope ya existente |
|--------------|---------------|-------------------|
| API Design | P3 Architecture | "API design, service boundaries, error handling, versionado" |
| Observability | P6 Operations | "observability, CI/CD" |
| Documentation Health | P4 Quality | "test coverage, naming, dead code, **documentation**" |
| Accessibility | P5 Experience | "accessibility, UX patterns" |

**Conclusión**: Crear agentes separados para estos duplicaría hallazgos y generaría contradicciones entre el agente nuevo y el existente. **No se recomienda.**

### 🟡 Overlap parcial (se puede absorber expandiendo)

| Dominio nuevo | Expandir en | Cómo |
|--------------|-------------|------|
| Data Layer | P3 Architecture | Sub-dominio: schemas, migraciones, integridad referencial, ORM usage |
| Supply Chain | P4 Quality | Flags opcionales: `--cve-scan`, `--license-check`, `--sbom` |
| Testing Strategy | P4 Quality | Sub-dominio: test pyramid balance, flaky test detection, mutation testing |
| Config Management | P1 Security | Sub-dominio: feature flags, env contract validation, config fail-fast |
| Resilience | P3 Architecture | Sub-dominio: circuit breakers, retries, timeouts, graceful degradation |

**Conclusión**: Se profundizan los agentes existentes sin cambiar su contrato. El nuevo scope es ortogonal al actual.

### 🟢 Sin conflicto (territorio virgen)

| Dominio nuevo | Agente | Nota |
|--------------|--------|------|
| **i18n** | P7 Internationalization | Nadie audita strings hardcodeados, RTL, formatos locale, extracción de traducciones |

---

## Recomendación final para v1.0 estable

### Lo que entra (3 cambios)

| # | Cambio | Tipo | Esfuerzo estimado |
|---|--------|------|-------------------|
| **1** | **P7 i18n** — agente nuevo | Nuevo agente | 4–6h (skill + templates + tests) |
| **2** | **P4 Quality: Deep Deps Scan** | Expansión con flags | 3–4h (CVE + licencias + SBOM) |
| **3** | **P3 Architecture: Resilience Patterns** | Expansión de scope | 2–3h (circuit breakers, retries, timeouts) |

```
Agentes que cambian:  3 (1 nuevo, 2 expandidos)
Agentes intactos:     P1, P2, P5, P6
Conflictos:            0
Hallazgos duplicados:  0
```

### Lo que se posterga a v1.1+

| Dominio | Razón |
|--------|-------|
| **Data Layer** como agente separado | La frontera con P2 (query perf) y P3 (data architecture) es difusa. Requiere definir bien el contrato para no pisar queries con diseño de esquema. |
| **Testing Strategy** como flag de P4 | Valioso pero requiere métricas de CI (flaky test history, test duration trends) que muchos proyectos no tienen. Mejor esperar a tener el baseline de P4 estable. |
| **Feature Flags** como sub-dominio de P1 | Nicho — muchos proyectos no usan feature flags. Como flag opcional de P1 no rompe nada pero suma complejidad sin alto uso esperado. |
| **Config fail-fast** como sub-dominio de P1 | Mismo razonamiento — valioso pero el overhead de implementación es mayor que el beneficio inmediato. |

---

## Roadmap visual

```
v0.1.0 (actual)          v1.0 (estable)             v1.1                     v1.2
─────────────            ────────────                ──────                   ──────
P1 Security              P1 Security                 P1 + Feature Flags
P2 Performance           P2 Performance              P2
P3 Architecture          P3 + Resilience ─────────→  P3 + Data Layer
P4 Quality               P4 + Deep Deps ──────────→  P4 + Testing Strategy
P5 Experience            P5 Experience
P6 Operations            P6 Operations
                         P7 i18n ✨                  
```

---

## Criterios de aceptación para v1.0

- [ ] P7 i18n se ejecuta sin modificar P1–P6
- [ ] P7 i18n no genera hallazgos que dupliquen los de ningún agente existente
- [ ] P3 Resilience no contradice hallazgos de P3 Architecture existentes
- [ ] P4 Deep Deps no contradice hallazgos de P4 Quality existentes
- [ ] Los 3 cambios pasan el test suite completo de Sequoia
- [ ] El Project Map actualizado refleja los nuevos agentes y sub-dominios
- [ ] Los docs de agentes se actualizan (skills, commands, SEQUOIA.md)

---

## Referencias

- [Audit actual — Master Report](../docs/sequoia/sequoia-master.md)
- [Audit actual — Health Score](../docs/sequoia/sequoia-score.md)
- [Audit actual — Tasks](../docs/sequoia/sequoia-tasks.md)
- [Sequoia Architecture](../docs/SEQUOIA.md)
