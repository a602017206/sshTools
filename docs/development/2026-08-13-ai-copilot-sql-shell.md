# 开发记录：AI Copilot（SQL / Shell）

## 实现内容

按 `docs/plans/2026-08-13-ai-copilot-sql-shell.md` 的 8 个任务推进，Task 1–7 为代码，Task 8 为本文档与变更记录。落点与规格第 4 节一致：在 Go 服务层新增 Copilot 内核，前端新增独立 AI 侧栏，执行复用既有 `ExecuteDatabaseQuery` 与 `SendSSHData`，不修改 Java `jdbc-agent`，不占用「文件 / 性能」dock。

### 后端 `internal/service/copilot/`

- `types.go`：`Message` / `ToolCall` / `Artifact` / `ChatRequest` / `ChatResponse` / `Result` 等结构体；`ChatRequest` 不含密码、私钥、DSN、API Key 字段。
- `classify.go`：本地规则危险分类，覆盖规格第 7 节名单（`DROP`/`DELETE`/`TRUNCATE`/无 WHERE 的 `UPDATE`、`rm`/`rm -rf`/`mkfs`/`dd`/`shutdown`/`reboot`/`kill -9`/`chmod 777`/`> /dev/sd`）。规则命中即危险，模型 `destructive` 字段仅作提示。
- `artifact.go`：从模型回复中取第一个 JSON 对象（可包在 markdown 代码块里），校验 `type` 与 `content`，再用 `Classify` 覆盖 `destructive`。
- `redact.go`：替换 `password=...` 与 PEM 私钥块。
- `probe.go`：SSH 探测白名单，trim 后整串精确匹配 `uname`、`pwd`、`df -h`、`cat /etc/os-release`，链式命令直接拒绝。
- `provider.go` / `provider_openai.go`：`Provider` 接口与 OpenAI 兼容实现；POST `{baseURL}/chat/completions`，body 中的 `model` 原样等于传入参数，不做映射、不拉模型列表。401 等错误返回可读文案且不含 API Key。
- `tools.go`：工具定义与分发，工具名固定 `list_databases` / `list_tables` / `get_table_schema` / `ssh_probe`；工具结果截断到 `MaxToolResultChars = 8000`。
- `service.go`：`Service.Chat` 跑只读工具循环，最多 `MaxToolRounds = 4` 轮；按 `sessionID` 用 mutex map 保证同一会话同时只跑一轮，重复请求返回「已有生成进行中」；`Cancel(sessionID)` 取消该会话的 context。database 模式不注册 `ssh_probe`，ssh 模式不注册 schema 工具。
- `config.go`：`ValidateConfig(baseURL, apiKey)`，缺 Key/Base URL 返回中文「请先在设置中填写 Base URL 和 API Key」。

### 后端接线 `app.go` / `internal/config` / `internal/service`

- `app.go`：`App` 增加 `credentialStore` 与 `copilotService`；导出 `CopilotChat` / `CopilotClassify` / `HasCopilotAPIKey` / `SetCopilotAPIKey` / `ClearCopilotAPIKey`。`CopilotChat` 在每次请求时读取最新 settings 与 Key 现建 Provider，避免改设置不生效；Chat 超时 60s。`CloseSSH` / `CloseDatabase` 调 `copilotService.Cancel(sessionID)`。填充 `ChatRequest` 的 Host/User/DBType/Database/WorkingDir 时从已有 session/config 读取，不拷密码。
- `internal/config/config.go`：`AppSettings` 增加 `CopilotProvider` / `CopilotBaseURL` / `CopilotModel`（json：`copilot_provider` / `copilot_base_url` / `copilot_model`）；`DefaultSettings` 中 `CopilotProvider` 为 `openai_compatible`，URL 与模型为空；`UpdateSettings` 忽略 `copilot_api_key`，API Key 不进 `config.json`。
- `internal/service/session_service.go`：新增 `ExecuteCommand` 包装，委托 `sessionManager.ExecuteCommand`，供 Copilot SSH 探测经另开会话执行，不进用户 PTY。

### 前端

- `frontend/src/lib/copilotApply.js`：填入事件名与执行辅助。`COPILOT_APPLY_SQL` / `applySqlEvent` / `shellExecutePayload`（去尾换行再补单个 `\n`）/ `peekSqlEvent` 等。
- `frontend/src/stores/copilot.js`：侧栏开关 `{ open, width }` 与 `toggle()`；对话数组按 `sessionID` 存内存，关会话删除。
- `frontend/src/components/AIPanel.svelte`：独立可折叠侧栏。无 session 显示「先连接主机或数据库」；缺 Key 提示去设置；输入框发送 `CopilotChat` 并带上该 session 的历史（role/content，不含 Key）；回复若有 `artifact` 显示摘要、代码块与「填入 / 执行」按钮；生成中禁用发送。
- `frontend/src/components/GlobalSettingsDialog.svelte`：新增「AI Copilot」分段，三个手填框（Base URL、模型名称、API Key password input）+「清除密钥」按钮；说明文案提示按服务商官方文档填写模型名。
- `frontend/src/settings/appearance.js`：默认设置增加 `copilot_provider` / `copilot_base_url` / `copilot_model`，不放 API Key。
- `frontend/src/App.svelte`：顶栏增加 AI 按钮（`title="AI Copilot"`），打开可折叠列放在主舞台与 SSH 工具坞之间，数据库模式也显示，默认收起；`persistAppSettings` / `handleSaveGlobalSettings` 处理三个 copilot 字段，非空 Key 调 `SetCopilotAPIKey`，不把 Key 写入 `UpdateSettings`。
- `frontend/src/components/DatabasePanel.svelte` / `DatabaseTablePanel.svelte`：监听 `COPILOT_APPLY_SQL`，`sessionId` 匹配才赋值 `query`。
- `frontend/src/components/TerminalPanel.svelte`：新增 `insertCopilotText(sessionId, text)`，对该 session 调 `SendSSHData`（填入不换行）。

### 执行语义（第一版锁死）

- SQL「填入」：派发 `copilot:apply-sql`，查询面板设置 `query`。
- SQL「执行」：先 `CopilotClassify('sql', sql)`；危险则 `ConfirmDialog`；确认后用当前编辑器 `query` 调 `ExecuteDatabaseQuery`。无打开的查询面板时先填入再执行同一内容。
- Shell「填入」：`SendSSHData(sessionId, content)`，不加换行。
- Shell「执行」：分类确认后 `SendSSHData(sessionId, shellExecutePayload(content))`，输出留在终端。

## 修改文件

后端：

- `internal/service/copilot/types.go`、`classify.go`、`classify_test.go`、`artifact.go`、`artifact_test.go`、`redact.go`、`redact_test.go`、`probe.go`、`probe_test.go`、`provider.go`、`provider_openai.go`、`provider_openai_test.go`、`tools.go`、`service.go`、`service_test.go`、`config.go`、`config_test.go`（新增）
- `internal/config/config.go`、`internal/config/config_test.go`（修改）
- `internal/service/session_service.go`（修改）
- `app.go`（修改）

前端：

- `frontend/src/lib/copilotApply.js`、`frontend/test/copilotApply.test.js`、`frontend/src/stores/copilot.js`、`frontend/src/components/AIPanel.svelte`（新增）
- `frontend/src/settings/appearance.js`、`frontend/src/components/GlobalSettingsDialog.svelte`、`frontend/src/App.svelte`、`frontend/src/components/DatabasePanel.svelte`、`frontend/src/components/DatabaseTablePanel.svelte`、`frontend/src/components/TerminalPanel.svelte`（修改）

文档：

- `docs/superpowers/specs/2026-08-13-ai-copilot-sql-shell-design.md`（状态改为已落地）
- `docs/changes/features/2026-08-13-ai-copilot-sql-shell.md`（新增变更记录）
- `docs/development/2026-08-13-ai-copilot-sql-shell.md`（本文档）

## 验证

```bash
go test ./internal/service/copilot ./internal/config -count=1
cd frontend && node --test test/copilotApply.test.js
cd frontend && npm run build
```

实测结果（2026-08-19）：

- `go test ./internal/service/copilot ./internal/config -count=1`：`ok  AHaSSHTools/internal/service/copilot`、`ok  AHaSSHTools/internal/config`，全部通过。
- `cd frontend && node --test test/copilotApply.test.js`：4 个测试全部 pass（`shellExecutePayload` 单换行、`applySqlEvent` 携带 sessionId、execute sql 事件 carried handled、peek sql 事件暴露 out）。
- `cd frontend && npm run build`：Vite 前端打包成功（仅 chunk 体积与动态导入告警，均为既有项），JDBC agent Gradle `shadowJar` 成功。

## 对照规格覆盖

- 第 2 节目标与非目标：SQL/Shell 生成 + 填入 + 执行、只读工具循环、OpenAI 兼容、模型名手填原样上传、Ollama 仅预留接口——均按计划落地；非目标项（不自动多轮执行用户产物、不落盘对话、不改 jdbc-agent、不占文件/性能 dock、不做 Flutter/MCP）均未越界。
- 第 7 节错误处理与安全：缺 Key/Base URL 提示、可读错误、非 JSON 当普通回复、SSH 白名单拒绝、探测失败跳过、JDBC 复用既有错误码、执行门禁、隐私截断——均有对应实现与测试。
- 第 9 节测试：自动化测试已覆盖（database 工具循环、SSH 非法探测拒绝、合法产物才可执行、危险规则优先、prompt 不含敏感字段、config.json 不含 Key、模型名原样上传、同 sessionID 并发只跑一轮）。**手工回归四条尚未在 `wails dev` 中执行**，见剩余风险。
- 第 10 节验收：需在真实连接的 MySQL/PostgreSQL/Oracle/金仓与 SSH 会话中手工验证，本次 SDD 未执行。

## 已知限制与后续

- 本地 Ollama Provider 仅预留接口，第一版未接通。
- 对话历史仅存内存，关会话即丢，不落盘。
- 不拉取可用模型列表，模型名由用户按官方文档手填。
- `frontend/wailsjs` 的 Copilot 绑定在 `wails dev` 时再生；在此之前前端通过 `window.go.main.App` 兜底调用。
- Task 1–7 review 期间记录的若干 Minor（如 `ParseArtifact` 取首个 `{`、`\bdd\b`/`\brm\b` 边界假阳性、`SetCopilotAPIKey("")` 仍 Has=true、生成中未禁用「填入」、SSH 模式收到 sql artifact 静默 no-op、`confirmDanger` 并发覆写、peek→confirm→execute 期间活跃面板可能切换等）已登记在 `.superpowers/sdd/progress.md`，留待最终评审或后续迭代处理。
