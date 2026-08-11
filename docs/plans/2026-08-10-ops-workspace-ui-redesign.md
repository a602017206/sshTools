# 运维工作区 UI 分阶段重构 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将前端改为「SSH 会话 | 数据库」双模式亮色专业壳：文件/性能进入 SSH 会话工具坞；去掉概览平级页；不破坏现有 Wails 会话能力。

**Architecture:** `activeMode: 'ssh' | 'database'`；`sshToolTab: 'files' | 'performance'`。`TerminalPanel` 跨模式常驻。`FileManager` / `ServerMonitor` 仅在 SSH 模式右侧坞渲染一份。

**Tech Stack:** Svelte 4、Vite、Tailwind、Wails bindings、Node test runner。

**规格:** `docs/superpowers/specs/2026-08-10-ops-workspace-ui-redesign-design.md`

**视觉参考:** 亮色专业 SSH 工作台 / 数据库模式原型（会话内已出图）。

---

## 期次

| 期 | Tasks | 主题 |
|---|---|---|
| 1 | 1–5 | 模式模型、顶栏分段、去掉概览、TerminalPanel 常驻、亮色 token |
| 2 | 6–8 | 数据库模式空状态/错误/文案；隐藏 SSH 坞 |
| 3 | 9–11 | 会话工具坞绑定、空状态、标签状态、文档收尾 |

未获用户明确要求前跳过所有 `git commit` 步骤。

---

### Task 1: 双模式状态模块（TDD）

**Files:**
- Modify: `frontend/src/lib/workspaceTabs.js`
- Modify: `frontend/src/lib/workspaceTabs.test.js`

**Step 1: 重写测试为双模式**

```javascript
import assert from 'node:assert/strict';
import test from 'node:test';
import {
  APP_MODES,
  resolveMode,
  modeForAsset,
  resolveSshToolTab,
  SSH_TOOL_TABS
} from './workspaceTabs.js';

test('exposes ssh and database modes only', () => {
  assert.deepEqual(APP_MODES.map((m) => m.id), ['ssh', 'database']);
});

test('resolveMode falls back to ssh', () => {
  assert.equal(resolveMode('database'), 'database');
  assert.equal(resolveMode('dashboard'), 'ssh');
  assert.equal(resolveMode('files'), 'ssh');
});

test('modeForAsset routes by type', () => {
  assert.equal(modeForAsset({ type: 'database' }), 'database');
  assert.equal(modeForAsset({ type: 'ssh' }), 'ssh');
});

test('ssh tool tabs are files and performance', () => {
  assert.deepEqual(SSH_TOOL_TABS.map((t) => t.id), ['files', 'performance']);
  assert.equal(resolveSshToolTab('performance'), 'performance');
  assert.equal(resolveSshToolTab('nope'), 'files');
});
```

**Step 2:** `node --test frontend/src/lib/workspaceTabs.test.js` → 预期 FAIL

**Step 3: 实现**

```javascript
export const APP_MODES = [
  { id: 'ssh', label: 'SSH 会话' },
  { id: 'database', label: '数据库' }
];

export const SSH_TOOL_TABS = [
  { id: 'files', label: '文件' },
  { id: 'performance', label: '性能' }
];

const modes = new Set(APP_MODES.map((m) => m.id));
const sshTools = new Set(SSH_TOOL_TABS.map((t) => t.id));

/** 兼容旧工作区 id → 新模式 */
const legacyModeMap = {
  dashboard: 'ssh',
  terminal: 'ssh',
  files: 'ssh',
  performance: 'ssh',
  docker: 'ssh',
  logs: 'ssh',
  database: 'database'
};

export function resolveMode(mode) {
  if (modes.has(mode)) return mode;
  return legacyModeMap[mode] || 'ssh';
}

export function modeForAsset(asset) {
  return asset?.type === 'database' ? 'database' : 'ssh';
}

export function resolveSshToolTab(tab) {
  return sshTools.has(tab) ? tab : 'files';
}

// 可选：保留 resolveWorkspace 作 deprecate 包装 → resolveMode
export function resolveWorkspace(workspace) {
  return resolveMode(workspace);
}
```

**Step 4:** 测试 PASS

---

### Task 2: 顶栏改为分段控件；去掉概览入口

**Files:**
- Modify: `frontend/src/components/WorkspaceNavigation.svelte` → 可改名为模式切换，或原地改为渲染 `APP_MODES`
- Modify: `frontend/src/App.svelte` 顶栏与 `activeWorkspace` → `activeMode`

**Step 1:** `WorkspaceNavigation` 渲染两个分段按钮，选中态用 `var(--ops-accent)` 浅底 + 底线。

**Step 2:** `App.svelte`：

```javascript
let activeMode = 'ssh';
let sshToolTab = 'files';

function selectMode(mode) {
  activeMode = resolveMode(mode);
}
```

删除默认渲染 `OpsDashboard` 的分支（组件文件可暂留，不再挂到主路径）。

**Step 3:** `npm run build` PASS

---

### Task 3: 单一会话壳 + TerminalPanel 常驻

**Files:**
- Modify: `frontend/src/App.svelte`（751+ 主内容）

**Step 1:** 布局始终为：左资源树 + 中 TerminalPanel +（仅 `activeMode === 'ssh'` 时）右工具坞。

**Step 2:** 删除 `files` / `performance` / `database` 各自再挂一份 FileManager/ServerMonitor/TerminalPanel 的分支。

**Step 3:** `activeMode === 'database'` 时右侧坞 `display: none` 或不开；中区仍是同一个 TerminalPanel（数据库标签在内）。

**Step 4:** 手工：SSH 连上后切到数据库再切回，终端会话仍在。

---

### Task 4: 连接落点

**Files:**
- Modify: `frontend/src/App.svelte` `handleConnect` / `handleDatabaseConnect`

```javascript
import { modeForAsset } from './lib/workspaceTabs.js';

function handleConnect(asset) {
  activeMode = modeForAsset(asset);
  // database → handleDatabaseConnect; ssh → terminalPanelRef.handleConnect
}
```

数据库成功打开面板时 `activeMode = 'database'`。

---

### Task 5: 亮色专业 token + SSH 工具坞 UI

**Files:**
- Modify: `frontend/src/styles/app.css`
- Create: `frontend/src/components/SessionToolDock.svelte`
- Modify: `frontend/src/App.svelte`

**Step 1:** `:root` 对齐规格亮色表（Canvas/Surface/Accent `#0E7490` 等）；`.dark` 映射深色从属值。

**Step 2:** `SessionToolDock.svelte`：

- 标题「会话工具」+「绑定 · {sessionName}」
- 分段 文件 | 性能
- slot / 条件渲染 `FileManager` 或 `ServerMonitor`
- 无绑定 SSH 会话时空状态文案

**Step 3:** 坞绑定规则（写死）：取「最近一个 `type !== 'database' && connected` 的会话」，若无则空状态。不要绑到数据库 panel 会话。

**Step 4:** 构建 + 文档 `docs/development/2026-08-10-ops-workspace-ui-redesign-phase1.md`

---

### Task 6–8: 数据库模式（同前计划精神）

- Task 6: `DatabaseWorkspaceEmpty`（无 DB 会话）
- Task 7: `formatConnectionError` + 替换粗 `alert`
- Task 8: 文案统一 运行/保存/断开；phase2 开发记录

---

### Task 9–11: SSH 坞体验与收尾

- Task 9: 切 SSH 会话时坞标题与 FileManager/Monitor 跟随 `activeSessionId`（仅 SSH）
- Task 10: 会话标签文字状态；上传条用 accent token
- Task 11: phase3 文档；更新 `docs/changes/features/2026-08-10-ops-workspace-ui-redesign.md`；全量测试

---

## 硬约束

1. 禁止顶栏再加「概览 / 文件 / 性能」平级项。  
2. 禁止模式切换卸载 TerminalPanel。  
3. 禁止概览假遥测回归。  
4. 文档中文；提交需用户明确要求。
