# JDBC 驱动管理闭环 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 补齐 JDBC 首版验收中未完成的 agent 生产接线、崩溃重连、托管 JRE 在线/离线安装和推荐驱动在线安装，使桌面应用能够从空环境完成安装、连接、查询和恢复。

**架构：** 构建阶段把 Java shadow jar 放入 Wails 嵌入资源，启动时原子部署到 `~/.sshtools/agent`。Go 侧新增惰性 `JDBCAgentSupervisor`，统一选择运行时、启动进程、建立 gRPC client、重启和关闭；`ManagedJDBCGateway` 保存会话配置，在 agent 不可用时重建进程并恢复目标会话。下载和归档安装使用共享的 checksum、临时目录、路径穿越防护和原子替换逻辑。

**技术栈：** Go 1.24、Wails v2、Svelte、Java 21、Gradle 8.5、gRPC、Protocol Buffers、JDBC、H2、Go `testing`。

---

## 实施约束

- 严格按任务顺序执行；每个行为变更先写失败测试并确认红灯，再实现、确认绿灯、写中文变更记录并提交。
- 所有下载必须限制响应大小、校验 SHA-256、写入临时目录并原子提交；失败时不能破坏已安装版本。
- 所有 zip 和 tar.gz 解压必须拒绝绝对路径、`..` 路径穿越、符号链接和目标目录外写入。
- agent 只监听 `127.0.0.1`，token 只保存在内存；错误和日志不得输出密码、token 或完整 JDBC URL。
- 厂商授权不允许自动分发的驱动只保留离线导入入口，不伪造在线来源。
- 遇到 gRPC、protoc、Gradle 或 Java agent 工具链问题，记录阻塞点和最小修复，不更换架构绕过。

## 任务 1：把 Java agent 作为 Wails 构建资源部署

**文件：**

- 创建：`internal/service/jdbc_agent_artifact.go`
- 创建：`internal/service/jdbc_agent_artifact_test.go`
- 创建：`scripts/stage-jdbc-agent.sh`
- 修改：`frontend/package.json`
- 修改：`.gitignore`
- 创建：`docs/changes/features/2026-07-10-jdbc-agent-artifact.md`

**步骤 1：写失败测试**

新增 `TestAgentArtifactInstallerWritesJarAtomically`：使用临时目录和固定 jar 字节调用 `Install`，断言返回路径为 `<AgentDir>/jdbc-agent.jar`、内容一致、权限不是全局可写；再次安装相同内容不改变文件。

**步骤 2：确认失败**

```bash
go test ./internal/service -run TestAgentArtifactInstallerWritesJarAtomically -v
```

预期：`NewAgentArtifactInstaller` 未定义。

**步骤 3：实现**

实现：

```go
type AgentArtifactInstaller struct { paths JDBCPaths }
func NewAgentArtifactInstaller(paths JDBCPaths) *AgentArtifactInstaller
func (s *AgentArtifactInstaller) Install(jar []byte) (string, error)
```

`Install` 计算现有文件与输入内容的 SHA-256；内容不同时写入同目录临时文件、`Sync`、`Chmod(0o600)` 后 `Rename`。`scripts/stage-jdbc-agent.sh` 执行 `jdbc-agent/gradlew shadowJar` 并把 jar 复制到 `frontend/build/jdbc-agent.jar`；`npm run build` 在 Vite 构建后执行该脚本。忽略生成的嵌入 jar。

**步骤 4：确认通过**

```bash
go test ./internal/service -run TestAgentArtifactInstallerWritesJarAtomically -v
cd frontend && npm run build
test -f frontend/build/jdbc-agent.jar
```

**步骤 5：提交**

```bash
git add .gitignore frontend/package.json internal/service/jdbc_agent_artifact.go internal/service/jdbc_agent_artifact_test.go scripts/stage-jdbc-agent.sh docs/changes/features/2026-07-10-jdbc-agent-artifact.md
git commit -m "feat: package jdbc agent artifact"
```

## 任务 2：实现惰性 agent supervisor 和真实 gRPC client

**文件：**

- 创建：`internal/service/jdbc_agent_supervisor.go`
- 创建：`internal/service/jdbc_agent_supervisor_test.go`
- 修改：`internal/service/jdbc_agent_process.go`
- 创建：`docs/changes/features/2026-07-10-jdbc-agent-supervisor.md`

**步骤 1：写失败测试**

新增 `TestJDBCAgentSupervisorStartsOnceAndReturnsAuthenticatedClient`：注入假的运行时选择器、进程启动器和 dialer；连续调用两次 `Client`，断言只启动和拨号一次，dialer 收到 `127.0.0.1`、动态端口和 handle token。新增 `TestJDBCAgentSupervisorRestartClosesClientAndRotatesProcess`，断言旧连接和旧进程都被关闭后创建新 token。

**步骤 2：确认失败**

```bash
go test ./internal/service -run TestJDBCAgentSupervisor -v
```

预期：`NewJDBCAgentSupervisor` 未定义。

**步骤 3：实现**

定义小接口 `JDBCRuntimeSelector`、`JDBCAgentStarter`、`JDBCAgentDialer`，生产适配器复用 `RuntimeService`、`AgentProcessManager` 和 `NewGRPCJdbcAgentClient`。`Client(ctx)` 在互斥锁内惰性选择 Java、启动进程并带超时拨号；缺少运行时映射为 `RUNTIME_MISSING`。`Restart(ctx)` 和 `Close()` 必须先关 gRPC 连接再停止进程，并清空内存状态。

**步骤 4：确认通过**

```bash
go test ./internal/service -run TestJDBCAgentSupervisor -v
```

**步骤 5：提交**

```bash
git add internal/service/jdbc_agent_supervisor.go internal/service/jdbc_agent_supervisor_test.go internal/service/jdbc_agent_process.go docs/changes/features/2026-07-10-jdbc-agent-supervisor.md
git commit -m "feat: supervise jdbc agent runtime"
```

## 任务 3：实现可恢复 gateway 并接入应用生命周期

**文件：**

- 创建：`internal/service/jdbc_managed_gateway.go`
- 创建：`internal/service/jdbc_managed_gateway_test.go`
- 修改：`internal/service/database_service.go`
- 修改：`internal/service/database_service_test.go`
- 修改：`app.go`
- 修改：`main.go`
- 创建：`app_jdbc_test.go`
- 创建：`docs/changes/features/2026-07-10-jdbc-live-gateway.md`

**步骤 1：写失败测试**

- `TestManagedJDBCGatewayReconnectsSessionAfterAgentUnavailable`：第一次 query 返回 `AGENT_UNAVAILABLE`，断言 supervisor 重启、原连接配置重新打开、query 只重试一次并成功。
- `TestDatabaseServiceTestConnectionUsesGateway`：有 gateway 时使用临时 session 执行 connect/close，不调用 Go 原生 DSN。
- `TestBuildJDBCServicesInjectsManagedGateway`：构造临时根目录和 fake supervisor，断言返回的 `DatabaseService` 通过 managed gateway 调用。

**步骤 2：确认失败**

```bash
go test ./internal/service -run 'TestManagedJDBCGateway|TestDatabaseServiceTestConnectionUsesGateway' -v
go test . -run TestBuildJDBCServicesInjectsManagedGateway -v
```

**步骤 3：实现**

`ManagedJDBCGateway` 保存成功连接的 `sessionID -> DatabaseConfig`，每次调用通过 supervisor 获取当前 client 并创建带 token 的 `JdbcGatewayService`。仅在错误码为 `AGENT_UNAVAILABLE` 时重启一次、重新打开目标 session 并重试原操作；其他错误直接返回。应用启动时从嵌入资源读取 `frontend/build/jdbc-agent.jar` 并部署，构造 supervisor 和 managed gateway；Wails `OnShutdown` 调用 supervisor `Close`。`RestartJDBCAgent` 统一走 supervisor，不再直接操作旧 manager。

**步骤 4：确认通过**

```bash
go test ./internal/service -run 'TestManagedJDBCGateway|TestDatabaseServiceTestConnectionUsesGateway' -v
go test . -run TestBuildJDBCServicesInjectsManagedGateway -v
go test ./...
```

**步骤 5：提交**

```bash
git add internal/service/jdbc_managed_gateway.go internal/service/jdbc_managed_gateway_test.go internal/service/database_service.go internal/service/database_service_test.go app.go app_jdbc_test.go main.go docs/changes/features/2026-07-10-jdbc-live-gateway.md
git commit -m "feat: connect app to managed jdbc agent"
```

## 任务 4：实现安全的离线 JRE 归档导入

**文件：**

- 修改：`internal/service/jdbc_runtime.go`
- 修改：`internal/service/jdbc_runtime_test.go`
- 创建：`internal/service/archive_extract.go`
- 创建：`internal/service/archive_extract_test.go`
- 创建：`docs/changes/features/2026-07-10-jdbc-runtime-import.md`

**步骤 1：写失败测试**

- `TestRuntimeServiceImportsJREArchive`：创建包含 `jdk-test/bin/java` 的 zip，导入后断言 Java 位于 `RuntimesDir/jre-<version>/bin/java` 且可执行。
- `TestArchiveExtractorRejectsPathTraversal`：归档含 `../escape` 时返回错误且目标目录不存在。
- `TestRuntimeServiceRollsBackInvalidArchive`：没有 `bin/java` 时不留下安装目录。

**步骤 2：确认失败**

```bash
go test ./internal/service -run 'TestRuntimeServiceImports|TestArchiveExtractor|TestRuntimeServiceRollsBack' -v
```

预期：现有 `ImportRuntimeArchive` 返回“暂未实现”。

**步骤 3：实现**

支持 `.zip`、`.tar.gz` 和 `.tgz`，只提取普通文件和目录，拒绝符号链接及路径穿越。自动查找唯一的 `bin/java`，把其 JRE 根目录原子移动到 `RuntimesDir/jre-<version>-<os>-<arch>`，设置 java 可执行权限并返回 `RuntimeKindManaged`。

**步骤 4：确认通过**

```bash
go test ./internal/service -run 'TestRuntimeServiceImports|TestArchiveExtractor|TestRuntimeServiceRollsBack' -v
```

**步骤 5：提交**

```bash
git add internal/service/jdbc_runtime.go internal/service/jdbc_runtime_test.go internal/service/archive_extract.go internal/service/archive_extract_test.go docs/changes/features/2026-07-10-jdbc-runtime-import.md
git commit -m "feat: import managed jdbc runtime"
```

## 任务 5：实现托管 JRE 在线安装

**文件：**

- 创建：`internal/service/artifact_download.go`
- 创建：`internal/service/artifact_download_test.go`
- 创建：`internal/service/jdbc_runtime_provider.go`
- 创建：`internal/service/jdbc_runtime_provider_test.go`
- 修改：`internal/service/jdbc_runtime.go`
- 修改：`app.go`
- 创建：`docs/changes/features/2026-07-10-jdbc-runtime-download.md`

**步骤 1：写失败测试**

- `TestArtifactDownloaderValidatesChecksumAndCommitsAtomically`：使用 `httptest.Server` 返回归档，断言 checksum 正确时提交、错误时目标不存在。
- `TestAdoptiumRuntimeProviderSelectsCurrentPlatformPackage`：使用固定 API JSON，断言选择当前 `runtime.GOOS/runtime.GOARCH` 的 JRE 21 包和 checksum。
- `TestRuntimeServiceInstallsManagedRuntime`：注入 provider/downloader，断言下载后复用任务 4 的导入路径。

**步骤 2：确认失败**

```bash
go test ./internal/service -run 'TestArtifactDownloader|TestAdoptiumRuntimeProvider|TestRuntimeServiceInstallsManagedRuntime' -v
```

**步骤 3：实现**

下载器只允许 `https`，测试通过显式选项允许本地 HTTP；设置 1 GiB 上限、连接和总超时、SHA-256 校验。Adoptium provider 使用官方 API 获取 Java 21 JRE 包元数据，不在代码中硬编码临时下载 URL。`App.InstallJDBCManagedRuntime()` 调用安装并返回新状态。

**步骤 4：确认通过**

```bash
go test ./internal/service -run 'TestArtifactDownloader|TestAdoptiumRuntimeProvider|TestRuntimeServiceInstallsManagedRuntime' -v
go test ./...
```

**步骤 5：提交**

```bash
git add internal/service/artifact_download.go internal/service/artifact_download_test.go internal/service/jdbc_runtime_provider.go internal/service/jdbc_runtime_provider_test.go internal/service/jdbc_runtime.go app.go docs/changes/features/2026-07-10-jdbc-runtime-download.md
git commit -m "feat: install managed jdbc runtime"
```

## 任务 6：实现推荐 JDBC 驱动在线安装和内置清单

**文件：**

- 创建：`internal/service/jdbc_builtin_manifest.json`
- 修改：`internal/service/jdbc_catalog.go`
- 修改：`internal/service/jdbc_catalog_test.go`
- 修改：`internal/service/jdbc_install.go`
- 修改：`internal/service/jdbc_install_test.go`
- 修改：`app.go`
- 创建：`docs/changes/features/2026-07-10-jdbc-online-driver-install.md`

**步骤 1：写失败测试**

- `TestDriverCatalogBootstrapsBuiltinManifest`：用户清单不存在时复制内置清单并可列出首批八类数据库。
- `TestDriverInstallDownloadsProfileJarsAtomically`：使用本地测试服务器和两个 jar，断言全部 checksum 正确后写 `driver.json`；第二个 jar 错误时不留下目标版本。
- `TestDriverInstallRejectsProfileWithoutOfficialURLs`：受限驱动没有 URL 时返回 `DRIVER_MISSING` 并提示离线导入。

**步骤 2：确认失败**

```bash
go test ./internal/service -run 'TestDriverCatalogBootstraps|TestDriverInstallDownloads|TestDriverInstallRejects' -v
```

**步骤 3：实现**

用 `go:embed` 保存内置清单；只为可依法公开下载且有 SHA-256 的 jar 配置在线来源，Oracle、达梦、人大金仓等无可靠公开分发来源时保留 profile 但不配置 URL。在线安装逐个下载到临时版本目录，全部通过 checksum 后原子替换。`App.InstallJDBCDriver` 从 catalog 解析指定版本并调用安装器。

**步骤 4：确认通过**

```bash
go test ./internal/service -run 'TestDriverCatalogBootstraps|TestDriverInstallDownloads|TestDriverInstallRejects' -v
go test ./...
```

**步骤 5：提交**

```bash
git add internal/service/jdbc_builtin_manifest.json internal/service/jdbc_catalog.go internal/service/jdbc_catalog_test.go internal/service/jdbc_install.go internal/service/jdbc_install_test.go app.go docs/changes/features/2026-07-10-jdbc-online-driver-install.md
git commit -m "feat: install recommended jdbc drivers"
```

## 任务 7：完成运行时 UI、文件选择和 agent 状态

**文件：**

- 修改：`app.go`
- 修改：`frontend/src/components/JDBCDriverManager.svelte`
- 修改：`frontend/src/components/DatabasePanel.svelte`
- 修改：`frontend/wailsjs/go/main/App.js`
- 修改：`frontend/wailsjs/go/main/App.d.ts`
- 修改：`frontend/wailsjs/go/models.ts`
- 创建：`docs/changes/features/2026-07-10-jdbc-management-actions.md`

**步骤 1：写失败后端测试**

新增 `TestJDBCManagementAPIReturnsAgentAndRuntimeState`，断言状态包含 agent 的 `stopped/starting/running/failed`、运行时类型和最后错误；文件选择方法由 Wails runtime 适配器注入，测试不启动 GUI。

**步骤 2：确认失败**

```bash
go test . -run TestJDBCManagementAPIReturnsAgentAndRuntimeState -v
```

**步骤 3：实现后端和前端**

增加 `InstallJDBCManagedRuntime`、`ImportJDBCRuntimeArchive`、`SelectJDBCRuntimeArchive`、`SelectJDBCDriverPackage`、`GetJDBCAgentStatus`。前端移除路径 `prompt`，改用文件选择器；“安装 JRE”调用真实在线安装；状态条显示 supervisor 状态和最后错误。重新生成 Wails bindings，不手工伪造签名。

**步骤 4：确认通过**

```bash
go test . -run TestJDBCManagementAPIReturnsAgentAndRuntimeState -v
/Users/dingwei/go/bin/wails generate module
cd frontend && npm run build
```

**步骤 5：提交**

```bash
git add app.go frontend/src/components/JDBCDriverManager.svelte frontend/src/components/DatabasePanel.svelte frontend/wailsjs docs/changes/features/2026-07-10-jdbc-management-actions.md
git commit -m "feat: complete jdbc management actions"
```

## 任务 8：补充崩溃恢复集成测试和最终验收

**文件：**

- 修改：`internal/service/jdbc_integration_test.go`
- 修改：`scripts/test-jdbc-agent.sh`
- 修改：`scripts/build-mac.sh`
- 修改：`docs/development/2026-07-09-jdbc-driver-management-implementation.md`
- 修改：`docs/changes/features/2026-07-09-jdbc-driver-management-rollout.md`
- 创建：`docs/changes/features/2026-07-10-jdbc-driver-management-completion.md`

**步骤 1：写失败集成测试**

新增 `TestJDBCAgentRecoversSessionAfterCrash`：通过 supervisor 连接 H2、建表，终止 agent，执行查询；断言返回结果来自自动重启后的新会话。测试数据库使用文件模式，避免进程重启后内存库数据丢失。

**步骤 2：确认失败**

```bash
./scripts/test-jdbc-agent.sh
```

预期：崩溃后 query 返回 `AGENT_UNAVAILABLE`，没有恢复会话。

**步骤 3：补齐测试脚本与构建流程**

确保集成脚本先生成 shadow jar；macOS 构建脚本校验应用内嵌 agent 资源存在。更新中文实施记录和发布清单，只把有直接证据的项目标记为通过。

**步骤 4：完整验证**

```bash
go test ./...
cd jdbc-agent && ./gradlew test
./scripts/test-jdbc-agent.sh
cd frontend && npm run build
/Users/dingwei/go/bin/wails build
test -f build/bin/AHaSSHTools.app/Contents/MacOS/AHaSSHTools
```

手工检查：

- 空运行时状态可安装托管 JRE，或选择离线归档/系统 Java。
- 可在线安装有官方公开来源的推荐驱动；受限驱动明确要求离线包。
- 离线导入 H2 后可连接和查询。
- 杀死 agent 后界面显示恢复状态，目标 session 自动重连一次。
- 应用退出后 agent 进程被清理。

**步骤 5：提交**

```bash
git add internal/service/jdbc_integration_test.go scripts/test-jdbc-agent.sh scripts/build-mac.sh docs/development/2026-07-09-jdbc-driver-management-implementation.md docs/changes/features/2026-07-09-jdbc-driver-management-rollout.md docs/changes/features/2026-07-10-jdbc-driver-management-completion.md
git commit -m "test: verify complete jdbc management flow"
```

