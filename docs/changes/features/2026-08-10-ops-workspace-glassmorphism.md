# 变更：Glassmorphism 视觉落地

## 背景

双模式工作区信息架构已定，需按 Aether-SSH 式玻璃拟态设计稿统一沉浸背景、毛玻璃面板、弹层遮罩与微交互，提升观感与操作流畅度，且不改动连接/会话业务逻辑。

## 范围

- 全局 glass token、沉浸渐变壳、顶栏/侧栏/主面板/工具坞毛玻璃。
- 模式切换 pill、会话标签栏、Dialog/Confirm、上传飞出层、右键菜单玻璃化。
- 终端视口仍保持深色 `ops-terminal-surface`。
- 不改 Go/Wails 后端。

## 修改文件

- `frontend/src/styles/app.css`
- `frontend/src/App.svelte`
- `frontend/src/components/WorkspaceNavigation.svelte`
- `frontend/src/components/SessionToolDock.svelte`
- `frontend/src/components/DatabaseWorkspaceEmpty.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/ui/Dialog.svelte`
- `frontend/src/components/ui/ConfirmDialog.svelte`
- `frontend/src/components/ui/Button.svelte`
- `docs/development/2026-08-10-ops-workspace-glassmorphism.md`
- `docs/changes/features/2026-08-10-ops-workspace-glassmorphism.md`（本文件）

## 验证

- `node --test frontend/src/lib/workspaceTabs.test.js frontend/src/lib/formatConnectionError.test.js`
- `cd frontend && npx vite build`
- 建议手工：`wails dev` 切换明暗主题与 SSH/数据库模式，打开连接对话框与工具坞，确认毛玻璃与过渡正常。

## 剩余风险

- 低端 GPU / 部分 WebView 对 `backdrop-filter` 支持较弱时，会回退为半透明面板，层次略弱。
- FileManager 等深层面板仍有部分硬编码灰阶类名，后续可继续统一到 glass token。
- 若背景色块不够鲜艳或面板 alpha 过高，半透明会看起来像「实心白」——玻璃效果依赖「可折射的氛围层 + 足够低的不透明度」，不依赖贴图。
