# JDBC agent 查询闭环

## 背景

JDBC agent 建立健康检查后，需要完成首个真实数据库操作闭环。该能力用于验证 agent 可以加载 JDBC 驱动、建立本地连接，并把查询结果通过 gRPC 返回给 Go 网关。

## 范围

- 新增 `DriverLoader`，通过 `URLClassLoader` 按 profile 的 jar 路径加载 JDBC 驱动类。
- 新增 `ConnectionRegistry`，按 session ID 保存 JDBC `Connection`。
- 新增 `QueryServiceImpl`，实现 `OpenSession` 和 `ExecuteQuery`。
- 查询结果统一转为字符串，避免首版引入复杂动态类型。
- 更新 agent 应用入口，注册包含健康检查和查询能力的服务实现。
- 新增 H2 查询闭环测试，覆盖 `select 1 as id, 'ok' as name` 的列和值返回。
- 在 Gradle 测试依赖中加入 H2 JDBC 驱动。

## 修改文件

- `jdbc-agent/build.gradle`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/DriverLoader.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/ConnectionRegistry.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/QueryServiceImpl.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/HealthServiceImpl.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/JdbcAgentApplication.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/QueryServiceImplTest.java`

## 验证

- 已运行 `cd jdbc-agent && ./gradlew test --tests '*QueryServiceImplTest'`，结果通过。

## 剩余风险

- 当前 URL template 只做简单占位符替换，尚未处理转义或数据库类型差异。
- 连接属性只合并基础 `user`、`password` 和请求 map，后续需要和 profile 属性定义联动。
- `DriverLoader` 当前未管理 `URLClassLoader` 生命周期，后续会话关闭和驱动卸载需要补充资源回收策略。
