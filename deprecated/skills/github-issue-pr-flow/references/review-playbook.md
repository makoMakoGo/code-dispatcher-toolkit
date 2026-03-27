# Review Triage Playbook

## Objective
Handle each review comment with an explicit decision: fix or rebut.

## Host Decision Authority
1. Treat all external reviewers as inputs, not as decision owners.
2. Make final fix/rebut decisions from code evidence, test results, and project conventions.
3. Escalate only when ambiguity or risk crosses escalation policy in `../SKILL.md`.

## Rebuttal Template
```markdown
Thanks for the review.

Decision: rebuttal
Reason:
1. <factual evidence>
2. <code path or test evidence>

Verification:
- <test/command/output summary>
```

## Fix Template
```markdown
Thanks for the review.

Decision: accepted and fixed
Changes:
1. <what changed>
2. <why this resolves the concern>

Verification:
- <test/command/output summary>
```

## Triage Checklist
1. Is the reported issue reproducible?
2. Does it violate project constraints or conventions?
3. Is there a smaller fix that preserves scope?
4. Is rebuttal supported by concrete evidence?

## Evidence Priority
1. Reproducible failure and CI breakage
2. Failing or missing tests tied to the finding
3. Concrete code-path impact and runtime risk
4. Reviewer wording or severity label

Never let reviewer label override stronger contradictory evidence.

## Rebuttal Minimum Bar
A rebuttal is acceptable only if it includes:
1. Explicit decision (`rebuttal`) and one-sentence conclusion.
2. At least one falsifiable technical reason (code path, invariant, contract, or test behavior).
3. Verification summary (existing test evidence or rerun result).

## Required Signals Before Triage
Collect these signals with any suitable `gh` query pattern:
1. Review-level summary: `author`, `state`, `submittedAt`, `body` (allow truncated body).
2. Line-level review comments/conversations for actionable findings.
3. Latest check-run states for CI gating.

Do not start fix-or-rebut decisions if review body text is missing.

## CodeRabbit Rate-Limit Fallback
When CodeRabbit reports rate limiting, fall back to `code-dispatcher --backend codex` for review.
Ensure findings end up as GitHub PR comments (either codex posts directly, or host agent collects and posts) so they are visible in the PR history.
Treat fallback findings as normal review signals and continue the fix/rebut loop.
