# JDBC Agent 退出后状态未刷新

## 背景

真实桌面验收中终止 JDBC Agent 子进程后，进程已经不存在，但驱动管理页面经过多个 2 秒轮询周期仍显示“运行中”。前端轮询正常执行，后端状态 API 返回了 supervisor 的缓存状态。

## 范围

- 为 Agent 进程句柄增加存活状态查询，并由 `exec.Cmd.Wait` 回收进程、发布退出状态。
- 进程管理器健康检查在子进程退出后返回错误并清除失效句柄。
- supervisor 在运行状态刷新时执行健康检查，关闭失效 gRPC 连接并把状态更新为 `failed`。
- App 状态 API 返回刷新后的状态，同时保持 `starting` 等非运行状态查询不被启动锁阻塞。

## 修改文件

- `internal/service/jdbc_agent_process.go`
- `internal/service/jdbc_agent_process_test.go`
- `internal/service/jdbc_agent_supervisor.go`
- `internal/service/jdbc_agent_supervisor_test.go`
- `app.go`

## 验证

- 修复前真实终止 Agent 后，`pgrep` 无进程但页面仍显示“运行中”。
- 新增 `TestAgentProcessManagerHealthDetectsExitedProcess` 和 `TestJDBCAgentSupervisorRefreshStatusDetectsExitedProcess`；实现前 supervisor 测试因缺少刷新方法失败。
- 实现后两个定向测试通过。
- `TestJDBCManagementAPIReturnsAgentAndRuntimeState` 验证 `starting` 状态查询不阻塞并通过。
- 沙箱内首次运行测试时无法访问 Go 构建缓存；最小修复是在沙箱外重跑相同命令，没有修改 Go 工具链。
- 运行 `go test ./...`，全部 Go 测试通过。
- 运行 `./scripts/test-jdbc-agent.sh`，真实 Java、gRPC、H2 与崩溃恢复集成测试通过；Gradle 仍报告既有的 9.0 兼容性弃用警告。
- 运行 `/Users/dingwei/go/bin/wails build`，包含修复的生产应用打包成功。
- 真实桌面复验中终止运行中的 Agent，页面通过 2 秒轮询自动显示“启动失败 / JDBC agent 进程已退出”。

## 剩余风险

- 状态刷新依赖页面 2 秒轮询，进程退出到界面更新之间仍存在最多一个轮询周期的延迟。
