# JDBC agent 工程和健康检查

## 背景

数据库模块迁移到 JDBC agent 架构后，需要一个独立的 Java 子进程承载 JDBC 驱动加载、连接和查询。首个 Java agent 任务先建立 Gradle 工程、gRPC 协议和健康检查接口，为后续查询与元数据能力铺底。

## 范围

- 新增 `jdbc-agent` Gradle 工程配置。
- 新增 Gradle wrapper，并配置使用本机 JDK 21。
- 新增 `jdbc_agent.proto`，定义健康检查、连接、查询、元数据和关闭会话的 gRPC 接口。
- 新增 `JdbcAgentApplication`，支持读取 `--port` 和 `--token` 并启动本地 gRPC server。
- 新增 `HealthServiceImpl`，校验 token 后返回 `OK`、agent 版本和 Java 版本。
- 新增健康检查单元测试，覆盖 token 正确和错误两种路径。
- 更新 `.gitignore`，忽略 `jdbc-agent/build/`。

## 修改文件

- `.gitignore`
- `jdbc-agent/settings.gradle`
- `jdbc-agent/build.gradle`
- `jdbc-agent/gradle.properties`
- `jdbc-agent/gradlew`
- `jdbc-agent/gradlew.bat`
- `jdbc-agent/gradle/wrapper/gradle-wrapper.jar`
- `jdbc-agent/gradle/wrapper/gradle-wrapper.properties`
- `jdbc-agent/src/main/proto/jdbc_agent.proto`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/JdbcAgentApplication.java`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/HealthServiceImpl.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/HealthServiceImplTest.java`

## 验证

- 已运行 `cd jdbc-agent && ./gradlew test --tests '*HealthServiceImplTest'`，结果通过。
- 初次验证阻塞点：仓库缺少 `gradlew`，计划指定命令无法启动。
- 第二个阻塞点：系统默认 `java` 指向 x86_64 Java 8，系统 `gradle` 即使设置 JDK 21 也因 `libnative-platform.dylib` 加载失败无法初始化。
- 最小修复：复用本机已有 Gradle wrapper jar 和脚本，新增 wrapper 配置，并通过 `jdbc-agent/gradle.properties` 指定 JDK 21。
- 第三个阻塞点：用户级 `~/.gradle/init.gradle` 注入仓库，与 `RepositoriesMode.FAIL_ON_PROJECT_REPOS` 冲突。
- 最小修复：移除 `settings.gradle` 中的强制仓库模式，保留 `mavenCentral()`。

## 剩余风险

- gRPC/protobuf 依赖已能在当前环境解析，但后续 CI 或干净机器仍需要可访问 Gradle distribution 和 Maven 仓库。
- agent 当前只注册健康检查服务，查询和元数据接口将在后续任务实现。
