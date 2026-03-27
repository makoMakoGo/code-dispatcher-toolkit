# Branch and Commit History Policy

## Goal
Keep each PR history clean after squash merge and prevent old commits from leaking into new PRs.

## Core Rules
1. Treat a squash-merged PR branch as closed history.
2. Start every new PR from the latest remote base branch, not from the old feature branch.
3. Keep one PR branch mapped to one primary issue chain.
4. Avoid stacking new work on top of a branch that has already been squash-merged.

## Why This Matters
Squash merge rewrites integration history on base:
- Base gets one new squashed commit.
- Original feature branch keeps the old commit chain.
- Reusing that old branch often reintroduces stale commits into the next PR.

ASCII model:
```text
before squash:
main:    A---B
feature:      \---c1---c2---c3

after squash merge:
main:    A---B---S
feature(local):  \---c1---c2---c3   (stale lineage)
```

`S` contains combined changes, but `c1/c2/c3` still exist on the old branch.

## New PR Workflow (Best Practice)
1. Sync local base to latest remote.
2. Create a fresh branch from that synced base.
3. Apply only required changes for the new issue.
4. Open PR from the fresh branch.

## Follow-up Changes After Squash Merge
If a follow-up PR needs part of previous work:
1. Create a fresh branch from latest base.
2. Cherry-pick only required commits or re-implement minimal delta.
3. Drop unrelated or already-squashed commits.

If commit lineage is messy, prefer re-implementation on a clean branch over aggressive history surgery.

## Cleanup Checklist After Merge
1. Delete merged local branch.
2. Prune remote-tracking refs.
3. Confirm current branch is base (or another active clean branch).
4. Start next issue from latest base snapshot.

## Escalation Cases
Escalate before proceeding when:
1. Work requires force-push on a shared branch.
2. Branch lineage is unclear and commit ownership cannot be trusted.
3. Follow-up change appears to depend on unreleased commits from another open PR.
