# JDBC agent H2 端到端集成测试

## 背景

Go gateway、Java agent、离线驱动导入和 JDBC 查询能力已经分别具备单元测试。为了验证首版闭环，需要增加一个带 build tag 的端到端测试，真实启动 agent、导入 H2 driver package、连接内存数据库并完成查询、元数据和关闭会话。

## 范围

- 新增 `integration` build tag 测试 `TestJDBCAgentH2EndToEnd`。
- 测试内动态生成 H2 离线驱动包，并复用 `DriverInstallService` 导入。
- 新增真实 gRPC client 包装 `NewGRPCJdbcAgentClient`。
- Gradle agent 工程新增 Shadow 插件，生成可运行的 fat jar。
- 新增 `printH2Jar` Gradle task，让脚本获取实际解析的 H2 jar 路径。
- 新增 `scripts/test-jdbc-agent.sh`，串联 Java 测试、agent 打包和 Go integration 测试。

## 修改文件

- `internal/service/jdbc_grpc_client.go`
- `internal/service/jdbc_integration_test.go`
- `jdbc-agent/build.gradle`
- `scripts/test-jdbc-agent.sh`

## 验证

- 初次运行 `go test -tags=integration ./internal/service -run TestJDBCAgentH2EndToEnd -v`，失败于 `NewGRPCJdbcAgentClient` 未定义，符合失败测试预期。
- 实现真实 gRPC client 后，直接运行失败于缺少 `H2_JAR` 环境变量，符合脚本前置未完成状态。
- 初次运行 `./scripts/test-jdbc-agent.sh`，失败于 `shadowJar` task 不存在。
- 最小修复：新增 Shadow 插件并固定输出 `sshtools-jdbc-agent-all.jar`。
- 第二次运行脚本时 Java 测试和 `shadowJar` 通过，但 H2 jar 路径查找失败。
- 最小修复：新增 `printH2Jar` Gradle task，由 Gradle 输出实际 test runtime classpath 中的 H2 jar。
- 已运行 `./scripts/test-jdbc-agent.sh`，结果通过。

## 剩余风险

- 集成测试依赖本机 JDK 21 路径 `/Library/Java/JavaVirtualMachines/jdk-21.jdk/Contents/Home/bin/java`，后续需要配置化。
- Shadow 插件首次使用需要访问 Gradle 插件仓库，干净环境可能需要网络。
- 端到端测试当前覆盖 H2 内存数据库，SQLite 文件数据库仍需后续补充。
