# Docker 列表搜索

## 背景

Docker 工作区已支持容器与镜像/网络/卷四类资源列表，以及容器「仅运行中」勾选过滤。容器与资源数量增多后，缺少按文本快速定位的能力，用户需滚动整表查找。

## 已定口径

1. **范围**：四个 Tab（容器 / 镜像 / 网络 / 卷）各自独立搜索词；切 Tab 保留，不互相覆盖。
2. **匹配**：不区分大小写的子串包含；关键词为空则不过滤。
3. **字段**：匹配当前 Tab **可见列**的展示文本（含格式化后的大小、创建时间等），不匹配详情侧栏专有字段。
4. **实现**：纯前端过滤；不改 Go / Wails / Docker API。
5. **组合**：容器先应用「仅运行中」，再应用搜索（AND）。

## 目标

- 在工具栏用搜索框即时缩小列表，便于按名称、镜像、状态、端口、ID 等可见信息定位。
- 空结果时能区分「引擎上没有资源」与「搜索无匹配」。

## 非目标

- 搜索词持久化到 `config.json`
- 后端 Docker filters / 远程服务端过滤
- 空格分词 AND、模糊匹配、正则
- 高亮命中片段、搜索历史、快捷键聚焦（可后续加）

---

## 交互

### 搜索框

- 位置：工具栏右侧操作区，位于「仅运行中」（仅容器 Tab）与「刷新」之间或紧邻过滤器；样式对齐现有 `docker-btn` / `docker-filter`。
- Placeholder 随 Tab 变化，例如「搜索容器…」「搜索镜像…」。
- 输入即时过滤（无需回车）；可一键清空（原生 clear 或空串即可）。
- 四个 Tab 各一份 `searchQueries[kind]`；切换 Tab 显示对应词。

### 过滤与空态

- 列表数据仍为刷新拿到的全量；搜索只改前端可见行。
- 全量非空但过滤后为空：空态文案提示「没有匹配的…」，并保留搜索框便于改词。
- 全量为空：沿用现有「没有容器/镜像/…」文案。
- 刷新不重置搜索词；切 Tab 不重置他 Tab 的词。

### 选中与详情

- 当前选中行若被过滤隐藏：侧栏可保持打开（不强制关闭），列表无高亮即可；再次点行再选。不做额外自动关侧栏。

---

## 匹配字段（可见列）

| Tab | 参与匹配的展示文本 |
| --- | --- |
| 容器 | Name、Image、Status、Ports、ID（短）、Created（`formatContainerCreated`） |
| 镜像 | 标签（`imageDisplayTags`）、ID（去 `sha256:` 后短 ID）、大小（`formatDockerSize`）、Created |
| 网络 | Name、ID（短）、Driver、Scope |
| 卷 | Name、Driver、Mountpoint、Created |

任一字段包含关键词即命中。比较前对关键词与字段值做 `trim` + `toLowerCase()`。

---

## 实现

### 纯函数

- `dockerWorkspace.js`：扩展或新增 `filterContainers(list, { onlyRunning, query })`（或在 `visibleContainers` 上增加 `query` 参数），单测覆盖空串、大小写、多字段命中、「仅运行中」+ 搜索组合。
- `dockerResources.js`：新增 `filterResourceRows(kind, rows, query)`，按 kind 组装可见列文本后子串匹配；单测覆盖四类资源与空串。

### UI

- `DockerWorkspace.svelte`：
  - 状态：`searchQueries = { containers: '', images: '', networks: '', volumes: '' }`
  - `$:` 当前 `query = searchQueries[resourceKind]`
  - `$:` 容器 `rows` / 资源 `filteredResourceRows` 经纯函数过滤后再 `#each`
  - 工具栏 `<input type="search">`，`bind` 到当前 Tab 的 query

### 非改动

- 不改列表拉取、Inspect、删除、Exec、侧栏拖拽。
- 不改终端复制粘贴相关文件。

---

## 验收

1. 四 Tab 各自输入不同关键词，来回切换词仍在。
2. `Nginx` 与 `nginx` 等价命中。
3. 容器勾选「仅运行中」后，搜索只在运行中集合内再滤。
4. 可见列任一字段命中即显示；详情侧栏独有字段不参与。
5. 有数据但无匹配时显示搜索空态；无数据时仍为原空态。
6. 纯函数单测通过。

---

## 测试计划

- 单元：`frontend` 下对过滤纯函数的 `node --test`（新建或扩展现有 docker 相关测试文件）。
- 手工：连接 Docker → 各 Tab 搜索、切换、与「仅运行中」组合、刷新后词保留。
