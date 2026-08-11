# Unified Ops Workspace Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将现有 SSH 工具重构为具有资源树、概览、终端、文件、数据库和性能工作区的统一运维桌面界面。

**Architecture:** 新增一个纯 JavaScript 工作区状态模块，使用 Svelte store 管理活动页。`OpsDashboard` 只消费现有的资产与会话 store；`App.svelte` 继续负责所有 Wails 生命周期和连接逻辑，仅替换主壳层的布局与面板编排。

**Tech Stack:** Svelte 4、Vite、Tailwind CSS、现有 Wails bindings。

---

### Task 1: 工作区状态

**Files:**
- Create: `frontend/src/lib/workspaceTabs.js`
- Create: `frontend/src/lib/workspaceTabs.test.js`

**Step 1: Write the failing test**

验证 `resolveWorkspace` 对有效工作区返回原值、对未知值返回 `dashboard`。

**Step 2: Run test to verify it fails**

Run: `node --test frontend/src/lib/workspaceTabs.test.js`
Expected: FAIL，因为模块尚不存在。

**Step 3: Write minimal implementation**

导出 `WORKSPACE_TABS` 与 `resolveWorkspace`，只允许概览、终端、文件、数据库、Docker、性能和日志七个值。

**Step 4: Run test to verify it passes**

Run: `node --test frontend/src/lib/workspaceTabs.test.js`
Expected: PASS。

### Task 2: 概览和资源导航

**Files:**
- Create: `frontend/src/components/OpsDashboard.svelte`
- Create: `frontend/src/components/WorkspaceNavigation.svelte`
- Modify: `frontend/src/App.svelte`

**Step 1:** 以现有资产和会话派生连接概览与空状态。

**Step 2:** 以原生按钮实现工作区标签、快捷操作和资源树，快捷操作进入对应工作区或打开新建连接对话框。

**Step 3:** 在 `App.svelte` 中按活动工作区复用 TerminalPanel、FileManager、ServerMonitor 和数据库入口。

**Step 4:** 执行 `npm run build`，确认 Svelte 编译成功。

### Task 3: 视觉系统和记录

**Files:**
- Modify: `frontend/src/styles/app.css`
- Create: `docs/development/2026-08-05-unified-ops-workspace.md`
- Create: `docs/changes/features/2026-08-05-unified-ops-workspace.md`

**Step 1:** 添加限定在 `ops-workspace` 的 token 和响应式布局，避免影响既有弹窗。

**Step 2:** 验证窄宽度、键盘焦点和减少动画规则。

**Step 3:** 运行 `npm run build` 与 `git diff --check`。
