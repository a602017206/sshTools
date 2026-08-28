# Redis 与 Elasticsearch 原生读写能力

## 背景

原生 Redis / Elasticsearch 工作区此前仅支持资源浏览与受限只读预览，无法修改数据，ES 也不能执行自定义 DSL 查询。

## 范围

### 后端

- 新增 `ExecuteQuery` / `MutateResource` 服务契约与 Wails 绑定
- Redis：增强 hash/list/set/zset 预览；支持 `set` / `delete`
- Elasticsearch：支持 `_search` DSL、`index_document` / `update_document` / `delete_document`

### 前端

- `NativeDatabasePanel` 增加 ES 查询编辑器、结果列表、文档写入/更新/删除
- Redis Inspector 支持 string 键编辑与删除
- `nativeDatabaseWorkspace` 声明 `canQuery` / `canWrite` / `canDelete`

## 修改文件

- `internal/service/native_database_operations.go`（新增）
- `internal/service/native_redis_operations.go`（新增）
- `internal/service/native_elasticsearch_operations.go`（新增）
- `internal/service/native_redis.go`
- `internal/service/native_elasticsearch.go`
- `app.go`
- `frontend/src/lib/nativeDatabaseOperations.js`（新增）
- `frontend/src/lib/nativeDatabaseWorkspace.js`
- `frontend/src/components/NativeDatabasePanel.svelte`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/models.ts`
- 测试与 design 文档

## 验证

- [x] `go test ./internal/service/... -run 'Redis|Elasticsearch|NativeDatabase'`
- [x] `node --test frontend/test/nativeDatabaseOperations.test.js`
- [ ] 手工：Redis string 键保存/删除
- [ ] 手工：ES 执行 DSL 查询并写入/删除文档

## 剩余风险

- Redis 非 string 类型仍不支持直接编辑，仅预览
- ES 未支持 HTTPS / API Key 认证
- 大批量结果仍受 size 上限约束（100 条）
