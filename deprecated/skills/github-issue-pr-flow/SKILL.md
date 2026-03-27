---
name: github-issue-pr-flow
description: "Deprecated archive. Former autonomous GitHub issue-to-PR delivery workflow kept only for reference."
---

# GitHub Issue → PR Delivery Flow

> Deprecated: archived because this skill mixed issue decomposition, implementation, PR lifecycle automation, review triage, and merge closure into one overly coupled workflow. Prefer smaller workflows instead.

## Autonomy Directive
This skill is a **fully autonomous pipeline**. Once invoked, execute all phases (0–6) for ALL issues without pausing for user confirmation between phases or between issues.
- Do NOT stop after completing one issue to ask "should I continue to the next one?" — proceed to the next issue immediately.
- Do NOT present intermediate results and wait for instructions.
- Do NOT summarize partial progress and then stop — finish all issues, then give one final summary.
- The ONLY situations where you may stop are listed in the Escalation Policy.

## Note
When creating issues or PRs with long body content, prefer writing the body to a temp file and using `--body-file` (e.g. `gh issue create --body-file /tmp/gh-body.md`) instead of passing it inline via `--body`. This avoids shell escaping issues and potential model stalling. Clean up the temp file after the command succeeds.

## Operating Principles
- Keep full autonomy for decomposition, branch planning, triage, and merge execution.
- Treat external review tools as advisory signals; host agent is final decision-maker.
- Keep traceability: every PR body must include `Closes #<primary_issue>`.
- Keep scope discipline: one PR should close one primary issue chain.
- Use `code-dispatcher --backend codex` only as backup for `critical`/`high` findings or hard debugging.
- When review automation is rate-limited, switch to `code-dispatcher --backend codex` fallback review and publish findings to GitHub threads.

## Context Discovery
- Detect repository and default base branch from `gh repo view`.
- Do not ask user for repo/base unless `gh` metadata is unavailable.
- Load templates from `references/issue-templates.md` and review reply patterns from `references/review-playbook.md` when needed.

## Autonomous Workflow
Execute phases in order with no user confirmation gates unless escalation rules are hit.

### Phase 0: Sync Base Branch
- Fetch and fast-forward base before creating or updating working branches.
- If `git pull --ff-only` fails, stop and escalate to user. Never hard reset or force push automatically.

### Phase 1: Decompose Into Issues (Optional)
Decide issue structure based on task scope and complexity:
- Simple/small tasks: skip issue creation entirely, go straight to Phase 2.
- Medium tasks: create a single issue, no epic needed.
- Large/multi-deliverable tasks: decompose into multiple issues; create an epic only when multiple child issues need a parent umbrella.

When decomposing:
- Decompose by deliverables, not by file counts.
- Prefer one issue = one user-verifiable outcome with acceptance criteria and test evidence.
- Split work when merge/review risk is high or when tasks can land independently.

### Phase 2: Implement With Branch Discipline
- Create branch names as `feat/issue-<number>-<slug>` or `fix/issue-<number>-<slug>`.
- Keep one active PR branch per issue chain.
- Keep commits scoped to issue intent; avoid mixed-purpose commits.
- Follow `references/branch-history.md` for squash-merge aftermath and new-PR history hygiene.

### Phase 3: Open PR and Link Issues
- Ensure PR body contains `Closes #<primary>` and optional `Relates #<secondary>`.
- Keep PR summary testable and concise.

### Phase 4: Collect Reviews and CI Signals
- Wait a reasonable review window, then collect latest review and check-run states.
- Do not block forever if external review tools stay silent; CI and maintainers remain the gate.
- If CodeRabbit reports rate limit exhaustion, trigger fallback review flow in `references/review-playbook.md` and continue without waiting for CodeRabbit recovery.

### Phase 5: Triage Review Findings
- Collect findings from **both** review bodies (parse `<details>` / nitpick / issue blocks in `reviews[].body`) AND line-level comment threads. Findings = union of both sources.
- Decide each finding independently: fix, rebut with evidence, or align with repo convention.
- Reply under each review comment and resolve each thread after handling. For body-only findings with no line-level thread, post a scoped PR comment.
- Accumulate all code fixes locally before pushing — a premature push re-triggers CodeRabbit and wastes minutes.
- After push, re-fetch both review bodies and line-level comments to detect new findings. Do not only check line-level comments.
- Limit autonomous rebut/fix loops to 3 rounds, then escalate if blocking findings remain.

### Phase 6: Squash Merge and Closure
- Merge using squash and delete remote branch.
- Verify linked issues are auto-closed; close residual issues manually if needed.
- Apply post-merge local cleanup and next-PR branch reset rules from `references/branch-history.md`.

## Escalation Policy
Escalate to user only when:
1. Base branch sync cannot be fast-forwarded.
2. A destructive action is required (`force-push`, history rewrite on shared branch, closing unrelated issues).
3. Two triage rounds still leave blocking review/CI failures.
4. Fix requires unavailable credentials or infrastructure access.

Everything else should proceed autonomously.

## Failure Handling
- External reviews absent: proceed with CI and maintainer review signals.
- External review feedback conflicts: prioritize reproducible failures and test evidence.
- CodeRabbit rate limit exceeded: run fallback PR review via `code-dispatcher --backend codex`, require it to post findings on GitHub, then triage those findings with the same standards.
- Scope creep appears mid-flight: split to new issue(s) and keep current PR focused.
- `gh` command fails: retry once; if still failing, report the blocking error and stop.
