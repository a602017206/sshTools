# 通用 JDBC 表结构

## 背景

通用 JDBC 对象树已能显示 Oracle、SQL Server、达梦和 SQLite 的表，但这些表不能打开字段定义，浏览体验不完整。

## 范围

JDBC 模式下，除 MySQL 和 PostgreSQL 兼容类型外的数据库也使用现有 `ListColumns` 元数据生成 DDL。通用对象树中“表”节点单击会打开表结构面板，并携带 Schema 名称以消除同名表歧义。

## 修改文件

- `internal/service/database_service.go`
- `internal/service/database_service_test.go`
- `frontend/src/components/GenericJDBCObjectTree.svelte`
- `frontend/build/assets/index.js`
- 本变更记录。

## 验证

先执行 `go test ./internal/service -run TestDatabaseServiceGetTableDDLUsesJDBCSchemaForGenericDriver -v`，确认 Oracle 类型报“不支持”；实现后执行该测试和 Schema 表结构测试，均通过。执行 `npm run build`，前端编译与 JDBC agent 暂存通过。

## 剩余风险

生成的 DDL 由通用字段元数据构造，不包含索引、约束名称、分区、注释和特定厂商存储参数。对象树中的视图与系统表仍只浏览名称，未提供定义查看。
