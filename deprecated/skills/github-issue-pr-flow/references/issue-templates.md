# Issue Templates

Both epic and sub-issue layers are optional. Use the structure that matches task scope: skip issues for trivial tasks, use a single issue for medium tasks, and add an epic wrapper only for multi-issue deliveries.

## Epic Template
```markdown
# Epic: <title>
## Goal
<business/engineering goal>

## Scope
- In scope: ...
- Out of scope: ...

## Delivery Plan
- [ ] #<issue-a>
- [ ] #<issue-b>

## Acceptance Criteria
- ...
```

## Issue Template
```markdown
# <title>
## Background
<why this is needed>

## Requirements
- ...

## Acceptance Criteria
- [ ] ...
- [ ] ...

## Dependencies
- Parent: #<epic_issue> (optional)
- Blocks: #<another_issue> (optional)
```

## PR Body Template
```markdown
## Summary
- ...

## Linked Issues
Closes #<primary_issue>
Relates #<secondary_issue>

## Testing
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual verification

## Risk
- ...
```
