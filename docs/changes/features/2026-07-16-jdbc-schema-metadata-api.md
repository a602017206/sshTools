# JDBC Schema 元数据接口

## 背景

为 Oracle、SQL Server、达梦、SQLite 等 JDBC 数据库提供统一对象树，需要可靠读取 Schema，而不应在前端针对每种数据库猜测系统表。

## 范围

扩展 JDBC agent 的 gRPC 协议，新增 `ListSchemas`；Java agent 通过 `DatabaseMetaData.getSchemas` 返回 Schema 名称，Go gateway、Wails API 和前端对象树后续复用该接口。

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
- `frontend/src/components/TerminalPanel.svelte`
- 本变更记录。

## 验证

已先执行 `./gradlew test --tests '*MetadataServiceImplTest'`，新增测试因 `listSchemas` 尚未实现而失败；Java 实现已补充。执行 Go 协议生成脚本时，`protoc` 曾失败，提示缺少 `/opt/homebrew/opt/abseil/lib/libabsl_die_if_null.2508.0.0.dylib`。修复后已重新生成协议代码，并通过 `go test ./internal/service -run 'TestJdbcGatewayListSchemasPassesCatalogToAgent|TestManagedJDBCGatewayReconnectsSessionAfterAgentUnavailable' -v`、`./gradlew test --tests '*MetadataServiceImplTest'`、前端对象树测试和 `npm run build`。

## 剩余风险

`DatabaseMetaData.getSchemas` 的返回结果依赖 JDBC 驱动实现；没有 Schema 概念的驱动会在界面中显示“默认 Schema”。通用树当前以表为对象分类，Oracle、SQL Server、达梦的视图和例程分类仍需后续扩展。原 `protoc` 动态库问题已通过恢复匹配的 Homebrew `abseil`/protobuf 运行库解决，并已重跑 `./scripts/generate-jdbc-proto.sh`；没有手工编辑生成的 Go 协议文件。
