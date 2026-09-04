# 开发记录：数据库 Schema 右键菜单与运行 SQL 文件

## 实现内容

- `SplitSQLScript` / `EachSQLStatement` 按 `;`（字符串与注释外）和 Oracle 独立行 `/` 流式拆句，单条上限 32MB。
- `ExecuteSQLFile` 打开本机文件后先按方言执行 `USE` / `CURRENT_SCHEMA` / `search_path`，再逐条走 JDBC `ExecuteQuery`；5 分钟单条超时；出错即停并带回语句摘要。
- `StartSQLFile` 在 goroutine 中执行并立即返回；`sqlfile:progress` 带 `sessionId`、已读字节、语句数；`CancelSQLFile` 取消该会话上下文。
- 客户端与 JDBC agent 的 gRPC 入站/调用消息上限提到 64MB。
- 侧栏库/Schema 右键：新建查询派发 `database:new-query`；运行 SQL 文件先 `SelectSQLFile` 再 `StartSQLFile`；刷新重载树；断开复用 `database:disconnect`。
- `App.svelte` 订阅进度并弹出可取消对话框；`TerminalPanel` 监听窗口 `database:new-query` 打开查询页。

## 验证

- `go test ./internal/service -count=1 -run 'TestSplitSQLScript|TestSQLFilePreamble|TestDatabaseServiceExecuteSQLFile'`：通过。
- `cd frontend && node --test test/databaseSchemaMenu.test.js`：2 项通过。
- `go build -o /dev/null .`：主包编译通过。

## 剩余风险

大脚本仍受 JDBC 单次结果/参数大小和拆句规则限制。未在真实桌面会话中跑 91MB 初始化脚本。
