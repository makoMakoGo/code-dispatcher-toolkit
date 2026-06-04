---
name: dev
description: "Claude Code /dev manual-only end-to-end development workflow orchestrator. Use only when the user explicitly invokes /dev. Orchestrates 7 steps: backend selection, requirements, code-dispatcher analysis, dev plan, parallel execution, coverage validation (>=90%), and summary. All code changes and codebase analysis are delegated to code-dispatcher. Depends on the code-dispatcher skill."
disable-model-invocation: true
---

# /dev Workflow

## Hard Rules

1. ALL code changes go through code-dispatcher — never edit files directly
2. ALL backend calls go through code-dispatcher — never invoke codex or claude directly
3. Steps 0 and 1 require user input before continuing
4. Step 3 requires explicit user confirmation before Step 4

Violation → stop and restart.

## Step 0: Backend Selection [FIRST ACTION]

Ask which backends are allowed for this run (multi-select):
- `codex` — complex logic, refactoring, debugging (default)
- `claude` — quick fixes, config, docs, review, UI/UX implementation, copy/polish

Store as `allowed_backends`. If only codex selected → all tasks forced to codex.
Guidance: for non-trivial logic or multi-file refactors, recommend enabling at least codex.

## Step 1: Requirements Clarification

Ask targeted questions on scope, I/O, constraints, testing expectations. 2–3 rounds.
Create task tracking list before proceeding.

## Step 2: Analysis via code-dispatcher

Invoke code-dispatcher in a shell (backend: prefer codex from allowed_backends; fallback codex → claude).
Do NOT explore the codebase directly — delegate all exploration to code-dispatcher.

```bash
code-dispatcher --backend {analysis_backend} - <<'EOF'
Analyze the codebase for implementing [feature].

Requirements: [from Step 1]

Deliverables:
1. Codebase structure and existing patterns
2. Implementation options with trade-offs (if multiple valid approaches)
3. Architectural decisions with justification
4. Task breakdown: 2–5 tasks with ID, description, file scope, dependencies, type
5. UI detection: needs_ui true/false with evidence (.css/.tsx/.vue presence)
EOF
```

Task types: `default` | `ui` | `quick-fix` | `docs`

**Skip when**: single obvious approach, ≤2 files, clear requirements.

## Step 3: Development Plan

Generate `dev-plan.md` following [references/dev-plan-template.md](references/dev-plan-template.md).
Output path: `.specs/{feature_name}/dev-plan.md`

- Present summary to user: task count, types, file scopes, dependency graph, backend routing
- Ask user to confirm before execution
- If user wants adjustments → return to Step 1 or Step 2

## Step 4: Parallel Execution

Build ONE `--parallel` config covering all tasks from dev-plan.md. Submit once via code-dispatcher in a shell.

**Backend routing by task type**:
- `default` → codex; fallback `codex → claude`
- `ui` → claude; fallback `claude → codex` (Claude owns component/layout/style/interaction work)
- `quick-fix` → claude; fallback `claude → codex`
- `docs` → claude; fallback `claude → codex`
- Missing type → treat as `default`

Choose the first backend in the task type's fallback chain that is present in `allowed_backends`.

```bash
code-dispatcher --parallel --backend {analysis_backend} <<'EOF'
---TASK---
id: task-1
backend: {routed_backend}
workdir: .
dependencies:
---CONTENT---
Task: task-1
Reference: @.specs/{feature_name}/dev-plan.md
Scope: [file scope from plan]
Test: [test command from plan]
Deliverables: code + unit tests + coverage ≥90%

---TASK---
id: task-2
backend: {routed_backend}
workdir: .
dependencies: task-1
---CONTENT---
Task: task-2
Reference: @.specs/{feature_name}/dev-plan.md
Scope: [file scope from plan]
Test: [test command from plan]
Deliverables: code + unit tests + coverage ≥90%
EOF
```

Use `workdir: .` unless a task requires a specific subdirectory.

## Step 5: Coverage Validation

All tasks must hit ≥90%. Retry failures max 2 rounds, then report to user.

## Step 6: Summary

Completed tasks, coverage per task, key file changes.

## Error Handling

- code-dispatcher failure → retry once → ask user
- Coverage <90% after retries → report to user
- Dependency cycles → code-dispatcher detects and fails; revise task breakdown
- Backend unavailable → follow fallback priority above
