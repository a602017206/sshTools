# 原生数据库工作区改造 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use `executing-plans` to implement this plan task-by-task.

**Goal:** 让 Redis、Elasticsearch 与其余原生数据库连接在浅色和深色主题下均可用，并把仅有的资源树升级为可理解、可浏览的工作区。

**Architecture:** 原生数据库继续共用 `NativeDatabasePanel`，但该面板拥有自身的内容背景，不再继承终端黑底。前端以纯函数定义各类型资源的标题、说明和交互能力；Redis 的逻辑库可展开显示键，ES 的索引与其它叶子资源以只读资源卡片展示，避免让无子资源的条目看起来可点击。

**Tech Stack:** Svelte 4、Vite、Node 内建测试、Go 原生数据库 provider。

---

### Task 1: 定义原生资源展示模型

**Files:**
- Create: `frontend/src/lib/nativeDatabaseWorkspace.js`
- Create: `frontend/test/nativeDatabaseWorkspace.test.js`

**Step 1:** 写一个失败测试，约束 Redis 的逻辑库可展开、Elasticsearch 索引为叶子资源以及未知类型的安全回退。

**Step 2:** 使用 `node --test test/nativeDatabaseWorkspace.test.js` 验证测试因模块缺失失败。

**Step 3:** 实现最小展示模型函数。

**Step 4:** 重新运行该测试确认通过。

### Task 2: 重构共享原生数据库面板

**Files:**
- Modify: `frontend/src/components/NativeDatabasePanel.svelte`

**Step 1:** 使用 Task 1 的展示模型决定资源可展开性、标题和空状态文案。

**Step 2:** 让面板显式使用 `--bg-primary`、`--border-primary`、`--bg-secondary`，从终端容器黑底中隔离；为资源、子资源、加载/错误/空状态提供可辨识布局。

**Step 3:** 对 Redis 显示“逻辑库—键”浏览提示；对 ES 显示“索引”只读提示；对无二级资源类型显示明确的只读资源说明。

### Task 3: 添加回归文档与验证

**Files:**
- Create: `docs/designs/2026-07-23-native-database-workspace.md`
- Create: `docs/development/2026-07-23-native-database-workspace.md`
- Create: `docs/changes/bugs/2026-07-23-native-database-panel-visibility.md`

**Step 1:** 记录问题、范围、方案和限制，明确本次不新增 Redis 写入或 ES 文档编辑。

**Step 2:** 运行 `node --test test/nativeDatabaseTypes.test.js test/nativeDatabaseWorkspace.test.js`。

**Step 3:** 运行 `npm run build`，确认 Svelte 编译与前端构建无误。
