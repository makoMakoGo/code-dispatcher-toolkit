<div align="center">

<h1>CODE-DISPATCHER TOOLKIT</h1>
<p><strong>Multi-Backend AI Coding Toolkit</strong></p>
<p>Dispatch tasks across Codex and Claude with<br>reusable Skills and workflow tooling.</p>

<p>
  <a href="README.md">中文</a> | <strong>English</strong>
</p>

<p>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Python-3.9+-3776AB?logo=python&logoColor=white" alt="Python">
  <img src="https://img.shields.io/badge/Bash-4.0+-4EAA25?logo=gnu-bash&logoColor=white" alt="Bash">
  <br>
  <img src="https://img.shields.io/badge/Backend-Codex-412991?logo=data:image/svg%2Bxml;base64,PHN2ZyBmaWxsPSJ3aGl0ZSIgcm9sZT0iaW1nIiB2aWV3Qm94PSIwIDAgMjQgMjQiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHRpdGxlPk9wZW5BSTwvdGl0bGU+PHBhdGggZD0iTTIyLjI4MTkgOS44MjExYTUuOTg0NyA1Ljk4NDcgMCAwIDAtLjUxNTctNC45MTA4IDYuMDQ2MiA2LjA0NjIgMCAwIDAtNi41MDk4LTIuOUE2LjA2NTEgNi4wNjUxIDAgMCAwIDQuOTgwNyA0LjE4MThhNS45ODQ3IDUuOTg0NyAwIDAgMC0zLjk5NzcgMi45IDYuMDQ2MiA2LjA0NjIgMCAwIDAgLjc0MjcgNy4wOTY2IDUuOTggNS45OCAwIDAgMCAuNTExIDQuOTEwNyA2LjA1MSA2LjA1MSAwIDAgMCA2LjUxNDYgMi45MDAxQTUuOTg0NyA1Ljk4NDcgMCAwIDAgMTMuMjU5OSAyNGE2LjA1NTcgNi4wNTU3IDAgMCAwIDUuNzcxOC00LjIwNTggNS45ODk0IDUuOTg5NCAwIDAgMCAzLjk5NzctMi45MDAxIDYuMDU1NyA2LjA1NTcgMCAwIDAtLjc0NzUtNy4wNzI5em0tOS4wMjIgMTIuNjA4MWE0LjQ3NTUgNC40NzU1IDAgMCAxLTIuODc2NC0xLjA0MDhsLjE0MTktLjA4MDQgNC43NzgzLTIuNzU4MmEuNzk0OC43OTQ4IDAgMCAwIC4zOTI3LS42ODEzdi02LjczNjlsMi4wMiAxLjE2ODZhLjA3MS4wNzEgMCAwIDEgLjAzOC4wNTJ2NS41ODI2YTQuNTA0IDQuNTA0IDAgMCAxLTQuNDk0NSA0LjQ5NDR6bS05LjY2MDctNC4xMjU0YTQuNDcwOCA0LjQ3MDggMCAwIDEtLjUzNDYtMy4wMTM3bC4xNDIuMDg1MiA0Ljc4MyAyLjc1ODJhLjc3MTIuNzcxMiAwIDAgMCAuNzgwNiAwbDUuODQyOC0zLjM2ODV2Mi4zMzI0YS4wODA0LjA4MDQgMCAwIDEtLjAzMzIuMDYxNUw5Ljc0IDE5Ljk1MDJhNC40OTkyIDQuNDk5MiAwIDAgMS02LjE0MDgtMS42NDY0ek0yLjM0MDggNy44OTU2YTQuNDg1IDQuNDg1IDAgMCAxIDIuMzY1NS0xLjk3MjhWMTEuNmEuNzY2NC43NjY0IDAgMCAwIC4zODc5LjY3NjVsNS44MTQ0IDMuMzU0My0yLjAyMDEgMS4xNjg1YS4wNzU3LjA3NTcgMCAwIDEtLjA3MSAwbC00LjgzMDMtMi43ODY1QTQuNTA0IDQuNTA0IDAgMCAxIDIuMzQwOCA3Ljg3MnptMTYuNTk2MyAzLjg1NThMMTMuMTAzOCA4LjM2NCAxNS4xMTkyIDcuMmEuMDc1Ny4wNzU3IDAgMCAxIC4wNzEgMGw0LjgzMDMgMi43OTEzYTQuNDk0NCA0LjQ5NDQgMCAwIDEtLjY3NjUgOC4xMDQydi01LjY3NzJhLjc5Ljc5IDAgMCAwLS40MDctLjY2N3ptMi4wMTA3LTMuMDIzMWwtLjE0Mi0uMDg1Mi00Ljc3MzUtMi43ODE4YS43NzU5Ljc3NTkgMCAwIDAtLjc4NTQgMEw5LjQwOSA5LjIyOTdWNi44OTc0YS4wNjYyLjA2NjIgMCAwIDEgLjAyODQtLjA2MTVsNC44MzAzLTIuNzg2NmE0LjQ5OTIgNC40OTkyIDAgMCAxIDYuNjgwMiA0LjY2ek04LjMwNjUgMTIuODYzbC0yLjAyLTEuMTYzOGEuMDgwNC4wODA0IDAgMCAxLS4wMzgtLjA1NjdWNi4wNzQyYTQuNDk5MiA0LjQ5OTIgMCAwIDEgNy4zNzU3LTMuNDUzN2wtLjE0Mi4wODA1TDguNzA0IDUuNDU5YS43OTQ4Ljc5NDggMCAwIDAtLjM5MjcuNjgxM3ptMS4wOTc2LTIuMzY1NGwyLjYwMi0xLjQ5OTggMi42MDY5IDEuNDk5OHYyLjk5OTRsLTIuNTk3NCAxLjQ5OTctMi42MDY3LTEuNDk5N1oiLz48L3N2Zz4=" alt="Codex">
  <img src="https://img.shields.io/badge/Backend-Claude-D4A27F?logo=anthropic&logoColor=white" alt="Claude">
</p>

</div>

A multi-backend AI coding toolkit built around the `code-dispatcher` CLI: executor + Skills.

## Why Dispatcher

Because the meaning of this word fits the core function of this tool:

<div align="center">
<strong>Receive Task</strong> &nbsp;→&nbsp;
<strong>Select Backend</strong> &nbsp;→&nbsp;
<strong>Build Args</strong> &nbsp;→&nbsp;
<strong>Dispatch Execution</strong> &nbsp;→&nbsp;
<strong>Collect Results</strong>
</div>

## Components

### Dispatcher CLI

`code-dispatcher` is a multi-backend task dispatcher that unifies `codex` and `claude` AI coding tools. Core capabilities include:

- Multi-backend support: switch between or parallelize multiple AI backends via `--backend`
- Parallel execution: DAG-based concurrent task scheduling
- Session resume: continue unfinished tasks after context resets
- Unified config: single-point configuration via `~/.code-dispatcher/.env` for all backends

Backend positioning (recommended only, can specify freely):

- `codex`: complex logic, bug fixes, optimization & refactoring
- `claude`: quick tasks, review, supplementary analysis, UI/UX implementation, docs and copy polish

> [!NOTE]
> Tool `code-dispatcher` core idea is based on the `codeagent wrapper` from [`cexll/myclaude`](https://github.com/cexll/myclaude), with substantial refactoring.

### Skills

Note: "Dependency" indicates whether the skill relies on the code-dispatcher CLI for scheduling and execution.

<table>
<tr>
  <th>Name</th>
  <th>Purpose</th>
  <th width="80">Dependency</th>
</tr>
<tr>
  <td><a href="docs/code-dispatcher.md"><code>code&#8209;dispatcher</code></a></td>
  <td>Executor usage guide; unified backends <code>codex/claude</code>; core mechanisms: parallel execution and session resume</td>
  <td>Required</td>
</tr>
<tr>
  <td><a href="docs/dev.md"><code>dev</code></a></td>
  <td>Requirements clarification → plan → select backend → parallel execution (DAG scheduling) → verification</td>
  <td>Required</td>
</tr>
<tr>
  <td><a href="docs/pr-review-reply.md"><code>pr&#8209;review&#8209;reply</code></a></td>
  <td>Autonomous bot-review triage on PRs (Gemini Code Assist / CodeRabbit etc.) → verify → fix or rebut → reply in thread → resolve</td>
  <td>Optional</td>
</tr>
</table>

## Installation

### Step 1: Code Dispatcher CLI Core Installation

The installation script supports Linux x86_64, Windows x86_64, and macOS Apple Silicon. By default, it downloads the current platform binary from GitHub Release `latest`:

```bash
python3 install.py
```

Optional parameters:

```bash
python3 install.py --install-dir ~/.code-dispatcher --force
python3 install.py --skip-dispatcher
```

The script installs the runtime configuration, per-backend prompt templates, and the dispatcher binary for the current platform. The binary layout, config path, and prompt filenames are derived from the runtime asset manifest; after installation, it prints the actual file locations and the PATH command for the current shell.

The documentation does not duplicate the binary directory or PATH commands. When using a custom `--install-dir` or manifest layout, use the command printed by `install.py`.

### Step 2: Install Code Dispatcher Skill

Copy the `code-dispatcher` skill to the appropriate directory based on your target code agent tool's configuration. Global installation is recommended. Typical locations:

- General: `~/.agents/skills`
- Claude Code: `~/.claude/skills`
- Codex CLI: `~/.codex/skills`
- OpenCode: `~/.config/opencode/skills`

### Step 3: Select Skills

Refer to individual skill docs for specific purposes, then select the functional modules you need:

**Skills**: Skills are cross-agent functional modules. The core is the `SKILL.md` definition file, and some also include `references/` documentation. When installing, copy the corresponding skill directory to the target agent's skills directory. Taking Claude as an example:
- Global: `~/.claude/skills/<skill-name>/`
- Project-level: `<path to your project>/.claude/skills/<skill-name>/`

The `dev` skill is a Claude Code manual-only skill, invoked explicitly with `/dev`. Typical usage:

```text
# Explicit trigger
/dev "I want to implement X"

# Keyword trigger
use dispatcher with codex to fix the bug we just discussed
```

### Optional Configuration

Runtime parameters are unified in `~/.code-dispatcher/.env`. Configurable items include:

- Executor timeout parameters
- Executor parallel worker limit
- Executor log output settings
- Invoked backend model override (codex only)

For full field definitions, see: [docs/runtime-config.en.md](docs/runtime-config.en.md). If not configured, default parameters will be used.

## Development / Testing

### Prerequisites

- Go: `1.21` (`code-dispatcher/go.mod`)
- Python: `3.9+` (`install.py` uses `list[str]`)
- Bash: for local build scripts (`scripts/build-dist.sh`)

### Running Tests

```bash
# Identical to the pull-request CI and release gate
bash scripts/verify.sh

# Verify the local cross-platform build
bash scripts/build-dist.sh

# Non-polluting install regression test (using a temporary directory)
tmpdir="$(mktemp -d)"
python3 install.py --install-dir "$tmpdir/.code-dispatcher" --skip-dispatcher --force
python3 install.py --install-dir "$tmpdir/.code-dispatcher" --force
rm -rf "$tmpdir"
```

## Community / Contributing

This is a non-commercial open-source project maintained independently. Bug reports, documentation fixes, usage feedback, and focused small PRs are welcome.

Before submitting changes, please run the relevant verification commands above.
