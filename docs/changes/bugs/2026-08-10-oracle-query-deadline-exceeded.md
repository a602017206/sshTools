# Oracle 表查询 DeadlineExceeded 超时

## 背景

修复 LIMIT 语法后，打开 Oracle 表数据仍可能失败，错误为：

```text
查询执行失败: [DB_CONNECT_FAILED] rpc error: code = DeadlineExceeded desc = context deadline exceeded
```

根因是打开表页时 `loadColumnMetadata`（ListColumns）与 `runDefaultQuery`（ExecuteQuery）并发打同一 JDBC Connection；Connection 非线程安全，Oracle 驱动易挂死直到 Go 侧 30 秒超时。此外 Oracle 把服务名 `pdb` 当作 JDBC catalog 传入元数据接口，无 schema 时可能触发极慢的全库扫描。超时还被误标为 `DB_CONNECT_FAILED`。

## 范围

- JDBC agent 按 session 串行化连接访问
- 表数据面板先加载元数据再执行查询；Oracle/达梦无 schema 时跳过列元数据扫描
- Managed gateway 对 Oracle/达梦清空 catalog
- 超时错误映射为 `QUERY_TIMEOUT` 并给出中文说明

## 修改文件

- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/ConnectionRegistry.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/QueryServiceImpl.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/ConnectionRegistryTest.java`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `internal/service/jdbc_errors.go`
- `internal/service/jdbc_errors_test.go`
- `internal/service/jdbc_managed_gateway.go`
- `internal/service/jdbc_managed_gateway_test.go`
- `docs/changes/bugs/2026-08-10-oracle-query-deadline-exceeded.md`
- `docs/designs/2026-08-10-jdbc-session-serialization.md`
- `docs/development/2026-08-10-oracle-query-deadline-exceeded.md`

## 验证

- `go test ./internal/service -run 'TestJDBCErrorMaps|TestManagedJDBCGatewayClearsOracleCatalog'` 通过
- `cd frontend && node --test test/tableQueryBuilder.test.js` 通过
- JDBC agent `shadowJar`：当前环境无法下载 protobuf Gradle 插件，未能重建 jar；agent 锁改动待网络可用后打包

## 剩余风险

- 未重建的 agent jar 不含 session 锁；依赖前端串行化与跳过无 schema 元数据作为主修复
- 打开表时若未选中 schema，Oracle 主键识别等依赖列元数据的能力暂不可用，但数据查询可先执行
- 真正超慢的 SQL 仍会在 30 秒超时；未调整全局查询超时阈值
