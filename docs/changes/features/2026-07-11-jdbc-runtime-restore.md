# JDBC 运行时选择恢复

## 背景

运行时选择已经写入应用配置，但 `RuntimeService` 仍使用单一布尔优先级，无法区分显式托管模式和自动选择，也不能在应用启动时安全恢复失效的系统 Java 配置。

## 范围

- 增加自动、托管和系统 Java三种运行时模式。
- 显式托管模式不回退系统 Java，自动模式保留原有托管优先规则。
- 系统 Java模式校验路径存在且为普通文件。
- 增加线程安全的模式快照和恢复能力。
- 应用构建 JDBC 服务时恢复持久化选择，但不主动启动 Agent。
- 失效或未知的持久化配置返回初始化错误，并保持运行时状态为 `missing`，不静默回退。

## 修改文件

- `internal/service/jdbc_runtime.go`
- `internal/service/jdbc_runtime_test.go`
- `app.go`
- `app_jdbc_test.go`

## 验证

- `go test ./internal/service -run 'TestRuntimeService(RestoresExplicitSystemJava|RejectsInvalidExplicitSystemJava|RestoresSnapshot)' -v`
- `go test . -run 'TestBuildJDBCServices(RestoresPersistedRuntimeMode|KeepsInvalidPersistedSystemRuntimeMissing)' -v`
- `go test ./internal/service -v`

## 剩余风险

- 当前只校验 Java 路径为普通文件，尚未执行 `java -version` 验证版本和可启动性；Agent 重启阶段会提供最终验证。
- 托管 JRE 仍按目录名称排序选择，没有固定具体版本。
