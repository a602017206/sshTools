# MySQL 测试连接握手超时提示不明确

## 背景

测试 MySQL 连接时，界面只显示 `[DB_CONNECT_FAILED] rpc error: code = DeadlineExceeded desc = context deadline exceeded`。本地复现确认两个目标都在固定 10 秒截止时间失败，但 TCP 端口可以建立连接。

进一步使用托管 JRE 的原始 Socket 探针发现：Java 可以完成 TCP 连接，但保持连接时始终收不到 MySQL 服务端首个握手包；只有客户端先关闭写端后，目标端口才返回 MySQL 握手和 `08S01`。这不符合 MySQL 服务端先发送握手的协议流程，通常由等待客户端数据后才连接上游的 TCP 代理、端口转发或错误缓冲导致。

## 范围

- MySQL 测试连接向 Connector/J 传入 8 秒 `connectTimeout` 和 `socketTimeout`，确保驱动在 Go 的 10 秒外层截止前返回。
- 将 gRPC `DeadlineExceeded`、Go 上下文截止和 Connector/J“未收到任何包”的通信失败统一映射为中文握手超时提示。
- 错误提示包含目标主机与端口，并明确要求使用 MySQL 原生端口或支持服务端先发送握手的 TCP 转发。
- 不记录、不显示或修改数据库密码。

## 修改文件

- `internal/service/database_service.go`
- `internal/service/database_service_test.go`

## 验证

- 修复前 `TestDatabaseServiceTestConnectionAddsMySQLHandshakeTimeouts` 因 properties 为空而失败。
- 修复前 `TestDatabaseServiceTestConnectionExplainsMySQLHandshakeTimeout` 因原样返回 `context deadline exceeded` 而失败。
- TCP 探针确认 `192.168.121.158:22306` 与 `192.168.1.2:22306` 可建立连接，但托管 JRE 在读取服务端握手时超时。
- 实现后三个定向测试均通过，确认 MySQL 测试连接携带 8 秒驱动超时，并把截止错误及 Connector/J“未收到任何包”错误转换为包含目标地址、原生端口和代理要求的中文提示。
- 运行 `go test ./...`，全部 Go 测试通过。
- 修复后使用保存凭据执行真实目标复验时，沙箱外审批因额度限制被拒绝；没有绕过审批，真实目标的最终提示仍待用户在应用中重新点击确认。

## 剩余风险

- 此修复改进超时边界和诊断信息，不会伪造连接成功，也不能修复外部 TCP 代理的协议死锁。
- 目标端口需要由网络或数据库管理员调整为标准 MySQL 透传行为后，JDBC 连接才能成功。
