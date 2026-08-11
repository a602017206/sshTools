# 变更：运维工作区 UI 分阶段重构

## 背景

去掉概览；文件/性能与终端平级会造成割裂。改为 SSH / 数据库双模式，亮色专业壳，文件与性能进入 SSH 会话工具坞，并打磨数据库空状态与主操作文案。

## 范围

- 三期壳层与体验改造均已落地（双模式、SessionToolDock、数据库空状态、错误短句、运行/保存/断开、标签状态、上传条 token）。
- 未改 Go/Wails 后端连接 API。

## 修改文件

- `frontend/src/lib/workspaceTabs.js` / `workspaceTabs.test.js`
- `frontend/src/lib/formatConnectionError.js` / `formatConnectionError.test.js`
- `frontend/src/components/WorkspaceNavigation.svelte`
- `frontend/src/components/SessionToolDock.svelte`
- `frontend/src/components/DatabaseWorkspaceEmpty.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/App.svelte`
- `frontend/src/styles/app.css`
- `docs/superpowers/specs/2026-08-10-ops-workspace-ui-redesign-design.md`
- `docs/designs/2026-08-10-ops-workspace-ui-redesign.md`
- `docs/plans/2026-08-10-ops-workspace-ui-redesign.md`
- `docs/development/2026-08-10-ops-workspace-ui-redesign-phase{1,2,3}.md`
- `docs/changes/features/2026-08-10-ops-workspace-ui-redesign.md`（本文件）

## 验证

- `node --test frontend/src/lib/workspaceTabs.test.js frontend/src/lib/formatConnectionError.test.js` 通过
- `cd frontend && npx vite build` 成功
- 建议手工：`wails dev` 切换 SSH/数据库、连库看空状态消失、故意连错看对话框短句、开文件坞与上传进度色

## 剩余风险

- ~~数据库与 SSH 会话仍共用 TerminalPanel 标签栏~~：已按 `activeMode` 过滤标签；会话内容仍保活挂载，互切不丢。
- `OpsDashboard` 等旧组件仍在仓库但未挂主路径。
- 完整 `npm run build` 含 JDBC Gradle，可能因网络失败，与本期 UI 无关。
