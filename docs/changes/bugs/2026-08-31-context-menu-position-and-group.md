# 修复右键菜单位置、选中文本与分组预填

## 背景

资产树文件夹右键「新建连接」后分组字段仍为空。所有右键菜单都偏离点击位置，Mac 触控板双指点按还会把文件夹名选成蓝色。

## 范围

- 右键菜单挂到 `document.body`，用视口坐标，避开毛玻璃面板的 `backdrop-filter` 把 `position: fixed` 困在侧栏里
- 树、标签、文件行禁止文本选择，右键时清掉选区
- 新建弹窗按本次请求再次写入 `preferredGroup`，避免重建实例时被空表单盖掉

不改分组存储模型。

## 修改文件

- `frontend/src/lib/contextMenu.js`（新建）
- `frontend/test/contextMenu.test.js`（新建）
- `frontend/src/lib/assetGroupTree.js`（无逻辑变更，单测补充）
- `frontend/test/assetGroupTree.test.js`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/FileManager.svelte`
- `frontend/src/components/FileManagerContextMenu.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/DatabaseWorkspaceEmpty.svelte`
- `frontend/src/components/OpsUnavailableWorkspace.svelte`
- `frontend/src/components/OpsDashboard.svelte`
- `frontend/src/styles/app.css`
- `docs/designs/2026-08-31-context-menu-position-and-group.md`
- `docs/changes/bugs/2026-08-31-context-menu-position-and-group.md`（本文）

## 验证

- `cd frontend && node --test test/contextMenu.test.js test/assetGroupTree.test.js test/fileManagerContextMenu.test.js`
- 手工：在 Mac 触控板上右键资产树文件夹，菜单应贴近指针，「新建连接」分组为该文件夹路径；标签栏和文件管理右键同样贴合，且不出现蓝色选中

## 剩余风险

- 本次无法在 Wails 窗口内自动点选验证
- 极少数仍带 `transform` 的祖先若被误放到 `body` 上，固定定位仍会偏移
