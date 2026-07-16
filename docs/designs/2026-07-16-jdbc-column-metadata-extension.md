# JDBC 列元数据扩展设计

## 背景

当前 JDBC agent 的 `Column` 协议仅包含字段名、类型、可空和主键。PostgreSQL、人大金仓、openGauss 的表结构展示因此缺少长度、精度、小数位和默认值；MySQL 的 `information_schema` 对象还因仅过滤 `TABLE` 类型而被遗漏。

## 方案

在 `Column` protobuf 消息中追加字段长度、十进制小数位和默认值，并保持原有字段编号不变。Java agent 从 `DatabaseMetaData.getColumns` 读取 `COLUMN_SIZE`、`DECIMAL_DIGITS`、`COLUMN_DEF`；Go gateway 映射到内部列模型，DDL 生成按数据库类型补全长度、精度和默认值。

`ListTables` 的对象类型过滤扩展为 `TABLE` 与 `SYSTEM TABLE`，使 MySQL `information_schema` 可见。该变更不改变用户表的既有列表行为。

## 工具链记录

修改 protobuf 后必须执行 Gradle 的 `generateProto` 生成 Java 文件，并使用仓库既有 protobuf 生成流程更新 Go bindings。最小修复方案是使用项目锁定的 Gradle wrapper 与现有 `protoc` 配置生成，不手工编辑生成的 `jdbcproto` 文件，不跳过 agent 构建。

本次执行时，Homebrew `protoc 33.2` 因缺失 `libabsl_die_if_null` 退出。Gradle 的 `generateProto` 已成功，因此最小修复是把 `PROTOC` 指向 Gradle 缓存中已下载的 `protoc 4.28.3 osx-aarch_64` 后重跑同一个 `scripts/generate-jdbc-proto.sh`，不升级 Homebrew 依赖或改写生成脚本。

## 验证

Java 元数据测试验证长度、精度、默认值和系统表类型；Go gateway 测试验证映射；Gradle 测试和前端构建验证生成的 agent 可用。

## 风险

不同 JDBC 驱动对 `COLUMN_DEF` 的文本格式可能不同。应用只原样显示默认值，不解析或执行默认表达式。
