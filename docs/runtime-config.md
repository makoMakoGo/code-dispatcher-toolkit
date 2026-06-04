# code-dispatcher 运行时配置

该文档是运行时行为的唯一可信来源，覆盖以下内容：

- 超时行为
- 并行执行与 worker 限制
- 配置加载模型

## 1）单一配置来源

所有运行时选项统一从以下文件读取：

```text
~/.code-dispatcher/.env
```

`code-dispatcher` 不再从 shell 环境变量读取这些控制项。

只有 `CODE_DISPATCHER_*` 是 dispatcher 控制项。`.env` 中其他键值会注入给后端进程；若同名键也存在于外层 shell，`.env` 值优先。

## 2）后端审批/跳过审批设置

所有后端默认都使用“绕过审批”模式（不可开关）：

- `codex`：`--dangerously-bypass-approvals-and-sandbox`
- `claude`：`--dangerously-skip-permissions`

## 3）`.env` 中的运行时字段

- `CODE_DISPATCHER_TIMEOUT`
  - 默认：`7200`（秒，2 小时）
  - 单位：秒

- `CODE_DISPATCHER_MAX_PARALLEL_WORKERS`
  - 默认：不限制（`0`）
  - 建议：`8`
  - 调度器硬上限：`100`

- `CODE_DISPATCHER_ASCII_MODE`
  - `1` / `true` / `yes` / `on`：使用 ASCII 状态码（`PASS/WARN/FAIL`）
  - `0` / `false` / `no` / `off` 或未知值：使用 Unicode 状态符号

- `CODE_DISPATCHER_LOGGER_CLOSE_TIMEOUT_MS`
  - 默认：`5000`
  - `<= 0`：表示无限等待

### 后端模型覆盖

- `CODE_DISPATCHER_CODEX_MODEL`
  - 可选；若设置，则向 codex CLI 透传 `-m <value>`
  - 示例：`gpt-5.4`

- Claude 不支持通过 dispatcher 覆盖模型。

## 4）提示词文件

提示词文件来自：

```text
~/.code-dispatcher/prompts/<backend>-prompt.md
```

支持的后端：`codex`、`claude`。

安装器不会覆盖已存在的提示词文件，除非使用 `python3 install.py --force`。升级默认 prompt contract 后，已有安装需要手动同步这些文件或使用 `--force` 重新安装。

## 5）超时分层（重要）

通常存在两层超时机制：

- 外层调用超时（例如工具调用超时）
- dispatcher 超时（来自 `.env` 的 `CODE_DISPATCHER_TIMEOUT`）

实际生效超时为率先触发的那一个。

## 6）编辑配置

直接编辑 `.env` 文件即可：

```bash
${EDITOR:-vi} ~/.code-dispatcher/.env
```
