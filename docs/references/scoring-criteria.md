# Scoring Criteria — Sequoia Health Score

Reference document for the canonical scoring formula used across all Sequoia documentation and agents.

---

## Canonical Formula

```
score = 100 − Σ(severity_weight × scope_multiplier)
```

Score is floored at 0 (cannot be negative).

---

## Severity Weights

| Severity | Weight |
|----------|--------|
| critical | 15 |
| high | 8 |
| medium | 4 |
| low | 2 |
| info | 0 |

---

## Scope Multiplier

| Condition | Multiplier |
|-----------|------------|
| Isolated finding (independent root cause) | 1.0 |
| Shared root cause (≥2 findings share same root) | 1.5 |

A finding qualifies for the 1.5 multiplier when the correlator (M1) has identified it as a manifestation of a shared root cause. The multiplier reflects that systemic problems are more severe than isolated ones.

---

## Severity Emoji Mapping (Presentation Only)

Emojis are used for visual display in reports. They do NOT affect score computation. Emojis are stripped before any calculation.

| Severity | Emoji |
|----------|-------|
| critical | 🔴 |
| high | 🟠 |
| medium | 🟡 |
| low | 🟢 |
| info | 🔵 |

---

## Worked Example

Audit produces:
- 1 critical finding — isolated root cause
- 2 high findings — shared root cause (same underlying problem)
- 1 low finding — isolated root cause

Calculation:

```
1 critical (×1.0)  = 15 × 1.0 = 15
2 high     (×1.5)  = 8 × 1.5 + 8 × 1.5 = 12 + 12 = 24
1 low      (×1.0)  = 2 × 1.0 = 2

Total deducted = 15 + 24 + 2 = 41
score = 100 − 41 = 59
```

Score 59 → Grade **D** (40–59 range).

This example is the canonical verification case. Any implementation of the formula MUST produce 59 for this input.

---

## Global Score (Multi-Category Audits)

When Sequoia runs a full audit, each category (security, performance, architecture, quality, experience, operations) receives its own score using the formula above. The global score is a weighted average:

```
global_score = Σ(category_score × weight) / Σ(applicable_weights)

weights:
  security:     0.25  (always applied)
  architecture: 0.20  (always applied)
  performance:  0.15  (if P2 ran)
  quality:      0.15  (always applied)
  experience:   0.10  (if P5 ran; 0 if not applicable, weight redistributed)
  operations:   0.15  (if P6 ran)
```

**Classification**:
| Score | Grade | Meaning |
|-------|-------|---------|
| 90–100 | A — Excellent | Production-ready, preventive maintenance |
| 75–89 | B — Good | Minor issues, improve gradually |
| 60–74 | C — Fair | Significant problems, action plan required |
| 40–59 | D — Deficient | Serious problems, priority action |
| 0–39 | F — Critical | Immediate risk, urgent action |

Categories not applicable to the project type (e.g., experience for a CLI) are excluded from both numerator and denominator. Their weight is redistributed proportionally among remaining categories.

---

## Implementation Notes

- The formula is identical across `SKILL.md`, `ARCHITECTURE.md`, and this file. Any divergence is a documentation bug.
- Emoji presence in a finding's display data does not alter which severity level is applied.
- The scope multiplier is determined by the correlator (M1) during Phase 4. Phase agents do not set multipliers — they report findings with severity only.
- info-severity findings never deduct points (weight = 0), but they appear in the report for informational purposes.
