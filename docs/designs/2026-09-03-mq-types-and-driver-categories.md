# 设计：消息队列类型扩展与驱动分组

## 背景

参考 dbx 的驱动侧栏分组，需要把 RocketMQ / RabbitMQ 纳入消息队列域，并为 JDBC 驱动管理增加分类导航。本仓库 Kafka 已是 Go 原生路径，MQ 不应绑 Java/JRE。

## 决策

1. **MQ 走 Native Provider**：RocketMQ（NameServer Admin）、RabbitMQ（AMQP + Management API），与 Kafka 一致，不进入 JDBC 安装清单。
2. **分组拆两层**：
   - 连接类型：`databaseTypeCatalog` 细分类（关系型 / 国产 / 轻量 / 文档缓存检索 / 图谱时序 / 消息队列）
   - JDBC 驱动：manifest `category` + 驱动管理左侧栏（不含 MQ）
3. **第一期能力**：连通、列一级资源、只读详情；支持认证=无；不做生产消费与完整 SASL。

## 关键落点

- 前端：`databaseTypeCatalog.js`、`assetDomain.js`、`AddAssetDialog`、`JDBCDriverManager`、native types/workspace
- 后端：`native_rabbitmq.go`、`native_rocketmq.go`、`app.go` 注册、`jdbc.go` + manifest category

## 风险

- RabbitMQ 列队列依赖 Management 插件（默认 15672）
- RocketMQ 依赖 `rocketmq-client-go` Admin；ACL 凭据映射为 AccessKey/SecretKey
