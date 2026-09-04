# 设计：Redis / Elasticsearch 运维工作区与 AI 模块（C+C）

## 背景

Redis / ES 虽有原生工作区与受控读写，但仍是轻量 MVP：键发现无 pattern、截断可误覆盖、无 CLI；ES 缺 Dev Tools / 索引管理 / 结果分页；AI 仅只读 list/describe，文案偏 SQL。需要并行升级「日常运维最小集」，并由 AI 在人在回路下做只读执行与确认后写入。

## 决策

1. **统一壳**：顶栏上下文 + 左列表 + 中栏 Tab（Redis：键详情/CLI；ES：Discover/Dev Tools/索引详情）+ 右 Inspector。不新增第三种 `activeMode`。
2. **Redis**：`SCAN MATCH` + 游标续扫；新建键；截断禁止静默保存；CLI 经 `ExecuteQuery` 白名单；批量删 ≤100。
3. **ES**：查询 `from/size` 分页；受限 Dev Tools（GET/POST，路径白名单）；建/删/刷新索引；文档编辑闭环。
4. **AI**：只读 tool `execute_native_query` 可自动跑；写入产出 `native_mutation` artifact，前端确认后调 `MutateNativeDatabaseResource`。不在服务端静默 mutate。
5. **非目标**：TLS/API Key、Cluster、ILM、无限 SCAN、Kafka。

## 接口增量

- `ListNativeDatabaseChildResourcesPage(sessionID, parent, pattern, cursor, limit) → NativeResourcePage`
- Redis `ExecuteQuery`：CLI / scan JSON
- ES `ExecuteQuery`：索引 `_search` 或 Dev Tools `{method,path,body}`
- Mutate：`create_key`、`delete_keys`、`create_index`、`delete_index`、`refresh_index`
- Copilot：`execute_native_query`；Artifact 类型 `native_query` / `native_mutation`

## 安全

- FLUSH*/DEBUG/CONFIG SET 拒绝；ES 路径仅 `_search|_mapping|_stats|_doc|_refresh|_cluster/health` 等白名单
- 查询结果条数/字节 clamp；写操作 ConfirmDialog
