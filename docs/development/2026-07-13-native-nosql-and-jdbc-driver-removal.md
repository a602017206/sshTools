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
- `wails build -platform darwin/arm64`：通过，已生成 `build/bin/AHaSSHTools.app`。

单元测试覆盖 provider 的连接委派、资源排序、错误传播与关闭。未执行各服务的真实手工连接，因为当前环境未提供 Redis、MongoDB、Elasticsearch、Memcached、Cassandra、Couchbase、InfluxDB、Neo4j 或 Kafka 的可用地址与凭据。

## Wails 工具链阻塞与修复

`wails build -platform darwin/arm64` 在绑定生成阶段失败，报错为：`package "github.com/neo4j/neo4j-go-driver/v6/neo4j" without types was imported from "AHaSSHTools/internal/service"`。

Wails 和命令行 Go 均使用 `go1.24.4`，因此不是 Go 可执行文件版本不一致。实际采用的最小修复是将 Neo4j 官方驱动从 v6 改为 v5 LTS 分支；该分支仍兼容目标服务器版本，并避免触发 Wails v2.11 的类型加载缺陷。修复后不跳过 bindings 重新运行 `wails build -platform darwin/arm64`，绑定生成、前端编译和应用编译均通过。

## 剩余风险

- 真实服务连通性、认证、TLS、代理和权限策略需在目标环境逐类验证。
- Kafka 首版仅支持单一明文 broker；Couchbase、InfluxDB 和 Neo4j 的 TLS/云端高级认证尚未暴露到表单。
- 真实数据库与 Kafka 的 TLS、代理及细粒度权限仍需在目标环境验证。
