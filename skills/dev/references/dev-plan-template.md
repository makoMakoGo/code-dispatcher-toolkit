# dev-plan.md Generation Template

## Output Path

`.specs/{feature_name}/dev-plan.md`

Where `{feature_name}` is kebab-case derived from the feature description.

## Document Structure

```markdown
# {Feature Name} - Development Plan

## Overview
[One-sentence description of core functionality]

## Task Breakdown

### Task 1: [Task Name]
- **ID**: task-1
- **type**: default|ui|quick-fix|docs
- **Description**: [What needs to be done]
- **File Scope**: [Directories or files involved, e.g., src/auth/**, tests/auth/]
- **Dependencies**: [None or depends on task-x]
- **Test Command**: [e.g., pytest tests/auth --cov=src/auth --cov-report=term]
- **Test Focus**: [Scenarios to cover]

### Task 2: [Task Name]
...

(2–5 tasks based on natural functional boundaries)

## Acceptance Criteria
- [ ] Feature point 1
- [ ] Feature point 2
- [ ] All unit tests pass
- [ ] Code coverage ≥90%

## Technical Notes
- [Key technical decisions]
- [Constraints to be aware of]
```

## Task Type Definitions

- `ui`: touches UI/style/component work (.css/.scss/.tsx/.jsx/.vue, tailwind, design tweaks)
- `quick-fix`: small, fast changes (config tweaks, small bug fix, minimal scope); NOT for UI work
- `docs`: documentation, README, API specs, technical notes
- `default`: everything else

Routing reference: `default`→codex, `ui`→claude, `quick-fix`→claude, `docs`→claude; missing type → default. UI routing means component/layout/style/interaction work goes to claude.

## Generation Rules

1. **Task Count**: 2–5 tasks based on natural functional boundaries. Prefer fewer well-scoped tasks over fragmentation. Each task independently completable.
2. **Required Fields**: Every task must have ID, type, description, file scope, dependencies, test command, test focus.
3. **Task Independence**: Maximize independence to enable parallel execution. Minimize dependencies.
4. **Test Commands**: Must include coverage parameters (`--cov=module --cov-report=term` for pytest, `--coverage` for npm, etc.).
5. **Coverage Threshold**: Always require ≥90% in acceptance criteria.
6. **Language Matching**: Output language matches user input (Chinese input → Chinese doc).
7. **Append UI Task**: If analysis marked `needs_ui: true` but no task has `type: ui`, append a dedicated UI task.

## Quality Checks

Before writing the file, verify:
- Every task has all required fields (ID, type, description, file scope, dependencies, test command, test focus)
- File scope is specific (not "all files")
- Testing focus is concrete (not "test everything")
- Test commands include coverage parameters
- Dependencies are explicitly stated
- Acceptance criteria includes ≥90% coverage

## Input Requirements

If input context is incomplete, request clarification on: feature scope, tech stack, testing framework, file structure. Do not generate a low-quality plan with guessed details.
