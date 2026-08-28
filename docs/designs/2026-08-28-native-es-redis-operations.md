# Redis 与 Elasticsearch 原生工作区读写能力设计

## 背景

2026-08-21 只读预览上线后，Redis 与 Elasticsearch 仍缺少日常运维所需的核心能力：Redis 无法修改/删除键，Elasticsearch 无法执行 DSL 查询或写入/删除文档。用户反馈与 JDBC 工作区体验差距过大。

## 目标

在现有 `NativeDatabaseService` 架构上，为 Redis 与 Elasticsearch 增加受控读写能力，并保持协议细节封装在各自 Provider 内。

### Redis

- 增强 `DescribeResource`：支持 hash、list、set、zset 的受限预览
- `set`：写入/更新 string 键，可选 TTL
- `delete`：删除键

### Elasticsearch

- `ExecuteQuery`：对指定索引执行 `_search` DSL（POST body）
- `index_document`：写入或覆盖文档（指定 `_id` 可选）
- `update_document`：按 `_id` 部分更新
- `delete_document`：按 `_id` 删除

## 架构

```text
NativeDatabasePanel
  → ExecuteNativeDatabaseQuery / MutateNativeDatabaseResource (Wails)
    → NativeDatabaseService.ExecuteQuery / MutateResource
      → optional NativeQueryExecutor / NativeResourceMutator on session client
        → Redis / Elasticsearch provider implementation
```

未实现接口的 Provider 返回 `ErrNativeOperationUnsupported`，前端根据 `nativeDatabaseWorkspace` 能力标志隐藏操作。

## 安全与限制

- Redis 值预览上限 4 KiB；hash/list/set/zset 各取前 50 项
- ES 查询默认 `size` 上限 100（服务端 clamp）
- 所有变更操作需前端确认对话框
- 不引入 TLS/SASL/API Key（后续独立变更）

## 前端

- `NativeDatabasePanel` 增加工作区模式：
  - Redis：Inspector 内联编辑值 + 删除键
  - ES：查询编辑器 + 结果表格 + 文档增删改
- `nativeDatabaseWorkspace.js` 声明 `canQuery`、`canWrite`、`canDelete`
