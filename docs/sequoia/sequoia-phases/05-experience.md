# P5 Experience — sequoia-ai v0.1.0

**Score**: N/A — Skipped

---

## Skip Reason

Sequoia is a CLI tool and plugin framework. It has a single interactive TUI screen (the installer with multi-select) that is not a user-facing product interface — it's a setup utility.

The TUI was evaluated structurally by P3 Architecture (internal/tui/ is cleanly separated) and P2 Performance (Lipgloss styles cached as a recommendation). A full UX/accessibility audit for a single installation screen would produce negligible value.

**Recommendation**: If the TUI installer evolves to have multiple screens with complex user flows, re-enable P5 Experience for a future audit.
