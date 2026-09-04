# 设计：搜索/缓存与关系库连接路径剥离

## 背景

新建/编辑连接可以继续共用「数据」资产表单（按域选 Redis、ES、MySQL 等），但实际连接后的侧栏与工作区仍把所有 `type=database` 当成 JDBC：资产树出现展开箭头，挂载 `DatabaseSidebarTree` 调用 `ListDatabases`，Elasticsearch / Redis 等原生类型必然「加载数据库失败」，体验很差。

## 决策

1. **表单仍合并。** 新建/编辑对话框、资产域分组（数据库 / 缓存 / 搜索 / 消息）保持现状。
2. **运行时按能力分流。** 用 `isNativeDatabaseType` / `assetSupportsJdbcSidebar`：仅 JDBC 关系库在资产树展开库/Schema；原生类型无展开箭头，双击/连接直接打开 `native-database` 工作区。
3. **文案按域。** 会话标签用工作区标题（如「Redis 键空间」「Elasticsearch 索引」）；顶栏模式改为「数据」；AI 助手标题按缓存/搜索/消息分流，不再一律「SQL 助手」。
4. **防御。** `DatabaseSidebarTree` 对非 JDBC 资产直接拒绝加载，避免误挂载后红字失败。

## 取舍

不拆第三种顶栏 Mode（仍 ssh / database），避免改标签栏与会话过滤；域筛选继续靠左侧 DomainRail。MongoDB 等虽归「数据库」域，但仍是原生协议，同样不展开 JDBC 树。

## 限制

资产仍存 `type: database`；历史已展开状态在升级后不再显示子树。OpenSearch 统一按 Elasticsearch 原生路径处理。
