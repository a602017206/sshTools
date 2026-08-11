# App 退出残留 JDBC agent 子进程

## 背景

关闭应用后，本机仍可能残留 `jdbc-agent`（Java）进程，数据库连接也未按预期释放。根因是：

1. `OnShutdown` 只调用了 `jdbcAgentSupervisor.Close()`，未先关闭数据库 / 原生数据库 / SSH·本地终端会话
2. agent 以独立进程启动，停止时仅对单 PID 发 `SIGKILL`，未按进程组清理，也未等待退出
3. 本地 shell 同样可能留下子进程

## 范围

- 扩展 `App.shutdown`：先关业务会话，再停 JDBC agent
- 为数据库 / 原生数据库 / SessionManager 增加 `CloseAll*`
- Unix 下 agent / 本地 shell 使用独立进程组，停止时 `SIGTERM` 进程组，超时再 `SIGKILL`，并等待退出
- 补充回归测试与本变更文档

## 修改文件

- `app.go`
- `app_shutdown_test.go`
- `app_jdbc_test.go`
- `internal/service/jdbc_agent_process.go`
- `internal/service/jdbc_agent_process_unix.go`
- `internal/service/jdbc_agent_process_windows.go`
- `internal/service/jdbc_agent_process_test.go`
- `internal/service/database_service.go`
- `internal/service/database_service_test.go`
- `internal/service/native_database.go`
- `internal/service/native_database_test.go`
- `internal/service/session_service.go`
- `internal/ssh/manager.go`
- `internal/ssh/manager_close_all_test.go`
- `internal/ssh/local_session_unix.go`
- `docs/changes/bugs/2026-08-11-app-shutdown-child-processes.md`
- `docs/development/2026-08-11-app-shutdown-child-processes.md`

## 验证

- `go test ./internal/service -run 'TestDatabaseServiceCloseAll|TestExecAgentProcessStop|TestAgentProcessManagerStop|TestNativeDatabaseServiceCloseAll' -count=1` 通过
- `go test . -run 'TestAppShutdown' -count=1` 通过
- `go test ./internal/ssh -run 'TestSessionManagerCloseAll' -count=1`：当前环境无法创建 PTY，测试 Skip（逻辑已接入 `CloseAllSessions`）

## 剩余风险

- 若进程被强制杀死（如 `kill -9` 主进程、`wails dev` 热重载中断），`OnShutdown` 仍可能来不及执行，历史孤儿 agent 需手工清理或下次启动前处理
- 本地 shell `CloseAll` 的真实 PTY 路径未能在本环境实测
