# 开发记录：Copilot 当前打开对象上下文

## 实现内容

- 前端用 `buildCopilotWorkspaceContext` 从活动标签、侧栏导航和按会话 focus 合成当前库/schema/表/索引/键；表标签优先于对象浏览器选中项。
- `AIPanel` 发话时把 `Schema`、`ObjectKind`、`ObjectName`、`ObjectParent`、`EditorContent` 一并交给 `CopilotChat`。面板标题显示“当前：”摘要。
- JDBC 工具按请求中的 Database/Schema 限定 `list_tables`；`get_table_schema` 未传表名时使用当前打开的表。
- Redis / Elasticsearch / Kafka 等原生类型不再注册 SQL schema 工具，改为 `list_resources`、`list_child_resources`、`describe_resource`。
- `fillCopilotRequest` 只在 Host/User/DBType/Database 为空时回填连接默认值，避免覆盖已打开的库。

## 验证

- `cd frontend && node --test test/copilotContext.test.js`：通过。
- `go test ./internal/service/copilot -count=1`：通过。

## 剩余风险

原生资源详情可能含业务数据；依赖用户在发话前已选中目标对象。桌面端需在真实会话中确认标题“当前：”与助手回复是否指向同一张表或同一个索引。
