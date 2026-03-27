# code-dispatcher Project Instructions

## Project Summary
code-dispatcher is a multi-backend AI coding toolkit: a Go CLI that dispatches tasks to `codex`/`claude`/`gemini` backends, paired with reusable Skills and Bundles for structured workflows such as planning, parallel execution, iterative waves, multi-reviewer code review, bot-review triage, and long-running task orchestration.

## Tech Stack
- Go: main dispatcher program
- Python: install/uninstall scripts
- Bash: build and distribution script

## Repository Structure
- `code-dispatcher/`: Go source (main package and backend dispatch logic)
- `skills/`: reusable modules
- `deprecated/`: archived skills and docs kept only for reference
- `bundles/`: packaged Claude Code assets such as `commands/`, `hooks/`, `settings`, and supporting docs
- Skills: `dev`, `wave`, `code-dispatcher`, `code-council`, `pr-review-reply`
- Bundles: `codex-review-loop`, `harness`
- `docs/`: documentation (`runtime-config.md`)
- `scripts/`: build scripts (`build-dist.sh`)
- `install.py` / `uninstall.py`: installer and uninstaller
- `dist/`: build outputs (gitignored)

## Development Commands
```bash
cd code-dispatcher && go test ./...
bash scripts/build-dist.sh
python3 install.py
```

## Code Conventions
- All backend execution must go through `code-dispatcher`; do not call `codex`, `claude`, or `gemini` directly.
- Single runtime config source: `~/.code-dispatcher/.env`.
- Prompt injection files: `~/.code-dispatcher/prompts/<backend>-prompt.md`.
