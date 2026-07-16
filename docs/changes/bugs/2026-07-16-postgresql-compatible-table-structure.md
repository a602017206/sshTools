# PostgreSQL 兼容数据库表结构加载失败

## 背景

JDBC 模式下的 `GetTableDDL` 仅实现了 MySQL 的 `SHOW CREATE TABLE`。人大金仓、PostgreSQL 和 openGauss 虽然 JDBC agent 已能读取列和主键信息，但在调用元数据前被直接拒绝，导致表结构面板无法打开。

## 范围

保留 MySQL 的原始精确 DDL 查询。对 PostgreSQL、人大金仓和 openGauss，通过 JDBC agent 的表字段元数据生成 PostgreSQL 风格的 `CREATE TABLE` 结构文本，包含字段名、类型、非空和主键标识。

## 修改文件

- `internal/service/database_service.go`
- `internal/service/database_service_test.go`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run 'TestDatabaseServiceGetTableDDLUsesJDBC(GatewayForMySQL|SchemaForKingbase)' -v`，验证 MySQL 既有 DDL 和人大金仓字段结构生成。

## 剩余风险

JDBC 的 `ListColumns` 协议未携带默认值、表注释、索引、外键和 schema 名称，因此 PostgreSQL 兼容类型生成的是字段结构文本，不等同于厂商导出的完整建表语句。schema、表、视图、函数等对象树需要后续由 JDBC agent 暴露完整元数据后再实现，不能用固定名称模拟。
