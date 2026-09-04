# 功能：Copilot 使用当前打开的表、索引与键

## 背景

数据库、缓存、搜索工作区里，用户已经打开了表、索引或键，但 AI 助手请求几乎只带会话 ID。模型看不到当前 catalog、schema、表名，Redis/ES 还被当成 JDBC 去调 `list_tables`，只能猜测。

## 范围

把当前工作区打开对象写入 Copilot 请求，并按 JDBC / 原生数据源分流只读工具。不改变生成结果的执行与二次确认流程。

## 修改文件

- `docs/designs/2026-09-01-copilot-open-object-context.md`
- `docs/development/2026-09-01-copilot-open-object-context.md`
- `frontend/src/lib/copilotContext.js`
- `frontend/test/copilotContext.test.js`
- `frontend/src/stores/copilot.js`
- `frontend/src/components/AIPanel.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/workspaces/RedisWorkspace.svelte`
- `frontend/src/components/workspaces/ElasticsearchWorkspace.svelte`
- `frontend/src/components/workspaces/KafkaWorkspace.svelte`
- `frontend/src/components/workspaces/GenericNativeWorkspace.svelte`
- `frontend/wailsjs/go/models.ts`
- `internal/service/copilot/service.go`
- `internal/service/copilot/tools.go`
- `internal/service/copilot/service_test.go`
- `internal/service/database_service.go`
- `internal/service/native_database.go`
- `app.go`
- `copilot_native.go`

## 验证

```bash
cd frontend && node --test test/copilotContext.test.js
go test ./internal/service/copilot -count=1
go test ./internal/service -count=1 -run 'TestDatabaseService_ListTables|TestDatabaseServiceGetTable'
```

前端 10 项 Copilot 上下文测试通过；Copilot 包测试通过，含打开对象 prompt、按 schema 列表面、默认当前表结构、原生 `describe_resource`。

未在桌面应用内对真实 Oracle/Redis/ES 会话做手工发话验证。

## 剩余风险

未选中任何对象时，助手仍只有连接级信息。原生 `describe_resource` 可能包含键值或文档片段，发送前会脱敏和截断，但仍可能把业务数据交给模型。Wails 绑定需在下次 `wails dev` / `wails generate` 时与 `ChatRequest` 新字段对齐。
