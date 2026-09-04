# 变更：搜索/缓存与关系库连接路径剥离

## 背景

ES、Redis 等在资产树被当成关系库展开，出现「数据库 / 加载数据库失败」，无法正常使用专用工作区体验。

## 范围

- 资产树：原生类型不再显示 JDBC 展开与 `DatabaseSidebarTree`
- 连接：原生类型展开手势改为打开专用工作区；密码提示按域文案
- 会话标签 / 顶栏「数据」/ 空态 / AI 助手标题按缓存、搜索、消息分流
- `opensearch` 纳入原生类型判定

## 修改文件

- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/lib/nativeDatabaseWorkspace.js`
- `frontend/src/lib/copilotContext.js`
- `frontend/src/lib/workspaceTabs.js`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/DatabaseSidebarTree.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/AIPanel.svelte`
- `frontend/src/components/DatabaseWorkspaceEmpty.svelte`
- `frontend/src/App.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- `frontend/test/nativeDatabaseWorkspace.test.js`
- `frontend/test/copilotContext.test.js`
- `docs/designs/2026-09-02-native-cache-search-decouple.md`
- `docs/development/2026-09-02-native-cache-search-decouple.md`
- `.github/workflows/release.yml`

## 验证

- `cd frontend && node --test test/nativeDatabaseTypes.test.js test/nativeDatabaseWorkspace.test.js test/copilotContext.test.js test/assetDomain.test.js` 通过
- 手工：ES/Redis 资产无展开箭头；双击打开原生工作区；MySQL 仍可展开库树

## 剩余风险

- 顶栏仍共用 `activeMode=database`，会话列表会混排关系库与原生会话，靠标签后缀区分
- 未跑完整前端构建与端到端连接（依赖本机实例）
