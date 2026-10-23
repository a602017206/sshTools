# 变更：Docker 容器连接实现（本机 / SSH / TCP）

## 背景

依据 `2026-10-23-docker-connection-design.md` 规格，Docker 域需要真实可用的接入能力：本机 socket / `DOCKER_HOST`、基于已保存 SSH 资产的隧道，以及无密码 TCP（2375）。此前仅有域导航与表单占位，无法连接并列出容器。

## 范围

- 后端：`internal/docker` 提供传输解析、Unix socket 隧道与 CLI 回退解析；`DockerService` 提供 `TestDockerConnection`、`ConnectDocker`、容器列表、`DetectLocalDocker`；`app.go` 暴露对应方法。
- 前端：添加连接表单按 transport（local / ssh / tcp）展示字段；新增 `DockerWorkspace` 列出容器；Docker 资产不映射为 SSH 工作区。
- 入口（本次任务 8）：
  - Docker 域内若 `DetectLocalDocker().ok` 且尚无 `metadata.transport=local` 的 Docker 资产，侧栏顶部显示「添加本机 Docker」，点击后以 `transport: local` 预置打开添加对话框。
  - SSH 资产右键菜单新增「添加 Docker」，预置 `transport: ssh` 与 `sshConnectionId`，名称默认 `<SSH 名>-Docker`。
  - `App.svelte` 的 `openAddConnection(group, defaults)` 支持传入 defaults，并透传给 `AddAssetDialog`。
- 非目标：容器日志、exec、完整 TLS 支持。

## 修改文件

- `frontend/src/components/AssetList.svelte`（本机探测提示、SSH 右键「添加 Docker」）
- `frontend/src/App.svelte`（`openAddConnection` defaults 透传）
- `internal/docker/types.go`（`ContainerSummary.Via`）、`internal/service/docker_service.go`（`CloseAll`、CLI 超时、`via: cli`）、`app.go`（shutdown 关闭 Docker 会话）
- `frontend/src/lib/sessionTabClose.js`、`TerminalPanel.svelte`（Docker 关闭确认文案）
- `docs/changes/features/2026-10-23-docker-connection-impl.md`（本文件）
- 此前任务已提交的后端与前端实现见 `git log` 中 `feat(docker)` 系列提交（`internal/docker`、`internal/service/docker_service.go`、`app.go`、`AddAssetDialog.svelte`、`DockerWorkspace.svelte`、`lib/dockerTransport.js` 等）。

## 验证

- `go test ./internal/docker ./internal/service -count=1`
- `cd frontend && node --test test/dockerTransport.test.js src/lib/workspaceTabs.test.js`
- `cd frontend && npm run build`（若因 gradle 沙箱限制失败，则以 `npx vite build` 验证）
- 终端复制粘贴契约测试（本次未改终端代码，作为回归）：`cd frontend && node --test test/terminalClipboardContract.test.js test/terminalSelection.test.js test/terminalShortcuts.test.js`
- 手工项：Docker 域无本机资产时出现提示；添加后提示消失；SSH 资产右键可打开预置 SSH 关联的 Docker 添加表单。

## 剩余风险

- 本机探测仅在首次进入 Docker 域时执行一次，Docker 启动晚于应用时需重启或切换后手动添加。
- 本机提示依赖 `metadata.transport === 'local'` 判断，历史手工创建且缺少该字段的 Docker 资产不会被识别。
- 隧道失败时依赖 CLI 回退，本机无 `docker` CLI 时错误提示可能不够细化。
- Windows 命名管道（`\\.\pipe\docker_engine`）目前仅作为默认路径解析，尚未实际 dial，Windows 本机 Docker 连接暂不可用。
- Docker 的 SSH 路径仅支持密码与无口令私钥；带 passphrase 的密钥不受支持（passphrase 不落盘，此路径无交互输入）。
- 本机探测为「一次性检测」（detect-once）：仅首次进入 Docker 域时执行，之后不会自动重试。
- 远程 CLI 回退的 `docker ps` 受 `dockerOpTimeout`（15 秒）限制，超时后返回错误；容器行带 `via: "cli"` 供界面提示。
- 已补充 `DockerService.CloseAll()`，应用退出时关闭全部 Docker 会话。
- 未在真实 Docker / 远程 SSH 环境做端到端手工验证。
