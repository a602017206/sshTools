# JDBC 表结构加载失败修复

## 背景

数据库连接迁移到 JDBC agent 后，表结构面板仍通过旧的 Go `database/sql` 连接获取 DDL。JDBC session 不持有 `sql.DB`，因此 MySQL 连接成功后加载表结构仍会失败，且前端无法展示 Wails 返回的字符串错误。

## 范围

JDBC gateway 模式下复用现有查询 RPC 执行 MySQL `SHOW CREATE TABLE`，解析返回的表名和建表语句。表结构面板同时兼容 `Error` 对象和字符串错误。此次修复不修改 gRPC/proto 或 Java agent。

## 修改文件

- `internal/service/database_service.go`
- `internal/service/database_service_test.go`
- `frontend/src/components/TableStructurePanel.svelte`
- `docs/changes/bugs/2026-07-13-jdbc-table-ddl.md`

## 验证

- `go test ./internal/service -run TestDatabaseServiceGetTableDDLUsesJDBCGatewayForMySQL -v`
- `go test ./internal/service -v`
- `go test ./...`
- `cd frontend && npm run build`
- `wails build`

首次执行 `cd frontend && npm run build` 时，Vite 编译成功，但 staging 脚本调用 Gradle wrapper 时因沙箱无权访问 `~/.gradle/wrapper/dists/gradle-8.5-bin/5t9huq95ubn472n8rpzujfbqh/gradle-8.5-bin.zip.lck` 而中止。最小修复方案是不修改 Gradle 或 Java agent 配置，仅在已授权访问现有 Gradle 缓存的环境中重跑同一构建命令。

## 剩余风险

JDBC gateway 下的 DDL 获取当前只覆盖 MySQL。PostgreSQL 及其他 JDBC 数据库仍会返回明确的不支持错误，需要后续按数据库方言实现 DDL 生成。
