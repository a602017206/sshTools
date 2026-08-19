# 变更：AI Copilot（SQL / Shell）

## 背景

运维工作区已能执行 SQL 与 SSH 命令，但缺少自然语言到可执行产物的一层。用户希望在当前数据库或 SSH 会话旁用自然语言生成 SQL 或 Shell，确认后再跑，且危险写操作要有二次确认。现有 Java `jdbc-agent` 是数据库 sidecar，不是 LLM Agent，本功能不另起 Agent 进程，而是在 Go 服务层新增 Copilot 内核，复用已有执行面。

## 范围

- 新增 Go 包 `internal/service/copilot`：危险分类、产物解析、脱敏、SSH 探测白名单、OpenAI 兼容 Provider、只读工具循环（最多 4 轮）、按 `sessionID` 的并发与取消控制。
- `app.go` 导出 `CopilotChat` / `CopilotClassify` / `HasCopilotAPIKey` / `SetCopilotAPIKey` / `ClearCopilotAPIKey`；`CloseSSH` / `CloseDatabase` 时取消该会话的进行中生成。
- `internal/config` 增加 `copilot_provider` / `copilot_base_url` / `copilot_model` 三个设置字段；API Key 只存 `CredentialStore` 键 `copilot:api_key`，不进 `config.json`。
- `internal/service/session_service.go` 新增 `ExecuteCommand` 包装，供 Copilot SSH 探测经另开会话执行，不进用户 PTY。
- 前端新增 `AIPanel` 独立可折叠侧栏（顶栏 AI 按钮打开，默认收起，数据库模式也显示），新增 `copilotApply.js` 填入/执行辅助与 `copilot` store。
- `GlobalSettingsDialog` 新增「AI Copilot」分段（Base URL、模型名称、API Key 三个手填框 + 清除密钥）；模型名按服务商官方文档手填，原样上传，不映射、不拉列表。
- `DatabasePanel` / `DatabaseTablePanel` 监听 `copilot:apply-sql` 填入编辑器；`TerminalPanel` 新增 `insertCopilotText` 填入终端不换行。
- 执行语义：SQL「填入」派发事件改 `query`，「执行」先 `CopilotClassify`，危险则 `ConfirmDialog`，确认后 `ExecuteDatabaseQuery`；Shell「填入」`SendSSHData` 不加换行，「执行」分类确认后 `SendSSHData` 发送命令 + 单换行。
- 不修改 Java `jdbc-agent`；不占用「文件 / 性能」dock；不接通 Ollama（仅预留接口）；对话不落盘。

## 修改文件

后端：

- 新增：`internal/service/copilot/types.go`、`classify.go`、`classify_test.go`、`artifact.go`、`artifact_test.go`、`redact.go`、`redact_test.go`、`probe.go`、`probe_test.go`、`provider.go`、`provider_openai.go`、`provider_openai_test.go`、`tools.go`、`service.go`、`service_test.go`、`config.go`、`config_test.go`
- 修改：`internal/config/config.go`、`internal/config/config_test.go`、`internal/service/session_service.go`、`app.go`

前端：

- 新增：`frontend/src/lib/copilotApply.js`、`frontend/test/copilotApply.test.js`、`frontend/src/stores/copilot.js`、`frontend/src/components/AIPanel.svelte`
- 修改：`frontend/src/settings/appearance.js`、`frontend/src/components/GlobalSettingsDialog.svelte`、`frontend/src/App.svelte`、`frontend/src/components/DatabasePanel.svelte`、`frontend/src/components/DatabaseTablePanel.svelte`、`frontend/src/components/TerminalPanel.svelte`

文档：

- 修改：`docs/superpowers/specs/2026-08-13-ai-copilot-sql-shell-design.md`（状态改为已落地）
- 新增：`docs/development/2026-08-13-ai-copilot-sql-shell.md`、`docs/changes/features/2026-08-13-ai-copilot-sql-shell.md`（本文档）

## 验证

```bash
go test ./internal/service/copilot ./internal/config -count=1
cd frontend && node --test test/copilotApply.test.js
cd frontend && npm run build
```

实测结果（2026-08-19）：

- `go test ./internal/service/copilot ./internal/config -count=1`：两个包均 `ok`，全部通过。
- `cd frontend && node --test test/copilotApply.test.js`：4 个测试全部 pass。
- `cd frontend && npm run build`：Vite 前端打包成功，JDBC agent Gradle `shadowJar` 成功（仅既有 chunk 体积与动态导入告警）。

## 剩余风险

- **手工回归未执行**：规格第 9 节四条手工回归（数据库自然语言→填入→执行→结果在面板；SSH 自然语言→填入终端不换行→执行后命令在 PTY；`DROP TABLE`/`rm -rf` 必须确认且取消后无执行；未配 Key、错误模型名、JDBC agent 不可用时提示可读）**未在 `wails dev` 中实际运行**，本次 SDD 仅完成自动化测试与前端构建。规格第 10 节验收项同样需在真实连接的 MySQL/PostgreSQL/Oracle/金仓与 SSH 会话中手工确认，尚未执行。
- `frontend/wailsjs` 的 Copilot 绑定在 `wails dev` 时再生；在此之前前端通过 `window.go.main.App` 兜底，未在桌面运行时验证过绑定完整性。
- Task 1–7 review 期间登记的若干 Minor（如 `ParseArtifact` 取首个 `{` 可能误抽、`\bdd\b`/`\brm\b` 词边界假阳性、生成中未禁用「填入」、SSH 模式收到 sql artifact 静默 no-op、`confirmDanger` 并发覆写、peek→confirm→execute 期间活跃面板可能切换等）未在本任务修复，留待最终评审或后续迭代。已修项：空 peek 误报「已执行」、MaxToolRounds 触顶无最终产物、`SetCopilotAPIKey("")` 仍 `Has=true`。
- 本地 Ollama Provider 仅预留接口，第一版未接通，无法验证本地模型路径。
- 对话历史仅存内存，关会话即丢；多窗口或异常退出时未做持久化兜底。
