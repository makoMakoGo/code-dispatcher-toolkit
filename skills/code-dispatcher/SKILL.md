---
name: code-dispatcher
description: Execute code-dispatcher for multi-backend AI code tasks. Use when the user explicitly requests a specific backend (Codex or Claude), mentions code-dispatcher, or when a skill or command definition declares a dependency on this skill. Supports pluggable backends (codex/claude), parallel task execution with DAG scheduling, session resume, and compact parallel summaries.
---

# code-dispatcher Integration

## Overview

Execute code-dispatcher commands with pluggable AI backends (Codex, Claude). Use it for backend-selected execution, file references via backend-native `@` syntax, parallel task execution, DAG dependencies, and session resume.

## When to Use

When the user explicitly requests a specific backend (Codex or Claude), mentions code-dispatcher, or when a skill or command definition explicitly declares a dependency on this skill.

Applicable scenarios include but are not limited to:
- Complex code analysis requiring deep understanding
- Large-scale refactoring across multiple files
- Automated code generation with backend selection


## Typical Usage for One-Round/New Tasks:

**1) Standard invocation: HEREDOC syntax (recommended)**
```bash
code-dispatcher --backend codex - [working_dir] <<'EOF'
<task content here>
EOF
```

**2) Single-line tasks (no heredoc)**
```bash
code-dispatcher --backend codex "simple task" [working_dir]
code-dispatcher --backend claude "simple task" [working_dir]
```

## Common Parameters

- Command notation and positional order
  - `[]` means optional. Do not type brackets literally.
  - Inline task: `code-dispatcher --backend <backend> "<task>" [working_dir]` — task text is passed in command args; best for short one-line prompts.
  - Stdin task: `code-dispatcher --backend <backend> - [working_dir]` — task text is read from stdin (`<<'EOF'`/pipe); best for multi-line or complex content.
  - These two forms are for new-task commands only.
  - Resume uses its own forms: inline `code-dispatcher --backend <backend> resume <session_id> "<follow-up task>"` and stdin `code-dispatcher --backend <backend> resume <session_id> -`.

- `--backend <backend>` (required)
  - Select backend explicitly: `codex | claude`.
  - Must be present in both new and resume modes.
  - In parallel mode, this is the global default backend.
  - If a task block defines `backend`, it overrides the global default for that task.

- `task` (required)
  - Task description for the backend.
  - Supports inline text or stdin marker `-`.
  - Supports `@file` references (backend-native feature, not processed by dispatcher).

- `working_dir` (optional)
  - Working directory for new task execution.
  - Omit it to use the current directory.
  - Typical values: `.`, `./subdir`, `/absolute/path`.
  - In resume mode, do not append `working_dir`; resume follows backend session context.

- Output modes (parallel execution only)
  - **Summary (default)**: Compact execution report with task status, log path, and any reported fields such as `Did`, `Files`, `Tests`, or `Coverage`.
  - Summary mode may omit fields that were not reported by the backend. Do not infer missing work from missing fields alone; inspect `Log:` or use `--full-output` when details matter.
  - **Full (`--full-output`)**: Complete per-task backend messages. Use when debugging failures, inspecting omitted details, or when raw task output is required.
  - Scope: `--full-output` is valid only with `--parallel`; single-task mode does not support this flag.
  - Backend behavior: mode selection is dispatcher-level and works the same for `codex | claude`.

## Utility Commands

- `code-dispatcher --help`
  - Print CLI usage, supported backends, common mistakes, runtime config hints, and exit codes.

- `code-dispatcher --cleanup`
  - Remove orphaned `code-dispatcher-*.log` files from the temp directory.
  - Safe behavior: logs for active dispatcher processes are kept; only stale/orphaned logs are deleted.
  - Use this when repeated runs leave many old temp logs behind.

## Return Format

Single-task mode returns the backend message directly, followed by `SESSION_ID` when available:

```text
Agent response text here...

---
SESSION_ID: <session_id>
```

Parallel summary mode returns a compact execution report:

```text
=== Execution Report ===
N tasks | N passed | N failed

## Task Results

### <task-id> ✓ <coverage>
Did: ...
Files: ...
Tests: ... passed
Log: ...

## Summary
- N/N completed successfully
```

Failed tasks include `FAILED`, `Exit code`, `Error`, `Detail`, and `Log`. Below-target coverage includes `below <target>%` and `Gap`.

Parallel full mode returns raw per-task backend messages:

```text
=== Parallel Execution Summary ===
Total: N | Success: N | Failed: N

--- Task: <task-id> ---
Status: SUCCESS
Coverage: ...
Session: ...
Log: ...

<raw backend message>
```

## Backends Selection Guide

**Note**: This backends selection guide applies only when the user has not explicitly requested a specific backend. If the user specifies a backend, always follow the user's instructions.


Quick look at the differences between backends:

| Backend | Command | Description | Best For |
|---------|---------|-------------|----------|
| codex | `--backend codex` | OpenAI Codex (default) | Code analysis, complex development, debugging |
| claude | `--backend claude` | Anthropic Claude | Quick fixes, documentation, prompts, review, UI/UX and copy polish |

For detailed guidance:

**Codex**:
- Deep code understanding and complex logic implementation
- Large-scale refactoring with precise dependency tracking
- Algorithm optimization and performance tuning
- Example: "Analyze the call graph of @src/core and refactor the module dependency structure"

**Claude**:
- Quick feature implementation with clear requirements
- Technical documentation, API specs, README generation
- Professional prompt engineering (e.g., product requirements, design specs)
- UI/UX implementation: component scaffolding, layout, styling, accessibility, and interaction behavior
- Design-system wiring, UI copy, labels, empty states, review notes, and polish tasks
- Example: "Generate a comprehensive README for @package.json with installation, usage, and API docs"

A Typical Backend Switching Example:
- Start with Codex for analysis and complex implementation; switch to Claude for UI-heavy work, documentation, review notes, and copy polish.
- Use per-task backend selection in parallel mode to optimize for each task's strengths

## Resume Session

Supported backends all support resume mode: `codex | claude`.

**1) Standard resume (HEREDOC)**
```bash
code-dispatcher --backend codex resume <session_id> - <<'EOF'
<follow-up task>
EOF
```

**2) Single-line resume (no heredoc)**
```bash
code-dispatcher --backend claude resume <session_id> "follow-up task"
```

**3) Parallel resume (supported)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: resume-a
backend: claude
session_id: sid_claude_1
---CONTENT---
follow-up for claude session

EOF
```

In parallel mode, any task that provides `session_id` runs in resume mode.

Resume mode relies on backend session context.
- Do not append `[working_dir]` in resume commands.
- If you need a different directory, start a new session instead of resume.

 Resume identifier contract:
- Use the dispatcher-returned `SESSION_ID` as the source of truth for follow-up resume commands.
- Standard form: `code-dispatcher --backend <backend> resume <SESSION_ID> ...`.


## Parallel Mode (`--parallel`)

`--parallel` is the execution entry for submitting multiple tasks in one run.

- Each task must define a unique `id`.
- `--backend` is a required global fallback; tasks without `backend` use it. Usually set to `codex`.
- `backend` inside a task block overrides the global fallback for that task.
- Tasks without `dependencies` are independent and can run concurrently.

Supported metadata fields inside each `---TASK---` block:
- `id` (required): unique task identifier.
- `backend`: override the global `--backend` for this task.
- `workdir`: working directory for this task. Default: `.`.
- `dependencies`: comma-separated task IDs this task depends on.
- `session_id`: resume an existing session; automatically switches this task to resume mode.

Validation rules:
- `---CONTENT---` is required for every task block.
- Lines without `:` are ignored.
- Unknown metadata keys with `:` fail fast with a parse error.
- `id` must be present and unique.
- `workdir: -` is invalid.
- `session_id` cannot be empty when present.

**1) Basic parallel (independent tasks, global backend fallback)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: scan-api
---CONTENT---
scan @src/api and summarize route structure

---TASK---
id: scan-db
---CONTENT---
scan @src/db and summarize data access patterns
EOF
```

**2) Parallel with per-task backend override (mixed backends)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: analyze
---CONTENT---
analyze code structure

---TASK---
id: document
backend: claude
---CONTENT---
write technical notes based on the analysis

---TASK---
id: follow-up
backend: claude
---CONTENT---
write implementation notes from the same requirements
EOF
```

Output styles in parallel mode:

**1) Summary mode (default, no flag)**, optimized for context efficiency. It may show only task status and `Log:` when optional fields are not reported. Use `--full-output` or inspect the task log when you need raw output, omitted details, or debugging context.
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: t1
---CONTENT---
analyze @src and summarize architecture changes
EOF
```

**2) Full mode (`--full-output`)**, for debugging failures, inspecting omitted details, or when full per-task messages are required.
```bash
code-dispatcher --parallel --backend codex --full-output <<'EOF'
---TASK---
id: t1
---CONTENT---
analyze @src and summarize architecture changes
EOF
```

## DAG Scheduling (inside `--parallel`)

When a task declares `dependencies`, `--parallel` runs tasks as a DAG scheduler.

- `dependencies: a, b` means this task waits for tasks `a` and `b` to succeed.
- Tasks in the same DAG layer run concurrently; the next layer starts only after the current layer finishes.
- If a dependency fails, dependent tasks are skipped.
- Invalid dependency IDs or dependency cycles fail fast before execution starts.

ASCII execution model:
```text
layer 0: task1      taskX
           |          |
layer 1: task2      taskY
             \      /
layer 2:      task3
```

**1) Dependency scheduling (DAG with mixed backends)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: task1
workdir: /path/to/dir
---CONTENT---
analyze code structure

---TASK---
id: task2
backend: claude
dependencies: task1
---CONTENT---
design architecture based on task1 analysis

---TASK---
id: task3
backend: codex
dependencies: task2
---CONTENT---
generate implementation code
EOF
```

**2) Minimal DAG example (annotated)**
```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: prep
# uses global backend codex
---CONTENT---
scan @src and list key modules

---TASK---
id: plan
backend: claude
# overrides global backend for this task
dependencies: prep
---CONTENT---
write implementation plan based on prep output
EOF
```

## Runtime Config and Patterns

- Runtime settings and approval policy are operator-configured before invocation.
- Do not set or override runtime environment values from this skill.
- Treat dispatcher-internal timeout as a long fallback configured by the operator.
- Control actual waiting budget via the host tool-call timeout.
- Timeout tiers for tool-call invocations:
  - Simple tasks: `600` s (10 minutes) minimum.
  - Normal tasks: `1800` s (30 minutes) recommended default.
  - Complex Codex tasks: `7200` s (2 hours).
- Do not use short timeouts like `300` s (5 minutes) for normal or complex tasks.

Invocation Pattern:

**Single Task**:
```
Host-agnostic tool-call template (field names vary by runtime):
- command payload (`command` or `cmd`):
  code-dispatcher --backend <backend> - [working_dir] <<'EOF'
  <task content>
  EOF
- timeout field (`timeout` / `timeout_ms` / equivalent): choose by tier (`600` / `1800` / `7200`)
- description field: optional

Field names depend on the host tool schema.
Timeout policy: always set explicit timeout by task complexity; do not rely on implicit defaults.

Note: `--backend` is required; supported values: `codex | claude`
```

**Parallel Tasks**:
```
Host-agnostic tool-call template (field names vary by runtime):
- command payload (`command` or `cmd`):
  code-dispatcher --parallel --backend <backend> <<'EOF'
  ---TASK---
  id: task_id
  backend: <backend>  # Optional, overrides global
  workdir: /path
  dependencies: dep1, dep2
  ---CONTENT---
  task content
  EOF
- timeout field (`timeout` / `timeout_ms` / equivalent): choose by tier (`600` / `1800` / `7200`)
- description field: optional

Field names depend on the host tool schema.
Timeout policy: always set explicit timeout by task complexity; do not rely on implicit defaults.

Note: Global --backend is required; per-task backend is optional
```

## Exit Codes

- `0`: Success.
- `1`: General error (for example: missing args, no output, invalid config).
- `124`: Timeout.
- `127`: Backend command not found in `PATH`.
- `130`: Interrupted (`Ctrl+C` / `SIGINT`).
- Other non-zero codes: passthrough from the backend process.

## Critical Rules

**NEVER kill code-dispatcher processes by default.** Long-running tasks are normal. Instead:

1. **Check task status via log file**:
   ```bash
   # Log path is printed to stderr at startup
   # Format: $TMPDIR/code-dispatcher-{PID}[-{taskID}].log
   tail -f "$TMPDIR"/code-dispatcher-*.log
   ```

2. **Wait with tiered timeout (host-runtime API)**:
  - Use the host runtime's blocking wait API (for example: TaskOutput/wait-result equivalents).
  - Choose timeout by complexity:
    - Simple: `600` s (10m)
    - Normal: `1800` s (30m)
    - Complex Codex: `7200` s (2h)
  - If the wait call times out, do not kill the process; re-check logs/process and continue waiting.

  - Pseudocode (adapt field names to your host runtime):
   ```text
   TaskOutput(task_id="<id>", block=true, timeout=600)
   TaskOutput(task_id="<id>", block=true, timeout=1800)
   TaskOutput(task_id="<id>", block=true, timeout=7200)
   ```

3. **Check process without killing**:
   ```bash
   ps aux | grep code-dispatcher | grep -v grep
   ```

**Why:** code-dispatcher tasks often take 5-120 minutes. Killing them wastes API costs and loses progress.

## Emergency Stop (User-Requested Only)

- Hard rule: kill/terminate is allowed **only when the user explicitly requests it**.
- Do not kill processes automatically because of long runtime or wait timeout.
- Use staged termination and stop escalation as soon as processes exit.
- Name-based global cleanup (`pkill -x codex/claude`) is prohibited.

1. **Graceful stop dispatcher first**:
   ```bash
   # Inspect running dispatcher processes
   pgrep -fa code-dispatcher

   # Soft interrupt first
   pkill -INT -f '(^|/)code-dispatcher( |$)'
   ```

2. **Escalate only if still running**:
   ```bash
   pkill -TERM -f '(^|/)code-dispatcher( |$)'
   sleep 2
   pkill -KILL -f '(^|/)code-dispatcher( |$)'
   ```

3. **Cleanup only descendants of the target code-dispatcher PID (safe default)**:
   ```bash
   # Pick target code-dispatcher PID first (example: newest one)
   DISPATCHER_PID=$(pgrep -n -f '(^|/)code-dispatcher( |$)')

   # TERM direct children of this dispatcher only
   pkill -TERM -P "$DISPATCHER_PID" 2>/dev/null || true
   sleep 2

   # If still present, escalate to KILL for direct children only
   pkill -KILL -P "$DISPATCHER_PID" 2>/dev/null || true
   ```

4. **Post-check**:
   ```bash
   pgrep -fa code-dispatcher
   pgrep -fa 'codex|claude'
   ```
