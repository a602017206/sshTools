# 统一运维工作区界面改造

## 背景

用户希望将 SSH 与数据库桌面工具复刻为更完整的现代运维工作台：资源树、概览、终端、文件和性能能力可在同一界面内切换。

## 范围

- 新增顶层工作区导航和可交互概览页。
- 文件、性能、数据库、Docker 与日志均有独立工作区呈现。
- 使用现有资产、SSH 会话和监控能力提供真实状态入口。
- 更新深色工作区的视觉 token 与响应式规则。
- 不新增 Docker、Redis 或云平台控制后端。

## 修改文件

- `frontend/src/App.svelte`
- `frontend/src/components/OpsDashboard.svelte`
- `frontend/src/components/WorkspaceNavigation.svelte`
- `frontend/src/components/OpsUnavailableWorkspace.svelte`
- `frontend/src/lib/workspaceTabs.js`
- `frontend/src/lib/workspaceTabs.test.js`
- `frontend/src/styles/app.css`
- `docs/designs/2026-08-05-unified-ops-workspace.md`
- `docs/plans/2026-08-05-unified-ops-workspace.md`
- `docs/development/2026-08-05-unified-ops-workspace.md`

## 验证

- 运行 `node --test src/lib/workspaceTabs.test.js`，3 项测试通过。
- 运行 `npm run build`，Vite 构建与 JDBC Agent 打包成功。

## 剩余风险

- 项目既有组件产生的 Svelte 无障碍告警仍存在，未在本次视觉重构中扩散修改范围。
- 工作区标签为新的壳层；Docker 等尚未接入后端的分类会展示为空值，后续接入时需要补充数据源和操作权限。
