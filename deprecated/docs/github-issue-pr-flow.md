# github-issue-pr-flow

Archived.

This skill has been moved out of the active `skills/` catalog into `deprecated/skills/github-issue-pr-flow/`.

Reason:

- It coupled issue decomposition, implementation, PR opening, review triage, and merge closure into one monolithic workflow.
- It overlapped heavily with `pr-review-reply`, which is narrower and easier to maintain.
- The repository now prefers smaller composable workflows over one skill owning the entire GitHub delivery lifecycle.

If you need PR review handling, use `pr-review-reply`.
