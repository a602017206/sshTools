# 变更：Redis / ES 运维工作区与 AI 模块（C+C）

## 背景

Redis / Elasticsearch 工作区能力偏薄，AI 仅只读 list/describe，日常查改体验不足。

## 范围

- Redis：MATCH 扫描与续扫、新建键、截断禁保存、CLI 白名单、批量删
- ES：Discover 分页、Dev Tools、建/删/刷新索引、文档编辑闭环
- AI：`execute_native_query`、`propose_native_mutation`；`native_query` / `native_mutation` artifact，确认后执行
- 设计文档与发布说明

## 修改文件

- `internal/service/native_*.go`、`native_ops_guard_test.go`
- `app.go`、`copilot_native.go`、`internal/service/copilot/*`
- `frontend/src/components/workspaces/RedisWorkspace.svelte`、`ElasticsearchWorkspace.svelte`
- `frontend/src/lib/nativeDatabaseOperations.js`、`copilotApply.js`、`AIPanel.svelte`
- `docs/designs/2026-09-02-redis-es-ops-and-copilot.md` 等

## 验证

- `go test ./internal/service/ ./internal/service/copilot/`
- `node --test test/nativeDatabaseOperations.test.js test/nativeDatabaseWorkspace.test.js test/copilotContext.test.js`
- 未连真实 Redis/ES 做端到端（剩余风险）

## 剩余风险

- Wails 绑定若未热更新，需重启 `wails dev` 才能看到 `ListNativeDatabaseChildResourcesPage`
- Dev Tools / CLI 白名单可能需按现场命令再扩展
- AI mutation 依赖模型输出合法 artifact JSON
