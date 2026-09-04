# 功能：数据库 Schema 右键菜单与运行 SQL 文件

## 背景

Navicat 可在库/Schema 节点上新建查询、刷新、断开，并直接执行上百 MB 的初始化脚本。当前应用里该节点没有右键；查询页和 `ExecuteDatabaseQuery` 按单条 SQL、30 秒、结果表格设计，不能承载 91MB 级脚本。

## 范围

- 在左侧 `DatabaseSidebarTree` 的库（MySQL）或 Schema（Oracle/PostgreSQL）节点上增加右键菜单：新建查询、运行 SQL 文件…、刷新、断开。
- 选完 `.sql` 文件后由 Go 后台流式拆句并经 JDBC `ExecuteQuery` 逐条执行，不把文件内容填进查询编辑器。
- 推送 `sqlfile:progress` 进度，支持取消；单条超时 5 分钟；gRPC 消息上限 64MB。
- 不做转储、命令行、数据字典、逆向模型；原生 Redis/ES 树不在本批范围。

## 修改文件

- `docs/designs/2026-09-01-database-schema-sql-file.md`
- `docs/development/2026-09-01-database-schema-sql-file.md`
- `internal/service/sql_script.go`
- `internal/service/sql_script_test.go`
- `internal/service/sql_file.go`
- `internal/service/database_service_test.go`
- `internal/service/jdbc_grpc_client.go`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/JdbcAgentApplication.java`
- `sql_file.go`
- `app.go`
- `frontend/src/lib/databaseSchemaMenu.js`
- `frontend/test/databaseSchemaMenu.test.js`
- `frontend/src/components/DatabaseSidebarTree.svelte`
- `frontend/src/components/SQLFileProgressDialog.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/App.svelte`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`

## 验证

```bash
go test ./internal/service -count=1 -run 'TestSplitSQLScript|TestSQLFilePreamble|TestDatabaseServiceExecuteSQLFile'
cd frontend && node --test test/databaseSchemaMenu.test.js
```

Go 相关测试通过（`ok AHaSSHTools/internal/service`）；前端 2 项菜单与进度归一化测试通过。未在桌面应用内对真实 91MB Oracle/MySQL 初始化脚本做手工执行验证。

## 剩余风险

存储过程或 PL/SQL 块中夹杂分号时，第一版拆句可能把过程体切碎。逐条 gRPC 比 Navicat 进程内执行慢。同一会话同时只能跑一份 SQL 文件。Wails 绑定需在下次 `wails dev` 时与导出方法对齐。
