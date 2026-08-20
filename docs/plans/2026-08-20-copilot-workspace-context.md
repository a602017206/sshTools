# Copilot 工作目录与上下文修复 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让 Shell Copilot 能安全识别当前目录内的启动脚本，并在同一会话中获得可靠的工作目录与近期终端上下文。

**Architecture:** 将 SSH 工具从“任意但极小白名单命令”调整为结构化、无参数的目录清单工具；工具层只在已验证的 `WorkingDir` 中执行固定的只读命令。会话管理器统一暴露会话类型和工作目录，拒绝在本地会话上执行 SSH 探测；前端把当前终端输出保留为限长、脱敏的请求上下文，并保持已有的按会话历史隔离。

**Tech Stack:** Go、Wails v2、Svelte、Node test runner。

---

### Task 1: 为目录浏览工具建立失败用例

**Files:**
- Modify: `internal/service/copilot/probe_test.go`
- Modify: `internal/service/copilot/service_test.go`
- Modify: `internal/service/copilot/probe.go`
- Modify: `internal/service/copilot/tools.go`

**Step 1: Write the failing test**

新增测试，要求 SSH 模式暴露 `list_working_directory` 工具；模型调用它时，CommandRunner 收到固定、只读的目录清单命令，且命令在请求的工作目录中运行。另新增非法/空工作目录被拒绝的测试。

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service/copilot -run 'TestService.*WorkingDirectory|TestWorkingDirectory' -count=1`

Expected: FAIL，因为该工具和目录命令构造函数尚不存在。

**Step 3: Write minimal implementation**

添加无参数 `list_working_directory` 工具。它只接受绝对、无 NUL 的 `WorkingDir`，使用 Go 的 shell 单引号转义构造 `cd -- <dir> && LC_ALL=C command ls -la --`。保留原有系统探测白名单，不提供模型可控的通用命令。

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service/copilot -run 'TestService.*WorkingDirectory|TestWorkingDirectory' -count=1`

Expected: PASS。

### Task 2: 防止本地会话进入 SSH 执行路径

**Files:**
- Modify: `internal/ssh/manager.go`
- Modify: `internal/ssh/manager_test.go`

**Step 1: Write the failing test**

新增本地 `ManagedSession` 的测试，要求 `ExecuteCommand` 返回清晰错误，不访问空的 SSH session。

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ssh -run TestExecuteCommandRejectsLocalSession -count=1`

Expected: FAIL 或 panic，证明现状错误。

**Step 3: Write minimal implementation**

在 `SessionManager.ExecuteCommand` 中先校验 `Type == SessionTypeSSH` 和 `Session != nil`，否则返回“仅支持 SSH 会话”的错误。

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ssh -run TestExecuteCommandRejectsLocalSession -count=1`

Expected: PASS。

### Task 3: 将当前目录和近期终端输出传入 Copilot

**Files:**
- Modify: `frontend/src/stores/copilot.js`
- Modify: `frontend/src/components/TerminalPanel.svelte`
- Modify: `frontend/src/components/AIPanel.svelte`
- Modify: `frontend/test/copilotContext.test.js`

**Step 1: Write the failing test**

为独立的上下文辅助模块写测试：按会话追加输出、仅保留最近固定字符数、发送前可读取对应 session 的 tail。

**Step 2: Run test to verify it fails**

Run: `cd frontend && node --test test/copilotContext.test.js`

Expected: FAIL，因为上下文辅助模块尚不存在。

**Step 3: Write minimal implementation**

增加可测试的终端 tail 缓冲；SSH 与本地输出均写入。AIPanel 发送请求时传入当前会话 tail；后端现有 `Redact` 继续作为出站脱敏防线。工作目录继续只由后端填充，避免浏览器伪造。

**Step 4: Run test to verify it passes**

Run: `cd frontend && node --test test/copilotContext.test.js`

Expected: PASS。

### Task 4: 修复本地会话工作目录同步

**Files:**
- Modify: `internal/ssh/manager.go`
- Modify: `internal/ssh/manager_test.go`
- Modify: `app.go`
- Modify: `frontend/src/components/TerminalPanel.svelte`

**Step 1: Write the failing test**

新增本地会话初始工作目录测试，要求会话创建时记录 `os.Getwd()`，`GetCurrentWorkingDirectory` 不调用 SSH 命令。

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ssh -run TestCreateLocalSessionTracksInitialWorkingDirectory -count=1`

Expected: FAIL。

**Step 3: Write minimal implementation**

创建本地会话时写入初始 cwd；允许前端从 OSC 7 更新本地会话 cwd（移除本地会话提前返回），并在 `fillCopilotRequest` 中读取该状态。远程端仍保持既有 CWD 跟踪。

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ssh -run 'Test(CreateLocalSessionTracksInitialWorkingDirectory|ExecuteCommandRejectsLocalSession)' -count=1`

Expected: PASS。

### Task 5: 文档、全量验证与提交

**Files:**
- Create: `docs/designs/2026-08-20-copilot-workspace-context.md`
- Create: `docs/development/2026-08-20-copilot-workspace-context.md`
- Create: `docs/changes/bugs/2026-08-20-copilot-workspace-context.md`
- Modify: `docs/README.md`（如文档地图需新增入口）

**Step 1: 完成中文设计、开发记录和变更说明**

明确安全边界：生成阶段只读、目录被固定为当前会话目录、用户仍需确认后才能执行产物；记录本地会话不支持 SSH 探测。

**Step 2: Run verification**

Run:

```bash
go test ./internal/service/copilot ./internal/ssh -count=1
cd frontend && node --test test/copilotApply.test.js test/copilotContext.test.js
cd frontend && npm run build
```

Expected: 全部通过。

**Step 3: Commit**

```bash
git add internal app.go frontend docs
git commit -m "fix: add copilot workspace context"
```
