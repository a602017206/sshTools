# Docker 容器详情侧栏可拖拽调宽

## 背景

Docker 工作区二期的容器详情侧栏使用固定宽度（400px，`max-width: 45%`，`min-width: 320px`），无法按需调整，查看长日志或 Inspect 内容时空间不足。本次补上二期漏做的拖拽调宽，交互对齐 `SelectedDatabaseObjects` 详情栏的拖拽手柄方案。

## 范围

- 纯前端改动，不涉及 Go 后端与 Wails 绑定。
- 在 `DockerWorkspace` 中为详情侧栏增加左侧拖拽手柄（`role="separator"`，`aria-label="调整详情栏宽度"`）。
- 宽度由 `dockerWorkspace.js` 中新增的 `clampDockerDetailWidth` 约束：默认 420，最小 320，最大 720，且不超过主区宽度约 55%。
- `.docker-detail-shell` 增加常驻 `max-width: max(320px, 55%)`，窗口缩小时侧栏自动收回到主区 55% 内；开始拖拽时起始宽度先按当前主区宽度钳制，避免缩窗后拖拽出现死区。
- 宽度仅在组件会话内记忆（切换容器、关闭再打开详情保持拖后宽度），不落盘。
- `DockerContainerDetail` 去掉固定宽度，铺满外层；左边框改由外层 `docker-detail-shell` 承担，避免双线。
- 不修改日志、Inspect 等详情功能逻辑，仅调整布局。

## 修改文件

- `frontend/src/lib/dockerWorkspace.js`：新增 `DEFAULT_DOCKER_DETAIL_WIDTH` 与 `clampDockerDetailWidth`（默认 420、最小 320、最大 720、不超过主区约 55%）。
- `frontend/test/dockerWorkspace.test.js`：新增宽度钳制相关单元测试（默认值、最小/最大、主区 55% 约束）。
- `frontend/src/components/DockerWorkspace.svelte`：引入 `clampDockerDetailWidth`、`DEFAULT_DOCKER_DETAIL_WIDTH`；新增拖拽状态与 `startDetailResize` / `resizeDetail` / `stopDetailResize`；详情外包 `docker-detail-shell`（含 `max-width: max(320px, 55%)`）与手柄；开始拖拽时起始宽度先钳制；`docker-main` 绑定 `dockerMainEl`；组件销毁清理并入既有 `onMount` 返回的清理函数。
- `frontend/src/components/DockerContainerDetail.svelte`：`.docker-detail` 改为 `flex: 1 1 auto; width: 100%; min-width: 0`，去除左边框。
- `docs/changes/features/2026-10-27-docker-detail-sidebar-resize.md`：本文档。

## 验证

```bash
cd frontend && node --test test/dockerWorkspace.test.js test/dockerLogs.test.js test/dockerContainerActions.test.js
```

实际结果（2026-10-27）：共 19 项测试，19 通过，0 失败；其中 `dockerWorkspace.test.js` 5/5、`dockerLogs.test.js` 7/7、`dockerContainerActions.test.js` 7/7 通过。

拖拽交互属 UI 接线，需手工验收：拖动手柄宽度连续变化且松手保持；极限宽度符合 320–720 与主区约 55%；切换容器、关闭再打开详情宽度保持；日志 Tab 跟流可用。

## 剩余风险

- 手工 UI 验收尚未在真实窗口中执行。
- 使用 `pointer` 事件挂在 `window` 上，若指针捕获被系统中断（如 `pointercancel`）可能遗留 `pointermove` 监听，直至下次 `pointerup` 或组件销毁。
- 宽度不持久化，重新打开工作区后回到默认 420。
- 终端复制粘贴受保护区域未触及。
