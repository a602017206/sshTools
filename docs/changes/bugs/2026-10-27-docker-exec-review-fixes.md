# Docker Exec 终审修复

## 背景

Docker 容器 Exec 功能（`docs/changes/features/2026-10-27-docker-exec-impl.md`）终审发现四处重要问题：

1. 详情栏头部 Exec 按钮永远禁用：`isContainerRunning` 只认列表的 `Up…`，而 Inspect 详情的 `State/Status` 为 `running`。
2. 首包与结束事件竞态：前端在 `StartDockerExec` 返回后才订阅事件，首个提示符或 `docker:exec-closed` 可能先到而丢失，表现为空白终端或一直显示已连接。
3. `UnixTunnel` 泄漏：一侧转发结束后另一侧 `io.Copy` 仍阻塞，exec 结束或关闭后连接与 goroutine 不释放，且本地读不到 EOF。
4. `App.svelte` 等处把 `docker-exec` 当作 SSH 会话：`sshSessions`、`ensureSshSessionActive`、状态墙已连接匹配与指标采集只排除了 `database`/`docker`。

## 范围

- 问题 1：`isContainerRunning` 额外识别 `Status`/`State` 以 `running` 开头。
- 问题 2：采用「前端生成 ID、先订阅后启动」方案。`StartDockerExec` 绑定新增 `execSessionID` 参数（空串时后端生成），后端校验格式与唯一性；前端用 `newDockerExecSessionId()` 生成后先订阅再调用；早到的结束事件使页签以已断开状态创建，输出与断开提示统一按序排队冲刷。未采用「后端缓冲到首次 resize/超时」方案，因为它依赖时序猜测，超时内前端未订阅仍会丢失。
- 问题 3：抽出 `pipeBidirectional`，任一方向结束即关闭两端。
- 问题 4：上述过滤统一改用 `sessionMatchesMode(session, 'ssh')`。

## 修改文件

- `internal/docker/tunnel.go`、`internal/docker/tunnel_test.go`
- `internal/service/docker_exec.go`、`internal/service/docker_exec_test.go`
- `app.go`、`frontend/wailsjs/go/main/App.js`、`frontend/wailsjs/go/main/App.d.ts`
- `frontend/src/lib/dockerWorkspace.js`、`frontend/src/lib/dockerExecSession.js`、`frontend/src/lib/assetWall.js`
- `frontend/src/components/TerminalPanel.svelte`、`frontend/src/App.svelte`、`frontend/src/components/OpsDashboard.svelte`
- `frontend/test/dockerWorkspace.test.js`、`frontend/test/dockerExecSession.test.js`、`frontend/test/assetWall.test.js`
- `docs/changes/features/2026-10-27-docker-exec-impl.md`、`docs/designs/2026-10-27-docker-phase3-exec-resources-design.md`

## 验证

- `go test ./internal/docker ./internal/service -run 'Exec|DockerExec|Tunnel|UnixTunnel' -count=1`：通过；新增 `TestUnixTunnelClosesClientWhenRemoteEnds`、`TestUnixTunnelClosesRemoteWhenClientEnds`、`TestPipeBidirectionalReturnsWhenEitherSideCloses` 在旧实现上超时失败，修复后通过。
- 新增 `TestDockerExecWithClientIDEmitsUnderThatID`（自带 ID、重复/非法 ID、空 ID 自动生成、重复启动失败不误发结束事件）与 `TestDockerExecClosedAfterBufferedOutput`（结束回调在输出之后）。
- 前端：`dockerWorkspace`、`dockerContainerActions`、`dockerExecSession`、`assetWall`、`workspaceTabs` 与三项剪贴板契约测试共 60 项通过；`vite build` 输出到临时目录编译通过。
- 前端全量 `node --test` 中 `nativeDatabasePanelLayout.test.js` 的 Redis/Kafka 两项失败在修复前的 HEAD 上同样存在，与本次无关。

## 剩余风险

- 关闭 exec 只断开 attach 流，容器内 `/bin/sh` 可能残留（moby#9098）。
- `StartDockerExec` 绑定签名变化，若有外部脚本按旧 4 参数调用会失败（当前仓库内仅 `TerminalPanel` 一处调用，已同步）。
- 未做真实 Docker 主机手工联调。
