---
description: "Develop a task using SDD with configured strategies. Auto-detects simple tasks for fast-forward mode."
argument-hint: "<task-id> [--ff | --full]"
allowed-tools: Read, Glob, Grep, Bash, Edit, Write, mem_search, mem_get_observation, mem_save
---

# /sequoia-dev

> **SDD task executor** — reads a task, computes complexity, and launches the appropriate SDD flow (fast-forward or full cycle). Configured via `~/.config/opencode/.sequoia-dev.yaml`.

Develops a Sequoia task using the SDD (Spec-Driven Development) workflow. It auto-detects task complexity and chooses between `sdd-ff` (fast-forward for simple tasks) and the full SDD cycle (`sdd-propose → sdd-spec → sdd-design → sdd-tasks → sdd-apply → sdd-verify → sdd-archive`). All strategies (TDD mode, delivery, chain) are read from your config.

## Purpose

`/sequoia-dev` bridges the gap between Sequoia's audit output and SDD's structured development process. When you run `/sequoia audit`, tasks are generated — this command takes a task ID and launches the right development workflow to implement it.

It reads the task metadata (risk, effort, files involved, dependencies) and applies a complexity heuristic to decide whether the task is simple enough for fast-forward (`sdd-ff`) or needs the full SDD cycle. You can override this with `--ff` or `--full`.

## Precondition

SDD must be initialized in this project. **Validate blocking**: search Engram for `sdd-init/{project}`. If no result is found, error:

```
SDD has not been initialized for this project. Run `/sdd-init` first to set up SDD workspace, testing capabilities, and registry.
```

Do NOT proceed without SDD init.

## Config Loading

**CRITICAL — Path resolution**: The config file lives at `~/.config/opencode/.sequoia-dev.yaml`, which is OUTSIDE the project workspace. You MUST use an absolute path to read it. Do NOT use workspace-scoped tools like `glob` — they will only search within the project directory and silently miss the config file.

1. **Resolve the absolute path**: On Linux/macOS, expand `~` to `$HOME`. On Windows, resolve to `$env:USERPROFILE\.config\opencode\.sequoia-dev.yaml`.
2. **Read the file directly** using the absolute path (e.g., `read` with the full resolved path).
3. If the file does not exist → use defaults silently (first run, no config yet).
4. If the file exists → parse it and merge with defaults. User config values override defaults, missing keys fall back to defaults, unknown keys are ignored silently.

| Key | Default | Description |
|-----|---------|-------------|
| `sdd.execution_mode` | `auto` | Execution mode: `auto` (run all phases back-to-back) or `interactive` (pause after each phase) |
| `sdd.artifact_store` | `engram` | Artifact store: `engram` (persistent memory), `openspec` (repo files), or `both` |
| `sdd.tdd` | `strict` | TDD mode: `strict` (red-green-refactor) or `standard` (tests optional) |
| `sdd.delivery` | `ask-on-risk` | Delivery strategy for PR review workflow |
| `sdd.chain` | `stacked-to-main` | Chain strategy when PRs are split |
| `sdd.review_budget_lines` | `400` | Max changed lines before stopping for review approval |
| `sdd.auto_ff` | `true` | Auto fast-forward for simple tasks |
| `sdd.complexity.ff_max_score` | `2` | Max score for fast-forward eligibility |
| `paths.tasks_dir` | `docs/sequoia/tasks/` | Directory where task files are stored |

**Config error handling**:
- **Config file missing**: Use defaults silently (first run, no config yet).
- **Config syntax error** (invalid YAML): Warn the user with the specific parse error, then fall back to defaults.
- **Unknown keys**: Ignore silently — forward compatibility.
- **Invalid values** (e.g., `tdd: banana`): Warn the user and fall back to the default for that key only.

## SDD Session Preflight Bypass (CRITICAL)

**You are running under `/sequoia-dev`, which provides ALL SDD Session Preflight answers via `.sequoia-dev.yaml`. Do NOT ask the user the preflight questions.** Construct the preflight answer block silently from the config and proceed.

The SDD orchestrator normally requires a 4-group preflight. `/sequoia-dev` satisfies every group from config:

### Preflight Answer Construction

Map config keys to preflight canonical values as follows. If a config key is missing, use the default shown in the table above.

| Preflight Group | Config Key | Mapping |
|----------------|------------|---------|
| **A. Pace (execution_mode)** | `sdd.execution_mode` | `auto` → `A2 Automatic`, `interactive` → `A1 Interactive` |
| **B. Artifacts (artifact_store)** | `sdd.artifact_store` | `engram` → `B2 Engram`, `openspec` → `B1 OpenSpec`, `both` → `B3 Both` |
| **C. PRs (delivery)** | `sdd.delivery` | `ask-on-risk` → `C1 Ask me`, `auto-chain` → `C3 Chained`, `single-pr` → `C2 Single PR`, `exception-ok` → `C2 Single PR` |
| **D. Review (review_budget_lines)** | `sdd.review_budget_lines` | `400` → `D1 400 lines`, `800` → `D2 800 lines`, other number → `D3 Other` (use the number) |

### Preflight Block Format

After loading config, construct this preflight block internally. Use it as the canonical session preflight when invoking any SDD phase (propose, spec, design, tasks, apply, verify, archive). Include it in every sub-agent prompt that expects SDD preflight values:

```
SDD Session Preflight (from .sequoia-dev.yaml — do NOT ask user):
  execution_mode: {canonical value, e.g. auto}
  artifact_store: {canonical value, e.g. engram}
  delivery_strategy: {canonical value, e.g. ask-always}
  chain_strategy: {sdd.chain value, e.g. stacked-to-main}
  review_budget_lines: {number, e.g. 400}
```

**HARD RULE**: Since all preflight answers are deterministically derived from `.sequoia-dev.yaml`, the preflight is already satisfied. Never interrupt the flow to ask these questions. If a sub-agent or SDD phase triggers the preflight prompt, respond with the preflight block above and continue.

### Canonical Value Mapping (full reference)

| Config value | Preflight canonical |
|-------------|-------------------|
| `execution_mode: auto` | `auto` |
| `execution_mode: interactive` | `interactive` |
| `artifact_store: engram` | `engram` |
| `artifact_store: openspec` | `openspec` |
| `artifact_store: both` | `both` |
| `delivery: ask-on-risk` | `ask-always` |
| `delivery: auto-chain` | `force-chained` |
| `delivery: single-pr` | `single-pr-default` |
| `delivery: exception-ok` | `single-pr-default` (with size exception) |
| `chain: stacked-to-main` | `stacked-to-main` |
| `chain: feature-branch-chain` | `feature-branch-chain` |

## Task Location

1. Parse `<task-id>` from user input. Task IDs follow the pattern `{PHASE}-{NNN}` (e.g., `P1-003`, `M2-001`).
2. Map the phase prefix to the area file:
   - `P1` → `security.md`
   - `P2` → `performance.md`
   - `P3` → `architecture.md`
   - `P4` → `quality.md`
   - `P5` → `experience.md`
   - `P6` → `operations.md`
   - `M1`, `M2` → `index.md`
3. Search `{paths.tasks_dir}{area}.md` for the task heading matching `### {TASK-ID} ·`.
4. Parse task metadata from the task block:

| Metadata | Source | Values |
|----------|--------|--------|
| Implementation risk | `**Implementation risk**: Low / Medium / High` | Low=0, Medium=1, High=3 |
| Effort | `**Effort**: small / medium / large` | small=0, medium=1, large=3 |
| Files involved | Count `- \`path/to/file.ext\`` bullet items in **Files involved** section | count |
| Dependencies | Check if **Dependencies** section has content beyond "None" | none=0, has_deps=1 |

5. If metadata is missing from the task, use safe defaults: `risk=Medium (1)`, `effort=medium (1)`, `files=3`, `deps=none (0)`.

## Complexity Computation

Apply the following heuristic to determine the implementation path:

```
score = implementation_risk + effort + files_score + dependencies

Where:
  implementation_risk: Low=0, Medium=1, High=3
  effort: small=0, medium=1, large=3
  files_score: 1-2 files=0, 3-4 files=1, 5+ files=2
  dependencies: none=0, has_deps=1
```

**Decision**:
- **score ≤ `sdd.complexity.ff_max_score` AND `sdd.auto_ff = true`** → Fast-forward mode (`sdd-ff`)
- **score > `sdd.complexity.ff_max_score` OR `sdd.auto_ff = false`** → Full SDD cycle

### Complexity Heuristic Table

| Factor | Value | Score |
|--------|-------|-------|
| Risk | Low | 0 |
| Risk | Medium | 1 |
| Risk | High | 3 |
| Effort | small | 0 |
| Effort | medium | 1 |
| Effort | large | 3 |
| Files | 1-2 | 0 |
| Files | 3-4 | 1 |
| Files | 5+ | 2 |
| Dependencies | none | 0 |
| Dependencies | has deps | 1 |

- **Score 0-2**: Fast-forward (`sdd-ff`)
- **Score 3+**: Full SDD cycle

## SDD Launch

Based on the complexity decision and configured strategies, launch the appropriate flow:

### Fast-Forward (`sdd-ff`)
When the score is ≤ `ff_max_score` or `--ff` is passed:
1. Read the spec scenarios and design decisions (if they exist in Engram)
2. Read the task's acceptance criteria and verification steps
3. Implement the task directly (apply phase)
4. Verify the implementation (verify phase)

Pass `execution_mode={sdd.execution_mode}`, `artifact_store={sdd.artifact_store}`, `tdd={sdd.tdd}`, `delivery={sdd.delivery}`, `chain={sdd.chain}`, `review_budget_lines={sdd.review_budget_lines}` to all subagents.

### Full SDD Cycle
When the score is > `ff_max_score` or `--full` is passed:
1. `sdd-propose` → Define scope and approach
2. `sdd-spec` → Write requirements and scenarios
3. `sdd-design` → Create technical design
4. `sdd-tasks` → Break into implementation tasks
5. `sdd-apply` → Implement each task
6. `sdd-verify` → Verify against specs
7. `sdd-archive` → Sync delta specs

Pass `execution_mode={sdd.execution_mode}`, `artifact_store={sdd.artifact_store}`, `tdd={sdd.tdd}`, `delivery={sdd.delivery}`, `chain={sdd.chain}`, `review_budget_lines={sdd.review_budget_lines}` to all subagents.

## Post-Implementation: Task Marking & Commit (MANDATORY)

After ALL implementation phases complete (apply + verify + archive), you MUST perform these two steps in order. This ensures the task documents stay synchronized and every completed task is committed with a clean, verified state.

### Step 1: Mark Task as Completed in Task File

After `sdd-verify` passes AND `sdd-archive` completes successfully:

1. **Locate the task file**: `{paths.tasks_dir}{area}.md` (same file where the task was originally found).
2. **Read the task file** to find the exact heading and content block for the task ID.
3. **Apply three edits** to the task file:

   a. **Update the heading**: Change `### {TASK-ID}` → `### ✅ {TASK-ID}`.  
      If the heading already has a prefix (e.g., `### ✅ P1-003`), skip — the task was already marked.

   b. **Mark acceptance criteria**: Change every `- [ ]` in the task's criteria list to `- [x]`.  
      Only edit lines within the task's own block (between its heading and the next `---` or `### ` heading). Do NOT touch criteria in other tasks.

   c. **Append resolution metadata**: After the last line of the task block (before the closing `---`), add a new line:
      ```
      **Resuelto**: YYYY-MM-DD (SDD {fast-forward|full cycle}, verify PASS {passed}/{total})
      ```
      Use today's date in ISO format. Extract `{passed}/{total}` from the verify report result. The mode comes from the complexity decision earlier.

   Example transformation:
   ```markdown
   ### P1-004: Remove init() auto-registration side effects
   ...
   **Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: CORR-002
   ```
   Becomes:
   ```markdown
   ### ✅ P1-004: Remove init() auto-registration side effects
   ...
   **Effort**: medium (4-8h) | **Risk**: medium | **Blocks**: CORR-002
   **Resuelto**: 2026-05-26 (SDD fast-forward, verify PASS 3/3)
   ```

4. **Handle edge cases**:
   - **Task already marked** (heading has ✅ prefix): Do not re-mark. Report: "Task {id} was already completed."
   - **Verify fails**: Do NOT mark the task. Report the failure and stop. Do not proceed to Step 2.
   - **Archive fails**: Do NOT mark the task. Report the failure and stop. Do not proceed to Step 2.

### Step 2: Commit After Verification

After Step 1 completes successfully (task marked, verify passed, archive succeeded):

1. **Wait for all SDD phases to finish**. Never commit mid-implementation. The gate is: verify PASS + archive complete + task marked.

2. **Stage the right files**. Use `git add` with these explicit paths — never use `git add -A` or `git add .`:
   - All source files modified/created during implementation (track them from the task's "Files involved" list)
   - The task file that was just marked: `{paths.tasks_dir}{area}.md`
   - Any spec/design artifacts if using OpenSpec mode (e.g., `openspec/` directory)
   - Test files that were added or modified

3. **Create the commit** using the work-unit-commits pattern:
   ```
   {phase}: {task-id} — {short description}

   Implements the task using SDD {fast-forward|full cycle}.
   Verify: PASS {passed}/{total}.
   Task marked as completed in {paths.tasks_dir}{area}.md.
   ```
   
   Example:
   ```
   P1: P1-004 — Remove init() auto-registration side effects

   Implements the task using SDD fast-forward.
   Verify: PASS 3/3.
   Task marked as completed in docs/sequoia/tasks/security.md.
   ```

4. **Commit gate rules**:
   - **Verify must PASS**: If verify reports any CRITICAL or WARNING, do NOT commit. Report the issues and stop.
   - **Archive must succeed**: If archive fails, do NOT commit. The delta specs must be synced first.
   - **Task must be marked**: The task file edit must succeed before committing. If the file can't be edited, report the error.
   - **Never commit if any phase failed**: This ensures the repo always contains verified, working code.

5. **Do NOT push** automatically. The commit is local. Let the user decide when to push, or follow the configured `delivery`/`chain` strategy for PR creation as a separate step. If the user wants auto-push, they'll say so explicitly.

### Commit Sequence Summary

```
sdd-apply  →  sdd-verify  →  sdd-archive  →  mark task  →  git commit
                                                              ↑
                                              ONLY if verify PASS + archive OK
```

If any step before the arrow fails, stop. Do not commit half-verified work.

## Override Flags

These flags bypass the complexity heuristic:

- **`--ff`** (fast-forward): Force `sdd-ff` regardless of complexity score. Use when you KNOW the task is simple, even if the heuristic says otherwise.
- **`--full`** (full cycle): Force the full SDD cycle regardless of complexity score. Use when a task looks simple on paper but you want the full rigor.

Only one override flag is accepted. If both `--ff` and `--full` are passed, error: "Cannot use both --ff and --full. Choose one."

## Error Handling

| Error condition | Behavior |
|-----------------|----------|
| **Task not found** | Search ALL area files (`*.md` in `tasks_dir`) for the task ID. If found in a different area, suggest: "Task P1-003 not found in security.md. Did you mean P1-003 in architecture.md?" If not found anywhere, error: "Task {id} not found in {tasks_dir}. Run `/sequoia audit` to regenerate tasks." |
| **Config file missing** | Use defaults silently. First-time users won't have it. |
| **Config syntax error** | Warn: "⚠️ .sequoia-dev.yaml has a syntax error: {parse_error}. Falling back to defaults." Continue with defaults. |
| **Task missing metadata** | Use safe defaults: risk=Medium (1), effort=medium (1), files=3, deps=none (0). Warn: "⚠️ Task {id} is missing some metadata. Using safe defaults." |
| **SDD init not found** | Error: "SDD has not been initialized. Run `/sdd-init` first." |
| **Both --ff and --full passed** | Error: "Cannot use both --ff and --full. Choose one." |
| **Verify fails** | Do NOT mark task. Do NOT commit. Report verify failures (CRITICAL/WARNING) and stop. |
| **Archive fails** | Do NOT mark task. Do NOT commit. Report archive error and stop. |
| **Task already marked (✅)** | Skip marking. Report: "Task {id} was already completed." Still commit if implementation files changed. |
| **Task file not editable** | Report error: "Cannot update {task_file}: {reason}." Do NOT commit. |
| **Commit after marking fails** | Task is already marked. Report: "Task marked in {task_file} but commit failed: {reason}. Commit manually." |

## Feedback to User

Always report BEFORE launching any subagents:

```
🔧 /sequoia-dev {task-id}
   Mode: {fast-forward | full SDD cycle}
   TDD: {strict | standard}
   Artifacts: {engram | openspec | both}
   Delivery: {ask-on-risk | auto-chain | single-pr | exception-ok}
   Chain: {stacked-to-main | feature-branch-chain}
   Review budget: {n} lines
   Complexity score: {score} (risk={r} + effort={e} + files={f} + deps={d})

{What happens next — brief description of the flow}
```

After implementation completes, report the final status:

```
✅ /sequoia-dev {task-id} — COMPLETED
   Verify: PASS {passed}/{total}
   Task marked: {task_file} → ✅ {task-id}
   Commit: {commit_hash} — "{commit_message_summary}"
   {Elapsed or summary}
```

## Configuration Reference

Your `~/.config/opencode/.sequoia-dev.yaml` controls all strategies:

```yaml
sdd:
  execution_mode: auto     # auto | interactive
  artifact_store: engram   # engram | openspec | both
  tdd: strict              # strict | standard
  delivery: ask-on-risk    # ask-on-risk | auto-chain | single-pr | exception-ok
  chain: stacked-to-main   # stacked-to-main | feature-branch-chain
  review_budget_lines: 400 # max changed lines
  auto_ff: true            # true | false
  complexity:
    ff_max_score: 2        # 0-9

paths:
  tasks_dir: docs/sequoia/tasks/
```

This file is created automatically by `sequoia install` from the default template. You can edit it anytime — `/sequoia-dev` reads it fresh on every invocation.
