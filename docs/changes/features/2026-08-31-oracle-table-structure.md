# Oracle 表结构加载与修改

## 背景

打开 Oracle 表的「设计表」时，字段加载失败并提示 `ORA-17008: 已关闭连接`，同时页面声明不支持 Oracle 改表。

设计文档：`docs/designs/2026-08-31-oracle-table-structure.md`。

## 范围

- JDBC 会话在连接已关闭时自动重开并重试
- 表设计器支持 Oracle 查看与 `CREATE`/`ALTER` 预览、保存
- 顺序加载表结构，避免对已关闭连接并发查询

不包含达梦、SQL Server 表设计器。

## 修改文件

- `internal/service/jdbc_errors.go`
- `internal/service/jdbc_errors_test.go`
- `internal/service/jdbc_managed_gateway.go`
- `internal/service/jdbc_managed_gateway_test.go`
- `frontend/src/lib/tableDefinitionSQL.js`
- `frontend/src/lib/tableAlterSQL.js`
- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/test/tableDefinitionSQL.test.js`
- `frontend/test/tableAlterSQL.test.js`
- `frontend/test/tableStructureDesigner.test.js`
- `docs/designs/2026-08-31-oracle-table-structure.md`
- `docs/changes/features/2026-08-31-oracle-table-structure.md`（本文）

## 验证

- `go test ./internal/service -count=1 -run 'TestJDBCClosedOracleConnectionIsStaleSession|TestManagedJDBCGatewayReopensSessionAfterOracleClosedConnection|TestManagedJDBCGatewayClearsOracleCatalogForMetadata'`
- `cd frontend && node --test test/tableDefinitionSQL.test.js test/tableAlterSQL.test.js test/tableStructureDesigner.test.js`
- 手工：在仍打开的 Oracle 会话上重新打开设计表，字段应能加载；修改一列后 DDL 预览应为 `ALTER TABLE "schema"."table" MODIFY (...)`

## 剩余风险

- 无法在本机连真实 Oracle 自动验证 `ORA-17008` 重连
- 部分 Oracle 版本对 `BOOLEAN` 等通用类型不支持，保存时会由数据库报错
