# JDBC agent 运行时监管

## 背景

应用原先直接持有静态 `AgentProcessManager`，没有根据当前 JRE 惰性启动 agent，也没有把进程返回的端口和 token 绑定到真实 gRPC client。重复调用还可能重复启动进程或遗留连接。

## 范围

- 新增线程安全的 `JDBCAgentSupervisor`。
- 首次请求时选择运行时、启动 agent 并连接 `127.0.0.1` 动态端口。
- 缓存带 token 的 client，后续请求复用同一连接。
- 重启和关闭时先关闭 gRPC 连接，再停止 Java 进程并清空状态。
- 为 `AgentProcessManager` 增加动态配置启动入口和并发保护。

## 修改文件

- `internal/service/jdbc_agent_process.go`
- `internal/service/jdbc_agent_supervisor.go`
- `internal/service/jdbc_agent_supervisor_test.go`
- `docs/changes/features/2026-07-10-jdbc-agent-supervisor.md`

## 验证

- 红灯：`go test ./internal/service -run TestJDBCAgentSupervisor -v` 在实现前因 `NewJDBCAgentSupervisor` 未定义而失败。
- 绿灯：`go test ./internal/service -run TestJDBCAgentSupervisor -v` 通过。
- 回归：`go test ./internal/service` 通过。

## 剩余风险

- supervisor 当前通过 gRPC 拨号成功判断 agent 就绪，尚未单独调用 health RPC 校验 agent 版本。
- agent 进程异常退出只能在下一次 gRPC 调用时被发现；主动进程退出通知在后续 managed gateway 任务处理。
