# Standard Finding Format

Mandatory format for ALL Sequoia agents. No exceptions.

## Markdown template

```markdown
### [PHASE-ID] · [Finding title]  [🔴 CRITICAL | 🟠 RISK | 🟡 ATTENTION]

**Status**: Confirmed | Partial | Not verifiable | Outdated

**Evidence**:
- `path/to/real/file.ext:line` — description of what was observed
- Detected behavior or absence

**Problem**:
What is wrong and why it technically matters. No generalities.

**Real impact**:
What can happen in production if this continues.

**Minimum high-leverage recommendation**:
What concrete change to make first and why specifically that one.

**Dependencies / blockers**:
Backend, infra, API contract, other modules, external team, etc.

**Implementation risk**: Low | Medium | High
Reason for the estimated risk.

**Acceptance criteria**:
How to verify that the finding was resolved.
```

## Guide for each field

### Finding ID

Format: `[AGENT-ID}-{NNN}]`

- `C1` = Context, `P1` = Security, `P2` = Performance, `P3` = Architecture, `P4` = Quality, `P5` = Experience, `P6` = Operations
- `M1` = Correlator, `M2` = Reporter
- `NNN` = sequential number within the phase
- Example: `[P1-001]`, `[P3-012]`, `[P4-003]`

### Severity

| Level | Emoji | When to use it |
|-------|-------|---------------|
| CRITICAL | 🔴 | Production blocker, exploitable security, data loss |
| RISK | 🟠 | Serious problem without active solution, likely degradation under load |
| ATTENTION | 🟡 | Prioritizable technical debt, quality improvement, future-proofing |

**Rule**: if you hesitate between two levels, use the lower one. It's better to under-estimate than generate alert fatigue.

### Status

| Status | Meaning |
|--------|-------------|
| Confirmed | Verified against real code, solid evidence |
| Partial | Partially verified, some aspect not verifiable |
| Not verifiable | Requires external access (infra, production DB, logs) |
| Outdated | The finding was valid but the code has since changed |

**Additional markers**:
- `[REQUIRES EXTERNAL ACCESS]` — when it cannot be verified without access to infra/logs
- `[ONLY IF SCALING]` — when the recommendation only applies if the project grows

### Evidence

**Mandatory** to cite real files with line numbers:

```markdown
**Evidence**:
- `src/auth/handler.ts:45` — token stored in localStorage without expiration
- `src/api/middleware.ts:12` — auth middleware doesn't validate refresh token
- Detected absence: `.env.example` file does not exist
```

If there is no file (detected absence), declare it explicitly.

### Problem

Technical description of the problem, not generic.

❌ Bad: "Token handling is not secure."
✅ Good: "The JWT is stored in localStorage, accessible by XSS. There is no revocation mechanism or refresh token rotation."

### Real impact

What happens in production, not in theory.

❌ Bad: "There could be security problems."
✅ Good: "An XSS attack can steal the session token. There is no way to invalidate compromised sessions without changing the secret."

### Recommendation

A single highest-impact action. Not a list of 5 things.

❌ Bad: "Improve authentication security."
✅ Good: "Move token to httpOnly cookie and add refresh token rotation. See `src/auth/handler.ts:45`."

### Implementation risk

| Level | When |
|-------|--------|
| Low | Localized change, no dependencies, easy rollback |
| Medium | Interface or contract change, requires coordination |
| High | Core flow change, affects multiple modules, complex rollback |

### Acceptance criteria

Verifiable and concrete. Not "improve X."

```markdown
**Acceptance criteria**:
- [ ] The token is no longer stored in localStorage
- [ ] The token is sent in httpOnly cookie with Secure + SameSite=Strict flags
- [ ] Refresh token endpoint exists with rotation
- [ ] E2E test verifies the token is not accessible via JS
```
