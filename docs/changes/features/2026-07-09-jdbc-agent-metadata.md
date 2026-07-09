# JDBC agent 元数据和会话关闭

## 背景

查询闭环完成后，前端数据库浏览需要获取表列表、字段信息和主键标记。同时 agent 需要支持显式关闭会话，避免 JDBC 连接长期驻留。

## 范围

- 新增 `MetadataServiceImpl`，在查询服务基础上实现元数据和关闭会话接口。
- `ListTables` 使用 `DatabaseMetaData.getTables` 返回表名。
- `ListColumns` 使用 `DatabaseMetaData.getColumns` 返回字段名、类型和可空信息。
- 主键信息通过 `DatabaseMetaData.getPrimaryKeys` 合并到字段结果。
- `CloseSession` 从 `ConnectionRegistry` 删除连接并关闭底层 JDBC 连接。
- 更新 agent 应用入口，注册包含健康检查、查询、元数据和关闭能力的服务实现。
- 新增 H2 元数据测试，覆盖建表、列元数据、主键和关闭后 `NOT_FOUND`。

## 修改文件

- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/JdbcAgentApplication.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/MetadataServiceImplTest.java`

## 验证

- 已运行 `cd jdbc-agent && ./gradlew test --tests '*MetadataServiceImplTest'`，结果通过。

## 剩余风险

- 当前表类型只查询 `TABLE`，暂未覆盖视图、系统表或数据库特定对象类型。
- schema 和 catalog 过滤直接透传 JDBC 驱动，后续需要按数据库类型补充默认 schema 策略。
- 字段类型直接使用 `TYPE_NAME`，不同数据库的展示格式可能需要前端或 Go 网关进一步归一化。
