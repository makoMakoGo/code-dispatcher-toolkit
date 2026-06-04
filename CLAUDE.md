# code-dispatcher Project Instructions

## Project Summary
code-dispatcher is a multi-backend AI coding toolkit: a Go CLI that dispatches tasks to `codex`/`claude` backends, paired with reusable Skills for dispatcher execution, end-to-end development orchestration, bot-review triage, and long-running task workflows.

## Tech Stack
- Go: main dispatcher program
- Python: install/uninstall scripts
- Bash: build and distribution script

## Repository Structure
- `code-dispatcher/`: Go source (main package and backend dispatch logic)
- `skills/`: reusable modules
- Skills: `dev`, `code-dispatcher`, `pr-review-reply`
- `docs/`: documentation (`runtime-config.md`, `runtime-config.en.md`, `code-dispatcher.md`, `dev.md`, `pr-review-reply.md`)
- `prompts/`: per-backend default prompt templates (`claude-prompt.md`, `codex-prompt.md`)
- `templates/`: scaffold templates (`runtime-config.env`)
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
- All backend execution must go through `code-dispatcher`; do not call `codex` or `claude` directly.
- Single runtime config source: `~/.code-dispatcher/.env`.
- Prompt injection files: `~/.code-dispatcher/prompts/<backend>-prompt.md`.
