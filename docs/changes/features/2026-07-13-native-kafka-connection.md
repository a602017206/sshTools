# Kafka 原生连接

## 背景

Apache Kafka 仅维护 Java 客户端，Go 使用社区客户端。选择 `franz-go v1.20.7`，它是纯 Go 的 Kafka 协议实现，兼容项目 Go 1.24，不依赖 `librdkafka` 或任意平台动态库。

## 范围

新增 Kafka provider，通过 broker 的 `Ping` 校验连通性，并发出只读 Metadata 请求列出非内部 topic。应用与前端静态注册 Kafka，默认端口为 `9092`。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_kafka.go`
- `internal/service/native_kafka_test.go`
- `app.go`
- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- `go.mod`
- `go.sum`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestKafkaNative -v`，验证 broker 连通性、topic 浏览、错误传播和关闭；执行前端类型测试验证原生路由。

## 工具链阻塞与最小修复

首次尝试 `franz-go v1.21.0` 时，该版本要求 Go 1.25，并将 `golang.org/x/crypto` 等传递依赖升级到 Go 1.25 基线。项目固定 Go 1.24 后，测试报出 `golang.org/x/crypto@v0.50.0 requires go >= 1.25.0`。最小修复为固定 `franz-go v1.20.7` 及其元数据模块，并将此前被升级的 `x/crypto`、`x/net`、`x/sys`、`x/text`、`x/sync`、`klauspost/compress` 还原到 Go 1.24 可编译版本；不升级项目 Go 指令。

## 剩余风险

首版只支持单个明文 broker，不暴露 SASL、TLS、多个 seed broker 或 consumer group。无描述 topic 权限的账号将只看到服务端允许返回的元数据。
