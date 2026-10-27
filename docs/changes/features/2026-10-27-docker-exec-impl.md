# Docker Exec 接入终端页签（前端接线）

## 背景

后端已提供 `StartDockerExec` / `SendDockerExecData` / `ResizeDockerExec` / `CloseDockerExec`，并通过 `docker:exec:{id}`、`docker:exec-closed:{id}` 事件推送输出与结束。前端已有 `canExec`、`buildDockerExecTabSession`、`sessionTabTitle`、`dockerExecCloseFlags` 等纯函数，但 `TerminalPanel` 与 `DockerWorkspace` 尚未接线，用户无法在界面里进入容器终端。

## 范围

- 容器行菜单「Exec」与详情栏头部「Exec」按钮，均触发 `onExec({ containerId, containerName })`。
- `TerminalPanel` 的 `openDockerExecSession`：调用 `StartDockerExec(父 Docker 会话, 容器 ID, 80, 24)`，成功后写入 `type: 'docker-exec'` 会话并激活页签；失败仅在面板顶部显示错误提示，不建页签。
- `docker-exec` 与 SSH 共用 `<Terminal>` 渲染分支；输入、resize、关闭分别走 `SendDockerExecData`、`ResizeDockerExec`、`CloseDockerExec`。
- 订阅 `docker:exec:{id}`（base64，与 `ssh:output` 一致）；订阅 `docker:exec-closed:{id}` 仅标记已断开并提示，保留页签。
- `viewportIsConsole` 改用 `sessionTabs.js` 的 `viewportIsConsole(session)`。
- 关闭父 Docker 页签不会级联关闭 exec 页签。
- 不在范围：Terminal.svelte、剪贴板受保护区域、ZMODEM、批量执行（exec 页签不参与）。

## 修改文件

- `frontend/src/components/DockerWorkspace.svelte`：新增 `onExec`，菜单 `exec` 与详情栏透传。
- `frontend/src/components/DockerContainerDetail.svelte`：新增 `onExec` 与头部 Exec 按钮（仅运行中可用）。
- `frontend/src/components/TerminalPanel.svelte`：`openDockerExecSession`、exec 输出/结束订阅、数据/resize/关闭分派、页签分组序号、命令提示与编码选择对 exec 关闭、错误提示条。

## 验证

- `node --test test/dockerContainerActions.test.js test/dockerExecSession.test.js test/sessionTabs.test.js test/sessionTabClose.test.js test/dockerWorkspace.test.js test/workspaceTabs.test.js`：27 项通过。
- 剪贴板契约：`node --test test/terminalClipboardContract.test.js test/terminalSelection.test.js test/terminalShortcuts.test.js`：23 项通过。
- `vite build`（输出到临时目录）编译通过。
- 未做真实 Docker 主机的手工联调。

## 剩余风险

- exec 会话 ID 由后端生成，前端只能在 `StartDockerExec` 返回后订阅；极端情况下首个提示符可能先于订阅到达而丢失。Terminal 未挂载时的输出已做缓存冲刷。
- 初始 PTY 尺寸为 80x24，挂载后会按实际终端尺寸再发一次 resize。
- 父 Docker 连接关闭后，后端会结束其 exec，页签保留为已断开状态，需手动关闭。
