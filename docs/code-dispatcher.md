# Code Dispatcher - 设计决策记录

Code Dispatcher 是一个多后端任务分发器，统一调度 `codex`、`claude` 两个 backend。

```
┌─────────────────────────────────────────────────────────────────┐
│                     Code Dispatcher 架构                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   用户请求 ──→ ┌─────────────┐ ──→ ┌─────────┐ ──→ 结果返回     │
│                │  Backend    │     │ Codex   │                  │
│                │  Router     │     │ Claude  │                  │
│                └─────────────┘     └─────────┘                  │
│                       │                                         │
│           ┌───────────┴───────────┐                             │
│           ▼                       ▼                             │
│   ┌──────────┐            ┌──────────┐                          │
│   │ Task A   │            │ Task B   │   ← --parallel           │
│   │ backend  │            │ backend  │      DAG 调度             │
│   │ :codex   │            │ :claude  │                          │
│   └──────────┘            └──────────┘                          │
│        │                       │                                │
│        └───────────┬───────────┘                                │
│                      │                                          │
│                      ▼                                          │
│              ┌─────────────┐                                    │
│              │  Session ID │  ← resume 断点续传                  │
│              └─────────────┘                                    │
│                                                                 │
│   三种模式: 单任务 / 并行执行 / 断点续传                         │
│   关键规则: 禁止默认终止 code-dispatcher 进程                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 设计决策记录

### 为什么用 Go？[DC-1]

需要原生并发能力——dispatcher 同时调度多个后端进程，goroutine + channel 是 Go 的核心优势，不需要额外依赖。

编译后单一二进制，分发到 WSL2/Linux/macOS 不需要运行时环境。跨平台进程组管理靠原生 syscall（`syscall.Kill(-pid, SIGTERM)`），Python/Node 做不到这种级别的进程控制。

### 为什么只留 codex/claude？[DC-2]

尝试过 ampcode、copilot，都删了。关键区别：codex/claude 是直接面向主要模型厂商 CLI 的 backend slot，ampcode/copilot 只是第三方工具包装。保留两个稳定后端，减少集成面和运行时差异。

| 后端 | 厂商 | 定位 | 最佳场景 |
|------|------|------|---------|
| **codex** | OpenAI | 复杂开发（默认） | 复杂逻辑、bug 修复、优化重构、大规模代码生成 |
| **claude** | Anthropic | 快速响应 | 快速修复、代码 review、补充分析、UI/UX 实现、文档编写、文案 polish |

### Backend 接口抽象 [DC-3]

`Backend` interface + `backendRegistry` map，标准抽象封装。删 ampcode/copilot 的经历说明后端会频繁变动，抽象层降低切换成本。每个后端只需实现 `Name()`、`Command()`、`BuildArgs()` 三个方法。

### HEREDOC 优先于 inline 参数 [DC-4]

AI 任务描述动辄含代码片段、文件路径、特殊字符，inline 参数有两个硬伤：

1. **`MAX_ARG_STRLEN` 128KB 内核硬限制**：`execve()` 单个参数上限 128KB（`PAGE_SIZE * 32`），无法通过 ulimit 修改。AI prompt 含上下文代码很容易超。
2. **Shell 展开会篡改内容**：`$VAR` 被变量替换、`` `cmd` `` 被命令替换、`!` 触发历史展开、双引号内嵌引号需要转义。传给后端的内容和用户写的不一样。

`<<'EOF'`（单引号 HEREDOC）走 stdin 路径，完全绕过 `execve()` 参数限制，零展开零转义。内容原样送达。

Go 侧默认做法是 `cmd.Stdin = strings.NewReader(content)`，不走参数拼接。

### --parallel 用 stdin 而不是配置文件 [DC-5]

这个 CLI 的调用方是 LLM agent，不是人类。agent 生成 stdin 文本流比生成文件再传路径简单得多。不需要类似 OMP agent swarm 那样支持 YAML/TOML 配置文件启动——那是面向人类操作者的模式。

stdin 无状态设计：dispatcher 不维护文件系统状态，不产生临时配置文件，和单任务模式接口风格一致。

### `---TASK---`/`---CONTENT---` 纯文本格式 [DC-6]

选分隔符格式而不是 YAML/JSON/TOML，核心原因是在 HEREDOC 里可靠：

- **JSON**：需要转义每个换行、引号、反斜杠，HEREDOC 里写 JSON 是噩梦
- **YAML**：对缩进极度敏感，HEREDOC 里混合 shell 缩进和 YAML 缩进容易出错；tab vs space 陷阱
- **分隔符格式**：只关心边界，内容区域可以粘贴任何东西（代码、特殊字符、任意缩进），不需要转义

Go 解析侧用 `strings.Split` 就够了，零外部依赖。类似模式在其他工具中也有先例：`git fast-import`、Debian control 文件、MIME boundary、markdown frontmatter。

### 单一全局超时而不是 per-backend [DC-7]

git 历史有 `792a3aa "unify timeout to seconds"`，说明之前可能有多套超时，统一了。

单一值 `CODE_DISPATCHER_TIMEOUT`（默认 7200 秒）覆盖所有后端。per-backend 超时增加配置碎片化，收益不大——后端自身有超时机制，dispatcher 只需要一个上限兜底。调用方在工具调用层按任务复杂度设外层 timeout，实际生效为先触发的那个。

建议分层：
- 简单任务：`600` 秒（10 分钟）
- 常规任务：`1800` 秒（30 分钟）
- 复杂 Codex 任务：`7200` 秒（2 小时）

### 退出码遵循 UNIX 惯例 [DC-8]

| 退出码 | 含义 | 来源 |
|--------|------|------|
| 0 | 全部成功 | UNIX 惯例 |
| 1 | 一般错误 | UNIX 惯例 |
| 124 | 超时 | coreutils `timeout` 命令标准 |
| 127 | 后端未找到 | shell 标准 |
| 130 | 用户中断（SIGINT） | shell 标志（128+2） |

Shell 脚本能直接 `$?` 判断，不需要额外错误码映射。启动时向 stderr 输出后端名、完整命令、PID、日志路径，方便调试。

### 禁止默认终止进程 + 优雅终止 [DC-9]

git 历史有 `a6ec7c9 "fix: gracefully terminate backend process trees on cancel"`。

AI 后端需要保存 session/刷新日志，强杀丢数据。信号先发到进程组（`-pid`），确保子进程也收到；dispatcher 先发送 SIGTERM，等待配置延迟后仍未退出才升级 SIGKILL。

唯一允许终止的三种情况写死在文档里，防止 agent 乱杀。禁止基于名称的全局清理（`pkill -x codex/claude`），只清理目标 dispatcher 进程的子进程。

### --full-output 只在 parallel 模式 [DC-10]

两条代码路径完全不同：

- **单任务**（`main.go:437-448`）：`fmt.Println(result.Message)` 直接透传后端 stdout。一个任务 = 一条流，没有"汇总/展开"的区分。
- **并行**（`main.go:270-289`）：多任务结果收集后走 `generateFinalOutputWithMode()`，默认做结构化提取（Summary 模式），`--full-output` 切回完整原始消息用于调试。

单任务不需要 flag，因为调用方直接拿到后端输出。并行需要，因为多份输出混在一起需要两种视图。

### Summary 报告的结构化提取是 best-effort [DC-11]

**当前实现只从 dispatcher 专用报告块提取结构化字段。** 这是针对 CLI headless 后端的显式 contract：不改变 `stream-json` 调用架构，也不依赖各后端的 schema/json-only 输出模式。

优先通道：后端默认 prompt 要求最终输出末尾追加：

```md
---CODE-DISPATCHER-REPORT---
Coverage: <number>% | NONE
Files: <comma-separated relative paths> | NONE
Tests: <passed> passed, <failed> failed | NONE
Summary: <one sentence>
---END-CODE-DISPATCHER-REPORT---
```

`NONE` 表示该字段不适用或未观测到，dispatcher 不渲染空字段。为避免 DAG/resume 输出引用上游报告块，解析时取最后一个 `---CODE-DISPATCHER-REPORT---` 块。

没有报告块时，Summary 模式不从自由文本猜测 Coverage、Files、Tests 或 Did；报告只显示任务状态和日志路径。任务本身的成功/失败仍由后端退出码和执行错误决定。

`utils.go` 里仍保留部分 best-effort 诊断提取，用于失败和覆盖率不足场景的辅助上下文：

| 提取目标 | 匹配方式 | 脆弱点 |
|----------|----------|--------|
| Coverage Gap | 找 `uncovered/not covered/missing coverage/branch not taken/function 0%` 等线索 | 覆盖率工具差异大 |
| Error Detail | 找 `error/fail/exception/assert/expected/timeout/not found/cannot/undefined` 等错误线索，fallback 取尾部输出 | 可能混入无关日志 |

设计思路：报告块是结构化字段的唯一来源。提取不到不报错，但不会用正则伪造结构化事实。报告的最低价值是“任务 ID + 成功/失败 + Log”，不依赖结构化提取。

### DAG 调度 [DC-12]

任务天然有依赖关系（先分析→再设计→再实现），但同层无依赖可并行。Kahn 算法拓扑排序，层内并发、层间串行。依赖失败时下游自动跳过。`CODE_DISPATCHER_MAX_PARALLEL_WORKERS` 从 `.env` 读取：未设置、无效或负数表示不限制；正数启用限流，并在超过 `100` 时封顶。

### Session/Resume 是为了复用上下文 [DC-13]

SESSION_ID 的核心目的是**复用对话上下文**，不是处理超时/中断（那是副作用）。长任务分多轮人机交互，resume 让后续指令在同一个对话里继续，不丢上下文。支持的后端都支持 resume，parallel 模式里也能混合 resume。

### 后端审批全跳过 [DC-14]

dispatcher 定位是自动化执行器，不是交互式工具。codex/claude 的 approval 提示会阻塞非交互进程，所以硬编码 bypass flags（Codex `--dangerously-bypass-approvals-and-sandbox`，Claude `--dangerously-skip-permissions`），不给开关。

### --cleanup 孤立日志清理 [DC-15]

每次执行在 tmpdir 创建日志文件，长期跑会累积。cleanup 检查 PID 是否还活着，只删孤儿日志。有 PID 复用防护（比较文件 mtime vs 进程启动时间）和 symlink 攻击防护（确保文件在 tmpdir 内且非符号链接）。

---

## 参考：命令行速查

`--backend` 在所有模式中均为必需参数。

### 单任务模式（new）

**HEREDOC（推荐，多行或复杂内容）**：

```bash
code-dispatcher --backend codex - [working_dir] <<'EOF'
<task content here>
EOF
```

**单行（简单任务）**：

```bash
code-dispatcher --backend codex "simple task" [working_dir]
```

参数说明：
- `--backend <codex|claude>`：选择后端（必需）
- `<task>`：任务描述，支持 inline 文本或 `-`（从 stdin 读取）。支持 `@file` 引用（后端原生功能，dispatcher 不处理）
- `[working_dir]`：工作目录（可选，默认 `.`）

### 断点续传（resume）

`resume` 是 positional 命令，不是 `--` flag。

```bash
code-dispatcher --backend codex resume <session_id> - <<'EOF'
<follow-up task>
EOF
```

```bash
code-dispatcher --backend claude resume <session_id> "follow-up task"
```

- Session ID 来自 dispatcher 输出的 `SESSION_ID`（UUID 格式）
- resume 模式不要追加 `working_dir`；如需不同目录，应开启新会话
- claude 会话按项目目录（cwd）存储：串行 resume 需在原任务 workdir 下运行 dispatcher；parallel 模式的 resume 任务应在 `workdir:` 中携带原任务目录，dispatcher 会用它启动 claude

### 并行模式（--parallel）

`--parallel` 从 stdin 读取任务配置，不接受文件参数。仅允许 `--backend` 和 `--full-output` 两个额外 flag。

```bash
code-dispatcher --parallel --backend <backend> <<'EOF'
---TASK---
id: task_id
backend: <backend>
workdir: /path
dependencies: dep1, dep2
---CONTENT---
task content
EOF
```

**元数据字段**：

| 字段 | 必需 | 说明 |
|------|------|------|
| `id` | 是 | 唯一任务标识 |
| `backend` | 否 | 覆盖全局 `--backend`，仅对该任务生效 |
| `workdir` | 否 | 工作目录，默认 `.` |
| `dependencies` | 否 | 逗号分隔的依赖任务 ID |
| `session_id` | 否 | resume 已有会话；设置后自动切换为 resume 模式 |

**验证规则**：`---CONTENT---` 必需分隔符；未知元数据键立即报错；重复或缺失 `id` 报错；`workdir: -` 无效；出现 `session_id` 时不能为空。

#### DAG 示例

```bash
code-dispatcher --parallel --backend codex <<'EOF'
---TASK---
id: task1
workdir: /path/to/dir
---CONTENT---
analyze code structure

---TASK---
id: task2
backend: claude
dependencies: task1
---CONTENT---
design architecture based on task1 analysis

---TASK---
id: task3
backend: codex
dependencies: task2
---CONTENT---
generate implementation code
EOF
```

#### 输出模式

**Summary 模式（默认）**：结构化报告（see DC-11 for extraction details）。

**Full 模式（`--full-output`）**：完整任务消息，调试失败时使用。仅 parallel 模式有效（see DC-10）。

### 工具命令

- `code-dispatcher --help`：打印 CLI 用法、支持的后端、常见错误和退出码
- `code-dispatcher --cleanup`：清理孤立日志文件（see DC-15）

### 返回格式

#### 单任务模式

```text
Agent response text here...

---
SESSION_ID: 019a7247-ac9d-71f3-89e2-a823dbd8fd14
```

#### Parallel Summary 模式

```text
=== Execution Report ===
3 tasks | 3 passed | 0 failed

## Task Results

### task-a ✓ 92%
Did: implemented requested change
Files: code-dispatcher/main.go
Tests: 12 passed
Log: /tmp/code-dispatcher-123-task-a.log

## Summary
- 3/3 completed successfully
```

失败任务会显示 `FAILED`、`Exit code`、`Error`、`Detail` 和 `Log`；覆盖率不足时会显示 `below <target>%` 和 `Gap`。

#### Parallel Full 模式

```text
=== Parallel Execution Summary ===
Total: 1 | Success: 1 | Failed: 0

--- Task: task-a ---
Status: SUCCESS
Coverage: 92%
Session: 019a7247-ac9d-71f3-89e2-a823dbd8fd14
Log: /tmp/code-dispatcher-123-task-a.log

<raw backend message>
```

### 后端选择指南

用户已明确指定后端时，始终遵循用户指令。以下仅适用于未指定的场景。

**Codex**（OpenAI）：深度代码理解、大规模重构、算法优化

**Claude**（Anthropic）：快速功能实现、UI/UX 实现、技术文档、prompt 编写、文案 polish

### 紧急停止

```bash
# 优雅中断
pkill -INT -f '(^|/)code-dispatcher( |$)'

# 升级
pkill -TERM -f '(^|/)code-dispatcher( |$)'
sleep 2
pkill -KILL -f '(^|/)code-dispatcher( |$)'
```
