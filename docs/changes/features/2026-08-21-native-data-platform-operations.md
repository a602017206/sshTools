# 原生数据平台只读详情能力

## 背景

Redis、Elasticsearch 与 Kafka 的原生连接此前仅支持连通性测试和资源名浏览，无法支撑对象级排障。

## 范围

新增统一的资源详情接口和界面详情面板。Redis 返回键类型、TTL 与最多 4 KiB 的字符串值预览；Elasticsearch 返回最多 20 条命中文档；Kafka 返回 Topic 分区、Leader、副本与 ISR 元数据。所有新增操作均为只读。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_redis.go`
- `internal/service/native_elasticsearch.go`
- `internal/service/native_kafka.go`
- `internal/service/native_resource_details_unsupported.go`
- `app.go`
- `frontend/src/components/NativeDatabasePanel.svelte`
- `frontend/src/lib/nativeDatabaseWorkspace.js`
- 对应 Go 与前端测试文件。

## 验证

新增 Redis、Elasticsearch、Kafka 的资源详情单元测试，并执行原生 provider 测试与前端构建验证。

## 剩余风险

当前未覆盖 TLS、Redis ACL 用户名、Elasticsearch API Key、Kafka SASL/SSL 等认证变体。Redis 非字符串键只展示类型与 TTL；Kafka 刻意不提供消息消费或生产。
