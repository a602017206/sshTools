# JDBC Agent 日志文件无输出

## 背景

真实桌面验收中日志对话框可以正常打开，但 Agent 多次启动和退出后固定日志文件仍不存在。根因是子进程启动器没有把 Java Agent 的标准输出和标准错误连接到约定的 `jdbc-agent.log`。

## 范围

- 在 Agent 启动配置中增加固定日志路径。
- supervisor 把 `NewJDBCLogPaths(paths).Agent` 传给进程管理器。
- 默认命令运行器以追加模式打开日志，目录权限为 `0700`，文件权限为 `0600`。
- Java 子进程的 stdout 和 stderr 共用日志文件，进程结束后由等待协程关闭文件句柄。
- 即使 Java Agent 没有输出，Go 启动器也会写入不含 token 和数据库信息的启动、退出记录。
- 已退出进程的 `Stop()` 按幂等成功处理，避免等待协程回收后阻断 Agent 崩溃恢复。
- 自定义测试 runner 不支持输出接线时保持原有启动行为。

## 修改文件

- `app.go`
- `internal/service/jdbc_agent_process.go`
- `internal/service/jdbc_agent_process_test.go`
- `internal/service/jdbc_agent_supervisor.go`
- `internal/service/jdbc_agent_supervisor_test.go`

## 验证

- 修复前桌面日志对话框显示 `0 B`，固定日志文件不存在。
- 新增 `TestAgentProcessManagerPassesConfiguredLogPath`，实现前因启动配置缺少 `LogPath` 而失败。
- 修改 supervisor 启动测试，要求固定日志路径传入进程配置。
- 新增 `TestExecAgentCommandRunnerWritesLifecycleLog`，实现前真实子进程日志文件为空，增加生命周期记录后通过。
- 最终集成回归曾因已回收进程返回 `os.ErrProcessDone` 而失败；增加幂等关闭断言并最小修复后重跑原命令。
- 实现后两个定向测试均通过。
- 运行 `go test ./...`，全部 Go 测试通过。
- 运行 `./scripts/test-jdbc-agent.sh`，真实 Java、gRPC、H2 与崩溃恢复集成测试通过；Gradle 保留既有的 9.0 兼容性弃用警告。
- 运行 `/Users/dingwei/go/bin/wails build`，生产应用打包成功。
- 真实桌面复验中日志对话框显示启动记录、文件大小和可用复制按钮；刷新与复制成功。
- 向本次验收日志追加 70000 个测试字节后，对话框显示“仅显示最近 64 KiB”；验收结束后已删除测试日志。

## 剩余风险

- 日志采用追加写入，不包含轮转策略；日志查看仍只读取受限尾部，长期磁盘管理需要后续单独设计。
