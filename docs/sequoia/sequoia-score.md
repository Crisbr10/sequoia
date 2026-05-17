# Sequoia Health Score — sequoia-ai

**Audit date**: 2026-05-12
**Mode**: Full
**Scope**: Full project (59 findings across 6 phases)
**Methodology**: `score = 100 − Σ(severity_weight × scope_multiplier)`

---

## Global Score

| | |
|---|---|
| **Health Score** | **28 / 100** |
| **Grade** | **F** |
| **Severity scale** | 🔴 CRÍTICO=15 · 🟠 ALTO=8 · 🟡 MEDIO=4 · 🟢 BAJO=2 · 🔵 INFO=0 |
| **Multiplier** | ×1.0 isolated · ×1.5 shared root cause (≥2 findings) |

---

## Score by Category

| Categoría | Deducción | Score | Grade | Peso | Ponderado |
|-----------|-----------|-------|-------|------|-----------|
| Security (P1) | -70.0 | **30** | F | ×1.3 | 39.0 |
| Architecture (P3) | -99.0 | **1** | F | ×1.1 | 1.1 |
| Performance (P2) | -54.5 | **46** | D | ×1.0 | 45.5 |
| Quality (P4) | -105.0 | **0** | F | ×1.0 | 0.0 |
| Operations (P6) | -61.0 | **39** | F | ×0.9 | 35.1 |
| i18n (P7) | -42.0 | **58** | D | ×0.9 | 52.2 |
| Experience (P5) | — | N/A | N/A | — | — |

**Global**: Σ ponderado / Σ pesos = 172.9 / 6.2 = **27.9 → 28 (F)**

---

## Findings by Severity

| Severidad | Count | % |
|-----------|-------|---|
| 🔴 CRÍTICO | 6 | 10% |
| 🟠 ALTO | 15 | 25% |
| 🟡 MEDIO | 37 | 63% |
| 🟢 INFO | 1 | 2% |
| **Total** | **59** | 100% |

---

## Top 5 Deductors

| # | Hallazgo | Severidad | Scope | Deducción |
|---|----------|-----------|-------|-----------|
| 1 | P3-001 + P4-001: Duplicación masiva de código en 5 adaptadores (70-85%) | 🔴 ×2 | ×1.5 | -45.0 |
| 2 | P3-002: Funciones de estrategia duplicadas byte-por-byte | 🔴 | ×1.5 | -22.5 |
| 3 | P2-001: 248 KB de plantillas embebidas duplicadas | 🔴 | ×1.5 | -22.5 |
| 4 | P4-002: `_template` compilable y auto-registrable en producción | 🔴 | ×1.5 | -22.5 |
| 5 | P6-001: CI con Go 1.22 pero go.mod dice 1.24.2 | 🔴 | ×1.0 | -15.0 |

---

## Root Causes Identified (M1 Correlator)

| # | Root Cause | Findings | Severity |
|---|-----------|-----------|----------|
| R1 | **No common adapter layer** — 5 adapters duplicate 70-85% of code | P3-001, P3-002, P4-001, P4-005 | 🔴 CRITICAL |
| R2 | **Minimal CI without automation** — Go version mismatch, no linting, no coverage, no Dependabot | P4-004, P6-001, P6-002, P6-003, P6-004, P6-005, P6-006, P6-007 | 🟠 HIGH |
| R3 | **Dead i18n** — Language selector exists but no adapter uses it, strings hardcoded | P1-006, P4-003, P7-001, P7-002, P7-005, P7-006 | 🟠 HIGH |
| R4 | **No error taxonomy** — Single sentinel error, no codes, no classification | P1-004, P3-010, P7-006 | 🟡 MEDIUM |
| R5 | **Over-designed pipeline** — 3 cosmetic steps vs 1 real step, dead spinner | P2-007, P3-006, P3-012, P4-005 | 🟡 MEDIUM |
| R6 | **`_template` package in production** — Compilable, auto-registers, 20 TODOs | P2-008, P4-002, P4-011 | 🟡 MEDIUM |

---

## Distribution Chart

```
🔴 CRÍTICO  ██████░░░░░░░░░░░░░░░░░░░░░░ 6
🟠 ALTO     ████████████████░░░░░░░░░░░░ 15
🟡 MEDIO    ████████████████████████████████████████ 37
🟢 INFO     █ 1
```

---

## Recommended Trajectory

Resolving R1 (common adapter layer) and R2 (minimal CI) in the current sprint would raise the score from ~28 to ~55 (D→C). Resolving R1-R6 completely would bring the score to ~78 (C→B).

```
Actual:  ██████████████░░░░░░░░░░░░░░░░░░░░░░░░░░ 28/100  F
Post-R1: ████████████████████░░░░░░░░░░░░░░░░░░░░ 40/100  D
Post-R2: █████████████████████████████░░░░░░░░░░░░ 55/100  D
Post-R3: █████████████████████████████████░░░░░░░░ 65/100  C
Post-R6: ██████████████████████████████████████░░░ 78/100  B
```

---

*Score calculado con la fórmula canónica de Sequoia v0.1.0 · references/scoring-criteria.md*
