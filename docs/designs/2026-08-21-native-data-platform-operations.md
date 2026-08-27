# Redis、Elasticsearch 与 Kafka 原生运维能力设计

## 背景

当前原生数据平台连接仅能测试连通性并浏览少量资源：Redis 只能枚举键名，Elasticsearch 只能枚举索引，Kafka 只能枚举 Topic。用户无法在客户端查看对象详情，也无法完成最基本的只读诊断操作，因此与 JDBC 工作区的可用性差距明显。

## 目标与范围

本次仅增加安全的只读能力，不引入删除、写入、生产消息或修改集群配置的操作。

- Redis：读取键类型、TTL 和受限长度的值预览。
- Elasticsearch：读取索引统计和受限数量的文档预览。
- Kafka：读取 Topic 分区、Leader、副本和 ISR 元数据。
- 统一通过现有原生数据库会话与 Wails bindings 暴露对象详情。
- 原生工作区在用户选择资源后显示详情，并保留既有资源树。

## 架构与取舍

在 `NativeDatabaseClient` 上增加 `DescribeResource`，使服务层和 UI 只依赖统一的详情模型；各 provider 保持各自协议客户端与具体实现。详情的值以 JSON 字符串传输，避免把 Redis 二进制值、ES 文档和 Kafka 元数据混入 UI 的协议判断。

预览会设置显式上限：Redis 值最多 4 KiB、Elasticsearch 最多 20 条文档；Kafka 不消费消息，只请求元数据。这样既满足日常排障，也避免误把客户端变成无边界的数据导出或消息消费工具。

## 风险

不同服务版本和权限模型可能拒绝部分只读 API。错误将原样包裹并显示在详情区域，资源浏览不受影响。TLS、SASL、Redis ACL、ES API key 等认证扩展不在本次范围内。
