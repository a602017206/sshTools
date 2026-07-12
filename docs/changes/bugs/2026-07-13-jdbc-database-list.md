# JDBC 数据库列表加载失败修复

## 背景

JDBC 连接建立成功后，数据库面板会调用 `ListDatabases` 加载数据库列表。原 managed gateway 将该调用转发到尚未实现的占位方法，导致界面显示“加载数据库失败”。

## 范围

本次修复复用现有查询 RPC，不修改 gRPC/proto 和 Java agent。MySQL 使用 `SHOW DATABASES`，PostgreSQL 使用 `pg_database` 系统目录查询，并将结果第一列转换为数据库名称列表。

## 修改文件

- `internal/service/jdbc_managed_gateway.go`
- `internal/service/jdbc_managed_gateway_test.go`
- `docs/changes/bugs/2026-07-13-jdbc-database-list.md`

## 验证

- `go test ./internal/service -run TestManagedJDBCGatewayListsMySQLDatabasesThroughQuery -v`
- `go test ./internal/service -v`
- `go test ./...`

## 剩余风险

当前数据库列表查询仅覆盖 MySQL 和 PostgreSQL。其他 JDBC 数据库类型仍会返回明确的不支持错误，后续需要按各数据库的目录语义补充查询或扩展元数据 RPC。
