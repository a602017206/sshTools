# JDBC 字段精度与 MySQL 系统表修复

## 背景

JDBC 列协议缺少长度、精度和默认值，导致 PostgreSQL 兼容 DDL 不够细致；MySQL `information_schema` 的系统表被类型过滤遗漏。

## 范围

扩展列元数据协议并生成 Java/Go bindings；DDL 输出长度与默认值；表列表包含 `SYSTEM TABLE`。

## 修改文件

- `jdbc-agent/src/main/proto/jdbc_agent.proto`
- Java agent、Go gateway、列模型与测试
- `internal/service/jdbcproto/`
- 本设计和变更记录。

## 验证

已通过 Go JDBC 测试与 `./gradlew test --tests '*MetadataServiceImplTest'`。

## 剩余风险

厂商驱动的默认值文本格式可能不同，应用保持原样展示。
