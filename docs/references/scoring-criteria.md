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

## Health Grade

| Score | Grade |
|-------|-------|
| 90–100 | A |
| 75–89 | B |
| 60–74 | C |
| 40–59 | D |
| 0–39 | F |

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
  security:     1.3  (always applied)
  architecture: 1.1  (always applied)
  performance:  1.0  (if P2 ran)
  quality:      1.0  (always applied)
  experience:   0.9  (if P5 ran)
  operations:   0.9  (if P6 ran)
```

Categories not applicable to the project type (e.g., experience for a CLI) are excluded from both numerator and denominator.

---

## Implementation Notes

- The formula is identical across `SKILL.md`, `ARCHITECTURE.md`, and this file. Any divergence is a documentation bug.
- Emoji presence in a finding's display data does not alter which severity level is applied.
- The scope multiplier is determined by the correlator (M1) during Phase 4. Phase agents do not set multipliers — they report findings with severity only.
- info-severity findings never deduct points (weight = 0), but they appear in the report for informational purposes.
