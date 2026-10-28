# Docker 工作区：镜像/网络/卷浏览与删除（前端实现）

## 背景

第三阶段设计要求在 Docker 工作区中除容器外，还能浏览并删除镜像、网络、数据卷。后端 Wails 绑定与 `dockerResources.js` 辅助函数已就绪，本次补齐工作区 UI。

## 范围

- 工作区新增资源 Tab：容器 / 镜像 / 网络 / 卷（默认容器）。
- 切换 Tab 时清空选中详情，按类型调用 `ListDockerImages/Networks/Volumes`；加载与错误复用现有横幅。
- 镜像/网络/卷表格按规格列展示；点击行调用对应 `InspectDocker*`（失败时回退为列表行字段并提示原因），在可拖拽详情栏中以 `DockerResourceDetail` 展示概览键值对，仅含标题与关闭，无日志/Exec。
- 行末「删除」经 `ConfirmDialog` 确认；镜像提供「强制删除」勾选，文案来自 `removeResourceConfirmCopy`；成功后刷新列表并关闭侧栏。
- 「仅运行中」过滤仅在容器 Tab 显示；刷新按钮只刷新当前 Tab。
- 容器 Tab 的表格、右键菜单、Exec、`DockerContainerDetail`（含日志）保持不变，并复用既有 `detailWidth` 与拖拽逻辑。

## 修改文件

- `frontend/src/components/DockerResourceDetail.svelte`（新增）
- `frontend/src/components/DockerWorkspace.svelte`
- `docs/changes/features/2026-10-27-docker-resources-impl.md`（本文件）

## 验证

- `cd frontend && node --test test/dockerResources.test.js test/dockerWorkspace.test.js test/dockerContainerActions.test.js`：通过。
- 终端剪贴板契约测试（`terminalClipboardContract` / `terminalSelection` / `terminalShortcuts`）：通过。
- `npx vite build`：构建成功。
- 未做真机手工点击验证（需连接 Docker 引擎）。

## 剩余风险

- 网络/卷/镜像被引用时的删除失败信息依赖引擎返回，仅以横幅展示。
- 未提供右键菜单删除，仅行末按钮。
- 详情栏宽度仅会话内记忆。
