# 设计：Copilot 当前打开对象上下文

## 背景

数据库、缓存、搜索工作区里，用户已经选中了库、表、索引或键，但 AI 助手请求几乎只携带 `SessionID` 和用户原话。后端虽然能从连接配置补 `Host` / `User` / `DBType` / 默认 `Database`，却拿不到对象浏览器或表标签里真正打开的 catalog、schema、表名，也拿不到 Redis 键、Elasticsearch 索引等原生资源。缓存/搜索会话仍走 JDBC 的 `list_tables` 工具，工具失败后模型只能猜测。

## 决策

- **前端是唯一的“当前打开对象”来源。** 后端会话只有连接默认库，没有 UI 选中态。对象浏览器、表/设计器标签、原生工作区在变化时写入按会话隔离的 focus；发话时由纯函数合成 `ChatRequest`。
- **表标签优先于侧栏选中。** 若当前标签是表数据或表设计器，使用该标签上的 `databaseName` / `schemaName` / `tableName`；否则使用侧栏导航与对象浏览器选中项。
- **JDBC 与原生分流工具。** `Mode` 仍为 `database`（工作区只有 ssh/database 两态）。`DBType` 属于 Redis / Elasticsearch / Kafka 等原生类型时，注册 `list_resources` / `list_child_resources` / `describe_resource`，不再暴露 SQL schema 工具。
- **只读、可默认当前对象。** `get_table_schema` 与 `describe_resource` 在模型未传名称时使用请求里的 `ObjectName` / `ObjectParent`。`list_tables` 按请求中的 Database + Schema 限定范围。不自动执行写操作。
- **后端不覆盖前端已填字段。** `fillCopilotRequest` 只在 Host / User / DBType / Database 为空时回填连接默认值，避免把已打开的库冲掉。

## 请求字段

在现有 `ChatRequest` 上增加：`Schema`、`ObjectKind`、`ObjectName`、`ObjectParent`。`EditorContent` 改为传入当前 SQL 或 ES DSL。字段不得包含密码。

## 取舍

不把完整表结构或索引 mapping 默认塞进 prompt，避免上下文膨胀；模型用只读工具按需拉取。不新增第三种工作区 Mode，以免改动标签栏。原生 `describe_resource` 的正文走既有脱敏与截断。

## 限制

未打开任何对象时，助手只能看到连接级信息。关闭标签后 focus 可能短暂残留在已关闭的 sessionId 上，下一次发话以当前活动标签为准。
