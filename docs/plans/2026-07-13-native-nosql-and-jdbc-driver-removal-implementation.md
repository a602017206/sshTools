# 原生 NoSQL 连接与 JDBC 驱动安全卸载 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 安全卸载未被引用的 JDBC profile，并为常用 NoSQL 和 Kafka 提供跨 Windows、macOS、Linux 的不依赖 JDBC 的只读连接和资源浏览。

**架构：** JDBC 卸载在 `App` 层检查已保存连接和活动 gateway 会话。原生服务由独立的 `NativeDatabaseService` 和内置 provider 注册表实现，通过 Wails API 暴露连接与资源浏览；前端按 `metadata.db_type` 路由到原生资源面板。

**技术栈：** Go、Wails、Svelte、`github.com/redis/go-redis/v9`、`go.mongodb.org/mongo-driver`、`github.com/elastic/go-elasticsearch/v9`、Couchbase/InfluxDB/Neo4j/Kafka Go 客户端、Go `testing`。

---

## 实施约束

- 每个任务先写失败测试并运行确认失败，再写最小实现并运行确认通过，最后单独提交。
- 新增或修改文档正文全部使用中文。
- 任何 Gradle、protoc、gRPC 或 Java agent 工具链阻塞，先在对应变更记录说明阻塞点和最小修复方案，再重试命令。
- 原生 NoSQL 不调用 JDBC agent，不安装 JDBC 驱动。
- 本次不实现外部插件加载；provider 接口与注册表作为后续扩展点。

### 任务 1：保护被引用的 JDBC profile

**文件：** 修改 `app.go`、`app_jdbc_test.go`；创建 `docs/changes/bugs/2026-07-13-jdbc-driver-removal-protection.md`。

1. 写失败测试，覆盖已保存显式 profile、旧连接推荐 profile、活动 JDBC 会话阻止卸载，以及未引用 profile 可卸载。
2. 运行 `go test . -run TestRemoveJDBCDriver -v`，确认因缺少引用检查而失败。
3. 在 `App` 增加 JDBC profile 引用解析和中文错误；在 `ManagedJDBCGateway` 增加只读会话配置快照方法。
4. 再次运行同一命令，确认通过。
5. 写中文变更记录并运行 `go test ./...`。
6. 提交 `fix: protect jdbc drivers in use from removal`。

### 任务 2：修正驱动管理器的 profile 操作状态

**文件：** 修改 `frontend/src/components/JDBCDriverManager.svelte`；创建 `docs/changes/bugs/2026-07-13-jdbc-driver-profile-actions.md`。

1. 写前端状态单测，证明 V8 已安装、V9 未安装时 V9 显示安装而不是卸载。
2. 运行测试确认失败。
3. 提取所选 profile 的操作状态，使用 `selectedProfile.installed` 而非驱动总状态；把“被连接使用”的后端错误原样显示为可操作提示。
4. 运行前端单测确认通过。
5. 写中文变更记录并运行 `npm run build`；若 Java agent 构建受阻，先记录阻塞。
6. 提交 `fix: use selected jdbc profile state for actions`。

### 任务 3：建立原生 NoSQL 会话模型与通用服务

**文件：** 创建 `internal/service/native_database.go`、`internal/service/native_database_test.go`；修改 `go.mod`、`go.sum`；创建 `docs/changes/features/2026-07-13-native-nosql-session-model.md`。

1. 写失败测试，覆盖会话注册、类型校验、关闭和未知会话错误。
2. 运行 `go test ./internal/service -run TestNativeDatabaseService -v`，确认失败。
3. 实现 `NativeDatabaseService`、`NativeDatabaseSession`、`NativeResource` 与协议适配器接口；加入三个官方客户端依赖。
4. 再次运行同一命令，确认通过。
5. 写中文变更记录并运行 `go test ./internal/service`。
6. 提交 `feat: add native nosql session service`。

### 任务 4：实现 Redis 原生连接与键浏览

**文件：** 创建 `internal/service/native_redis.go`、`internal/service/native_redis_test.go`；修改原生服务文件；创建 `docs/changes/features/2026-07-13-native-redis-connection.md`。

1. 写失败测试，覆盖 Ping、逻辑数据库列表、按数据库扫描键和超时错误包装。
2. 运行 `go test ./internal/service -run TestRedisNative -v`，确认失败。
3. 使用 `go-redis/v9` 实现 Redis provider，限制扫描数量并使用连接超时上下文。
4. 再次运行同一命令，确认通过。
5. 写中文变更记录并运行 `go test ./internal/service`。
6. 提交 `feat: add native redis connection`。

### 任务 5：实现 MongoDB 原生连接与集合浏览

**文件：** 创建 `internal/service/native_mongodb.go`、`internal/service/native_mongodb_test.go`；修改原生服务文件；创建 `docs/changes/features/2026-07-13-native-mongodb-connection.md`。

1. 写失败测试，覆盖 Ping、数据库列表、集合列表和关闭客户端。
2. 运行 `go test ./internal/service -run TestMongoNative -v`，确认失败。
3. 使用官方 MongoDB Go Driver 实现 provider 和注入式工厂。
4. 再次运行同一命令，确认通过。
5. 写中文变更记录并运行 `go test ./internal/service`。
6. 提交 `feat: add native mongodb connection`。

### 任务 6：实现 Elasticsearch 原生连接与索引浏览

**文件：** 创建 `internal/service/native_elasticsearch.go`、`internal/service/native_elasticsearch_test.go`；修改原生服务文件；创建 `docs/changes/features/2026-07-13-native-elasticsearch-connection.md`。

1. 写失败测试，覆盖健康检查、索引列表、HTTP 错误转换和客户端关闭。
2. 运行 `go test ./internal/service -run TestElasticsearchNative -v`，确认失败。
3. 使用官方 Elasticsearch Go Client 实现 provider；解析 `_cat/indices` JSON 响应并只保留公开索引名称。
4. 再次运行同一命令，确认通过。
5. 写中文变更记录并运行 `go test ./internal/service`。
6. 提交 `feat: add native elasticsearch connection`。

### 任务 7：暴露原生 Wails API 并接入连接配置

**文件：** 修改 `app.go`、`app_jdbc_test.go` 或新建 `app_native_database_test.go`、`frontend/wailsjs/go/main/App.js`、`frontend/wailsjs/go/models.ts`；创建 `docs/changes/features/2026-07-13-native-nosql-wails-api.md`。

1. 写失败测试，覆盖三种类型不会进入 JDBC gateway、API 连接与资源列表路由正确。
2. 运行 `go test . -run TestNativeDatabase -v`，确认失败。
3. 初始化原生服务，暴露连接、测试、一级/二级资源列表和关闭 API；重新生成 Wails bindings。
4. 再次运行同一命令，确认通过。
5. 写中文变更记录并运行 `go test ./...`。
6. 提交 `feat: expose native nosql database APIs`。

### 任务 8：扩展连接表单、资产连接路由和原生资源面板

**文件：** 修改 `frontend/src/components/AddAssetDialog.svelte`、`frontend/src/App.svelte`、`frontend/src/components/AssetList.svelte`、`frontend/src/components/TerminalPanel.svelte`；创建 `frontend/src/components/NativeDatabasePanel.svelte`、前端测试文件；创建 `docs/changes/features/2026-07-13-native-nosql-ui.md`。

1. 写失败测试，覆盖三种类型不显示 JDBC profile、保存正确的 `db_type`、连接调用原生 API、资源树显示正确标签。
2. 运行对应前端测试确认失败。
3. 增加三种连接类型和默认端口；按类型调整认证和数据库字段；新建资源浏览面板并接入会话关闭流程。
4. 运行前端测试确认通过。
5. 写中文变更记录，运行 `npm run build` 与 `go test ./...`；若 Java agent 构建受阻，先记录阻塞。
6. 提交 `feat: add native nosql connection interface`。

### 任务 9：端到端验证与开发记录

**文件：** 创建 `docs/development/2026-07-13-native-nosql-and-jdbc-driver-removal.md`、`docs/changes/features/2026-07-13-native-nosql-completion.md`。

1. 运行完整 Go 测试、前端构建和必要的 Wails 构建。
2. 对每种服务完成手工连接、资源浏览、关闭和保存连接检查；对 JDBC profile 完成“被引用禁止卸载”和“未引用可卸载”检查。
3. 记录实际验证结果、工具链阻塞（如有）和剩余风险。
4. 提交 `docs: verify native nosql connections and jdbc driver removal`。

### 任务 10：实现跨平台内置 provider 注册表

**文件：** 创建 `internal/service/native_database_registry.go`、对应测试；创建 `docs/changes/features/2026-07-13-native-nosql-provider-registry.md`。

1. 写失败测试，覆盖内置 provider 注册、重复 ID 拒绝、类型查找和跨平台不变的 provider ID。
2. 运行 `go test ./internal/service -run TestNativeDatabaseRegistry -v`，确认失败。
3. 实现 provider 注册表，不使用 Go `plugin`、动态库或 Unix 专有机制。
4. 再次运行同一命令，确认通过。
5. 写中文变更记录并运行 `go test ./internal/service`。
6. 提交 `feat: add cross-platform nosql provider registry`。

### 任务 11：实现首批常用 NoSQL provider

**文件：** 创建各 provider 与测试、修改 `go.mod`/`go.sum`；创建每类 provider 的中文变更记录。

1. 依次为 Redis/KeyDB、MongoDB、Elasticsearch/OpenSearch、Memcached、Cassandra、Couchbase、InfluxDB、Neo4j、Kafka 写失败测试。
2. 每个 provider 独立运行对应测试确认失败。
3. 逐个实现连接测试与只读资源浏览，优先使用厂商官方 SDK；无官方 Go SDK 时封装稳定公开协议并注明来源。
4. 每个 provider 独立确认通过、记录中文变更并提交。
5. 完成后运行 `go test ./internal/service`。
