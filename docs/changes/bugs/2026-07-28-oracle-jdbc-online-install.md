# Oracle JDBC 在线安装修复

## 背景

Oracle 驱动在管理界面中提供在线安装操作，但内置 profile 没有配置下载地址和 SHA-256 校验值，安装服务会拒绝该 profile，导致用户无法自动安装。

## 范围

将 Oracle `ojdbc11` profile 切换为 Oracle 官方发布到 Maven Central 的在线包，并提升内置驱动清单版本，使已生成的本地清单自动迁移到可在线安装的推荐版本。

## 修改文件

- `internal/service/jdbc_builtin_manifest.json`
- `internal/service/jdbc_catalog_test.go`

## 验证

新增驱动目录测试，校验 Oracle profile 具有 Maven Central HTTPS 地址和有效 SHA-256，并校验版本为 2 的旧本地清单能够迁移到新的 Oracle profile。

测试执行期间发现本机已有多个并发 Go 编译任务在重建 Redis、Kafka、MongoDB 和 Couchbase 依赖，单个测试启动阶段长时间无输出。最小处理方案是等待这些重复构建完成后，以单个 `go test` 进程重新执行相关测试；不修改 gRPC、protoc、Gradle 或 Java agent 工具链。

## 剩余风险

Oracle JDBC 包受 Oracle Free Use Terms and Conditions 约束，用户仍需自行确认其使用场景符合上游许可要求。下载时应用会继续校验固定 SHA-256，Maven Central 内容变化会被拒绝而不是静默安装。
