# JDBC 对象类型筛选

## 背景

通用 JDBC 对象树需要显示视图等标准数据库对象，不能只固定查询表。

## 范围

为既有 `ListTablesRequest` 增加 `types` 字段。未传递类型时保持原有“表、系统表”默认行为；传递类型时由 JDBC agent 原样传入 `DatabaseMetaData.getTables`，支持标准 JDBC `TABLE`、`VIEW`、`SYSTEM TABLE` 等类型。

## 修改文件

- `jdbc-agent/src/main/proto/jdbc_agent.proto`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/MetadataServiceImplTest.java`
- `internal/service/jdbcproto/jdbc_agent.pb.go`
- `internal/service/jdbcproto/jdbc_agent_grpc.pb.go`
- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_managed_gateway.go`
- `internal/service/database_service.go`
- `app.go`
- `frontend/src/components/GenericJDBCObjectTree.svelte`
- 本变更记录。

## 验证

先执行 `./gradlew test --tests '*MetadataServiceImplTest'`，验证新增 `VIEW` 筛选测试失败；实现后执行 `./scripts/generate-jdbc-proto.sh` 重新生成 Go 代码，并再次执行同一 Gradle 测试，结果通过。执行 `go test ./internal/service -run TestJdbcGatewayListObjectsPassesTypesToAgent -v` 验证 Go gateway 会传递对象类型；前端构建验证通用对象树。

## 剩余风险

不同 JDBC 驱动支持的对象类型名称可能不同。当前使用 JDBC 标准类型并保留默认兼容行为；通用对象树已展示表、视图和系统表，例程与函数需要后续使用 JDBC 的过程和函数元数据接口扩展。
