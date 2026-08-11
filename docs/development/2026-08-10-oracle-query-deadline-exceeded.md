# Oracle 查询超时修复实现说明

## 实现要点

- `ConnectionRegistry.withConnection` 串行化 session 级 JDBC 访问
- `DatabaseTablePanel`：`await loadColumnMetadata()` 后再 `runDefaultQuery()`
- `ManagedJDBCGateway.jdbcCatalog`：oracle/dm 返回空 catalog
- `newJDBCError`：识别 `DeadlineExceeded` → `QUERY_TIMEOUT`

## 验证

```bash
go test ./internal/service -run 'TestJDBCErrorMaps|TestManagedJDBCGatewayClearsOracleCatalog'
cd frontend && node --test test/tableQueryBuilder.test.js
```

均已通过。

JDBC agent 的 session 锁改动已写入源码；当前环境无法解析 `com.google.protobuf` Gradle 插件，`shadowJar` 未能重建，因此 `frontend/build/jdbc-agent.jar` 仍为旧产物。主修复不依赖新 jar：前端已串行化请求，并在 Oracle 无 schema 时跳过列元数据。成功执行 `node scripts/stage-jdbc-agent.mjs` 后重启应用即可启用 agent 侧锁。
