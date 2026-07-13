# 原生 NoSQL 跨平台 Provider 设计

## 背景

原需求从 Redis、MongoDB、Elasticsearch 扩展为常用 NoSQL 和 Kafka，并明确要求 Windows、macOS、Linux 都可使用。插件模式被确定为非必需能力。

## 范围

设计使用内置 provider 注册表，不实现外部插件加载。首批覆盖 Redis/KeyDB、MongoDB、Elasticsearch/OpenSearch、Memcached、Cassandra、Couchbase、InfluxDB、Neo4j 和 Kafka。

## 修改文件

- `docs/designs/2026-07-13-native-nosql-and-jdbc-driver-removal.md`
- `docs/plans/2026-07-13-native-nosql-and-jdbc-driver-removal-implementation.md`
- 本变更记录。

## 验证

已审查 provider 方案不使用 Go `plugin`、Unix socket、Shell 或 macOS 专有接口，连接模型和 provider ID 在三类目标系统保持一致。

## 剩余风险

SDK 可用性与服务端版本兼容性将在各 provider 的实现任务中逐一验证；外部插件需求若重新启用，需要单独设计跨系统分发与安全机制。
