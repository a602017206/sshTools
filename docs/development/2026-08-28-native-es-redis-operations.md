# Redis / Elasticsearch 原生读写实现说明

## 开发分支

- 分支：`feature/native-es-redis-ops`
- Worktree：`.worktrees/native-es-redis-ops`

## API

### ExecuteNativeDatabaseQuery(sessionID, parent, name, query)

- Elasticsearch：`name` 为索引名，`query` 为 DSL JSON
- 返回 `{ summary, content }`，`content` 含 `hits` 与原始响应

### MutateNativeDatabaseResource(sessionID, parent, name, operation, payload)

| 类型 | operation | payload |
|------|-----------|---------|
| Redis | `set` | `{"value":"...", "ttlSeconds":60}` |
| Redis | `delete` | `{}` |
| ES | `index_document` | `{"id":"optional","document":{...}}` |
| ES | `update_document` | `{"id":"required","document":{...}}` |
| ES | `delete_document` | `{"id":"required"}` |

## 前端入口

连接 Redis / ES 后进入 `native-database` 面板：

- Redis：选中键 → 右侧 Inspector 编辑 string 值或删除
- ES：选中索引 → 「查询」页签执行 DSL → 结果中编辑/删除文档

## 合并前

```bash
cd .worktrees/native-es-redis-ops
go test ./...
node --test frontend/test/nativeDatabaseOperations.test.js
wails dev  # 重新生成 bindings（如与手工维护不一致）
```
