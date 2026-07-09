# JDBC agent Go 网关

## 背景

Java agent 已能完成健康检查、查询和元数据操作后，Go/Wails 后端需要具备启动本地 agent、生成访问 token、调用 gRPC 接口并把结果转换为现有数据库 API 返回结构的能力。

## 范围

- 新增 `AgentProcessManager`，负责选择本地端口、生成 32 字节 token，并通过 Java `-jar` 启动 agent。
- 新增可替换的命令 runner，单元测试可用 fake runner 避免启动真实 Java 进程。
- 新增 `scripts/generate-jdbc-proto.sh`，统一生成 Go gRPC client。
- 给 `jdbc_agent.proto` 增加 Go package 配置，并生成 `internal/service/jdbcproto`。
- 新增 `JdbcGatewayService`，封装 agent client 的连接、查询、表、字段和关闭会话调用。
- 查询结果从 agent 的字符串行转换为现有 `QueryResult`。
- 新增基础 `JDBCError`，先覆盖 `DRIVER_MISSING` 映射。

## 修改文件

- `go.mod`
- `go.sum`
- `jdbc-agent/src/main/proto/jdbc_agent.proto`
- `scripts/generate-jdbc-proto.sh`
- `internal/service/jdbc_agent_process.go`
- `internal/service/jdbc_agent_process_test.go`
- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_gateway_test.go`
- `internal/service/jdbcproto/jdbc_agent.pb.go`
- `internal/service/jdbcproto/jdbc_agent_grpc.pb.go`

## 验证

- 已运行 `go test ./internal/service -run 'TestAgentProcessManager|TestJdbcGateway' -v`，结果通过。
- 初次检查发现系统 `protoc` 因 Homebrew `abseil` 动态库缺失无法启动。
- 初次检查发现 `protoc-gen-go` 和 `protoc-gen-go-grpc` 不在当前 `PATH`。
- 最小修复：复用 Gradle 缓存中的 `protoc-4.28.3-osx-aarch_64.exe`，并在生成脚本中追加 `$(go env GOPATH)/bin` 到 `PATH`。
- gRPC 依赖初次补齐时拉取到需要 Go 1.25 的 `genproto`，与仓库 Go 1.24 不兼容。
- 最小修复：固定 `google.golang.org/grpc` 到本地已有且兼容 Go 1.24 的 `v1.65.0`，并固定 `google.golang.org/genproto/googleapis/rpc` 到兼容版本。

## 剩余风险

- `AgentProcessManager.Health` 目前只检查进程句柄存在，尚未实际调用 agent `Health` RPC。
- `JdbcGatewayService` 的真实 gRPC 连接创建尚未接入 agent 地址，当前单元测试通过 fake client 覆盖转换逻辑。
- `ListDatabases` 暂未在 Java agent 协议中实现，当前 Go gateway 返回未实现错误。
- proto 生成依赖本机可用 `protoc`，干净机器需要先修复或安装 protobuf 工具链。
