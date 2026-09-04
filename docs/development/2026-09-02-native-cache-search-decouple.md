# 实现：搜索/缓存与关系库连接路径剥离

## 做法

- 新增 `assetSupportsJdbcSidebar`：仅非原生 `database` 资产走库树。
- `AssetList` 用该判定控制展开按钮与子树；原生类型点击展开改为 `openPanel: true` 连接。
- `databaseSessionTabLabel`、`copilotAssistantTitle` 提供会话与 AI 文案。
- `isNativeDatabaseType` 规范化大小写并包含 `opensearch`。

## 发布说明要点

资产树中 Redis / Elasticsearch 等不再出现「加载数据库失败」的 JDBC 子树；请双击打开对应工作区。
