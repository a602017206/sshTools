# JDBC 运行时持久化与自动恢复设计

## 背景

JDBC 驱动管理已经具备托管 JRE 在线安装、运行时归档导入、系统 Java 选择、Agent supervisor 和崩溃后 session 恢复能力，但运行时选择只保存在进程内。应用重启后会重新按默认规则选择 Java，用户无法确认当前选择是否被恢复。JDBC 设置页的 Agent 状态只在页面加载和显式操作后刷新，日志按钮也只展示约定路径，不能直接查看诊断内容。

本设计补齐运行时选择持久化、应用启动恢复、切换后立即验证、Agent 状态定时刷新和内置日志尾部查看，形成可诊断的运行时管理闭环。

## 目标

- 系统 Java 或托管 JRE 的选择持久化到现有应用配置。
- 应用启动时恢复用户选择，并对无效路径给出明确状态与错误。
- 运行时切换、托管 JRE 安装和归档导入后立即重启并验证 Agent。
- JDBC 设置页打开期间自动刷新运行时与 Agent 状态。
- 用户可在应用内查看、刷新和复制 JDBC Agent 最近日志。
- 不暴露任意文件读取能力，不把 token、密码或完整 JDBC URL写入新增日志接口。

## 非目标

- 本阶段不实现自定义 JDBC profile、Maven 坐标解析或驱动更新检测。
- 本阶段不改为 Wails event 推送状态，仍使用低频轮询。
- 本阶段不实现日志搜索、下载、清空或长期归档。
- 本阶段不持久化 Agent 的瞬时状态和历史错误。

## 方案选择

采用“统一应用配置 + 状态轮询 + 固定日志尾部读取”方案。

- 配置继续写入 `~/.ahasshtools/config.json`，避免形成第二套设置生命周期。
- JRE、驱动和日志实体继续保存在 `~/.sshtools/`。
- JDBC 页面每 2 秒查询一次状态，组件销毁时停止定时器。
- 日志接口只读取 `JDBCPaths.LogsDir` 下固定的 `jdbc-agent.log`，不接受调用方路径。

相比独立 `runtime.json`，统一配置更容易复用现有加载、默认值和原子保存路径。相比 Wails event，轮询实现边界更小，足以覆盖低频的进程状态变化。

## 配置模型

在 `config.AppSettings` 增加：

```go
JDBCRuntimeMode     string `json:"jdbc_runtime_mode"`
JDBCSystemJavaPath string `json:"jdbc_system_java_path"`
```

约束如下：

- `jdbc_runtime_mode` 只允许 `managed`、`system` 或空字符串。
- 空字符串表示尚未显式选择，兼容旧配置，并使用现有自动选择规则。
- `system` 必须配合非空 `jdbc_system_java_path`。
- `managed` 不保存具体 JRE 目录，运行时服务仍选择本地最新有效托管 JRE。
- 配置加载时缺少新字段不视为错误，保持向后兼容。

配置更新必须通过 `ConfigManager` 的专用方法完成，并复用现有 `Save`。不得由 JDBC service 直接操作配置文件。

## 启动恢复

应用启动顺序调整为：

1. 初始化 `ConfigManager`。
2. 创建 JDBC 路径、运行时服务、Agent supervisor 和 gateway。
3. 读取 `AppSettings` 中的 JDBC 运行时选择。
4. 将选择应用到现有 `RuntimeService` 实例，不能替换实例，以保持 supervisor 引用有效。
5. 不在应用启动时主动启动 Agent；首次数据库操作或用户点击重启时仍采用懒启动。

恢复规则：

- `managed`：关闭系统 Java优先级，选择最新有效托管 JRE；若不存在，状态为 `missing`。
- `system`：配置指定路径并启用系统 Java；路径不存在或不是普通文件时，状态为 `missing`，同时返回可诊断错误，不静默改用其他 Java。
- 空字符串：沿用自动选择，优先有效托管 JRE，其次默认系统 Java。
- 未知模式：按配置错误处理，状态为 `missing`，不猜测用户意图。

## 运行时切换事务

`SetJDBCRuntimeMode` 的顺序为：

1. 校验模式和系统 Java 路径。
2. 更新当前 `RuntimeService` 实例。
3. 持久化配置。
4. 使用 15 秒超时立即重启 Agent。
5. 返回最新运行时状态和 Agent 状态。

若第 3 步持久化失败，恢复切换前的内存选择，不重启 Agent。若第 4 步重启失败，不回滚已保存的用户选择；返回错误，并由 supervisor 保留 `failed` 和最后错误，便于用户安装或重新选择运行时。

托管 JRE 在线安装和归档导入成功后执行相同流程：切换到 `managed`、持久化、立即重启。下载或导入失败时不改变当前选择。

## API 模型

新增组合结果模型：

```go
type JDBCRuntimeActivationResult struct {
    Runtime RuntimeStatus   `json:"runtime"`
    Agent   JDBCAgentStatus `json:"agent"`
}
```

调整或新增 Wails API：

- `SetJDBCRuntimeMode(mode, path) (JDBCRuntimeActivationResult, error)`
- `InstallJDBCManagedRuntime() (JDBCRuntimeActivationResult, error)`
- `ImportJDBCRuntimeArchive(path) (JDBCRuntimeActivationResult, error)`
- `GetJDBCAgentLogTail(maxBytes) (JDBCLogTail, error)`

日志结果模型：

```go
type JDBCLogTail struct {
    Content   string `json:"content"`
    Truncated bool   `json:"truncated"`
    Size      int64  `json:"size"`
}
```

`maxBytes` 默认 64 KiB，最小 1 KiB，最大 1 MiB。日志不存在时返回空内容而不是错误；目录不可访问或文件不是普通文件时返回错误。

## 状态轮询

`JDBCDriverManager.svelte` 在挂载时立即加载状态，并启动 2 秒定时器。轮询只调用运行时和 Agent 状态 API，不刷新驱动 catalog，避免频繁扫描目录。

行为约束：

- 页面不可见或组件销毁时清理定时器。
- 用户操作进行中时仍允许状态轮询，但不得覆盖当前操作错误。
- 连续轮询失败只更新状态区域，不弹出重复错误。
- “刷新”按钮继续刷新驱动、运行时和 Agent 全部数据。

## 日志查看

错误操作中的“查看日志”打开内置模态框。模态框包含：

- 最近日志内容，使用等宽字体并保留换行。
- “刷新”命令。
- “复制”命令，复制当前已展示内容。
- 截断提示和当前文件大小。
- 空日志和读取失败状态。

日志读取固定在 `jdbc-agent.log`。后端从文件尾部读取，不先把完整文件载入内存。返回内容按 UTF-8 展示，非法字节使用替换字符处理。

## 错误处理

- 无效系统 Java：`RUNTIME_MISSING`，信息包含路径不可用但不包含环境变量或其他本机敏感信息。
- 配置保存失败：保留原运行时选择，返回中文持久化错误。
- Agent 重启失败：保留新选择，返回 `AGENT_UNAVAILABLE`，顶部状态显示最后错误。
- 日志读取失败：仅影响日志模态框，不覆盖 Agent 主状态。
- 所有日志和 UI 错误禁止显示认证 token、密码和未经脱敏的完整 JDBC URL。

## 测试策略

### Go 单元测试

- 旧配置缺少 JDBC 字段时使用空模式并可正常加载。
- 系统 Java和托管模式可保存、重载并恢复。
- 未知模式、空系统路径、目录路径和不存在路径被拒绝。
- 配置保存失败时恢复内存选择且不重启 Agent。
- 切换、在线安装和归档导入成功后持久化并调用一次 supervisor 重启。
- Agent 重启失败时配置仍保留新选择，状态为 `failed`。
- 日志不存在、短文件、超限截断、最大值限制和非普通文件处理正确。

### 前端与生成验证

- Wails bindings 由生成命令更新，不手工伪造。
- 前端生产构建通过。
- 组件销毁时清除状态定时器。
- 日志模态框可打开、刷新、复制和关闭。

### 桌面验收

- 选择系统 Java后立即重启，重开应用仍显示相同路径。
- 安装或导入托管 JRE 后立即切换并重启，重开应用仍为托管模式。
- Agent 从 `starting` 进入 `running` 或 `failed`，页面无需手动刷新。
- 日志模态框显示最近日志，超限时显示截断提示。
- 应用退出后 Agent 子进程被清理。

## 工具链约束

继续沿用现有 gRPC/protoc/Gradle/Java agent 工具链。若 Gradle wrapper、Go 构建缓存、Wails bindings 或 Java 版本导致失败，必须记录具体阻塞点和最小修复方案，并重跑原验证命令，不得绕过 Agent 构建、手工伪造 bindings 或缩小验收范围。

## 剩余风险

- 2 秒轮询会产生少量 Wails 调用；若未来状态维度增加，再迁移到事件推送。
- 配置文件保存目前不是临时文件加原子替换，本阶段复用既有机制；后续可统一增强 `ConfigManager.Save`。
- 托管 JRE 仍按目录名称排序选择最新版本，尚未实现语义化版本比较和固定具体版本。
- 日志内容来自本地 Agent；若上游日志未来包含敏感字段，需要在写入端增加统一脱敏。
