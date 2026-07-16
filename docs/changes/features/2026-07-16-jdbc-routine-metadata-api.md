# JDBC 例程元数据接口

## 背景

通用对象树需要支持存储过程和函数，不能只显示表与视图。

## 范围

JDBC agent 新增 `ListRoutines` RPC。接口使用 `DatabaseMetaData.getProcedures` 和 `DatabaseMetaData.getFunctions`，以 Schema 和 catalog 过滤例程名称。

## 修改文件

- `jdbc-agent/src/main/proto/jdbc_agent.proto`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/MetadataServiceImplTest.java`
- `internal/service/jdbcproto/jdbc_agent.pb.go`
- `internal/service/jdbcproto/jdbc_agent_grpc.pb.go`
- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_grpc_client.go`
- `internal/service/jdbc_managed_gateway.go`
- `internal/service/database_service.go`
- `app.go`
- `frontend/src/components/GenericJDBCObjectTree.svelte`
- 本变更记录。

## 验证

先运行 `./gradlew test --tests '*MetadataServiceImplTest'`，新增测试因 RPC 尚未实现失败；实现后运行 `./scripts/generate-jdbc-proto.sh` 并再次运行同一 Gradle 测试，结果通过。

## 剩余风险

已接入 Go gateway、Wails API 与通用对象树。驱动可能不支持过程或函数元数据，调用失败时界面会显示加载错误；当前未提供例程定义查看。
