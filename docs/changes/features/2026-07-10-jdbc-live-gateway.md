# 应用接入可恢复 JDBC gateway

## 背景

应用启动时原先向 `DatabaseService` 注入的是 nil client gateway，`RestartJDBCAgent` 也只启动进程而不建立新的 gRPC client。数据库测试连接仍走 Go 原生驱动，因此 Oracle、SQLite 和国产数据库无法使用统一 JDBC 路径。

## 范围

- 新增 `ManagedJDBCGateway`，保存成功连接配置并在 agent 不可用时恢复目标 session 后重试一次。
- `DatabaseService.TestConnection` 在配置 gateway 时使用临时 JDBC session。
- Wails 启动时读取嵌入的 agent jar，原子部署并构造真实 supervisor/gateway。
- profile resolver 补齐驱动实际安装目录。
- agent 重启统一走 supervisor，Wails 退出时关闭 gRPC client 和 Java 进程。

## 修改文件

- `app.go`
- `app_jdbc_test.go`
- `main.go`
- `internal/service/database_service.go`
- `internal/service/database_service_test.go`
- `internal/service/jdbc_managed_gateway.go`
- `internal/service/jdbc_managed_gateway_test.go`
- `docs/changes/features/2026-07-10-jdbc-live-gateway.md`

## 验证

- 红灯：service 定向测试在实现前因 `NewManagedJDBCGateway` 未定义而失败；应用测试因 `buildJDBCServices` 未定义而失败。
- 绿灯：`go test ./internal/service -run 'TestManagedJDBCGateway|TestDatabaseServiceTestConnectionUsesGateway' -v` 通过。
- 绿灯：`go test . -run TestBuildJDBCServicesInjectsManagedGateway -v` 通过。
- 回归：`go test ./...` 通过。

## 剩余风险

- 自动恢复只重建连接，不恢复事务、临时表或 session 级数据库状态。
- 运行时模式切换目前仍需要后续任务保证 supervisor 使用同一个可变运行时配置。
- catalog 内置清单和在线驱动安装将在后续任务补齐；清单不存在时 profile resolver 仍会返回驱动缺失。
