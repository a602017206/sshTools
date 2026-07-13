# 原生 NoSQL 连接与 JDBC 驱动卸载开发记录

## 实现内容

实现 JDBC profile 卸载保护：已保存连接、旧版推荐 profile 和活动 JDBC 会话引用的 profile 均不可卸载。驱动管理界面改为按所选 profile 显示安装或卸载操作。

原生连接使用静态内置 provider，不实现外部插件动态加载。已实现 Redis/KeyDB、MongoDB、Elasticsearch/OpenSearch、Memcached、Cassandra/ScyllaDB、Couchbase、InfluxDB、Neo4j 和 Kafka 的连接测试、只读资源浏览和会话关闭。所有 provider 均使用 Go 库、TCP 或 HTTP，不使用 Go `plugin`、系统动态库、shell 命令或 macOS 专有 API。

## 验证结果

- `go test ./...`：通过。
- `node --test test/jdbcDriverProfileState.test.js test/nativeDatabaseTypes.test.js`：通过。
- `npm run build`：通过，包含 Vite 和 Gradle JDBC agent 打包。
- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/sshTools-windows.test .`：通过。
- `GOOS=linux GOARCH=amd64 go test -c -o /tmp/sshTools-linux.test .`：通过。
- `wails build -platform darwin/arm64`：未通过，详见工具链阻塞。

单元测试覆盖 provider 的连接委派、资源排序、错误传播与关闭。未执行各服务的真实手工连接，因为当前环境未提供 Redis、MongoDB、Elasticsearch、Memcached、Cassandra、Couchbase、InfluxDB、Neo4j 或 Kafka 的可用地址与凭据。

## Wails 工具链阻塞

`wails build -platform darwin/arm64` 在绑定生成阶段失败，报错为：`package "github.com/neo4j/neo4j-go-driver/v6/neo4j" without types was imported from "AHaSSHTools/internal/service"`。

Wails 和命令行 Go 均使用 `go1.24.4`，因此不是 Go 可执行文件版本不一致。最小修复方案是升级或修补 Wails 使用的 `go/packages`/类型加载依赖，使其能够加载 Neo4j v6；修复后必须重新运行不跳过 bindings 的 `wails build -platform darwin/arm64`。本次不使用 `-skipbindings` 绕开该失败。

## 剩余风险

- 真实服务连通性、认证、TLS、代理和权限策略需在目标环境逐类验证。
- Kafka 首版仅支持单一明文 broker；Couchbase、InfluxDB 和 Neo4j 的 TLS/云端高级认证尚未暴露到表单。
- Wails 构建在 Neo4j v6 的绑定类型加载问题修复前不能完成 macOS 发行包验证。
