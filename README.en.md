<div align="center">

<h1>CODE-DISPATCHER TOOLKIT</h1>
<p><strong>Multi-Backend AI Coding Toolkit</strong></p>
<p>Dispatch tasks across Codex, Claude, and Gemini with<br>reusable Skills, Bundles, and workflow tooling.</p>

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
  <img src="https://img.shields.io/badge/Backend-Gemini-4285F4?logo=google&logoColor=white" alt="Gemini">
</p>

</div>

A multi-backend AI coding toolkit built around the `code-dispatcher` CLI: executor + Skills + Bundles.

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

`code-dispatcher` is a multi-backend task dispatcher that unifies `codex`, `claude`, and `gemini` AI coding tools. Core capabilities include:

- Multi-backend support: switch between or parallelize multiple AI backends via `--backend`
- Parallel execution: use `--parallel` for DAG-based concurrent task scheduling
- Session resume: use `--resume` to continue unfinished tasks after context resets
- Unified config: single-point configuration via `~/.code-dispatcher/.env` for all backends

Backend positioning (recommended only, can specify freely):

- `codex`: complex logic, bug fixes, optimization & refactoring
- `claude`: quick tasks, review, supplementary analysis
- `gemini`: frontend UI/UX prototyping, styling, and interaction polish

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
  <td><code>code&#8209;dispatcher</code></td>
  <td>Executor usage guide; unified 3 backends <code>codex/claude/gemini</code>; core mechanisms <code>--parallel</code> and <code>--resume</code></td>
  <td>Required</td>
</tr>
<tr>
  <td><code>dev</code></td>
  <td>Requirements clarification → plan → select backend → parallel execution (DAG scheduling) → verification</td>
  <td>Required</td>
</tr>
<tr>
  <td><code>wave</code></td>
  <td>Iterative wave-based parallel execution (host agent decomposes each wave → parallel dispatch → review results → next wave)</td>
  <td>Required</td>
</tr>
<tr>
  <td><code>code&#8209;council</code></td>
  <td>Multi-perspective parallel code review (2–3 AI reviewers in parallel + host agent final pass)</td>
  <td>Required</td>
</tr>
<tr>
  <td><a href="docs/github-issue-pr-flow.md"><code>github&#8209;issue&#8209;pr&#8209;flow</code></a></td>
  <td>Autonomous issue-to-PR delivery (decompose → implement → open PR → handle reviews → squash merge)</td>
  <td>Optional</td>
</tr>
<tr>
  <td><a href="docs/pr-review-reply.md"><code>pr&#8209;review&#8209;reply</code></a></td>
  <td>Autonomous bot-review triage on PRs (Gemini / CodeRabbit etc.) → verify → fix or rebut → reply in thread → resolve</td>
  <td>Optional</td>
</tr>
</table>

### Bundles

<table>
<tr>
  <th>Name</th>
  <th>Purpose</th>
  <th width="80">Dependency</th>
</tr>
<tr>
  <td><code>codex&#8209;review&#8209;loop</code></td>
  <td>Claude Code review loop bundle; composed of <code>commands/</code>, <code>hooks/</code>, <code>settings.json</code>; triggers Codex review on Stop</td>
  <td>No</td>
</tr>
<tr>
  <td><code>harness</code></td>
  <td>Claude Code long-running task bundle; provides state persistence, recovery, dependency scheduling, and SessionStart/Stop hooks</td>
  <td>No</td>
</tr>
</table>

> [!NOTE]
> Some bundles are not original to this repository — brought here for easier personal management and installation.
> - `harness` is ported from [`cexll/myclaude`](https://github.com/cexll/myclaude).
> - `codex-review-loop` is adapted from [`hamelsmu/claude-review-loop`](https://github.com/hamelsmu/claude-review-loop).
>
> See each bundle README for the specific upstream relationship, adaptation scope, and toolkit-specific packaging details.

## Installation

### Step 1: Code Dispatcher CLI Core Installation

The installation script is cross-platform: WSL2/Linux + macOS + Windows. By default, it downloads the current platform binary from GitHub Release `latest`:

```bash
python3 install.py
```

Optional parameters:

```bash
python3 install.py --install-dir ~/.code-dispatcher --force
python3 install.py --skip-dispatcher
```

The script installs the following:

- `~/.code-dispatcher/.env`: single runtime config source
- `~/.code-dispatcher/prompts/*-prompt.md`: default prompt templates for each backend (editable; leave empty to disable injection)
- `~/.code-dispatcher/bin/code-dispatcher` (`.exe` on Windows, etc.)

**Important**: Different platforms have different code agent shell environments, so they cannot all be assumed to follow the same execution pattern.

`install.py` outputs persistence commands for common shells (Windows: PowerShell, CMD, Git Bash; Non-Windows: Bash, Zsh, Fish):

```bash
Windows PowerShell:
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$HOME\.code-dispatcher\bin", "User")

Windows CMD:
setx PATH "%PATH%;%USERPROFILE%\.code-dispatcher\bin"

Git Bash (e.g. Claude Code on Windows):
echo 'export PATH="$HOME/.code-dispatcher/bin:$PATH"' >> ~/.bashrc

Bash (e.g. WSL/Linux):
echo 'export PATH="$PATH:$HOME/.code-dispatcher/bin"' >> ~/.bashrc

Zsh (e.g. macOS default):
echo 'export PATH="$PATH:$HOME/.code-dispatcher/bin"' >> ~/.zshrc

Fish:
echo 'set -gx PATH "$HOME/.code-dispatcher/bin" $PATH' >> ~/.config/fish/config.fish
```

### Step 2: Install Code Dispatcher Skill

Copy the `code-dispatcher` skill to the appropriate directory based on your target code agent tool's configuration. Global installation is recommended. Typical locations:

- General: `~/.agents/skills`
- Claude Code: `~/.claude/skills`
- Codex CLI: `~/.codex/skills`
- OpenCode: `~/.config/opencode/skills`
- Gemini CLI: `~/.gemini/skills`

### Step 3: Select Skills / Bundles

Refer to `docs/skills-and-bundles.en.md` for specific purposes, then select the functional modules you need:

**Skills**: Skills are cross-agent functional modules. The core is the `SKILL.md` definition file, and some also include `references/` documentation. When installing, copy the corresponding skill directory to the target agent's skills directory. Taking Claude as an example:
- Global: `~/.claude/skills/<skill-name>/`
- Project-level: `<path to your project>/.claude/skills/<skill-name>/`

The `dev` skill is recommended to be used with `templates/dev-skill-constraint.md` injected into user-level configuration. Typical usage:

**Bundles**: Bundles are Claude Code-specific packages, usually containing hooks and settings, and some also include commands or skill definitions. Installation steps vary by bundle:
- If containing `commands/`: copy to `.claude/commands/`
- If containing `hooks/`: copy to `.claude/hooks/`, and ensure scripts have execute permissions
- If containing `SKILL.md`: copy `SKILL.md` (and optional `references/`) to `.claude/skills/<bundle-name>/`
- Merge hooks configuration from `settings.json` into `.claude/settings.json` or `.claude/settings.local.json`

```text
# Explicit trigger
/dev "I want to implement X"

# Keyword trigger
use dispatcher --codex to fix the bug we just discussed
```


### Optional Configuration

Runtime parameters are unified in `~/.code-dispatcher/.env`. Configurable items include:

- Executor timeout parameters
- Executor parallel worker limit
- Executor log output settings
- Invoked backend model override (codex/gemini only)

For full field definitions, see: [docs/runtime-config.en.md](docs/runtime-config.en.md). If not configured, default parameters will be used.

## Development / Testing

### Prerequisites

- Go: `1.21` (`code-dispatcher/go.mod`)
- Python: `3.9+` (`install.py` uses `list[str]`)
- Bash: for local build scripts (`scripts/build-dist.sh`)

### Running Tests

```bash
cd code-dispatcher
go test ./...

# Verify local build
cd ..
bash scripts/build-dist.sh

# Verify install script syntax and template references
python3 -m py_compile install.py

# Non-polluting install regression test (using temp directory)
tmpdir="$(mktemp -d)"
python3 install.py --install-dir "$tmpdir/.code-dispatcher" --skip-dispatcher --force
python3 install.py --install-dir "$tmpdir/.code-dispatcher" --force
rm -rf "$tmpdir"
```
