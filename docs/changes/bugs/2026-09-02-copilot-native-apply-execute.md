# 变更：ES/Redis AI 助手恢复「填入 / 执行」

## 背景

原生会话系统提示仍要求 `sql`/`shell`，且 DSL 常以 JSON 对象放在 `content` 导致 artifact 解析失败，气泡没有填入/执行按钮；工作区也未监听 native apply 事件。

## 范围

- 解析 `content` 为对象的 `native_query` / `native_mutation`
- 原生 DBType 专用 system prompt；sql/shell 强制归一为 `native_query`
- ES/Redis 工作区监听填入/执行事件

## 修改文件

- `internal/service/copilot/artifact.go`、`service.go`、`artifact_test.go`
- `frontend/src/lib/nativeCopilotApply.js`、`test/nativeCopilotApply.test.js`
- `ElasticsearchWorkspace.svelte`、`RedisWorkspace.svelte`、`AIPanel.svelte`

## 验证

- `go test ./internal/service/copilot/`
- `node --test test/nativeCopilotApply.test.js`

## 剩余风险

- 旧对话历史里无 artifact 的消息不会补出按钮，需重新提问
