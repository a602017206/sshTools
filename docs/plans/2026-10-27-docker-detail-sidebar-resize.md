# Docker 详情侧栏可拖 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 Docker 容器详情/日志侧栏可拖拽改变宽度，日志区可读，且不改动日志跟流与 Inspect 逻辑。

**Architecture:** 在 `dockerWorkspace.js` 增加纯函数 `clampDockerDetailWidth`；`DockerWorkspace.svelte` 持有宽度状态与 pointer 拖拽（对齐 `SelectedDatabaseObjects` 详情栏）；外层包裹详情并设像素宽度；`DockerContainerDetail` 去掉写死宽度，改由外层控制。

**Tech Stack:** Svelte、`node --test`

**规格:** `docs/superpowers/specs/2026-10-27-docker-phase3-exec-resources-design.md`（刀 1）

## Global Constraints

- 默认宽 420px；范围 320–720px；不超过主区约 55%（用 `docker-main` 客户端宽度作 viewport）
- 宽度仅组件内记忆，不写 `config.json`
- 不改日志跟流、菜单、后端；不改 Terminal 剪贴板受保护区
- 文档正文中文；变更写 `docs/changes/features/`
- 工作在当前 `master`，不强制新分支
- 本计划**仅刀 1**；exec / 资源管理另开计划

## 文件结构

| 文件 | 职责 |
|------|------|
| `frontend/src/lib/dockerWorkspace.js` | 新增 `clampDockerDetailWidth`、`DEFAULT_DOCKER_DETAIL_WIDTH` |
| `frontend/test/dockerWorkspace.test.js` | 宽度夹紧单测 |
| `frontend/src/components/DockerWorkspace.svelte` | 宽度状态、拖拽手柄、外层 style |
| `frontend/src/components/DockerContainerDetail.svelte` | 去掉固定 width/max-width |
| `docs/changes/features/2026-10-27-docker-detail-sidebar-resize.md` | 实现变更记录 |

---

### Task 1: clampDockerDetailWidth 纯函数

**Files:**
- Modify: `frontend/src/lib/dockerWorkspace.js`
- Modify: `frontend/test/dockerWorkspace.test.js`

**Interfaces:**
- Produces:
  - `export const DEFAULT_DOCKER_DETAIL_WIDTH = 420`
  - `export const MIN_DOCKER_DETAIL_WIDTH = 320`
  - `export const MAX_DOCKER_DETAIL_WIDTH = 720`
  - `export function clampDockerDetailWidth(width, mainWidth)`  
    - `width` / `mainWidth` 为 number；非法或非有限数时回退默认 420  
    - 先按 `min(720, max(320, width))` 夹硬限  
    - 若 `mainWidth` 为有限正数：再夹到 `Math.floor(mainWidth * 0.55)` 与 `MIN` 的较大者之上限（即最终 `<= max(MIN, floor(mainWidth*0.55))` 且 `>= MIN`，且仍 `<= MAX`）  
    - 若 `mainWidth` 缺失/无效：只做 320–720 硬限

- [ ] **Step 1: 在 `frontend/test/dockerWorkspace.test.js` 追加失败测试**

```javascript
import {
  clampDockerDetailWidth,
  DEFAULT_DOCKER_DETAIL_WIDTH,
  MAX_DOCKER_DETAIL_WIDTH,
  MIN_DOCKER_DETAIL_WIDTH
} from '../src/lib/dockerWorkspace.js';

test('clampDockerDetailWidth 默认与硬限', () => {
  assert.equal(DEFAULT_DOCKER_DETAIL_WIDTH, 420);
  assert.equal(MIN_DOCKER_DETAIL_WIDTH, 320);
  assert.equal(MAX_DOCKER_DETAIL_WIDTH, 720);
  assert.equal(clampDockerDetailWidth(Number.NaN, 1000), 420);
  assert.equal(clampDockerDetailWidth(200, 1000), 320);
  assert.equal(clampDockerDetailWidth(900, 1000), 720);
  assert.equal(clampDockerDetailWidth(500, 1000), 500);
});

test('clampDockerDetailWidth 受主区 55% 约束', () => {
  // 800 * 0.55 = 440 → 上限 440
  assert.equal(clampDockerDetailWidth(600, 800), 440);
  // 500 * 0.55 = 275 → 与 MIN 取大得 320，宽度 400 被夹到 320
  assert.equal(clampDockerDetailWidth(400, 500), 320);
  // mainWidth 无效时仅硬限
  assert.equal(clampDockerDetailWidth(600, 0), 600);
  assert.equal(clampDockerDetailWidth(600, Number.NaN), 600);
});
```

- [ ] **Step 2: 跑测确认失败**

Run: `cd frontend && node --test test/dockerWorkspace.test.js`  
Expected: FAIL（`clampDockerDetailWidth` 未导出）

- [ ] **Step 3: 在 `frontend/src/lib/dockerWorkspace.js` 末尾实现**

```javascript
export const DEFAULT_DOCKER_DETAIL_WIDTH = 420;
export const MIN_DOCKER_DETAIL_WIDTH = 320;
export const MAX_DOCKER_DETAIL_WIDTH = 720;

/** 将详情侧栏宽度夹到规格范围，并可选按主区宽度限制为约 55%。 */
export function clampDockerDetailWidth(width, mainWidth) {
  const raw = Number(width);
  const base = Number.isFinite(raw) ? raw : DEFAULT_DOCKER_DETAIL_WIDTH;
  let next = Math.min(MAX_DOCKER_DETAIL_WIDTH, Math.max(MIN_DOCKER_DETAIL_WIDTH, base));
  const main = Number(mainWidth);
  if (Number.isFinite(main) && main > 0) {
    const cap = Math.max(MIN_DOCKER_DETAIL_WIDTH, Math.floor(main * 0.55));
    next = Math.min(next, Math.min(MAX_DOCKER_DETAIL_WIDTH, cap));
    next = Math.max(MIN_DOCKER_DETAIL_WIDTH, next);
  }
  return next;
}
```

- [ ] **Step 4: 再跑测确认通过**

Run: `cd frontend && node --test test/dockerWorkspace.test.js`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/dockerWorkspace.js frontend/test/dockerWorkspace.test.js
git commit -m "$(cat <<'EOF'
feat(docker): add detail sidebar width clamp helper

EOF
)"
```

---

### Task 2: DockerWorkspace 拖拽手柄 + 详情去固定宽

**Files:**
- Modify: `frontend/src/components/DockerWorkspace.svelte`
- Modify: `frontend/src/components/DockerContainerDetail.svelte`（`.docker-detail` 样式）
- Create: `docs/changes/features/2026-10-27-docker-detail-sidebar-resize.md`
- Create: `docs/changes/process/2026-10-27-docker-detail-sidebar-resize-plan.md`（本计划落库说明，可与实现变更分开；若只写一份 features 实现记录亦可，但至少要有 features 实现记录）

**Interfaces:**
- Consumes: `clampDockerDetailWidth`、`DEFAULT_DOCKER_DETAIL_WIDTH`（Task 1）
- Produces: 侧栏可拖；会话内宽度记忆

- [ ] **Step 1: 改 `DockerContainerDetail.svelte` 样式**

将：

```css
.docker-detail {
  flex-shrink: 0;
  width: 400px;
  max-width: 45%;
  min-width: 320px;
  ...
}
```

改为：

```css
.docker-detail {
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border-left: none; /* 左边框改由外层/手柄承担，避免双线；若外层已有 border-left 则此处保持无左边框 */
  background: var(--glass-bg-strong);
  color: var(--text-primary);
  font-size: 12px;
}
```

说明：外层面板负责 `border-left` 与像素宽度；详情组件铺满外层。

- [ ] **Step 2: 在 `DockerWorkspace.svelte` script 中增加导入与拖拽状态**

在现有 `dockerWorkspace.js` 导入中增加：

```javascript
import {
  // ...existing
  clampDockerDetailWidth,
  DEFAULT_DOCKER_DETAIL_WIDTH
} from '../lib/dockerWorkspace.js';
```

在 `selectedDetail` 等状态旁增加：

```javascript
let detailWidth = DEFAULT_DOCKER_DETAIL_WIDTH;
let detailResizeStartX = 0;
let detailResizeStartWidth = 0;
let dockerMainEl = null;

function startDetailResize(event) {
  event.preventDefault();
  detailResizeStartX = event.clientX;
  detailResizeStartWidth = detailWidth;
  window.addEventListener('pointermove', resizeDetail);
  window.addEventListener('pointerup', stopDetailResize, { once: true });
}

function resizeDetail(event) {
  const mainWidth = dockerMainEl?.clientWidth || 0;
  const proposed = detailResizeStartWidth - (event.clientX - detailResizeStartX);
  detailWidth = clampDockerDetailWidth(proposed, mainWidth);
}

function stopDetailResize() {
  window.removeEventListener('pointermove', resizeDetail);
}

// 在既有 onDestroy 中调用 stopDetailResize；若尚无 onDestroy，增加：
import { onMount, onDestroy } from 'svelte'; // 若已有则合并
onDestroy(() => {
  stopDetailResize();
  // ...保留既有 destroy 逻辑
});
```

注意：文件顶部若已 `import { onMount, onDestroy } from 'svelte'`，勿重复导入；把 `stopDetailResize()` 并入现有 `onDestroy`。

- [ ] **Step 3: 改模板——详情外包一层带手柄**

将现有：

```svelte
{#if connected && selectedContainerId}
  {#key selectedContainerId}
    <DockerContainerDetail
      {sessionId}
      containerId={selectedContainerId}
      detail={selectedDetail}
      {detailError}
      onClose={closeDetail}
    />
  {/key}
{/if}
```

替换为：

```svelte
{#if connected && selectedContainerId}
  <div
    class="docker-detail-shell"
    style={`width: ${detailWidth}px;`}
  >
    <div
      class="docker-detail-resizer"
      role="separator"
      aria-orientation="vertical"
      aria-label="调整详情栏宽度"
      on:pointerdown={startDetailResize}
    ></div>
    {#key selectedContainerId}
      <DockerContainerDetail
        {sessionId}
        containerId={selectedContainerId}
        detail={selectedDetail}
        {detailError}
        onClose={closeDetail}
      />
    {/key}
  </div>
{/if}
```

并在 `<div class="docker-main">` 上增加 `bind:this={dockerMainEl}`：

```svelte
<div class="docker-main" bind:this={dockerMainEl}>
```

- [ ] **Step 4: 追加样式（`DockerWorkspace.svelte` `<style>`）**

```css
.docker-detail-shell {
  flex-shrink: 0;
  display: flex;
  flex-direction: row;
  min-width: 0;
  min-height: 0;
  border-left: 1px solid var(--glass-border);
  background: var(--glass-bg-strong);
}

.docker-detail-resizer {
  width: 6px;
  flex: 0 0 6px;
  cursor: col-resize;
  touch-action: none;
  background: transparent;
}

.docker-detail-resizer:hover,
.docker-detail-resizer:focus-visible {
  background: color-mix(in srgb, var(--ops-signal, #1687d4) 28%, transparent);
}
```

- [ ] **Step 5: 跑前端相关测试**

Run:

```bash
cd frontend && node --test test/dockerWorkspace.test.js test/dockerLogs.test.js test/dockerContainerActions.test.js
```

Expected: 全部 PASS

- [ ] **Step 6: 写变更文档 `docs/changes/features/2026-10-27-docker-detail-sidebar-resize.md`**

正文须含：背景、范围、修改文件、验证、剩余风险（中文）。要点：补二期漏做的可拖；纯前端；不落盘；验证命令同上。

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/DockerWorkspace.svelte \
  frontend/src/components/DockerContainerDetail.svelte \
  docs/changes/features/2026-10-27-docker-detail-sidebar-resize.md
git commit -m "$(cat <<'EOF'
feat(docker): make container detail sidebar resizable

EOF
)"
```

---

## 手工验收（实现后）

1. 打开 Docker 工作区 → 点开容器详情。
2. 拖动手柄：宽度连续变化；松手后保持。
3. 拉到极限：不小于约 320、不大于 720，窄窗口下不超过主区约一半。
4. 切换容器 / 关闭再打开详情：宽度仍为拖后值。
5. 日志 Tab 跟流仍可用。

## Spec 覆盖自检

| 规格要点 | 任务 |
|----------|------|
| 拖动手柄 + aria | Task 2 |
| 默认 420 / 320–720 / 55% | Task 1 + Task 2 |
| 会话内记忆、不落盘 | Task 2（组件状态） |
| 不改日志/Inspect | Task 2 仅布局 |
| 纯函数单测 | Task 1 |
