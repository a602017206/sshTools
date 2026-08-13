# AI Copilot（SQL / Shell）设计规格

> 状态：设计已确认，实现计划已写  
> 关联设计摘要：`docs/designs/2026-08-13-ai-copilot-sql-shell.md`  
> 实现计划：`docs/plans/2026-08-13-ai-copilot-sql-shell.md`  
> 日期：2026-08-13

## 1. 背景

sshTools 已具备 SSH 终端、`ExecuteCommand`、JDBC 查询、表/列元数据和统一运维工作区。用户希望用自然语言生成 SQL 和 Shell 命令，并在当前会话里确认后执行。

现有 Java `jdbc-agent` 是数据库 sidecar，不是 LLM Agent。本功能在 Go 服务层新增 Copilot，复用已有执行面，不另起 Agent 进程。

## 2. 目标与非目标

### 目标

- 数据库会话：自然语言生成方言正确的 SQL，填入当前查询编辑器，用户点「执行」后走现有 `ExecuteDatabaseQuery`。
- SSH 会话：自然语言生成命令/脚本，填入当前终端缓冲（不自动换行），用户点「执行」后经 `SendSSHData` 发送到当前 PTY。
- 生成阶段允许只读工具调用，以读取真实 schema 或主机环境。
- 设置使用 OpenAI 兼容接口；模型名称由用户按官方文档手填，应用不做探测或映射。
- 预留本地模型（Ollama）的 Provider 接口，第一版不接通。

### 非目标（第一版）

- 不自动多轮执行用户产物。
- 不做行内补全、`Cmd+K` 命令面板。
- 对话不落盘；关闭会话即丢历史。
- 不拉取「可用模型列表」。
- 不实际接通 Ollama。
- 不修改 Java `jdbc-agent`。
- 不把 AI 塞进 SSH「文件 / 性能」工具坞。
- 用户 Shell 不改走另开会话执行。
- 不做 Flutter 客户端、不做 MCP 对外暴露。

## 3. 关键决策

| 项 | 选择 |
| --- | --- |
| 入口 | SQL 工作区 + SSH 终端，共用一套 Copilot 内核 |
| 交互 | 跟随当前标签的侧边对话；行内补全以后再做 |
| 执行 | 先填入，再点「执行」；危险写操作二次确认 |
| 模型接入 | 用户自备 OpenAI 兼容 API；设置预留本地模型 |
| 模型名称 | 纯文本手填，原样传给接口，不猜测、不映射 |
| 生成策略 | 只读工具循环，最多 4 轮 |
| 面板位置 | 独立 AI 侧栏，不占用文件/性能 dock |

## 4. 架构

第一版做成桌面应用内的 Copilot 服务。

```text
前端 AIPanel（独立侧栏，跟随当前标签）
        │  CopilotChat / CopilotClassify
        ▼
Go CopilotService（internal/service/copilot/）
        ├─ Provider 接口
        │     └─ OpenAI Compatible（首版）
        │     └─ Ollama（接口预留，未实现）
        ├─ 只读工具
        │     ├─ 数据库：ListDatabases / ListTables / GetTableSchema
        │     └─ SSH：白名单探测，经 ExecuteCommand 另开会话
        └─ 危险分类器（规则优先于模型）
                │
                ▼ 用户确认后
        现有执行面：ExecuteDatabaseQuery / SendSSHData
```

职责边界：

- **`AIPanel`**：对话 UI、填入、执行按钮、确认框。对话按 `sessionID` 隔离。
- **`CopilotService`**：注入上下文、跑工具循环、抽出产物、标记危险。`app.go` 导出 `CopilotChat`、`CopilotClassify`。填入在前端完成；执行复用已有 bindings。
- **`Provider`**：Chat Completions + tool call。API Key 走 `CredentialStore`（固定键 `copilot:api_key`），`config.json` 只存 `copilot_provider`、`copilot_base_url`、`copilot_model`。
- **只读工具**：全部包装现有服务。SSH 探测不进用户 PTY。
- **危险分类器**：本地规则，模型 `destructive` 字段只作提示；规则说危险则以规则为准。

密码、私钥、DSN、API Key 不进入 prompt 或工具结果。

## 5. 设置

全局设置新增「AI Copilot」分段，三项手填：

- **Base URL**：例如 `https://api.openai.com/v1`、`https://api.deepseek.com/v1`
- **模型名称**：官方文档中的模型 id，例如 `gpt-4o`、`deepseek-chat`、`qwen-plus`
- **API Key**：写入 `CredentialStore`，不进 `config.json`

请求时把用户填写的 `model` 原样放入请求体。填错由对方 API 报错，对话展示可读原因。第一版不拉模型列表。本地 Ollama 以后同样手填 `model`（如 `qwen2.5:7b`），只换 `base_url` 与 `provider`。

未配置 API Key 或 Base URL 时，侧栏提示去设置，不发请求。

## 6. 数据流

1. **发话**：侧栏提交当前 `sessionID`、模式（`ssh` / `database`）、用户消息、会话内历史。可选附带：编辑器已有 SQL，或终端最近输出（截断并脱敏）。
2. **组装上下文**：连接名、主机、库类型/当前库，或 SSH 用户名与当前工作目录。不含密钥。
3. **只读工具循环（最多 4 轮）**  
   - 数据库：`list_databases` → `list_tables` → `get_table_schema`  
   - SSH 白名单：`uname`、`pwd`、`df -h`、`cat /etc/os-release`  
   - 超轮次或工具失败：用已有信息继续生成，并在对话说明未读到完整环境。
4. **抽出产物**：模型须返回 `{ type: "sql"|"shell", content, summary, destructive: bool }`。抽不出则当普通回复，不出现「填入 / 执行」。
5. **填入**：SQL 写入当前数据库编辑器；Shell 写入当前终端输入且不换行。
6. **执行**：  
   - 先走 `CopilotClassify`；危险则弹现有 `ConfirmDialog`。  
   - 取消确认则不发送。  
   - SQL → `ExecuteDatabaseQuery`，结果表仍在数据库面板，对话只留摘要（行数/报错）。  
   - Shell → `SendSSHData` 发送命令 + 换行，输出留在终端；对话不截获 PTY 流。
7. **改错**：执行失败后，用户可将报错发回对话，再生成一版。
8. **会话**：历史按 `sessionID` 存在内存；关会话取消进行中的 LLM 请求并丢历史。切换标签只换当前对话。

同一 `sessionID` 同时只跑一轮生成。

## 7. 错误处理与安全

### 模型与配置

- 缺 Key / Base URL：提示去设置。
- 超时、4xx/5xx、余额不足、未知模型名：对话显示可读原因，保留已生成草稿。
- 非 JSON 回复：当普通文本，不出现执行按钮。

### 工具

- SSH 探测不在白名单：拒绝，对话记「工具被拒绝」，不执行。
- 探测超时或失败：跳过该工具，继续生成。
- 数据库未连接或 JDBC agent 不可用：复用现有错误码（如 `AGENT_UNAVAILABLE`），不把 SQL 当已执行。

### 执行门禁

- 「填入」只改缓冲，不运行。
- 「执行」前本地再判危险模式。覆盖至少：`DROP`、`DELETE`、`TRUNCATE`、无 WHERE 的 `UPDATE`、`rm` / `rm -rf`、`mkfs`、`dd`、`shutdown`、`reboot`、`kill -9`、`chmod 777`、`> /dev/sd`。
- SQL 执行沿用现有 30s 超时。
- 用户命令只写入当前 PTY，不另开会话跑。

### 隐私与截断

- Prompt / 工具结果 / 历史不含密码、私钥、DSN、API Key。
- Schema 按需通过工具读取；单次工具结果截断到 8000 字符，避免整库 dump。
- 回传给模型的终端输出过滤 `password=`、密钥块等。

## 8. 前端落点

- 新组件 `AIPanel`：可折叠独立侧栏，默认收起。SSH 与数据库模式共用同一入口。
- 从顶栏或右侧 rail 打开；展开时作为独立列，不占用、不改写 `SessionToolDock` 的「文件 | 性能」分段。
- 无当前会话时：空状态「先连接主机或数据库」。
- 设置入口：`GlobalSettingsDialog` 增加 AI 分段。
- 危险确认：复用 `ConfirmDialog`。
- `DELETE` / `DROP` / `UPDATE` 等写操作一律先确认，即使带 WHERE。

## 9. 测试

### 自动化（Go，假 HTTP Provider）

- 数据库工具循环只调 schema 类工具。
- SSH 非法探测被拒绝且不进 PTY。
- 合法结构化产物才可填入/执行；非 JSON 无执行按钮语义。
- 危险分类规则优先于模型；取消确认不执行。
- 发给 Provider 的消息不含密码、DSN、API Key。
- `config.json` 不含 API Key。
- 模型名原样上传。
- 同一 `sessionID` 并发只跑一轮；关会话取消请求。

### 手工回归

- 数据库：自然语言 → 填入编辑器 → 执行 → 结果在数据库面板。
- SSH：自然语言 → 填入终端（不换行）→ 执行才发送。
- 写操作弹确认；取消后无副作用。
- 未配 Key、错误模型名、JDBC agent 不可用时的提示。

## 10. 验收

- 在已连接的 MySQL/PostgreSQL/Oracle/金仓会话中，能根据自然语言生成可执行 SQL 并填入编辑器。
- 在已连接的 SSH 会话中，能生成命令并填入终端，点执行后出现在 PTY。
- 危险语句必须确认；只读探测不得进入用户终端。
- 用户按官方文档填写 Base URL 与模型名即可切换供应商，无需改代码。
