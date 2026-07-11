# JDBC 运行时持久化与自动恢复实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 持久化 JDBC 运行时选择，在启动和切换后可靠恢复并验证 Agent，同时提供自动状态刷新和安全的内置日志查看。

**架构：** JDBC 选择写入现有 `config.AppSettings`，由应用层协调 `ConfigManager`、`RuntimeService` 和 `JDBCAgentSupervisor`，不替换 supervisor 已持有的运行时实例。前端使用 2 秒低频轮询读取状态，日志通过固定路径的受限尾部读取服务返回，禁止任意路径访问。

**技术栈：** Go 1.24、Wails v2、Svelte 4、Java 21 JDBC Agent、gRPC、Gradle 8.5、Go 标准库文件 API。

---

## 执行约束

- 严格按任务顺序执行，不跳任务。
- 每个任务都执行：写失败测试、确认失败、最小实现、确认通过、补中文变更文档、提交。
- Wails bindings 必须使用 `/Users/dingwei/go/bin/wails generate module` 生成，不手工伪造。
- 遇到 gRPC、protoc、Gradle、Java agent 或 Wails 工具链问题，先记录阻塞点和最小修复，再重跑原命令。
- 不修改或删除用户现有 `~/.ahasshtools/config.json`、`~/.sshtools/` 驱动和运行时数据。

## 任务 1：扩展 JDBC 运行时配置模型

**文件：**

- 修改：`internal/config/config.go`
- 修改：`internal/config/config_test.go`
- 创建：`docs/changes/features/2026-07-11-jdbc-runtime-settings.md`

**步骤 1：写失败测试**

新增：

```go
func TestConfigManagerPersistsJDBCRuntimeSettings(t *testing.T) {
    cm := newTestConfigManager(t)
    if err := cm.UpdateJDBCRuntimeSettings("system", "/opt/jdk-21/bin/java"); err != nil {
        t.Fatal(err)
    }
    reloaded := loadTestConfigManager(t, cm.configPath)
    settings := reloaded.GetSettings()
    if settings.JDBCRuntimeMode != "system" ||
        settings.JDBCSystemJavaPath != "/opt/jdk-21/bin/java" {
        t.Fatalf("settings=%+v", settings)
    }
}
```

同时增加：

- `TestConfigManagerLoadsLegacySettingsWithoutJDBCFields`：旧配置加载后两个新字段为空，既有设置不变。
- `TestConfigManagerRejectsInvalidJDBCRuntimeMode`：未知模式和 system 空路径被拒绝。

**步骤 2：确认失败**

```bash
go test ./internal/config -run 'TestConfigManager(PersistsJDBCRuntimeSettings|LoadsLegacySettingsWithoutJDBCFields|RejectsInvalidJDBCRuntimeMode)' -v
```

预期：编译失败，字段或 `UpdateJDBCRuntimeSettings` 不存在。

**步骤 3：最小实现**

在 `AppSettings` 增加：

```go
JDBCRuntimeMode     string `json:"jdbc_runtime_mode"`
JDBCSystemJavaPath string `json:"jdbc_system_java_path"`
```

实现专用保存方法，模式只允许空、`managed`、`system`；system 必须有路径。保存失败时恢复内存中的旧值。

**步骤 4：确认通过**

```bash
go test ./internal/config -run 'TestConfigManager(PersistsJDBCRuntimeSettings|LoadsLegacySettingsWithoutJDBCFields|RejectsInvalidJDBCRuntimeMode)' -v
go test ./internal/config -v
```

**步骤 5：文档和提交**

```bash
git add internal/config/config.go internal/config/config_test.go docs/changes/features/2026-07-11-jdbc-runtime-settings.md
git commit -m "feat: persist jdbc runtime settings"
```

## 任务 2：增加运行时快照、校验和启动恢复

**文件：**

- 修改：`internal/service/jdbc_runtime.go`
- 修改：`internal/service/jdbc_runtime_test.go`
- 修改：`app.go`
- 修改：`app_jdbc_test.go`
- 创建：`docs/changes/features/2026-07-11-jdbc-runtime-restore.md`

**步骤 1：写失败测试**

新增：

- `TestRuntimeServiceRestoresExplicitSystemJava`：有效普通文件恢复为 system。
- `TestRuntimeServiceRejectsInvalidExplicitSystemJava`：目录、不存在路径和未知模式被拒绝。
- `TestRuntimeServiceRestoresSnapshot`：模式修改后可恢复原快照。
- `TestBuildJDBCServicesRestoresPersistedRuntimeMode`：bundle 创建时应用配置，但不启动 Agent。

核心断言：

```go
before := runtimeService.Snapshot()
if err := runtimeService.ApplyMode("managed", ""); err != nil { t.Fatal(err) }
runtimeService.Restore(before)
if runtimeService.Snapshot() != before { t.Fatal("snapshot not restored") }
```

**步骤 2：确认失败**

```bash
go test ./internal/service -run 'TestRuntimeService(RestoresExplicitSystemJava|RejectsInvalidExplicitSystemJava|RestoresSnapshot)' -v
go test . -run TestBuildJDBCServicesRestoresPersistedRuntimeMode -v
```

预期：`ApplyMode`、`Snapshot`、`Restore` 或配置恢复依赖不存在。

**步骤 3：最小实现**

新增：

```go
type RuntimeModeSnapshot struct {
    Mode           string
    SystemJavaPath string
}
```

`RuntimeService` 增加 mutex 和显式模式字段。规则：

- system：路径存在且是普通文件。
- managed：只选择托管 JRE，不回退系统 Java。
- 空模式：沿用自动选择。
- 未知模式：返回错误。

调整 `buildJDBCServices` 接收持久化设置并调用 `ApplyMode`，不得调用 `Client` 或 `Restart`。

**步骤 4：确认通过**

```bash
go test ./internal/service -run 'TestRuntimeService(RestoresExplicitSystemJava|RejectsInvalidExplicitSystemJava|RestoresSnapshot)' -v
go test . -run TestBuildJDBCServicesRestoresPersistedRuntimeMode -v
go test ./internal/service -v
```

**步骤 5：文档和提交**

```bash
git add internal/service/jdbc_runtime.go internal/service/jdbc_runtime_test.go app.go app_jdbc_test.go docs/changes/features/2026-07-11-jdbc-runtime-restore.md
git commit -m "feat: restore jdbc runtime selection"
```

## 任务 3：实现持久化后立即激活 Agent 的事务

**文件：**

- 修改：`internal/service/jdbc_api_models.go`
- 修改：`app.go`
- 修改：`app_jdbc_test.go`
- 创建：`docs/changes/features/2026-07-11-jdbc-runtime-activation.md`

**步骤 1：写失败测试**

新增：

- `TestSetJDBCRuntimeModePersistsAndRestartsAgent`：保存一次、重启一次、返回组合状态。
- `TestSetJDBCRuntimeModeRollsBackWhenPersistenceFails`：恢复快照，重启次数为 0。
- `TestSetJDBCRuntimeModeKeepsSelectionWhenRestartFails`：保留新配置，Agent 为 failed。
- `TestManagedRuntimeInstallAndImportActivateManagedMode`：在线安装和归档导入都保存 managed 并重启。

**步骤 2：确认失败**

```bash
go test . -run 'Test(SetJDBCRuntimeMode|ManagedRuntimeInstallAndImportActivateManagedMode)' -v
```

预期：现有 API 不返回组合结果，也不持久化或统一激活。

**步骤 3：最小实现**

新增模型：

```go
type JDBCRuntimeActivationResult struct {
    Runtime RuntimeStatus   `json:"runtime"`
    Agent   JDBCAgentStatus `json:"agent"`
}
```

抽出 `activateJDBCRuntime(mode, path)`，顺序固定为：

1. 保存 runtime 快照。
2. 校验并应用新模式。
3. 持久化配置。
4. 使用 15 秒超时重启 Agent。
5. 返回最新组合状态。

持久化失败恢复快照且不重启；重启失败不回滚已保存选择。在线安装和归档导入成功后都调用 managed 激活路径。

**步骤 4：确认通过**

```bash
go test . -run 'Test(SetJDBCRuntimeMode|ManagedRuntimeInstallAndImportActivateManagedMode)' -v
go test ./...
```

**步骤 5：文档和提交**

```bash
git add internal/service/jdbc_api_models.go app.go app_jdbc_test.go docs/changes/features/2026-07-11-jdbc-runtime-activation.md
git commit -m "feat: activate persisted jdbc runtime"
```

## 任务 4：实现固定路径的 Agent 日志尾读

**文件：**

- 创建：`internal/service/jdbc_log_tail.go`
- 创建：`internal/service/jdbc_log_tail_test.go`
- 修改：`internal/service/jdbc_api_models.go`
- 修改：`app.go`
- 修改：`app_jdbc_test.go`
- 创建：`docs/changes/features/2026-07-11-jdbc-agent-log-viewer.md`

**步骤 1：写失败测试**

新增：

- `TestJDBCLogTailReturnsLastBytes`：2 KiB 文件读取后 1 KiB，后缀正确且 truncated。
- `TestJDBCLogTailHandlesMissingAndInvalidFiles`：缺失为空，目录或非普通文件报错。
- `TestJDBCLogTailClampsRequestedSize`：默认 64 KiB，范围 1 KiB 至 1 MiB。
- `TestGetJDBCAgentLogTailUsesConfiguredPath`：App API 只接收 maxBytes，不接收路径。

**步骤 2：确认失败**

```bash
go test ./internal/service -run TestJDBCLogTail -v
go test . -run TestGetJDBCAgentLogTailUsesConfiguredPath -v
```

预期：日志 service、结果模型和 App API 不存在。

**步骤 3：最小实现**

新增：

```go
type JDBCLogTail struct {
    Content   string `json:"content"`
    Truncated bool   `json:"truncated"`
    Size      int64  `json:"size"`
}
```

`JDBCLogTailService` 只使用 `NewJDBCLogPaths(paths).Agent`。通过 `Open`、`Stat`、`Seek` 从尾部读取，不加载完整文件。非法 UTF-8 使用替换字符。日志不存在返回空结果。

**步骤 4：确认通过**

```bash
go test ./internal/service -run TestJDBCLogTail -v
go test . -run TestGetJDBCAgentLogTailUsesConfiguredPath -v
go test ./internal/service -v
```

**步骤 5：文档和提交**

```bash
git add internal/service/jdbc_log_tail.go internal/service/jdbc_log_tail_test.go internal/service/jdbc_api_models.go app.go app_jdbc_test.go docs/changes/features/2026-07-11-jdbc-agent-log-viewer.md
git commit -m "feat: read jdbc agent log tail"
```

## 任务 5：前端状态轮询和日志模态框

**文件：**

- 创建：`frontend_contract_test.go`
- 修改：`frontend/src/components/JDBCDriverManager.svelte`
- 修改：`frontend/wailsjs/go/main/App.js`
- 修改：`frontend/wailsjs/go/main/App.d.ts`
- 修改：`frontend/wailsjs/go/models.ts`
- 修改：`frontend/build/assets/index.js`
- 创建：`docs/changes/features/2026-07-11-jdbc-runtime-recovery-ui.md`

**步骤 1：写失败测试**

新增 `TestJDBCDriverManagerIncludesPollingAndLogViewer`，读取组件源码并断言：

- 导入并使用 `onDestroy`。
- 存在 2000 毫秒轮询及 `clearInterval`。
- 导入并调用 `GetJDBCAgentLogTail`。
- 存在日志刷新、复制和关闭命令。
- 轮询不调用 `ListJDBCDrivers`。

**步骤 2：确认失败**

```bash
go test . -run TestJDBCDriverManagerIncludesPollingAndLogViewer -v
```

预期：源码契约断言失败。

**步骤 3：最小实现**

先运行：

```bash
/Users/dingwei/go/bin/wails generate module
```

组件实现：

- 挂载时全量加载并启动 2 秒状态轮询。
- 销毁时清理 timer。
- 轮询只调用 runtime 和 Agent 状态 API，不扫描驱动 catalog。
- 轮询错误写入独立状态提示，不覆盖当前操作错误。
- “查看日志”打开模态框，读取 64 KiB。
- 模态框提供刷新、复制、关闭、空状态、文件大小和截断提示。
- 激活 API 返回后立即更新 runtime 和 Agent 组合状态。

**步骤 4：确认通过**

```bash
go test . -run TestJDBCDriverManagerIncludesPollingAndLogViewer -v
/Users/dingwei/go/bin/wails generate module
cd frontend && npm run build
```

bindings 必须由 Wails 生成。记录既有 Svelte 和分块警告，但不跳过构建。

**步骤 5：文档和提交**

```bash
git add frontend_contract_test.go frontend/src/components/JDBCDriverManager.svelte frontend/wailsjs frontend/build/assets/index.js docs/changes/features/2026-07-11-jdbc-runtime-recovery-ui.md
git commit -m "feat: monitor jdbc runtime and logs"
```

## 任务 6：完整回归与真实桌面验收

**文件：**

- 修改：`docs/development/2026-07-09-jdbc-driver-management-implementation.md`
- 创建：`docs/changes/features/2026-07-11-jdbc-runtime-recovery-completion.md`

**步骤 1：建立待验证验收表**

先把以下项目标记为“待验证”，不得预填通过：

- 旧配置迁移。
- 系统 Java选择保存与应用重启恢复。
- 托管 JRE 选择保存与应用重启恢复。
- 切换后 Agent 立即进入 running 或 failed。
- 页面在 2 秒内自动刷新状态。
- 日志模态框读取、刷新、复制和截断提示。
- 应用退出后 Agent 子进程清理。

**步骤 2：完整自动验证**

```bash
go test ./...
cd jdbc-agent && ./gradlew test
./scripts/test-jdbc-agent.sh
cd frontend && npm run build
/Users/dingwei/go/bin/wails build
test -f build/bin/AHaSSHTools.app/Contents/MacOS/AHaSSHTools
```

任何命令失败都先记录阻塞点和最小修复，然后重跑原命令。

**步骤 3：真实桌面验收**

使用构建的 `.app`，且不卸载或覆盖用户驱动：

1. 选择有效系统 Java，确认立即重启并显示结果。
2. 重开应用，确认相同路径恢复。
3. 切换托管 JRE，重开应用后确认仍为 managed。
4. 终止 Agent，确认页面 2 秒内更新。
5. 打开日志模态框，确认内容、刷新、复制和截断提示。
6. 退出应用，确认 Agent 子进程结束。

测试配置优先使用临时 HOME 或专用测试路径。

**步骤 4：据实更新中文记录**

只把有命令输出或桌面观察证据的项目标记为通过。保留未验证项和工具链告警。

**步骤 5：最终提交**

```bash
git add docs/development/2026-07-09-jdbc-driver-management-implementation.md docs/changes/features/2026-07-11-jdbc-runtime-recovery-completion.md
git commit -m "test: verify jdbc runtime recovery flow"
```

## 分支收尾

全部任务完成后使用 `verification-before-completion` 和 `finishing-a-development-branch`：

1. 在功能分支重跑完整验证。
2. 对照设计审查 `git diff master...HEAD`。
3. 按用户既有要求本地快进合并到 `master`，不推送远端。
4. 在合并后的 `master` 上重跑 Go、前端和 Wails 生产构建。
5. 工作区干净后删除功能分支和工作树。

