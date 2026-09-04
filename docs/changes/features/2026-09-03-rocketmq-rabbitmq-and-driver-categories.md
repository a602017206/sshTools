# 变更：RocketMQ / RabbitMQ 与驱动分组

## 背景

需要补齐消息队列类型，并参考 dbx 为连接类型与 JDBC 驱动增加分组标签；MQ 不走 Java 驱动安装。

## 范围

- 连接类型细分组：关系型 / 国产 / 轻量 / 文档·缓存·检索 / 图谱·时序 / 消息队列
- `rocketmq`、`rabbitmq` 原生连接（无认证可选）+ 最小运维工作区
- JDBC 驱动 manifest / 管理页增加 category 侧栏（明确 MQ 无需安装 JDBC）

## 修改文件

- `frontend/src/lib/databaseTypeCatalog.js`（新）
- `frontend/src/lib/assetDomain.js`、`nativeDatabaseTypes.js`、`nativeDatabaseWorkspace.js`、`databaseTypeIcon.js`、`copilotContext.js`
- `frontend/src/components/AddAssetDialog.svelte`、`JDBCDriverManager.svelte`、`workspaces/KafkaWorkspace.svelte`
- `internal/config/jdbc.go`、`internal/service/jdbc_builtin_manifest.json`
- `internal/service/native_rabbitmq.go`、`native_rocketmq.go` 及测试
- `internal/service/native_database.go`、`copilot/tools.go`、`app.go`
- 文档：`docs/designs/2026-09-03-mq-types-and-driver-categories.md`

## 验证

- `cd frontend && node --test test/databaseTypeCatalog.test.js test/assetDomain.test.js test/nativeDatabaseTypes.test.js test/nativeDatabaseWorkspace.test.js`
- `go test ./internal/service/ -run 'RabbitMQ|RocketMQ|Kafka' -count=1`

## 剩余风险

- RabbitMQ 无 Management 插件时仅 AMQP 连通成功，列表会失败
- RocketMQ ACL / 多 NameServer 组合尚未覆盖
