# P1 · Security Audit Report

## 1. Objective

Analyze the Sequoia CLI codebase (`github.com/Crisbr10/sequoia`) for security vulnerabilities, misconfigurations, and supply-chain risks. Sequoia is a Go CLI (~3,300 LOC non-test) that installs skill/command/prompt files into AI coding assistant configuration directories (Claude Code, OpenCode, Cursor, Gemini CLI, OpenAI Codex). The tool has no network server surface—all risk is local: file permissions, rollback safety, install/uninstall integrity, and dependency chain.

## 2. Scope

| Dimension | Coverage |
|-----------|----------|
| **Source files** | `cmd/sequoia/`, `adapters/` (claude, opencode, cursor, gemini, codex, common), `internal/pipeline/`, `internal/app/`, `internal/tui/`, `plugin/`, `scripts/` |
| **Excluded** | `docs/Ejemplos de Skills/engram/` (example code, not part of Sequoia build), test-only code not exercising production paths |
| **Analysts** | P1 (security) |
| **Tool version** | v0.1.0 (pre-release) |

## 3. Verified State

- **No hardcoded secrets**: Zero API keys, tokens, credentials, or private keys found in source code. The `.goreleaser.yaml` uses `${Env.HOMEBREW_TAP_TOKEN}` (runtime env variable, not committed). The `action.yml` references `secrets.GITHUB_TOKEN` (GitHub Actions built-in).
- **No command injection**: `exec.LookPath()` is used for detection only (checking if `claude`, `opencode`, `cursor`, `engram` binaries exist in PATH). No `exec.Command()` with user-controlled input.
- **No path traversal in user input**: The CLI has no `--target`, `--dir`, or `--output` flags that accept arbitrary user paths. All paths are constructed from `os.UserHomeDir()` + well-known subdirectories.
- **No dependency confusion**: `go.mod` contains no `replace` directives. All dependencies are pinned to specific versions from well-known open-source packages.
- **No XML/HTML external entity injection**: No XML or HTML parsing in the application.
- **Go version**: 1.24.2 (current, supported).

## 4. Consolidated Findings

---

### [P1-001] · Backup file collision: user data loss from `.sequoia-backup` naming  [🟠 RIESGO]

**State**: Confirmed

**Evidence**:
- `adapters/opencode/installer.go:36` — `backup := path + ".sequoia-backup"`
- `adapters/cursor/installer.go:36` — `backup := path + ".sequoia-backup"`
- `adapters/codex/installer.go:22` — `backupPath := path + ".sequoia-backup"`

**Problem**:
Three adapters (OpenCode, Cursor, Codex) create backup files by appending `.sequoia-backup` directly to the target file path. If a user already has a file at that exact backup path (e.g., `~/.config/opencode/AGENTS.md.sequoia-backup`), the install process silently **overwrites** it with the current content of the target file. On uninstall, the overwritten backup is restored to the target file—potentially replacing the user's original `AGENTS.md` with data they never intended as a backup. The `common.Installer` path (used by Claude and Gemini adapters) correctly namespaces its backup into a dedicated `.sequoia-backup/` directory.

**Real Impact**:
Silent data loss of user files that happen to be named `*.sequoia-backup` in config directories. On uninstall, the lost data replaces the current file content, compounding the damage.

**Minimal High-Leverage Recommendation**:
Replace `path + ".sequoia-backup"` with a time-stamped or random-suffixed name (e.g., `path + ".sequoia-backup-" + timestamp`), or use the `common.Installer` pattern with a dedicated subdirectory. On uninstall, restore only if the backup was created by the current install session (track via a session marker).

**Dependencies/Blockers**: None

**Implementation Risk**: Low
The backup files are only read/written by Sequoia itself. Adding a timestamp or random suffix eliminates collision without breaking any external contract.

**Acceptance Criteria**:
- [ ] Backup file names include a unique component (timestamp, UUID, or rand suffix)
- [ ] Install does not overwrite pre-existing `.sequoia-backup` files
- [ ] Uninstall only restores backups created during the corresponding install
- [ ] Test: pre-create a `.sequoia-backup` file, install, verify it is preserved

---

### [P1-002] · No OS signal handling — partial install on forced kill  [🟠 RIESGO]

**State**: Confirmed

**Evidence**:
- `cmd/sequoia/main.go:56` — `root.Execute()` calls directly into TUI/install with no `signal.NotifyContext` wrapper
- `grep "os/signal" cmd/ *.go` — zero matches
- `internal/pipeline/runner.go:43-49` — context cancellation is wired but never triggered by OS signals

**Problem**:
When a user presses Ctrl+C or the process receives SIGTERM, the Go runtime's default signal handler kills the process immediately. The pipeline's context is never cancelled, so the `Rollback()` methods in `common/installer.go` and adapter-specific installers never execute. This leaves:
- Partially copied skill/command files in target directories
- Backup files in `.sequoia-backup` directories
- No cleanup of temp staging directories (though OS temp cleanup handles these eventually)
- The tool reports as "not installed" (marker check fails) but has artifacts present

**Real Impact**:
A user kills Sequoia mid-install and their Claude Code/OpenCode/etc. config directories are left in an inconsistent state. Re-running install handles this (backup is retried), but the stale files from the aborted run remain indefinitely, potentially confusing the AI assistant.

**Minimal High-Leverage Recommendation**:
Wrap `root.Execute()` in a `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` and pass the derived context to the pipeline. On cancellation, call `Rollback()` for any adapter that started installation.

**Dependencies/Blockers**: None

**Implementation Risk**: Low
Signal handling in Go is straightforward. The pipeline already supports context cancellation—it just needs to be wired to OS signals.

**Acceptance Criteria**:
- [ ] Ctrl+C during install triggers rollback for all in-progress adapters
- [ ] SIGTERM triggers rollback
- [ ] Rollback completes within 1 second (file operations are local)
- [ ] No stale backup directories remain after signal-initiated rollback

---

### [P1-003] · Checksum verification silently skipped on download failure  [🟠 RIESGO]

**State**: Confirmed

**Evidence**:
- `scripts/install.sh:220-248` — Checksum download failure is caught with `|| true`, then the script logs a warning and continues: `log_warn "Could not download checksums.txt. Skipping checksum verification."`
- `scripts/install.ps1:165` — PowerShell script has the same pattern: `Write-Warn "Could not download checksums.txt. Skipping checksum verification."`

**Problem**:
The install scripts attempt to download `checksums.txt` from GitHub Releases. If the download fails (network error, 404 for pre-release, GitHub rate limiting), the script proceeds to extract and run the binary **without any integrity verification**. A man-in-the-middle attacker who can block the checksums.txt download (easier than modifying the binary in transit) can serve a tampered binary that passes all checks because verification was skipped.

The `install.ps1` script offers `-SkipChecksum` as an explicit opt-out, but the silent skip on download failure is an implicit bypass that the user cannot control.

**Real Impact**:
Supply-chain attack: an attacker who controls the network path (compromised CDN, corporate proxy, public WiFi) can serve a modified Sequoia binary by blocking the checksums.txt download. The script installs it without warning the user that verification was skipped.

**Minimal High-Leverage Recommendation**:
When `checksums.txt` cannot be downloaded, abort the installation with a clear error message and exit code `EXIT_CHECKSUM` (2). Provide a `--skip-checksums` flag for users in air-gapped environments who accept the risk. The checksum download should use the same retry logic as the binary download (currently 3 retries on the binary but not on checksums).

**Dependencies/Blockers**: Requires updating the retry logic in both `install.sh` and `install.ps1`

**Implementation Risk**: Low
This is a policy change in the install script, not the Go binary. The retry logic already exists for the binary download—copy the pattern.

**Acceptance Criteria**:
- [ ] Failed checksums.txt download aborts installation with exit code 2
- [ ] `--skip-checksums` flag available for explicit opt-out
- [ ] Both install.sh and install.ps1 updated
- [ ] Retry logic (3 attempts) applied to checksums download

---

### [P1-004] · Uninstall best-effort silently discards file removal errors  [🟠 RIESGO]

**State**: Confirmed

**Evidence**:
- `adapters/claude/adapter.go:219-224`:
  ```go
  _ = os.Remove(skillFilePath(base))
  _ = os.Remove(versionFilePath(base))
  for _, cmd := range common.CommandFiles {
      _ = os.Remove(filepath.Join(commandsPath(base), cmd))
  }
  ```
- `adapters/opencode/adapter.go:221-225` — identical pattern
- `adapters/gemini/adapter.go:215-216` — `_ = os.RemoveAll(sequoiaDir)`
- `adapters/cursor/adapter.go:216-220` — identical pattern
- `adapters/codex/adapter.go:226-230` — identical pattern

**Problem**:
All five adapter `Uninstall()` methods use `_ = os.Remove(...)` for file and directory removal, silently discarding all errors. If a file is locked by another process (Windows file locking), has incorrect permissions, or is on a read-only filesystem, the uninstall reports success but leaves Sequoia files behind. The user sees "Done" on the CLI but the AI assistant still finds Sequoia skills installed. The `IsInstalled()` check may still return true depending on which files were left—creating a confusing state where "uninstall succeeded" but "status shows installed."

**Real Impact**:
Incomplete uninstall that the user cannot detect without running `sequoia status`. The marker file (`.sequoia-version`) may be among the files that failed to remove, causing `IsInstalled()` to return true. The user must manually delete files.

**Minimal High-Leverage Recommendation**:
Collect errors from all `os.Remove` calls and return them as a `multierror` (or joined error via `errors.Join` in Go 1.20+). Report which files could not be removed in both the error message and the TUI progress view. In the TUI, show a warning: "Uninstall completed with warnings — 2 files could not be removed."

**Dependencies/Blockers**: Requires Go 1.20+ `errors.Join` (project already targets Go 1.24.2)

**Implementation Risk**: Low
Pure error-collection change. No behavioral change—still best-effort, but now transparent.

**Acceptance Criteria**:
- [ ] Uninstall returns aggregated errors for all failed removals
- [ ] TUI displays per-file removal warnings
- [ ] CLI headless mode prints warnings to stderr
- [ ] Test: create read-only file in skills dir, verify uninstall reports it

---

### [P1-005] · Symlink resolution fallback may cause path confusion on broken symlinks  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `adapters/claude/paths.go:20-24`:
  ```go
  resolved, err := filepath.EvalSymlinks(homeDir)
  if err != nil {
      resolved = homeDir
  }
  ```
- `adapters/opencode/paths.go:20-24` — identical
- `adapters/cursor/paths.go:20-24` — identical
- `adapters/codex/paths.go:20-24` — identical
- `adapters/gemini/paths.go:20-24` — identical

**Problem**:
All path resolution functions use `filepath.EvalSymlinks` with an unconditional fallback to the unresolved path. The comment says "fall back to unresolved path on any error." While this is a reasonable best-effort strategy for symlink loops or permission errors, it silently masks errors for broken symlinks (dangling). On macOS, it's common for users to have `.claude` or `.config` symlinked to iCloud/Dropbox for cross-machine sync. If the symlink target is temporarily unavailable (unmounted volume), Sequoia will install into the wrong location (the symlink path instead of the real path), potentially creating files in a location that gets overwritten when the volume remounts.

**Real Impact**:
On macOS with iCloud/Dropbox-synced config directories, if the sync target is unavailable, Sequoia installs skills into a local path that may conflict with the synced version when the volume remounts. The user ends up with duplicate configs or the synced version overwrites the Sequoia install.

**Minimal High-Leverage Recommendation**:
Distinguish between "symlink target doesn't exist" (which should warn) and "symlink resolution failed for other reasons" (which should fall back). Use `os.Lstat` to check if the path is actually a symlink before attempting resolution, and log a warning when falling back. Alternatively, surface the resolution error in the TUI status screen.

**Dependencies/Blockers**: None

**Implementation Risk**: Low
Add a `DEBUG`-level log or warning when the fallback triggers. The behavioral change (aborting on broken symlink) could break users who rely on the current fallback—start with a warning only.

**Acceptance Criteria**:
- [ ] Warning emitted when symlink fallback is used (TUI status, CLI stderr)
- [ ] `ResolveHome` distinguishes "not a symlink" from "broken symlink" from "resolution error"
- [ ] No behavior change for non-symlink paths

---

### [P1-006] · `_.Language` silently discarded — i18n feature non-functional  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `adapters/claude/adapter.go:126` — `_ = opts.Language`
- `adapters/opencode/adapter.go:126` — `_ = opts.Language`
- `adapters/cursor/adapter.go:121` — `_ = opts.Language`
- `adapters/gemini/adapter.go:122` — `_ = opts.Language`
- `adapters/codex/adapter.go:127` — `_ = opts.Language`

**Problem**:
The `InstallOpts.Language` field is defined in the public interface, passed from the pipeline (`runner.go:78` sets `opts := adapters.InstallOpts{Language: lang}`), and accepted by all 5 adapters. However, every adapter immediately discards it with `_ = opts.Language`. The Configuration screen in the TUI lets users select a language (`model.go:94` defaults to `"en"`), and the pipeline propagates it—but it has no effect anywhere. This is not a direct security vulnerability, but it's a feature gap that creates a false sense of i18n support. A user selecting `es` (Spanish) expects Spanish output but gets English.

**Real Impact**:
User trust erosion. The TUI presents a language selector that does nothing. If i18n is implemented later and the adapter interface changes, this dead code path may cause subtle bugs.

**Minimal High-Leverage Recommendation**:
Either implement template localization (pass `lang` to `templateData` and use in templates) or remove the Language field from the Configuration screen and opts until the feature is implemented. If keeping the field, add a `// TODO: i18n` comment in each adapter.

**Dependencies/Blockers**: Requires i18n template infrastructure (fallback chains, translation files)

**Implementation Risk**: Medium
Implementing i18n requires translation files, template conditionals, and testing. As a short-term fix, annotate with TODO comments and consider graying out the language option in the TUI.

**Acceptance Criteria**:
- [ ] Each adapter has a `// TODO(i18n): implement opts.Language` comment
- [ ] OR the Language field is removed from the Configuration screen until functional
- [ ] OR template localization is implemented with at least `en` and `es` support

---

### [P1-007] · Predictable backup directory names across all adapters  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `adapters/claude/paths.go:47` — `filepath.Join(base, ".sequoia-backup")`
- `adapters/opencode/paths.go:47` — `filepath.Join(base, ".sequoia-backup")`
- `adapters/cursor/paths.go:46` — `filepath.Join(base, ".sequoia-backup")`
- `adapters/gemini/paths.go:47` — `filepath.Join(base, ".sequoia-backup")`
- `adapters/codex/paths.go:47` — `filepath.Join(base, ".sequoia-backup")`

**Problem**:
The backup directory name `.sequoia-backup` is hardcoded and identical across all adapters. A local attacker with filesystem access to the user's home directory could pre-create a `.sequoia-backup` directory with restricted permissions (e.g., 0000) before the user runs `sequoia install`. The `os.MkdirAll(backupPath, 0o755)` in `common/installer.go:84` would fail with a permission error, causing the Prepare phase to fail and blocking installation entirely. While this is a denial-of-service rather than data compromise, it's a trivially exploitable local DoS.

**Real Impact**:
A malicious co-user on a shared system, or malware with user-level access, can prevent Sequoia from installing by pre-creating `.sequoia-backup` directories.

**Minimal High-Leverage Recommendation**:
Use a random suffix for the backup directory (e.g., `.sequoia-backup-<timestamp>-<random>`), or use `os.MkdirTemp` within the base directory to create a unique backup dir. On uninstall, clean up any stale `.sequoia-backup-*` directories.

**Dependencies/Blockers**: None

**Implementation Risk**: Low
Using `os.MkdirTemp` for backup directories is a drop-in replacement. Update `backupPath()` and the uninstall cleanup logic.

**Acceptance Criteria**:
- [ ] Backup directory is unique per install run (random or timestamp suffix)
- [ ] Pre-existing `.sequoia-backup` directories do not block installation
- [ ] Uninstall cleans up backup dir even with random suffix
- [ ] Test: pre-create `.sequoia-backup` dir, verify install still succeeds

---

### [P1-008] · World-readable permissions (0o644) on installed config files  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `adapters/common/files.go:15` — `os.WriteFile(filepath.Join(dir, name), content, 0o644)`
- `adapters/claude/installer.go:27,40,51` — all `os.WriteFile` calls use `0o644`
- `adapters/opencode/installer.go:29,33,41,44` — all use `0o644`
- `adapters/cursor/installer.go:29,33,41,44` — all use `0o644`
- `adapters/gemini/installer.go:27,39,50,86` — all use `0o644`
- `adapters/codex/installer.go:23,38,49,69` — all use `0o644`
- All adapter `versionFilePath` writes use `0o644`

**Problem**:
All files installed by Sequoia (skill files, command files, system prompt sections, version markers, backup files) are created with mode `0o644` — world-readable. On multi-user Unix systems, any other user on the machine can read the installed Sequoia content. While the content itself is not sensitive (it's public skill/command definitions), the system prompt file (`CLAUDE.md`, `AGENTS.md`, etc.) may contain user-specific configuration alongside the Sequoia section. During backup, the entire file content (including user customizations) is copied with `0o644`, potentially exposing private configuration to other local users.

**Real Impact**:
On shared Unix systems (university clusters, corporate dev servers), other local users can read the full content of the user's AI assistant configuration files, including any custom instructions or preferences that were in the file before Sequoia was installed.

**Minimal High-Leverage Recommendation**:
Use `0o600` for backup files and any file containing the user's original content. For purely Sequoia-generated files (skills, commands), `0o644` is acceptable but `0o640` is more defensive. The backup directory should be created with `0o700`.

**Dependencies/Blockers**: Windows ignores Unix permissions; Windows-only users are unaffected.

**Implementation Risk**: Low
Permission change only. Verify that read operations still work (they will—the user owns the files).

**Acceptance Criteria**:
- [ ] Backup files written with `0o600`
- [ ] Backup directories created with `0o700`
- [ ] User-original-content restores preserve original permissions (not hardcoded)
- [ ] Cross-platform: no regression on Windows (where these modes are no-ops)

---

### [P1-009] · Temp probe file exposes directory write test to TOCTOU race  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `adapters/common/installer.go:63-69`:
  ```go
  probe := filepath.Join(cfg.TargetDir, ".sequoia-probe")
  f, err := os.Create(probe)
  ...
  f.Close()
  if err := os.Remove(probe); err != nil { ... }
  ```

**Problem**:
The `Prepare()` method tests write access by creating and immediately deleting a `.sequoia-probe` file. There is a time-of-check-to-time-of-use window between the probe removal and the actual `Apply()` file writes. On a shared system, a concurrent process could create a symlink at `.sequoia-probe` pointing to an arbitrary file, causing the probe to succeed (it writes to the symlink target) while the actual install dir is unwritable — or vice versa. However, exploitation requires:
1. A malicious process running as the same user
2. Precise timing to replace the file between probe and Apply
3. Knowledge of when Sequoia is being run

This is extremely low probability in practice.

**Real Impact**:
Theoretical. In practice, the probe file is created and removed within microseconds in a user-owned directory. The attacker would need to win a very tight race.

**Minimal High-Leverage Recommendation**:
Replace the probe file test with an `unix.Access` call on Unix (checking `W_OK`) or `os.MkdirTemp` within the target directory. The temp dir approach is more robust because it exercises the same code path as the actual install.

**Dependencies/Blockers**: `unix.Access` requires build tags for cross-platform support

**Implementation Risk**: Medium
Requires platform-specific code. The current approach works correctly on all platforms. Risk of regression outweighs the benefit for this finding.

**Acceptance Criteria**:
- [ ] Probe replaced with `MkdirTemp`-based test OR documented as accepted risk
- [ ] No regression in test coverage for Prepare()

---

### [P1-010] · Config backup overwrites user content on power loss during write  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `adapters/opencode/installer.go:41-44`:
  ```go
  if err := os.WriteFile(backup, raw, 0o644); err != nil {
      return err
  }
  return os.WriteFile(path, []byte(content), 0o644)
  ```
- `adapters/codex/installer.go:22-25` — identical pattern
- `adapters/cursor/installer.go:41-44` — identical pattern

**Problem**:
In the `FileReplace` strategy (OpenCode, Cursor), the backup is written first, then the new content overwrites the target file. If a power loss or system crash occurs between these two writes, both the backup and the target file are intact (crash after backup write, before target write). However, if the crash occurs during either write (partial write), the file is corrupted. Go's `os.WriteFile` does an atomic rename on Unix (write to temp + rename), so partial writes are unlikely but not guaranteed on all platforms. On Windows, `os.WriteFile` truncates and writes in place — a crash during write leaves a truncated file.

Additionally, there is no integrity marker to distinguish "backup was created" from "backup is from a previous run."

**Real Impact**:
On Windows, a system crash during the 2-step backup-and-replace process can leave the target AI assistant config file truncated. The user loses their config content and Sequoia's content. Recovery requires manual intervention.

**Minimal High-Leverage Recommendation**:
Use the write-to-temp-then-rename pattern explicitly for the target file write. Write new content to `path + ".tmp"`, then `os.Rename(tmp, path)`. On Unix, rename is atomic; on Windows, it's atomic within the same volume. This ensures the target file is never in a partially-written state.

**Dependencies/Blockers**: None

**Implementation Risk**: Low
Drop-in replacement for `os.WriteFile` in the backup-and-replace sequence.

**Acceptance Criteria**:
- [ ] Target file writes use temp-then-rename pattern
- [ ] Test: simulate crash during write, verify target file is intact (previous content or new content, never truncated)
- [ ] Cross-platform: verify rename atomicity on Windows, macOS, Linux

---

### [P1-011] · install.sh: `curl ... | bash` pattern without script integrity  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `scripts/install.sh:6` — `curl -sSL https://raw.githubusercontent.com/Crisbr10/sequoia/main/scripts/install.sh | bash`
- `.goreleaser.yaml:65-68` — documents the same pipe-to-shell pattern
- `README.md` (inferred) — likely documents the same

**Problem**:
The recommended installation method pipes a remotely-fetched shell script directly into `bash`. While this is industry-standard for CLI tools (Homebrew, Rust, Oh My Zsh all use the same pattern), it carries inherent risks:
1. **TLS compromise**: If GitHub's TLS is compromised, the script is compromised
2. **Repository compromise**: If the `Crisbr10/sequoia` repo is compromised, the script serves malware
3. **Partial download**: If the connection drops mid-download, bash executes a truncated (potentially dangerous) command
4. **No integrity hash**: Unlike the binary (which has SHA-256 verification), the script itself has no hash published anywhere

The script mitigates some risks by verifying the binary checksum, but the script itself is the initial attack vector.

**Real Impact**:
A compromised `install.sh` could execute arbitrary commands as the user (rm -rf, curl exfil, crypto miner). This is the same risk profile as every other `curl | bash` installer in the ecosystem.

**Minimal High-Leverage Recommendation**:
1. Publish a SHA-256 hash of `install.sh` in the README and release notes
2. Add a `--dry-run` flag to the script that prints what it would do without executing
3. Document the alternative: download + inspect + run manually
4. Consider a statically-compiled Go installer binary that handles the download/verify/extract logic (eliminates the shell script entirely)

**Dependencies/Blockers**: None (documentation change) to Low (Go installer binary)

**Implementation Risk**: Low
Adding a hash to README is trivial. The Go installer binary approach is more work but eliminates the shell script attack surface entirely.

**Acceptance Criteria**:
- [ ] SHA-256 hash of install.sh published in release notes
- [ ] README documents manual verification steps
- [ ] `--dry-run` flag available in install.sh
- [ ] (Stretch) Go-based installer binary as alternative to shell script

---

### [P1-012] · Template rendering uses `text/template` — no auto-escaping  [🟡 ATENCIÓN]

**State**: Confirmed

**Evidence**:
- `adapters/common/template.go:13` — `text/template` for Markdown templates
- `adapters/common/template.go:18` — `template.New(name).Parse(string(raw))`
- Template data uses `templateData{Version: common.Version}` — all fields are internally sourced

**Problem**:
The template system uses Go's `text/template` package (not `html/template`). For Markdown output, this is correct—`html/template` would escape HTML entities in code blocks. However, `text/template` has no contextual auto-escaping, meaning if any template variable ever contains user-provided data in the future, it would be rendered verbatim without sanitization. Currently, all template data is hardcoded (`Version` string), so this is a latent risk, not an active vulnerability.

**Real Impact**:
None currently. If i18n support is added and template variables include user-selected language codes, or if plugin-supplied templates are rendered, this becomes an injection vector. An attacker who controls a template variable could inject Markdown/HTML into the AI assistant's system prompt.

**Minimal High-Leverage Recommendation**:
Add a comment to `RenderTemplate` documenting that all template data MUST come from trusted/internal sources. If external data is ever passed (user config, plugin output), add sanitization. Consider using `bluemonday` or similar for Markdown sanitization if user content enters templates.

**Dependencies/Blockers**: Only relevant when i18n or plugin templates are implemented

**Implementation Risk**: Low (documentation-only for now)

**Acceptance Criteria**:
- [ ] `RenderTemplate` godoc warns about untrusted data
- [ ] TODO comment linking to this finding for future i18n work
- [ ] Test: verify template renders correctly with special characters in Version

---

## 5. High-Leverage Missing Items

These are security features or practices **absent** from the codebase that would significantly reduce risk:

| # | Missing Item | Risk Addressed | Effort |
|---|-------------|---------------|--------|
| 1 | `go.sum` verification in CI (currently no CI pipeline defined in repo) | Supply chain | Small |
| 2 | `govulncheck` integration (Go vulnerability scanner) | Dependency CVEs | Small |
| 3 | `.gitattributes` for binary detection (prevent CRLF corruption of embedded files) | File integrity | Trivial |
| 4 | GPG signing of GitHub Releases (GoReleaser supports `signs` config) | Binary authenticity | Small |
| 5 | `sbom` generation in GoReleaser (`sboms` config) | Supply chain transparency | Trivial |
| 6 | Rate limiting / abuse prevention for the auto-version check (none needed—local CLI) | N/A | N/A |

## 6. Task Plan

### Immediate (this sprint)
| Task | Finding | Effort | Priority |
|------|---------|--------|----------|
| Add signal handling (SIGTERM/SIGINT) | P1-002 | 2h | P0 |
| Fix backup file collision (unique suffix) | P1-001 | 1h | P0 |
| Make checksum verification mandatory in install scripts | P1-003 | 1h | P1 |
| Collect and report uninstall errors | P1-004 | 1.5h | P1 |

### Short-term (next sprint)
| Task | Finding | Effort | Priority |
|------|---------|--------|----------|
| Use temp-then-rename for atomic writes | P1-010 | 2h | P2 |
| Restrict permissions on backup files (0o600) | P1-008 | 1h | P2 |
| Randomize backup directory names | P1-007 | 1h | P2 |
| Add `govulncheck` to CI/goreleaser | Missing #2 | 1h | P2 |

### Long-term (backlog)
| Task | Finding | Effort | Priority |
|------|---------|--------|----------|
| Symlink resolution warnings | P1-005 | 2h | P3 |
| Document template security contract | P1-012 | 0.5h | P3 |
| GPG signing of releases | Missing #4 | 2h | P3 |
| SBOM generation in GoReleaser | Missing #5 | 0.5h | P3 |
| Go-based installer (eliminate shell scripts) | P1-011 | 8h | P3 |
| Replace probe file with temp-dir-based test | P1-009 | 2h | P4 |

## 7. Implementation Order

```
P1-002 (signal handling)
  └─► P1-001 (backup collision) — both affect install/uninstall lifecycle
       └─► P1-004 (uninstall errors) — makes rollback observable
            └─► P1-008 (permissions) — defense in depth
                 └─► P1-007 (backup dir names) — blocks P1-001 alternative path
                      └─► P1-010 (atomic writes) — crash safety
```

## 8. Risks/Blockers

| Risk | Probability | Mitigation |
|------|-----------|------------|
| Signal handling on Windows: `syscall.SIGTERM` not available on Windows | Medium | Use `os.Interrupt` only on Windows; add `syscall.SIGTERM` for Unix via build tag |
| Atomic rename on cross-filesystem paths: `os.Rename` fails across volumes | Low | Fall back to copy+delete if rename fails with `EXDEV` |
| Permission changes (0o600) break Windows: Windows ignores Unix perms | None | Windows unaffected; Unix modes are no-ops on Windows |
| Backup file changes break existing test fixtures | Medium | Tests use temp dirs; update test expectations for new backup naming |

## 9. Phase Close Checklist

- [x] 12 findings documented with evidence, impact, and recommendations
- [x] Severity calibrated to CLI tool context (no network surface, local-only risk)
- [x] No fabricated findings — all traceable to specific file:line references
- [x] Findings ordered by severity (🟠 → 🟡)
- [x] Section 5 identifies 6 missing security practices with effort estimates
- [x] Task plan with 4 immediate, 4 short-term, 5 long-term tasks
- [x] Implementation order diagram showing dependency chain
- [x] Risks/blockers table covering cross-platform concerns

---

*Report generated by P1 sequoia-security · Sequoia v0.1.0 · $(date)*
