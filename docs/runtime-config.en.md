# code-dispatcher Runtime Config

This document is the single source of truth for runtime behavior covering:

- Timeout behavior
- Parallel execution and worker limits
- Config loading model

## 1) Single Config Source

All runtime options are loaded from:

```text
~/.code-dispatcher/.env
```

`code-dispatcher` no longer reads these control items from shell environment variables.

Only `CODE_DISPATCHER_*` keys are dispatcher control keys. Other keys in `.env` are injected into backend processes; when the same key exists in the inherited shell environment, the `.env` value wins.

## 2) Backend Approval / Bypass Approval Settings

All backends run with "bypass approval" mode by default (not toggleable):

- `codex`: `--dangerously-bypass-approvals-and-sandbox`
- `claude`: `--dangerously-skip-permissions`

## 3) Runtime Fields in `.env`

- `CODE_DISPATCHER_TIMEOUT`
  - Default: `7200` (seconds, 2 hours)
  - Unit: seconds

- `CODE_DISPATCHER_MAX_PARALLEL_WORKERS`
  - Default: unlimited (`0`)
  - Recommended: `8`
  - Dispatcher hard cap: `100`

- `CODE_DISPATCHER_ASCII_MODE`
  - `1` / `true` / `yes` / `on`: use ASCII status codes (`PASS/WARN/FAIL`)
  - `0` / `false` / `no` / `off` or unknown values: use Unicode status symbols

- `CODE_DISPATCHER_LOGGER_CLOSE_TIMEOUT_MS`
  - Default: `5000`
  - `<= 0`: means wait indefinitely

### Backend Model Override

- `CODE_DISPATCHER_CODEX_MODEL`
  - Optional; if set, passes `-m <value>` to codex CLI
  - Example: `gpt-5.4`

- Claude does not support model override via dispatcher.

## 4) Prompt Files

Prompt files are loaded from:

```text
~/.code-dispatcher/prompts/<backend>-prompt.md
```

Supported backends: `codex`, `claude`.

The installer does not overwrite existing prompt files unless `python3 install.py --force` is used. After default prompt contracts change, existing installations must sync these files manually or reinstall with `--force`.

## 5) Timeout Layering (Important)

There are usually two timeout layers:

- Outer caller timeout (e.g., tool invocation timeout)
- Dispatcher timeout (`CODE_DISPATCHER_TIMEOUT` from `.env`)

The effective timeout is whichever triggers first.

## 6) Editing Config

Edit the file directly:

```bash
${EDITOR:-vi} ~/.code-dispatcher/.env
```
