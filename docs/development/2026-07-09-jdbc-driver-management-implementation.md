# JDBC 驱动管理实现记录

## 实施状态

初始实施计划和后续补全计划均已按任务顺序执行。Go、Java agent、H2 端到端、崩溃恢复、前端和 Wails 生产构建均通过。推荐驱动在线安装、托管 JRE 安装与归档导入、真实文件选择器、Agent 状态和 session 自动恢复已经接线。

## 已实现架构

- Go 侧提供内置 JDBC manifest、driver profile、本地目录、checksum 校验、在线安装、离线驱动包导入、运行时安装与选择、agent supervisor 和托管 gRPC gateway。
- Java 21 agent 提供健康检查、JDBC classloader、连接注册、查询、表/字段元数据和会话关闭服务，并通过 shadow jar 打包。
- 原有 Wails 数据库 API 保持不变，`DatabaseService` 可把连接、查询、元数据和关闭请求委托给 JDBC gateway。
- Wails 暴露驱动列表、在线安装、离线导入、校验、删除、托管 JRE 安装、运行时归档导入、文件选择、Agent 状态和重启 API。
- Svelte JDBC 驱动管理页使用系统文件选择器，展示运行时和 Agent 实际状态，并提供错误分类后的恢复操作。
- H2 集成脚本覆盖离线包导入、真实 Java 子进程、gRPC、SQL、元数据、会话关闭和进程崩溃后的 session 恢复。

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
- 补全计划中首次执行 `wails generate module` 同样因 Go 构建缓存权限失败；最小修复是在允许访问缓存的环境中重跑原命令，未手工伪造 bindings。

### 前端构建与 Agent 暂存

- `npm run build` 的 Vite 阶段成功后，`stage-jdbc-agent.sh` 因沙箱无法创建 `~/.gradle` wrapper 锁文件而退出；最小修复是在允许访问 Gradle 缓存的环境中重跑完整命令，确保未跳过 agent 构建与暂存。

## 最终验证

| 验证项 | 结果 | 证据 |
| --- | --- | --- |
| Go 全量测试 | 通过 | `go test ./...` |
| Java agent 全量测试 | 通过 | `cd jdbc-agent && ./gradlew test` |
| H2 端到端测试 | 通过 | `./scripts/test-jdbc-agent.sh` |
| Agent 崩溃恢复 | 通过 | `TestJDBCAgentRecoversSessionAfterCrash` 杀死真实 Java 进程后重新查询成功 |
| 前端生产构建 | 通过 | `cd frontend && npm run build` |
| Wails 生产构建 | 通过 | `/Users/dingwei/go/bin/wails build` |

Wails 产物位于 `build/bin/AHaSSHTools.app`。前端仍有仓库既有的 Svelte 可访问性、大分块和静态/动态导入警告；Gradle 仍提示部分特性在 Gradle 9 将不兼容。

## 验收结论

| 验收项 | 结论 | 说明 |
| --- | --- | --- |
| 无 Java 环境时提示安装托管 JRE | 自动验证通过 | 在线安装、归档导入和文件选择 API 均有测试；未执行真实桌面点击。 |
| 离线导入 H2 包后查询 | 通过 | 集成测试已用真实 agent 完成导入、查询和元数据读取。 |
| 驱动页安装、校验、卸载、重新导入 | 自动验证通过 | 在线下载原子提交、离线导入、校验、删除和文件选择已接线；未执行真实桌面点击。 |
| 连接表单选择首批 JDBC 类型 | 通过 | 表单包含 MySQL、PostgreSQL、SQLite、Oracle、SQL Server、达梦、人大金仓和 openGauss，前端构建通过。 |
| agent 崩溃后恢复 session | 通过 | 集成测试杀死真实 agent 后由 supervisor 启动新进程、重新打开文件模式 H2 session 并完成原查询。 |

## 后续工作

1. 在发布候选包上执行真实 Wails 桌面点击验收，覆盖文件选择器、在线下载进度、错误提示和状态刷新。
2. 为系统 Java 与托管 JRE 的选择增加持久化配置，避免每次启动重新选择。
3. 处理仓库既有的 Svelte 可访问性和大分块告警，并评估 Gradle 9 升级。
