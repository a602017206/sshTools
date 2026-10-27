# Docker 容器 Exec（新页签）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 对运行中容器一键 Exec，顶栏新开完整 xterm 页签；API 与 CLI 回退双路径；关页签结束 exec。

**Architecture:** `internal/docker` 增加 Create/Start/Resize Exec（TTY hijack 原始流）；CLI 经 SSH/本机 PTY 跑 `docker exec -i -t … /bin/sh`；`DockerService` 维护独立 `execs` 表（execSessionID → 双向流），`CloseDocker`/`CloseAll`/`shutdown` 清理；`app.go` EventsEmit `docker:exec:{execSessionID}`（**base64**，对齐 `ssh:output`）；前端 `type: 'docker-exec'` 复用 `Terminal.svelte`。

**Tech Stack:** Go、Wails Events、Svelte、xterm、`go test`、`node --test`、薄 HTTP（无 moby SDK）

**规格:** `docs/superpowers/specs/2026-10-27-docker-phase3-exec-resources-design.md`（刀 2）

## Global Constraints

- 默认命令：`/bin/sh`（交互）；本期无 shell 选择器
- 仅运行中容器可 Exec；失败返回可读错误
- 事件 payload：**base64 编码字节**（与 SSH 终端解码路径一致），事件名 `docker:exec:{execSessionID}`
- 关 Docker 父页签**不**强制关已开 exec；关 exec 页签必须 `CloseDockerExec`
- `CloseDocker` / `CloseAll` / `shutdown` 必须清理该服务下全部未关 exec（防泄漏）
- 不引入 moby SDK；不改 Terminal 剪贴板受保护区；动到 TerminalPanel/Terminal 相关跑契约三件套
- 文档中文；变更写 `docs/changes/features/`；工作在当前 `master`
- 本计划**仅刀 2**（不含镜像/网络/卷）

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/docker/exec.go` | API：CreateExec、StartExec（hijack）、ResizeExec |
| `internal/docker/exec_test.go` | httptest 单测 |
| `internal/docker/cli_ops.go` | `BuildDockerExecCommand(id, cols, rows)` |
| `internal/ssh/client.go`（或新文件） | `StartPTYCommand(cmd, cols, rows) (ExecStream, error)` — 交互式 PTY |
| `internal/service/docker_exec.go` | exec 生命周期：Start/Send/Resize/Close + CloseDocker 联动 |
| `internal/service/docker_exec_test.go` | mock 流与清理 |
| `app.go` | 导出 Start/Send/Resize/CloseDockerExec + emit |
| `frontend/wailsjs/go/main/App.js` / `App.d.ts` | 手工补绑定 |
| `frontend/src/lib/dockerContainerActions.js` | `canExec`、菜单项 |
| `frontend/src/lib/dockerExecSession.js` | 页签标题/建会话对象纯函数 |
| `frontend/src/lib/sessionTabs.js` 等 | docker-exec 标题/副标题/批量排除/模式匹配 |
| `frontend/src/components/DockerWorkspace.svelte` | Exec 入口回调 |
| `frontend/src/components/DockerContainerDetail.svelte` | 详情头 Exec 按钮 |
| `frontend/src/components/TerminalPanel.svelte` | openDockerExecSession + I/O 路由 |
| `docs/changes/features/2026-10-27-docker-exec-impl.md` | 实现变更记录 |

### 关键接口（跨任务约定）

```go
// internal/docker
type ExecCreateOptions struct {
    Cmd          []string // 默认 ["/bin/sh"]
    AttachStdin  bool
    AttachStdout bool
    AttachStderr bool
    Tty          bool
}

func (c *Client) CreateExec(ctx context.Context, containerID string, opts ExecCreateOptions) (execID string, err error)
// POST /containers/{id}/exec  JSON → {"Id":"..."}

func (c *Client) StartExec(ctx context.Context, execID string, cols, rows int) (io.ReadWriteCloser, error)
// POST /exec/{id}/start  {"Detach":false,"Tty":true}  使用 http.Transport.Dial 后手动 hijack：
// 或 Client.Do + 检查 101/200 后返回 resp.Body 包装的双向 Conn（见下方实现注记）

func (c *Client) ResizeExec(ctx context.Context, execID string, cols, rows int) error
// POST /exec/{id}/resize?h={rows}&w={cols}

// CLI
func BuildDockerExecCommand(containerID string) string
// 返回: docker exec -i -t <id> /bin/sh
// （cols/rows 在 PTY 层设置，不塞进命令行）
```

```go
// DockerService
func (s *DockerService) StartDockerExec(dockerSessionID, containerID string, cols, rows int, emit func([]byte)) (execSessionID string, err error)
func (s *DockerService) SendDockerExecData(execSessionID string, data []byte) error
func (s *DockerService) ResizeDockerExec(execSessionID string, cols, rows int) error
func (s *DockerService) CloseDockerExec(execSessionID string) error
// Close / CloseAll 时 stop 全部 exec
```

```javascript
// 前端会话
{
  sessionId: execSessionID,           // 后端返回
  type: 'docker-exec',
  panelType: 'docker-exec',
  parentDockerSessionId,
  containerId,
  containerName,
  connection,                         // 复用 Docker 资产摘要（展示）
  connected: true,
  createdAt, lastActivity
}
```

**API StartExec 实现注记（Task 1 必读）：**

Docker Engine 对 `POST /exec/{id}/start`（Tty=true）返回 hijacked 连接：请求头加 `Connection: Upgrade`、`Upgrade: tcp`，成功后 `resp.Body` 即为原始双向字节流（TTY 模式**无** 8 字节 multiplex 头）。实现时：

1. 用 `c.transport.DialContext` 或 `http.NewRequest` + 自定义 `RoundTrip` 拿到底层 `net.Conn`；或
2. 简化路径：`req.Header.Set("Connection","Upgrade"); Set("Upgrade","tcp")`，`httpClient.Do`，若 Status 101 或 200，把 `resp.Body` 包成 `ReadWriteCloser`（写入需能写到同一连接——若 Body 只读，则必须 hijack `http.Hijacker` / 自建 dial）。

推荐实现：对 unix/tcp transport 使用 `net/http` 的 `httputil` 风格——`Dial` 连接后手写 HTTP 请求并 `bufio` 读响应头，然后把 `net.Conn` 作为 RWC 返回（与 moby 客户端同类做法，但不引入 SDK）。单测用 `httptest` **无法**完整测 hijack 时，改为：CreateExec/ResizeExec 用 httptest；StartExec 用自定义 `net.Listener` 假服务器返回 200 + body 管道（或抽 `startExecConn` 接口注入）。

---

### Task 1: Engine API CreateExec / ResizeExec（+ StartExec 可测核心）

**Files:**
- Create: `internal/docker/exec.go`
- Create: `internal/docker/exec_test.go`

**Interfaces:** 见上 `CreateExec` / `StartExec` / `ResizeExec`；默认 Cmd `[]string{"/bin/sh"}`，Attach* 与 Tty 全 true。

- [ ] **Step 1: 写失败测试**

`CreateExec`：httptest 断言 `POST /containers/abc/exec`，body 含 `"Tty":true` 与 `Cmd`，响应 `{"Id":"exec-1"}` → 返回 `exec-1`。

`ResizeExec`：断言 `POST /exec/exec-1/resize` 且 query 含 `h=` `w=`。

`StartExec`：至少断言发出 `POST /exec/exec-1/start` 且请求 JSON `Detach=false,Tty=true`（若完整 hijack 难测，可测「非 2xx 返回错误」+ 抽出 `parseExecStartResponse`；或用 `net.Pipe` 假服务）。

- [ ] **Step 2:** `go test ./internal/docker -run Exec -count=1` → FAIL

- [ ] **Step 3: 最小实现** `exec.go`

- [ ] **Step 4:** 测试 PASS

- [ ] **Step 5: Commit** `feat(docker): add Engine API exec create start resize`

---

### Task 2: CLI 命令构建 + SSH/本机交互式 PTY

**Files:**
- Modify: `internal/docker/cli_ops.go`（或 `cli_exec.go`）
- Modify: `internal/docker/cli_ops_test.go`
- Modify: `internal/ssh/client.go`（新增 `StartPTYCommand`）
- Create: `internal/ssh/pty_command.go`（可选拆分）+ 测试（无真实 SSH 时测未连接错误）

**Interfaces:**
- `BuildDockerExecCommand(containerID string) string` → ``docker exec -i -t <id> /bin/sh``（id 经既有校验/quoting：若已有 `shellQuote` 则复用，否则仅允许 `[A-Za-z0-9][A-Za-z0-9_.-]*` 与 validateContainerID 一致）
- `dockerSSHSession` 扩展：`StartPTYCommand(cmd string, cols, rows int) (io.ReadWriteCloser, func(cols, rows int) error, error)`  
  或返回小接口：
  ```go
  type interactiveStream interface {
    io.ReadWriteCloser
    Resize(cols, rows int) error
  }
  ```
- 本机 CLI（若 Docker 会话 mode=CLI 且无 SSH——当前产品路径主要是 SSH CLI；若仅 SSH CLI，本机 PTY 可放到「local API 失败」之外的后续；**本任务至少实现 SSH `StartPTYCommand`**，并用 `creack/pty` 或已有 `golang.org/x/crypto/ssh` RequestPty）

SSH 实现要点：
```go
session, _ := client.NewSession()
session.RequestPty("xterm-256color", rows, cols, modes)
stdin, _ := session.StdinPipe()
// Stdout/Stderr 合并到 pipe 或直接用 session 作为单一流
session.Start(cmd)
// Resize: session.WindowChange(rows, cols)
```

- [ ] **Step 1–4:** TDD：`BuildDockerExecCommand` 单测；`StartPTYCommand` 在未连接 client 上返回错误

- [ ] **Step 5: Commit** `feat(docker): add CLI exec command and SSH PTY helper`

---

### Task 3: DockerService exec 生命周期

**Files:**
- Create: `internal/service/docker_exec.go`
- Create: `internal/service/docker_exec_test.go`
- Modify: `internal/service/docker_service.go`（`dockerSession.close` / `DockerService.Close` / `CloseAll` 调用 `stopAllExecs`）
- Modify: `dockerSSHSession` 接口定义处，加入 `StartPTYCommand`（mock 同步更新）

**行为:**
1. `StartDockerExec`：`validateContainerID`；按 session mode：
   - API：`CreateExec` → `StartExec` → 登记
   - CLI：`BuildDockerExecCommand` + `ssh.StartPTYCommand`
2. 生成 `execSessionID`：`fmt.Sprintf("dexec-%s", uuid或短随机)`
3. 后台 goroutine 读流，`emit([]byte)`；EOF/错误后自动从 map 删除并 Close
4. `SendDockerExecData` → Write；`ResizeDockerExec` → API ResizeExec 或 stream.Resize
5. `CloseDockerExec`：Close 流、删 map
6. `stopAllExecs`：关闭该 **DockerService 全局** 或按 dockerSession 关联的全部 exec（推荐：exec 记录含 `dockerSessionID`，`close()` 某 docker 会话时关掉其子 exec；`CloseAll` 全关）

规格澄清落地：**CloseDocker 清理该会话下登记的 exec**；前端不关子页签，但后端流停掉后前端应能收到关闭（可选 emit `docker:exec-closed:{id}`——**本期可选**；若不做，Terminal 仅停输出。推荐最小：加 `docker:exec-closed:{execSessionID}` 空 payload，前端标 disconnected。）

- [ ] **Step 1:** 测试：mock API client / mock SSH；Start→Send→Close；CloseDocker 后 Send 失败

- [ ] **Step 2–4:** 实现至 PASS

- [ ] **Step 5: Commit** `feat(docker): manage interactive exec sessions in DockerService`

---

### Task 4: app.go 导出 + Wails 绑定

**Files:**
- Modify: `app.go`
- Modify: `frontend/wailsjs/go/main/App.js`、`App.d.ts`

**方法:**
```go
func (a *App) StartDockerExec(dockerSessionID, containerID string, cols, rows int) (string, error)
func (a *App) SendDockerExecData(execSessionID, data string) error
func (a *App) ResizeDockerExec(execSessionID string, cols, rows int) error
func (a *App) CloseDockerExec(execSessionID string) error
```

`StartDockerExec` 内 emit：
```go
encoded := base64.StdEncoding.EncodeToString(data)
runtime.EventsEmit(a.ctx, "docker:exec:"+execSessionID, encoded)
```
可选 closed 事件。确认 `shutdown`/`CloseAll` 路径已调 `dockerService.CloseAll`（若已有则复用）。

- [ ] **Step 1:** 手工补 JS/TS 绑定（与现有 Docker 方法同风格）

- [ ] **Step 2:** `go build -o /dev/null .`

- [ ] **Step 3: Commit** `feat(docker): expose Docker exec APIs to frontend`

---

### Task 5: 前端菜单 canExec + 纯函数

**Files:**
- Modify: `frontend/src/lib/dockerContainerActions.js`
- Modify: `frontend/test/dockerContainerActions.test.js`
- Create: `frontend/src/lib/dockerExecSession.js`
- Create: `frontend/test/dockerExecSession.test.js`
- Modify: `frontend/src/lib/sessionTabs.js` + 其测试（若有）
- Modify: `frontend/src/lib/workspaceTabs.js` + 测试：`sessionMatchesMode` 对 `docker-exec` 在 docker 模式下为 true
- Modify: `frontend/src/lib/sessionTabClose.js`：关闭文案/后端走 `CloseDockerExec`

**规则:**
- `canExec(container)` ≡ `isUp(container)`（与 `isContainerRunning` 同：`/^up\b/i`）
- 菜单在「详情」后插入 `{ id: 'exec', label: 'Exec', disabled: !canExec(container) }`
- `buildDockerExecTabSession({ execSessionId, parentDockerSessionId, containerId, containerName, connection })`
- `sessionTabTitle`：无自定义名时用 ``${containerName || shortId} · exec``
- `sessionTabSubtitle`：短 ID 或 `Docker Exec`
- `isBatchableSession`：排除 `docker-exec`
- `viewportIsConsole`：`docker-exec` 视为 console（true）

- [ ] **Step 1–4:** TDD 纯函数

- [ ] **Step 5: Commit** `feat(docker): add exec menu and session tab helpers`

---

### Task 6: TerminalPanel + DockerWorkspace 接线

**Files:**
- Modify: `frontend/src/components/TerminalPanel.svelte`
- Modify: `frontend/src/components/DockerWorkspace.svelte`
- Modify: `frontend/src/components/DockerContainerDetail.svelte`（可选 Exec 按钮，`export let onExec`）

**行为:**
1. `DockerWorkspace` 增加 `export let onExec = (payload) => {}`；菜单 `exec` → `onExec({ containerId, containerName })`
2. `TerminalPanel`：`onExec` → `openDockerExecSession`：
   - 调 `StartDockerExec(parentId, containerId, cols, rows)`（cols/rows 先用 80/24 或读默认）
   - 成功则 `connectionsStore` 写入 `type:'docker-exec'`，`activeSessionIdStore` 切过去
   - 失败：toast/banner，不建页签
3. 渲染：`session.type === 'docker-exec'` 走与 SSH 相同的 `<Terminal>` 分支（不要进 DockerWorkspace）
4. `handleTerminalData` / resize / close：`docker-exec` → `SendDockerExecData` / `ResizeDockerExec` / `CloseDockerExec`
5. 订阅 `docker:exec:${id}`：与 `ssh:output` 相同 base64 解码写入 terminal（**复用现有订阅函数或抽公共**；勿改剪贴板区）
6. `viewportIsConsole` 条件排除仅 `database`/`docker` workspace，**包含** `docker-exec`

关闭页签：已有 close 路径按 type 分支加 `CloseDockerExec`。

- [ ] **Step 1:** 实现接线

- [ ] **Step 2:** `cd frontend && node --test test/dockerContainerActions.test.js test/dockerExecSession.test.js test/sessionTabs.js`（若测试文件名不同则按实际）  
  以及：`node --test test/terminalClipboardContract.test.js test/terminalSelection.test.js test/terminalShortcuts.test.js`

- [ ] **Step 3:** 写 `docs/changes/features/2026-10-27-docker-exec-impl.md`

- [ ] **Step 4: Commit** `feat(docker): open exec sessions in terminal tabs`

---

### Task 7: 文档收尾 + release 说明（可选条目）

**Files:**
- Modify: `.github/workflows/release.yml`（更新说明加「容器 Exec」一句，若该文件仍用中文说明列表）
- 确认 specs/plans 勾选或 process 变更一份

- [ ] **Step 1:** 更新 release 说明

- [ ] **Step 2: Commit** `docs(docker): note container exec in release notes`

---

## 手工验收

1. 本机 Docker：运行中容器 → Exec → 出现页签 → 输入 `echo hi` 有回显 → 拉窗 resize 正常
2. 停止容器：Exec 菜单禁用
3. 关 exec 页签：后端无残留（再 Start 同容器仍可）
4. SSH CLI 回退环境（若有）：同样可 Exec
5. 关 Docker 工作区页签：exec 页签可仍在；后端流可能已停（若 CloseDocker 清 exec）——UI 可显示断开

## Spec 覆盖自检

| 规格 | 任务 |
|------|------|
| Exec 新页签 + 多开 | Task 5–6 |
| 仅运行中 | Task 5 |
| `/bin/sh` | Task 1–2 |
| API + CLI | Task 1–3 |
| Wails API 表 | Task 4 |
| Close 清理 | Task 3 |
| 复用 Terminal / 剪贴板不动 | Task 6 + 契约测试 |
| 不强制关子页签 UX | Task 6（关父不遍历关子） |
