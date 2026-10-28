# Docker 列表搜索 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Docker 工作区四个资源 Tab 上提供各自独立的前端即时搜索，按可见列不区分大小写子串过滤列表。

**Architecture:** 在 `dockerWorkspace.js` / `dockerResources.js` 增加纯过滤函数与单测；`DockerWorkspace.svelte` 用 per-tab `searchQueries` 驱动工具栏搜索框，列表与空态基于过滤结果渲染。不改后端。

**Tech Stack:** Svelte、`node --test`

**规格:** `docs/superpowers/specs/2026-10-28-docker-list-search-design.md`

## Global Constraints

- 四 Tab 搜索词独立；空串不过滤；大小写不敏感子串
- 容器：先「仅运行中」再搜索；匹配可见列展示文本（含格式化时间/大小）
- 不持久化搜索词；不改 Go / Wails；不改 Terminal 剪贴板受保护区
- 文档正文中文；变更写 `docs/changes/features/`
- 工作在当前 `master`

## 文件结构

| 文件 | 职责 |
|------|------|
| `frontend/src/lib/dockerWorkspace.js` | `filterContainersByQuery`、容器 haystack |
| `frontend/test/dockerWorkspace.test.js` | 容器搜索单测 |
| `frontend/src/lib/dockerResources.js` | `filterResourceRows`、资源 haystack |
| `frontend/test/dockerResources.test.js` | 资源搜索单测 |
| `frontend/src/components/DockerWorkspace.svelte` | 搜索框、per-tab 状态、空态 |
| `docs/changes/features/2026-10-28-docker-list-search.md` | 实现变更记录 |

---

### Task 1: 容器列表搜索纯函数

**Files:**
- Modify: `frontend/src/lib/dockerWorkspace.js`
- Modify: `frontend/test/dockerWorkspace.test.js`

**Interfaces:**
- `export function containerSearchText(container)` → 可见列拼接串（Name、Image、Status、Ports、ID、`formatContainerCreated(Created)`），字段用空格连接
- `export function filterContainersByQuery(list, query)` → `query` trim 后为空则返回原列表副本或原数组；否则保留 `containerSearchText` 含 `query.toLowerCase()` 子串的项
- `visibleContainers` **保持** `(containers, onlyRunning)` 签名不变；UI 层组合：`filterContainersByQuery(visibleContainers(...), query)`

- [ ] **Step 1: 追加失败测试**

在 `frontend/test/dockerWorkspace.test.js` 增加：

```javascript
import {
  // ...existing
  containerSearchText,
  filterContainersByQuery,
  formatContainerCreated,
} from '../src/lib/dockerWorkspace.js';

test('filterContainersByQuery 空串不过滤；大小写不敏感子串', () => {
  const list = [
    { ID: 'abc123def456', Name: 'web', Image: 'nginx:latest', Status: 'Up 1h', Ports: '80:80/tcp', Created: 1700000000 },
    { ID: 'fff111aaa222', Name: 'db', Image: 'postgres:15', Status: 'Exited (0)', Ports: '', Created: 0 },
  ];
  assert.equal(filterContainersByQuery(list, '').length, 2);
  assert.equal(filterContainersByQuery(list, '  ').length, 2);
  assert.deepEqual(filterContainersByQuery(list, 'NGINX').map((c) => c.Name), ['web']);
  assert.deepEqual(filterContainersByQuery(list, 'exited').map((c) => c.Name), ['db']);
  assert.deepEqual(filterContainersByQuery(list, '80:80').map((c) => c.Name), ['web']);
  assert.deepEqual(filterContainersByQuery(list, 'abc123').map((c) => c.Name), ['web']);
  // Created 展示串可命中（与 formatContainerCreated 一致）
  const created = formatContainerCreated(1700000000);
  assert.deepEqual(filterContainersByQuery(list, created.slice(0, 10)).map((c) => c.Name), ['web']);
  assert.equal(filterContainersByQuery(list, 'no-such').length, 0);
  assert.ok(containerSearchText(list[0]).toLowerCase().includes('nginx'));
});

test('visibleContainers + filterContainersByQuery 组合：仅运行中再搜索', () => {
  const list = [
    { ID: 'a', Name: 'web', Status: 'Up 1h', Image: 'nginx' },
    { ID: 'b', Name: 'web-old', Status: 'Exited (0)', Image: 'nginx' },
  ];
  const running = visibleContainers(list, true);
  assert.deepEqual(filterContainersByQuery(running, 'nginx').map((c) => c.ID), ['a']);
  assert.deepEqual(filterContainersByQuery(running, 'web-old').map((c) => c.ID), []);
});
```

- [ ] **Step 2: 跑测确认失败**

```bash
cd frontend && node --test test/dockerWorkspace.test.js
```

Expected: FAIL（`filterContainersByQuery` / `containerSearchText` 未导出）

- [ ] **Step 3: 最小实现**

在 `dockerWorkspace.js` 末尾附近增加：

```javascript
export function containerSearchText(container) {
  if (!container) return '';
  const parts = [
    containerField(container, 'Name'),
    containerField(container, 'Image'),
    containerField(container, 'Status'),
    containerField(container, 'Ports'),
    containerField(container, 'ID'),
    formatContainerCreated(containerField(container, 'Created')),
  ];
  return parts.map((p) => String(p ?? '')).join(' ');
}

export function filterContainersByQuery(list, query) {
  const q = String(query ?? '').trim().toLowerCase();
  const rows = list || [];
  if (!q) return rows;
  return rows.filter((c) => containerSearchText(c).toLowerCase().includes(q));
}
```

- [ ] **Step 4: 跑测确认通过**

```bash
cd frontend && node --test test/dockerWorkspace.test.js
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/dockerWorkspace.js frontend/test/dockerWorkspace.test.js
git commit -m "$(cat <<'EOF'
feat(docker): add container list search filter helpers

EOF
)"
```

---

### Task 2: 镜像/网络/卷搜索纯函数

**Files:**
- Modify: `frontend/src/lib/dockerResources.js`
- Modify: `frontend/test/dockerResources.test.js`

**Interfaces:**
- `export function resourceSearchText(kind, row)` → 按 kind 拼可见列文本
  - images: `imageDisplayTags(row)`、去 `sha256:` 后短 ID（与表格一致 `replace(/^sha256:/,'').slice(0,12)`）、`formatDockerSize(Size)`、`formatCreatedForSearch(Created)`（与列表一致：可用 `formatContainerCreated` 从 dockerWorkspace 导入，或在本文件用相同格式；**推荐**在 `dockerResources.js` 内对 Created 复用与 UI 相同的展示——为避免循环依赖，把时间格式化为与 `formatContainerCreated` 相同的 `YYYY-MM-DD HH:mm` 小函数，或从 `dockerWorkspace.js` import `formatContainerCreated`）
  - networks: Name、短 ID、Driver、Scope
  - volumes: Name、Driver、Mountpoint、Created 展示
  - 未知 kind：空串
- `export function filterResourceRows(kind, rows, query)` → 同容器：空串返回原列表；否则 haystack includes

**依赖注意：** `dockerResources.js` 已可被 `dockerWorkspace` 之外独立测试。优先在 `dockerResources.js` 内写本地 `formatSearchCreated(seconds)` 复制 `formatContainerCreated` 逻辑（几行），避免 `dockerResources` → `dockerWorkspace` 新依赖；或反过来把 `formatContainerCreated` 留在 workspace、resources 重复小函数。计划采用 **resources 内联相同格式函数**（DRY 次要，依赖清晰优先）。

- [ ] **Step 1: 追加失败测试**

```javascript
import {
  // existing...
  filterResourceRows,
  resourceSearchText,
} from '../src/lib/dockerResources.js';

test('filterResourceRows 各类型可见列子串匹配', () => {
  assert.equal(filterResourceRows('images', [{ ID: 'sha256:abcdef0123456789', Tags: ['nginx:latest'], Size: 1024, Created: 1700000000 }], '').length, 1);
  assert.deepEqual(
    filterResourceRows('images', [{ ID: 'sha256:abcdef0123456789', Tags: ['nginx:latest'], Size: 1024, Created: 0 }], 'NGINX').map((r) => r.Tags[0]),
    ['nginx:latest']
  );
  assert.equal(filterResourceRows('images', [{ ID: 'sha256:abcdef0123456789', Tags: [], Size: 2048, Created: 0 }], '2.0 KB').length, 1);
  assert.equal(filterResourceRows('networks', [{ ID: 'netid1234567', Name: 'bridge', Driver: 'bridge', Scope: 'local' }], 'LOCAL').length, 1);
  assert.equal(filterResourceRows('volumes', [{ Name: 'data', Driver: 'local', Mountpoint: '/var/lib/docker/volumes/data', Created: 0 }], 'mountpoint-miss').length, 0);
  assert.equal(filterResourceRows('volumes', [{ Name: 'data', Driver: 'local', Mountpoint: '/var/lib/docker/volumes/data', Created: 0 }], 'volumes/data').length, 1);
  assert.ok(resourceSearchText('networks', { ID: 'x', Name: 'bridge', Driver: 'bridge', Scope: 'local' }).includes('bridge'));
});
```

- [ ] **Step 2: 跑测失败**

```bash
cd frontend && node --test test/dockerResources.test.js
```

Expected: FAIL

- [ ] **Step 3: 最小实现**

在 `dockerResources.js`：

```javascript
function formatSearchCreated(seconds) {
  const value = Number(seconds);
  if (!value || value <= 0) return '-';
  const d = new Date(value * 1000);
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function shortId(id) {
  return String(id ?? '').replace(/^sha256:/, '').slice(0, 12);
}

export function resourceSearchText(kind, row) {
  if (!row) return '';
  switch (kind) {
    case 'images':
      return [imageDisplayTags(row), shortId(row.ID), formatDockerSize(row.Size), formatSearchCreated(row.Created)].join(' ');
    case 'networks':
      return [row.Name, shortId(row.ID), row.Driver, row.Scope].map((x) => String(x ?? '')).join(' ');
    case 'volumes':
      return [row.Name, row.Driver, row.Mountpoint, formatSearchCreated(row.Created)].map((x) => String(x ?? '')).join(' ');
    default:
      return '';
  }
}

export function filterResourceRows(kind, rows, query) {
  const q = String(query ?? '').trim().toLowerCase();
  const list = rows || [];
  if (!q) return list;
  return list.filter((row) => resourceSearchText(kind, row).toLowerCase().includes(q));
}
```

- [ ] **Step 4: 跑测通过**

```bash
cd frontend && node --test test/dockerResources.test.js
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/dockerResources.js frontend/test/dockerResources.test.js
git commit -m "$(cat <<'EOF'
feat(docker): add resource list search filter helpers

EOF
)"
```

---

### Task 3: DockerWorkspace UI 接入搜索

**Files:**
- Modify: `frontend/src/components/DockerWorkspace.svelte`

**要点：**

1. import `filterContainersByQuery`；import `filterResourceRows`
2. 状态：

```javascript
let searchQueries = { containers: '', images: '', networks: '', volumes: '' };
```

3. 响应式：

```javascript
$: searchQuery = searchQueries[resourceKind] ?? '';
$: rows = filterContainersByQuery(visibleContainers(containers, onlyRunning), searchQueries.containers ?? '');
$: visibleResourceRows = filterResourceRows(resourceKind, resourceRows, searchQueries[resourceKind] ?? '');
```

4. 工具栏在「仅运行中」与「刷新」之间加：

```svelte
<input
  type="search"
  class="docker-search"
  placeholder={resourceKind === 'containers' ? '搜索容器…' : resourceKind === 'images' ? '搜索镜像…' : resourceKind === 'networks' ? '搜索网络…' : '搜索卷…'}
  value={searchQueries[resourceKind] ?? ''}
  disabled={!connected}
  on:input={(e) => {
    searchQueries = { ...searchQueries, [resourceKind]: e.currentTarget.value };
  }}
/>
```

（或 `bind:value` 到中间变量再 `$:` 写回；注意 Svelte 对嵌套对象 bind 需整体赋值。）

5. 资源表 `#each` 改为 `visibleResourceRows`；空态分支：
   - `resourceRows.length === 0` → 原「没有 X」
   - `visibleResourceRows.length === 0` 且有 query → 「没有匹配的 X」/「试试其他关键词」
   - 否则显示表

6. 容器空态（现有 `rows.length === 0`）：
   - `containers.length === 0` → 没有容器
   - `searchQueries.containers.trim()` 非空 → 没有匹配的容器
   - `onlyRunning` → 没有运行中的容器（沿用）
   - 否则 → 没有容器（兜底）

7. 样式 `.docker-search`：高度约 30px、圆角 10px、边框 `var(--glass-border)`、字号 12px、宽度约 160–200px，与工具栏对齐。

8. **不要**在 `switchTab` 清空搜索词。

- [ ] **Step 1: 按上列改 UI**

- [ ] **Step 2: 跑相关单测**

```bash
cd frontend && node --test test/dockerWorkspace.test.js test/dockerResources.test.js
```

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/DockerWorkspace.svelte
git commit -m "$(cat <<'EOF'
feat(docker): wire per-tab list search into DockerWorkspace

EOF
)"
```

---

### Task 4: 变更文档与收尾

**Files:**
- Create: `docs/changes/features/2026-10-28-docker-list-search.md`
- Create: `docs/changes/process/2026-10-28-docker-list-search-plan.md`（可选短链到本计划）

- [ ] **Step 1: 写中文变更文档**（背景 / 范围 / 修改文件 / 验证 / 剩余风险）

- [ ] **Step 2: 全量相关测试再跑一遍**

```bash
cd frontend && node --test test/dockerWorkspace.test.js test/dockerResources.test.js
```

- [ ] **Step 3: Commit**

```bash
git add docs/changes/features/2026-10-28-docker-list-search.md docs/plans/2026-10-28-docker-list-search.md
git commit -m "$(cat <<'EOF'
docs: record Docker list search implementation

EOF
)"
```

---

## 手工验收清单

- [ ] 四 Tab 各输不同词，切换后词仍在
- [ ] `Nginx` / `nginx` 等价
- [ ] 容器「仅运行中」+ 搜索 AND
- [ ] 有数据无匹配 → 搜索空态；无数据 → 原空态
- [ ] 刷新不丢搜索词
