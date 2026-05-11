# P1 Security — sequoia-ai v0.1.0

**Score**: 85/100 (B+) | **Findings**: 7 (0C, 0H, 2M, 4L, 1I)

---

## Clean Checks ✅

| Category | Status | Notes |
|----------|--------|-------|
| Secrets | ✅ Clean | No hardcoded tokens, keys, or passwords found in non-test code |
| Authentication | ✅ N/A | CLI tool, no authentication mechanism |
| Injection | ✅ Clean | No command injection vectors; tool IDs validated against Registry |
| File Permissions | ✅ Correct | All dirs 755, files 644; no world-writable paths |
| Path Traversal | ✅ Clean | All paths derived from `os.UserHomeDir()` with hardcoded subdirectories |
| Input Validation | ✅ Clean | `--tool` flag validated via `DefaultRegistry.Get()` |
| Token Storage | ✅ N/A | No tokens persisted by the CLI |

---

## Findings

### P1-001 [MEDIUM] — Supply Chain: Pipe-to-shell install pattern
**Evidence**: `scripts/install.sh:6`, `scripts/install.ps1:19`

The documented one-liner installation pipes a remote script directly into a shell interpreter. If the GitHub repository is compromised or MITM succeeds against raw.githubusercontent.com, arbitrary code executes immediately with full user privileges.

**Impact**: Compromised repository or TLS MITM allows attackers to execute arbitrary code on user machines.

**References**: CWE-494, OWASP A06:2021

---

### P1-002 [MEDIUM] — Supply Chain: GitHub Action skips checksum verification
**Evidence**: `action.yml:92`

When the GitHub Action runs with a specific `sequoia-version`, it downloads the binary via `curl` without SHA-256 checksum verification. The install scripts DO verify checksums, but the Action skips this entirely for versioned downloads.

**Impact**: Compromised release or MITM could inject malicious binary into CI/CD pipelines with access to GITHUB_TOKEN.

**References**: CWE-494, SLSA Level 2

---

### P1-003 [LOW] — Supply Chain: Checksum verification silently skipped
**Evidence**: `scripts/install.sh:248`, `scripts/install.ps1:165`

Both installers silently skip SHA-256 verification when checksums.txt is unreachable. No user prompt or abort — the binary is installed regardless.

**Impact**: Users on networks where the checksum URL is blocked receive unverified binaries without meaningful warning.

**References**: CWE-494, CWE-345

---

### P1-004 [LOW] — Error messages expose internal filesystem paths
**Evidence**: `adapters/common/installer.go:193`, `cmd/sequoia/main.go:57`

Error messages include absolute paths like `/home/user/.claude/skills/sequoia/`. These propagate through the TUI error screen and stderr, unnecessarily exposing internal directory structure in logs.

**Impact**: Information disclosure via error logs — aids attackers understanding filesystem layout.

**References**: CWE-209, OWASP ASVS V7.4

---

### P1-005 [LOW] — Non-atomic file writes risk data corruption
**Evidence**: `adapters/common/installer.go:197`

Files are written directly to target paths with `os.Create`. A crash mid-write leaves partially-written corrupt files. The backup/rollback mitigates but the write itself is not atomic (no write-to-temp-then-rename pattern).

**Impact**: System crash during install could corrupt CLAUDE.md, AGENTS.md, GEMINI.md.

**References**: CWE-362

---

### P1-006 [LOW] — os.RemoveAll vs targeted removal in Gemini/Codex uninstall
**Evidence**: `adapters/gemini/adapter.go:212`, `adapters/codex/adapter.go:234`

Unlike Claude/OpenCode/Cursor which remove individual files, Gemini and Codex use `os.RemoveAll` on the entire `sequoia/` directory. Larger blast radius if path resolution ever changes.

**Impact**: Low likelihood — paths are hardcoded from UserHomeDir. TOCTOU symlink race potential.

**References**: CWE-367, CWE-59

---

### P1-007 [INFO] — install.ps1: LASTEXITCODE handling masks installation errors
**Evidence**: `scripts/install.ps1:238-243`

Redirection `2>&1` captures stderr, and the catch block logs a generic "completed with warnings" even if `sequoia install --no-tui` fails. The script reports "installed successfully" regardless of actual outcome.

**Impact**: Users may believe Sequoia is installed when it's not. Low severity because users can run `sequoia status` to verify.
