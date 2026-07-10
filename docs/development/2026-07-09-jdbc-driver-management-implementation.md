# JDBC 驱动管理实现记录

## 实施状态

实施计划中的任务 1 至 14 已按顺序执行并分别提交。Go、Java agent、H2 端到端、前端和 Wails 生产构建均通过。最终验收同时确认，在线驱动安装、托管 JRE 安装/归档导入和应用运行时 gateway 自动重连仍是占位或未接线状态，因此当前版本不应按“完整 JDBC 管理闭环”发布。

## 已实现架构

- Go 侧提供 JDBC manifest、driver profile、本地目录、checksum 校验、离线驱动包导入、运行时选择、agent 进程管理和 gRPC gateway。
- Java 21 agent 提供健康检查、JDBC classloader、连接注册、查询、表/字段元数据和会话关闭服务，并通过 shadow jar 打包。
- 原有 Wails 数据库 API 保持不变，`DatabaseService` 可把连接、查询、元数据和关闭请求委托给 JDBC gateway。
- Wails 暴露驱动列表、导入、校验、删除、运行时模式和 agent 重启 API。
- Svelte 增加 JDBC 驱动管理页、首批数据库类型、错误分类提示和恢复操作。
- H2 集成脚本覆盖离线包导入、真实 Java 子进程、gRPC、SQL、元数据和会话关闭。

## 错误与日志

稳定错误码包括：

- `RUNTIME_MISSING`
- `DRIVER_MISSING`
- `DRIVER_INVALID`
- `AGENT_UNAVAILABLE`
- `DB_CONNECT_FAILED`

日志路径约定包括：

- `~/.sshtools/logs/jdbc-agent.log`
- `~/.sshtools/logs/driver-install.log`
- `~/.sshtools/logs/runtime-install.log`

## 工具链阻塞与最小修复

### Gradle 与 JDK

- 初始工程没有 `gradlew`，最小修复是复用本机已有 Gradle wrapper 脚本和 jar，并提交固定的 wrapper 配置。
- 默认 Java 为 x86_64 Java 8，无法稳定运行当前 Gradle；最小修复是在 `jdbc-agent/gradle.properties` 固定本机 JDK 21 路径。
- 用户级 Gradle 初始化脚本与严格仓库模式冲突；最小修复是取消 `FAIL_ON_PROJECT_REPOS`，保留项目仓库声明。
- 最终验证时沙箱禁止写 `~/.gradle` wrapper 锁文件；最小修复是在沙箱外重跑同一条 `./gradlew test`，未修改项目配置。

### protoc 与 gRPC

- Homebrew `protoc` 因缺失 `abseil` 动态库无法启动，Go 插件也不在默认 `PATH`；最小修复是由生成脚本复用 Gradle 缓存中的 arm64 protoc，并补充 `GOPATH/bin`。
- 最新 gRPC 依赖链要求 Go 1.25，而仓库使用 Go 1.24；最小修复是固定 `google.golang.org/grpc v1.65.0` 及兼容的 `genproto` 版本。

### Wails

- 沙箱内 `wails build` 无法访问 `~/Library/Caches/go-build`，bindings 生成阶段失败；最小修复是在沙箱外重跑相同命令。第二次运行以退出码 0 完成应用编译和打包。

## 最终验证

| 验证项 | 结果 | 证据 |
| --- | --- | --- |
| Go 全量测试 | 通过 | `go test ./...` |
| Java agent 全量测试 | 通过 | `cd jdbc-agent && ./gradlew test` |
| H2 端到端测试 | 通过 | `./scripts/test-jdbc-agent.sh` |
| 前端生产构建 | 通过 | `cd frontend && npm run build` |
| Wails 生产构建 | 通过 | `/Users/dingwei/go/bin/wails build` |

Wails 产物位于 `build/bin/AHaSSHTools.app`。前端仍有仓库既有的 Svelte 可访问性、大分块和静态/动态导入警告；Gradle 仍提示部分特性在 Gradle 9 将不兼容。

## 验收结论

| 验收项 | 结论 | 说明 |
| --- | --- | --- |
| 无 Java 环境时提示安装托管 JRE | 部分通过 | 已有 `RUNTIME_MISSING` 按钮；托管 JRE 下载和归档导入仍未实现。 |
| 离线导入 H2 包后查询 | 通过 | 集成测试已用真实 agent 完成导入、查询和元数据读取。 |
| 驱动页安装、校验、卸载、重新导入 | 部分通过 | 离线导入、校验和删除 API 已接线；在线安装仍返回“暂未实现”，未进行真实 UI 点击验收。 |
| 连接表单选择首批 JDBC 类型 | 通过 | 表单包含 MySQL、PostgreSQL、SQLite、Oracle、SQL Server、达梦、人大金仓和 openGauss，前端构建通过。 |
| agent 崩溃后提示重连 | 未通过 | 已有 `AGENT_UNAVAILABLE` 操作，但应用初始化 gateway client 为空，重启后没有重新创建 gRPC client 并注入 gateway。 |

## 后续最小工作

1. 实现托管 JRE 下载、checksum 校验和离线运行时归档导入，并让 `SetJDBCRuntimeMode` 持久化选择。
2. 实现推荐驱动在线下载，同时保留现有离线包导入作为无网络路径。
3. 在应用启动和 agent 重启后创建真实 `GRPCJdbcAgentClient`，注入 gateway，并在进程退出时触发 `AGENT_UNAVAILABLE` 和重连。
4. 完成真实 Wails UI 点击验收，补充文件选择器和日志查看器，避免依赖文本路径输入。

