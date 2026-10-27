# Docker 容器 Exec（交互终端页签）

## 背景

Docker 三期设计（`docs/designs/2026-10-27-docker-phase3-exec-resources-design.md`）要求对运行中的容器提供交互 shell。实现计划见 `docs/plans/2026-10-27-docker-exec.md`。本文汇总整个功能涉及的后端、绑定、前端与发布改动，以及终审后的修复（详见 `docs/changes/bugs/2026-10-27-docker-exec-review-fixes.md`）。

## 范围

### 后端

- Engine API：`CreateExec`（`POST /containers/{id}/exec`，Attach Stdin/Stdout/Stderr + Tty）、`StartExec`（`POST /exec/{id}/start` 经 HTTP Upgrade 劫持为原始双向流，启动后按初始尺寸 resize 一次）、`ResizeExec`。
- CLI 回退：`BuildDockerExecCommand` 生成 `docker exec -i -t {id} /bin/sh`；`ssh.StartPTYCommand` 在 Docker 专用 SSH 连接上申请 PTY 运行该命令，resize 映射为 PTY window-change。
- `DockerService`：`StartDockerExec` / `StartDockerExecWithID`、`SendDockerExecData`、`ResizeDockerExec`、`CloseDockerExec`；exec 挂在所属 `dockerSession` 下，`CloseDocker` / `CloseAll` 会停止其全部 exec 并拒绝新建；EOF、出错或主动关闭时通过 `SetDockerExecClosedHandler` 回调一次结束事件。
- `StartDockerExecWithID` 接受调用方给定的 exec 会话 ID（须匹配 `^dexec-[A-Za-z0-9_-]{6,64}$`，重复 ID 拒绝），为空时自动生成；启动失败不发结束事件。
- `UnixTunnel`（SSH 隧道到远端 `docker.sock`）：任一方向转发结束即关闭两端，避免 exec 结束或关闭后另一方向 `io.Copy` 永久阻塞而泄漏连接与 goroutine。

### Wails 绑定与事件

- `app.go` 导出 `StartDockerExec(dockerSessionID, containerID, execSessionID, cols, rows)`、`SendDockerExecData`、`ResizeDockerExec`、`CloseDockerExec`；`execSessionID` 由前端预先生成，传空串时后端生成。
- 事件：`docker:exec:{execSessionID}`（base64 输出，与 `ssh:output` 一致）、`docker:exec-closed:{execSessionID}`（结束，在已缓冲输出之后发出）。
- `frontend/wailsjs/go/main/App.js`、`App.d.ts` 同步上述签名。

### 前端

- 容器行菜单「Exec」（`canExec`，仅 `Up…` 状态可用）与详情栏头部「Exec」按钮（`isContainerRunning` 同时识别列表 `Up…` 与 Inspect 的 `running`），均触发 `onExec({ containerId, containerName })`。
- `TerminalPanel.openDockerExecSession`：先用 `newDockerExecSessionId()` 生成 ID 并订阅输出/结束事件，再调用 `StartDockerExec`；成功后写入 `type: 'docker-exec'` 会话并激活页签，失败则退订并在面板顶部提示，不建页签。
- 输出与「[容器终端已断开]」提示按到达顺序排队，页签未建或 Terminal 未挂载时缓存，挂载后按序冲刷；结束事件早于页签创建时，页签直接以已断开状态创建。
- `docker-exec` 与 SSH 共用 `<Terminal>`；输入、resize、关闭分别走 `SendDockerExecData`、`ResizeDockerExec`、`CloseDockerExec`；不参与批量执行、ZMODEM、命令提示与编码切换。
- 页签：标题 `{容器名} · exec`、副标题短 ID、按容器分组序号；`viewportIsConsole(session)` 视 exec 为控制台；`sessionMatchesMode` 把 exec 归入容器模式。
- SSH 侧过滤：`App.svelte` 的 `sshSessions` / `ensureSshSessionActive`、`assetWall.js` 的已连接会话匹配、`OpsDashboard.svelte` 的指标采集均改用 `sessionMatchesMode(session, 'ssh')`，排除 `docker-exec`，避免把 exec 当成 SSH 会话（绑定文件面板、采集监控、状态墙误判已连接）。
- 关闭父 Docker 页签不会级联关闭 exec 页签。
- 不在范围：`Terminal.svelte`、剪贴板受保护区域。

### 发布

- `.github/workflows/release.yml` 发布说明新增「Docker 三期 — 容器 Exec」条目。

## 修改文件

- 后端：`internal/docker/exec.go`、`internal/docker/exec_test.go`、`internal/docker/cli_ops.go`、`internal/docker/cli_ops_test.go`、`internal/docker/tunnel.go`、`internal/docker/tunnel_test.go`、`internal/ssh/pty_command.go`、`internal/ssh/pty_command_test.go`、`internal/service/docker_exec.go`、`internal/service/docker_exec_test.go`、`internal/service/docker_service.go`、`internal/service/docker_service_test.go`、`app.go`。
- 绑定：`frontend/wailsjs/go/main/App.js`、`frontend/wailsjs/go/main/App.d.ts`。
- 前端组件：`frontend/src/components/TerminalPanel.svelte`、`DockerWorkspace.svelte`、`DockerContainerDetail.svelte`、`OpsDashboard.svelte`、`frontend/src/App.svelte`。
- 前端纯函数：`frontend/src/lib/dockerExecSession.js`、`dockerContainerActions.js`、`dockerWorkspace.js`、`sessionTabs.js`、`sessionTabClose.js`、`workspaceTabs.js`、`assetWall.js`。
- 前端测试：`frontend/test/dockerExecSession.test.js`、`dockerContainerActions.test.js`、`dockerWorkspace.test.js`、`assetWall.test.js`。
- 发布：`.github/workflows/release.yml`。
- 文档：本文、`docs/changes/bugs/2026-10-27-docker-exec-review-fixes.md`、`docs/designs/2026-10-27-docker-phase3-exec-resources-design.md`（API 签名）。

## 验证

- `go test ./internal/docker ./internal/service -run 'Exec|DockerExec|Tunnel|UnixTunnel' -count=1`：通过（Unix socket 相关用例需在允许 bind 的环境运行）。
- 隧道新用例在旧实现上会超时失败，修复后通过。
- `cd frontend && node --test test/dockerWorkspace.test.js test/dockerContainerActions.test.js test/dockerExecSession.test.js test/terminalClipboardContract.test.js test/terminalSelection.test.js test/terminalShortcuts.test.js test/assetWall.test.js src/lib/workspaceTabs.test.js`：60 项通过。
- `vite build`（输出到临时目录）编译通过。
- 未做真实 Docker 主机的手工联调。

## 剩余风险

- 容器内 shell 残留：关闭 exec 页签只会断开 attach 流，Docker 不会因此终止 exec 进程（moby#9098），`/bin/sh` 可能在容器内残留到容器停止；CLI 回退下关闭 SSH PTY 通常会使 `docker exec` 客户端退出，但容器内进程同样可能残留。
- 首包与结束事件依赖「前端先订阅、后启动」；若有其他调用方不传 `execSessionID`，仍会退回到旧的竞态。
- Terminal 挂载超过约 1 秒时排队的输出不会主动冲刷，需等下一块输出到达（所有会话常驻挂载，实际很少发生）。
- 初始 PTY 尺寸为 80x24，挂载后按实际终端尺寸再发一次 resize。
- 父 Docker 连接关闭后，后端会结束其 exec，页签保留为已断开状态，需手动关闭。
- 镜像未内置 `/bin/sh`（如 distroless）时只会看到错误输出并立即断开，暂不支持选择其他 shell。
