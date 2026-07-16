# PostgreSQL 兼容数据库 Schema 对象树

## 背景

PostgreSQL、人大金仓和 openGauss 的对象组织方式包含 Schema 层级。原有界面只展示数据库和表，无法表达 Schema、视图、物化视图、存储过程、函数及扩展等对象分类。

## 范围

新增 PostgreSQL 兼容数据库对象树组件。组件从 `pg_catalog.pg_namespace` 动态读取 Schema，并在用户展开节点时按需读取表、视图、物化视图、存储过程、函数和扩展。表节点会携带 Schema 名称打开对应表结构，避免同名表读取到错误定义；MySQL 等非 PostgreSQL 类型继续使用原数据库浏览组件。

## 修改文件

- `frontend/src/lib/databaseObjectTree.js`
- `frontend/src/components/PostgreSQLObjectTree.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/src/App.svelte`
- `frontend/src/stores.js`
- `app.go`
- `internal/service/database_service.go`
- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_managed_gateway.go`
- `frontend/test/databaseObjectTree.test.js`
- `internal/service/database_service_test.go`
- `internal/service/jdbc_gateway_test.go`
- `docs/designs/2026-07-16-postgresql-schema-object-tree.md`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run 'TestJdbcGatewayGetTableSchemaInSchemaPassesSchemaToAgent|TestDatabaseServiceGetTableDDLInSchemaUsesRequestedSchema' -v`，验证 Schema 名称会传入 JDBC `ListColumns` 请求；执行 `node --test test/databaseObjectTree.test.js test/tableMetadataQuery.test.js`，验证 Schema/Object 查询与分类模型。首次执行 `npm run build` 时，Vite 编译通过，但 JDBC agent 暂存被 Gradle 缓存锁文件 `/Users/dingwei/.gradle/wrapper/dists/gradle-8.5-bin/.../gradle-8.5-bin.zip.lck` 的沙箱权限阻断。

## 剩余风险

对象查询依赖 PostgreSQL 系统目录兼容性和当前连接用户权限。当前第一版使用 SQL 读取对象，不同数据库的统一对象树 API 仍需后续由 JDBC agent 暴露，以覆盖 MySQL、Oracle、SQL Server、达梦等厂商特有对象。Gradle 阻塞的最小修复方案是在受控本机环境中允许读取和写入既有 `~/.gradle` 包装器缓存后，重跑原 `npm run build` 命令；不修改构建脚本或绕开 JDBC agent 暂存。
