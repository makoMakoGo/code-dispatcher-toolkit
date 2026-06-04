<div align="center">

<h1>CODE-DISPATCHER TOOLKIT</h1>
<p><strong>Multi-Backend AI Coding Toolkit</strong></p>
<p>Dispatch tasks across Codex and Claude with<br>reusable Skills and workflow tooling.</p>

<p>
  <strong>中文</strong> | <a href="README.en.md">English</a>
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

基于 `code-dispatcher` CLI 构建的多后端 AI 编码工具集：执行器 + Skills。

## 为什么叫 Dispatcher

因为这个词的含义很符合这个工具的核心功能：

<div align="center">
<strong>接收任务</strong> &nbsp;→&nbsp;
<strong>选后端</strong> &nbsp;→&nbsp;
<strong>构建参数</strong> &nbsp;→&nbsp;
<strong>分发执行</strong> &nbsp;→&nbsp;
<strong>收集结果</strong>
</div>

## 组件

### Dispatcher CLI

`code-dispatcher` 是一个多后端任务分发器，统一调度 `codex`、`claude` 两个 AI 编码工具。核心能力包括：

- 多后端支持：通过 `--backend` 自由切换或并行调用多个 AI 后端
- 并行执行：基于 DAG 调度同时运行多个独立任务
- 会话恢复：上下文重置后继续执行未完成的任务
- 统一配置：单点配置 `~/.code-dispatcher/.env` 管理所有后端参数

后端定位（仅推荐，可自由指定）：

- `codex`：复杂逻辑、bug 修复、优化重构
- `claude`：快速任务、review、补充分析、UI/UX 实现、文档和文案 polish

> [!NOTE]
> 工具 `code-dispatcher` 核心思路基于 [`cexll/myclaude`](https://github.com/cexll/myclaude) 的 `codeagent wrapper`，经大量重构。

### Skills

注：依赖指是否依赖 code-dispatcher CLI 进行调度和执行。

<table>
<tr>
  <th>名称</th>
  <th>用途</th>
  <th width="60">依赖</th>
</tr>
<tr>
  <td><a href="docs/code-dispatcher.md"><code>code&#8209;dispatcher</code></a></td>
  <td>执行器使用说明；统一后端 <code>codex/claude</code>；核心机制并行执行和会话恢复</td>
  <td>必需</td>
</tr>
<tr>
  <td><a href="docs/dev.md"><code>dev</code></a></td>
  <td>需求澄清 → 计划 → 选择后端 → 并行执行（DAG 调度） → 验证</td>
  <td>必需</td>
</tr>
<tr>
  <td><a href="docs/pr-review-reply.md"><code>pr&#8209;review&#8209;reply</code></a></td>
  <td>自主处理 PR 上的 bot review（Gemini Code Assist / CodeRabbit 等）→ 验证 → 修复或反驳 → 回复线程 → resolve</td>
  <td>可选</td>
</tr>
</table>

## 安装

### Step 1: Code Dispatcher CLI 相关核心安装

安装脚本支持 Linux x86_64、Windows x86_64、macOS Apple Silicon。默认会从 GitHub Release 的 `latest` 标签下载当前平台二进制：

```bash
python3 install.py
```

可选参数：

```bash
python3 install.py --install-dir ~/.code-dispatcher --force
python3 install.py --skip-dispatcher
```

脚本会添加如下东西：

- `~/.code-dispatcher/.env`：运行时唯一配置源
- `~/.code-dispatcher/prompts/*-prompt.md`：每个后端的默认 prompt 模板（可编辑，置空则禁用注入）
- `~/.code-dispatcher/bin/code-dispatcher`（Windows 上是 `.exe`，以此类推）

特别需要注意的是：不同平台下 code agent 的 shell 环境并不完全一致，不能默认都按同一套假设执行。

`install.py` 会直接输出常见 shell 的持久化设置命令（Windows：PowerShell、CMD、Git Bash；非 Windows：Bash、Zsh、Fish）。下面是默认安装路径示例：

```bash
Windows PowerShell:
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$HOME\.code-dispatcher\bin", "User")

Windows CMD:
setx PATH "%PATH%;%USERPROFILE%\.code-dispatcher\bin"

Git Bash (e.g. Claude Code on Windows):
echo 'export PATH="$HOME/.code-dispatcher/bin:$PATH"' >> ~/.bashrc

Bash (e.g. WSL/Linux):
echo 'export PATH="$PATH:$HOME/.code-dispatcher/bin"' >> ~/.bashrc

Zsh (e.g. macOS 默认):
echo 'export PATH="$PATH:$HOME/.code-dispatcher/bin"' >> ~/.zshrc

Fish:
echo 'set -gx PATH "$HOME/.code-dispatcher/bin" $PATH' >> ~/.config/fish/config.fish
```

### Step 2: 安装 Code Dispatcher Skill 

根据目标 code agent 工具的配置目录，复制 `code-dispatcher` 这个 skill 到相关目录。建议全局安装，下面是一些典型位置：

- General： `~/.agents/skills`
- Claude Code : `~/.claude/skills`
- Codex CLI : `~/.codex/skills`
- OpenCode : `~/.config/opencode/skills`

### Step 3: 挑选 Skills

参考各技能的独立文档了解具体用途，然后挑选你需要的功能模块：

**Skills**：Skill 是跨 agent 通用的功能模块，核心是 `SKILL.md` 定义文件，部分还包含 `references/` 参考文档。安装时将对应的 skill 目录复制到目标 agent 的 skills 目录，以 Claude 为例：
  - 全局：`~/.claude/skills/<skill-name>/`
  - 项目级：`<path to your project>/.claude/skills/<skill-name>/`

其中 `dev` skill 是 Claude Code manual-only skill，通过 `/dev` 显式调用。典型用法：

```text
# 显式触发
/dev "我想实现一个xxx"

# 特定关键词触发
use dispatcher with codex to fix the bug we just discussed
```


### 可选配置项

运行时参数统一放在 `~/.code-dispatcher/.env`。可配置项包含：

- 执行器超时相关参数
- 执行器并行 worker 上限
- 执行器日志输出设置
- 被调用后端模型覆盖（仅 codex）

完整字段含义请先看：[docs/runtime-config.md](docs/runtime-config.md)，未配置时按默认参数运行。

## 开发/测试

### 先决环境

- Go：`1.21`（`code-dispatcher/go.mod`）
- Python：`3.9+`（`install.py` 使用 `list[str]`）
- Bash：用于本地构建脚本（`scripts/build-dist.sh`）

### 运行验证

```bash
cd code-dispatcher
go test ./...

# 验证本地构建
cd ..
bash scripts/build-dist.sh

# 验证安装脚本语法与模板引用
python3 -m py_compile install.py

# 不污染真实环境的安装回归（使用临时目录）
tmpdir="$(mktemp -d)"
python3 install.py --install-dir "$tmpdir/.code-dispatcher" --skip-dispatcher --force
python3 install.py --install-dir "$tmpdir/.code-dispatcher" --force
rm -rf "$tmpdir"
```

## 社区与贡献

本项目是独立维护的非商业开源项目。欢迎通过 issue 反馈 bug、文档问题和使用场景，也欢迎提交聚焦明确的小型 PR。

提交变更前，请先运行上面的相关验证命令。
