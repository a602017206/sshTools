# DatabaseService 切换到 JDBC gateway

## 背景

现有数据库功能直接依赖 Go `database/sql` 和本地驱动。JDBC agent 架构要求保留原有 Wails API 方法名称，同时把连接、查询、表列表、字段元数据和关闭会话路由到新的 `JdbcGatewayService`。

## 范围

- 新增 `DatabaseGateway` 接口，定义数据库服务和 JDBC gateway 之间的适配边界。
- 新增 `NewDatabaseServiceWithGateway` 构造函数。
- `DatabaseService` 在配置 gateway 时优先委托连接、查询、列表、字段元数据和关闭操作。
- 保留无 gateway 时的 legacy `database/sql` 回退路径，降低单元测试和兼容风险。
- `app.go` 启动时构建 JDBC 路径、驱动清单服务和 JDBC gateway 骨架，并注入 `DatabaseService`。
- `database_drivers.go` 增加说明，保留 legacy 驱动注册。

## 修改文件

- `app.go`
- `internal/service/database_service.go`
- `internal/service/database_service_test.go`
- `internal/service/database_drivers.go`
- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_gateway_test.go`

## 验证

- 已运行 `go test ./internal/service -run 'TestDatabaseServiceDelegates|TestDatabaseService_CloseDatabase' -v`，结果通过。
- 已运行 `go test ./...`，结果通过。

## 剩余风险

- `app.go` 当前注入的 JDBC gateway 仍未连接真实 agent client，未配置 client 时会返回 `AGENT_UNAVAILABLE`。
- 驱动 profile resolver 先读取推荐 profile，尚未合并本地安装状态。
- `GetTableDDL` 仍保留 legacy 实现，后续如需 JDBC agent 支持 DDL 需要扩展协议。
