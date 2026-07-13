# 原生 NoSQL 连接与 JDBC 驱动安全卸载设计

## 背景

JDBC 驱动管理页面的卸载操作直接删除驱动目录，没有判断已保存连接和活动会话是否依赖该 profile。达梦与人大金仓等多版本驱动并存时，页面还以驱动整体状态代替所选 profile 状态，可能展示错误的卸载动作。

Redis、MongoDB、Elasticsearch 是通用网络协议服务，不应经由 Java、JDBC agent 或 JDBC 驱动加载。当前数据库界面只适用于关系型数据库的 SQL、数据库和表，需要为三类服务提供各自的资源浏览模型。

## 目标与范围

本次实现以下首版能力：

- 禁止卸载仍被已保存数据库连接或活动 JDBC 会话使用的 JDBC profile，并返回可操作的中文错误信息。
- 按所选 JDBC profile 的安装状态展示安装、校验和卸载操作。
- 新增 Redis、MongoDB、Elasticsearch 原生连接测试、连接、关闭和资源浏览。
- Redis 浏览逻辑数据库和键；MongoDB 浏览数据库和集合；Elasticsearch 浏览索引。
- 保持首版只读：不提供 Redis 写命令、MongoDB 写入/删除或 Elasticsearch 写入/删除。

本次不实现 Redis Cluster、Sentinel，MongoDB SRV/副本集高级发现，Elasticsearch Cloud ID，数据编辑、批量导入导出或查询编辑器。

## 架构决策

### JDBC 卸载保护

`App.RemoveJDBCDriver` 在删除目录前调用单一的引用检查函数。该函数读取已保存的 `ConnectionConfig`，并读取 `ManagedJDBCGateway` 保存的活动 `DatabaseConfig`。

连接的 `metadata.db_type` 必须等于请求的 driver ID。显式设置 `metadata.driver_profile_id` 时只匹配该 profile；旧连接未设置 profile 时，按当时/当前 catalog 的推荐 profile 解析。错误中列出最多若干连接名称，并提示先修改连接版本或关闭会话。目录不存在时仍视为幂等卸载，不把它转换为错误。

### 原生 NoSQL 服务

新增 `NativeDatabaseService`，只负责 Redis、MongoDB、Elasticsearch。它维护独立会话表，以 `NativeDatabaseSession` 保存服务类型、显示名称和具体客户端。JDBC 的 `DatabaseService` 和 Java agent 不被调用。

服务 API 采用最小通用形状：测试连接、打开/关闭会话、列出一级资源、列出二级资源。返回值使用 `NativeResource`，带有稳定的 `kind` 和 `name` 字段，前端据此显示，而不将 Redis 键、Mongo 集合、ES 索引伪装成 SQL 表。

| 类型 | 一级资源 | 二级资源 | 原生客户端 |
| --- | --- | --- | --- |
| Redis | 逻辑数据库编号 | 键名 | `github.com/redis/go-redis/v9` |
| MongoDB | 数据库 | 集合 | `go.mongodb.org/mongo-driver` |
| Elasticsearch | 索引 | 无 | `github.com/elastic/go-elasticsearch/v9` |

连接配置继续使用 `type: "database"`，`metadata.db_type` 的新值为 `redis`、`mongodb` 和 `elasticsearch`。`metadata.database` 对 Redis 是可选逻辑数据库编号，对 MongoDB 是可选默认数据库，对 Elasticsearch 不使用。凭据沿用现有加密凭据存储。

### 前端路由

新增一种 `native-database` 面板：左侧资产和连接弹窗仍显示为数据库连接，连接时根据 `db_type` 选择 JDBC 或原生 Wails API。原生面板只显示资源树、刷新和关闭，不显示 SQL 编辑器、表字段或 DDL。关系型数据库的现有面板不改变。

## 错误处理与安全

- 所有原生连接使用带超时的 `context.Context`；错误保持底层原因并增加中文服务名和动作上下文。
- Redis 密码和 MongoDB/Elasticsearch 用户凭据不写入配置 JSON，仍由现有凭据存储保存。
- Elasticsearch 首版支持 HTTP/HTTPS 基本认证；TLS 默认使用系统证书校验。
- 驱动卸载保护只比较 profile 精确引用，不会阻止删除未被选中的其他版本。

## 验证策略

- JDBC 卸载测试覆盖显式 profile、旧连接回退到推荐 profile、活动会话和未引用 profile。
- 原生服务通过可注入客户端工厂进行单元测试，不依赖外部 Redis、MongoDB 或 Elasticsearch。
- 前端构建验证 Wails bindings 和 Svelte 编译；手工检查三类连接的表单字段、连接状态、资源树与关闭动作。

## 剩余风险

- 原生服务首版不支持集群和高级身份认证，遇到这些部署需在后续按服务类型扩展。
- 存量数据库连接缺少 `driver_profile_id` 时依据当前推荐 profile 判断引用；若用户在保存后改变推荐版本，需要在连接编辑页显式保存 profile 才能完全消除歧义。
