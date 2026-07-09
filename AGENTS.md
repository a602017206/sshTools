# CODEBUDDY.md

本文件为 CodeBuddy 在本仓库工作时提供指导。

## 常用开发命令

### 应用开发
- `go install github.com/wailsapp/wails/v2/cmd/wails@latest` - 安装 Wails CLI，构建桌面应用时需要。
- `cd frontend && npm install` - 安装 Svelte 前端依赖。
- `wails dev` - 以开发模式运行桌面应用，支持热更新。
- `wails build` - 构建生产二进制到 `build/bin/`。
- `wails build -platform darwin/arm64` - 构建 macOS Apple Silicon 版本；其他平台按需替换目标。

### 后端测试
- `go test ./...` - 运行全部 Go 测试。
- `go test -v ./internal/service` - 以详细模式运行 service 测试。
- `go test -v ./internal/service -run TestFormatJSON` - 运行指定测试函数。
- `go test ./internal/service -cover` - 运行测试并输出覆盖率。

### Go 工具
- `go fmt ./...` - 使用 gofmt 格式化 Go 代码。
- `go vet ./...` - 运行 Go vet 做代码分析。
- `go mod tidy` - 清理 Go module 依赖。

### macOS 分发
- `./scripts/build-mac.sh` - 构建用于分发的 macOS 包，使用 ad-hoc 签名并移除 quarantine 属性。

### Flutter 客户端
- `cd flutter_ui && flutter pub get` - 安装 Flutter 依赖。
- `cd flutter_ui && flutter run -d macos` - 在 macOS 上运行 Flutter 应用；其他设备按需替换。
- `cd flutter_ui && flutter test` - 运行 Flutter 测试。

## 架构概览

这是一个跨平台 SSH 桌面客户端，后端使用 Go，桌面框架使用 Wails。应用把 SSH 终端仿真、SFTP 文件管理、系统监控和可扩展开发者工具集整合到统一 UI 中。

### 后端架构（`internal/`）

**SSH 核心（`ssh/`）**：处理所有 SSH 协议操作。
- `client.go` - SSH 客户端，负责连接管理和认证（密码、keyboard-interactive、密钥）。
- `session.go` - PTY 会话处理、终端 I/O、双向通信。
- `manager.go` - 并发管理多个 SSH 会话。
- `sftp.go` - SFTP 客户端，负责上传、下载、删除、重命名、创建目录等文件操作。
- `transfer.go` - 文件传输任务管理，包含进度追踪和取消。
- `monitor.go` - 实时系统性能监控，包括 CPU、内存、磁盘、网络。

**服务层（`service/`）**：暴露给前端的业务逻辑。
- `devtools_service.go` - 开发者工具后端，支持 JSON 格式化、校验、压缩、转义。
- `devtools_service_test.go` - devtools 功能的 24 个单元测试。

**配置与存储**：
- `config/` - 管理连接配置和应用设置，持久化到 `~/.sshtools/config.json`。
- `store/` - 内存凭据存储；密码使用 AES-256-GCM 加密，存储在 `~/.sshtools/credentials.enc`。
- `crypto/` - 加密操作，包括 AES-GCM 加密和基于机器特征的密钥派生。

**终端（`terminal/`）**：用于 xterm.js 集成的终端仿真层。

### 前端架构（`frontend/src/`）

前端基于 Svelte + Vite，通过 Wails bindings 与 Go 后端通信。

**组件结构**：
- `App.svelte` - 主布局，协调标签页、连接侧栏和右侧面板。
- `TabBar.svelte` - 类浏览器的水平标签栏，支持重命名、关闭确认、自动切换。
- `Terminal.svelte` - 基于 xterm.js 的终端仿真器，支持 PTY 和实时输出。
- `ConnectionManager.svelte` - SSH 连接的增删改查、测试、保存、删除。
- `FileManager.svelte` - 可折叠右侧 SFTP 面板，包含面包屑导航、文件操作、传输进度。
- `MonitorPanel.svelte` - 实时系统指标，包括 CPU 核心、内存、磁盘分区、网络 I/O。
- `DevToolsPanel.svelte` - 可折叠工具面板，包含工具注册系统。
- `tools/JsonFormatter.svelte` - JSON 格式化工具，包含校验、语法高亮、压缩。

**状态管理（`stores/`）**：
- `theme.js` - 明暗主题状态。
- `fileManager.js` - 文件管理器面板状态，包括折叠、宽度、当前路径、文件列表。
- `monitor.js` - 监控面板状态和指标数据。
- `devtools.js` - 工具集状态和工具注册系统。

**工具扩展系统**：
- 新工具在 `frontend/src/tools/index.js` 注册。
- 每个工具是 `frontend/src/components/tools/` 下的 Svelte 组件。
- 工具由 DevToolsPanel 动态发现和渲染。
- 注册项必须包含：id、name、icon、component、category、order。

### 应用入口（`app.go`）

`App` 结构体是主控制器：
- 通过 Wails bindings 向前端导出 Go 方法，导出方法必须首字母大写。
- 通过 SessionManager 管理 SSH 会话生命周期。
- 使用 ConfigManager 管理配置持久化。
- 提供 CredentialStore 进行加密密码存储。
- 向前端发出事件：`ssh:output:{sessionID}`，用于终端数据。

关键导出方法包括：GetConnections、AddConnection、RemoveConnection、TestConnection、ConnectSSH、SendSSHData、ResizeSSH、CloseSSH，以及 devtools 方法（FormatJSON、ValidateJSON、MinifyJSON、EscapeJSON）。

### 配置与数据持久化

数据存储在 `~/.sshtools/`：
- `config.json` - 连接配置和应用设置（主题、侧栏和面板宽度），禁止提交。
- `credentials.enc` - 使用机器绑定密钥派生后的 AES-GCM 加密密码。

## 开发指南

### 变更文档要求
- 每次代码、配置、构建或文档变更，都必须在 `docs/changes/` 下包含对应变更文档。
- 功能和需求变更必须记录在 `docs/changes/features/`。
- Bug 修复必须记录在 `docs/changes/bugs/`。
- 流程、文档、仓库布局和工作流变更必须记录在 `docs/changes/process/`。
- 每份变更文档必须包含：背景、范围、修改文件、验证、剩余风险。
- 文档语言要求：所有新增或修改的文档正文必须使用中文。稳定技术标识、代码符号、文件路径、命令、API 名称和引用的上游名称可按需保留原文。不得新增全英文文档。
- 设计记录和开发记录必须分离：
  - 设计提案、取舍和架构决策放在 `docs/designs/`。
  - 实现说明、执行细节和上线说明放在 `docs/development/`。
- 非平凡变更必须先写或更新设计文档，再写实现说明。
- 不要把无关的功能工作和 bug 修复混在一份变更文档中。
- 如果某次变更故意跳过测试或构建验证，必须明确记录原因和风险。

### 文档布局
- `docs/README.md` 是文档地图，新增或移动文档分类时必须更新。
- `docs/audits/` 存放代码库扫描、评审报告和问题清单。
- `docs/plans/` 存放实现计划。
- `docs/archive/` 存放历史一次性总结、临时调试记录和已废弃报告。
- 根目录 Markdown 只应保留活跃入口和长期指南。
- 包含项目历史的过期文档优先归档，不要直接删除。

### 新增开发者工具
1. 在 `frontend/src/components/tools/YourTool.svelte` 创建 Svelte 组件。
2. 在 `internal/service/devtools_service.go` 添加后端方法。
3. 在 `app.go` 中以导出方法暴露。
4. 在 `frontend/src/tools/index.js` 注册工具，包含 id、name、icon、component、order。
5. 前端在 `wails dev` 时会自动重新生成 bindings。

### 代码风格
- Go：导出符号使用 `CamelCase`，未导出符号使用 `camelCase`，遵循 `gofmt`。
- Svelte：组件文件使用 `PascalCase` 文件名（如 `ConnectionManager.svelte`），store 模块使用 `camelCase`。

### 测试
- Go 单元测试与代码放在同目录，文件名为 `*_test.go`。
- 后端变更提交前运行 `go test ./internal/service -v`。
- UI 回归按 `TESTING_GUIDE.md` 手工检查。

## 平台说明

### macOS 分发
为避免分发时出现 “app is damaged” 错误：
- 运行 `./scripts/build-mac.sh`，该脚本会执行 ad-hoc 签名并移除 quarantine 属性。
- 构建输出在 `build/bin/AHaSSHTools.app`。
- 以 zip 分发；用户下载后运行 `xattr -cr sshTools.app && open sshTools.app`。
- 正式分发需要 Apple Developer Program 签名和 notarization，见 `MACOS_SIGNING.md`。

## 安全注意事项
- 禁止提交 `~/.sshtools/` 目录内容，其中包含真实配置和加密凭据。
- SSH key passphrase 不存盘。
- 密码加密使用 AES-256-GCM，并基于机器特征派生密钥。
- 所有通信都通过本地 Wails bindings 进行，不暴露网络服务。
