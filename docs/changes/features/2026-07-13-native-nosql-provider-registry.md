# 原生 NoSQL Provider 注册表

## 背景

常用 NoSQL 和 Kafka 需要按稳定类型 ID 逐步加入，但本次不实现外部插件加载。

## 范围

新增线程安全的内置 provider 注册表，支持注册、查询和防御性快照。注册表不使用动态库、Go `plugin`、Shell 或系统专有接口。

## 修改文件

- `internal/service/native_database_registry.go`
- `internal/service/native_database_registry_test.go`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestNativeDatabaseRegistry -v`，覆盖注册、查询、重复 ID、空 ID 和未知 ID。

## 剩余风险

注册表目前只提供内置扩展点；若未来重新启用第三方插件，需要单独设计跨 Windows、macOS、Linux 的分发和安全模型。
