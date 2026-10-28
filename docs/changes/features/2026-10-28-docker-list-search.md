# Docker 工作区四 Tab 列表搜索

## 背景

Docker 工作区已支持容器、镜像、网络、卷四类资源列表及容器「仅运行中」过滤。资源增多后缺少按文本快速定位能力。规格见 `docs/superpowers/specs/2026-10-28-docker-list-search-design.md`；实现计划见 `docs/plans/2026-10-28-docker-list-search.md`；计划入库见 `docs/changes/process/2026-10-28-docker-list-search-plan.md`。

## 范围

- 四个 Tab 各自独立搜索词；切换 Tab 保留，互不覆盖。
- 不区分大小写的子串匹配；关键词 trim 后为空则不过滤。
- 匹配当前 Tab 表格可见列的展示文本（含 `formatContainerCreated`、`formatDockerSize`、`imageDisplayTags` 等格式化结果）。
- 容器：先 `visibleContainers(onlyRunning)`，再 `filterContainersByQuery`（AND）。
- 纯前端过滤；不持久化搜索词；不改 Go / Wails / Docker API。
- 全量非空但过滤后为空时显示「没有匹配的…」空态；全量为空沿用原「没有容器/镜像/…」文案；刷新不重置搜索词。

## 修改文件

- `frontend/src/lib/dockerWorkspace.js`：`containerSearchText`、`filterContainersByQuery`
- `frontend/test/dockerWorkspace.test.js`：容器搜索与「仅运行中 + 搜索」组合单测
- `frontend/src/lib/dockerResources.js`：`resourceSearchText`、`filterResourceRows`（按 kind 组装可见列 haystack）
- `frontend/test/dockerResources.test.js`：四类资源搜索单测
- `frontend/src/components/DockerWorkspace.svelte`：`searchQueries`、工具栏 `<input type="search">`、过滤后列表与空态

实现提交（按任务拆分）：

- `713d1b6` — 容器过滤纯函数与单测
- `fc5f737` — 镜像/网络/卷过滤纯函数与单测
- `084a59f` — DockerWorkspace UI 搜索框与 per-tab 状态

## 验证

```bash
cd frontend && node --test test/dockerWorkspace.test.js test/dockerResources.test.js
```

- 结果：15 项全部通过（含 `filterContainersByQuery`、`filterResourceRows` 及既有 docker 工作区/资源用例）。
- 未在本轮连接真实 Docker 引擎做手工点击验收；建议按规格手工核对：四 Tab 独立词、大小写等价、容器「仅运行中」+ 搜索、两种空态、刷新保留搜索词。

## 剩余风险

- 选中行被搜索隐藏时侧栏可仍打开、列表无高亮（设计如此）；用户需改词或清空搜索后再点选。
- 大列表全量在前端过滤，极端数量时或略增渲染开销（未做虚拟滚动变更）。
- 搜索词不写入 `config.json`，重启应用后丢失。
- 不支持分词 AND、正则、命中高亮与快捷键聚焦（规格非目标，可后续迭代）。
