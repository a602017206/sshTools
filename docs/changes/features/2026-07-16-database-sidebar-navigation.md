# 数据库左侧导航

## 背景

用户需要在左侧看到完整的数据库对象层级，右侧只展示当前选中数据库的对象列表。

## 范围

左侧数据库资产展开后使用数据库、Schema、表、视图、系统表、存储过程和函数的层级树。右侧数据库主区改为选中数据库的对象分类列表，移除重复的主区导航树。

## 修改文件

- `frontend/src/stores.js`
- `frontend/src/components/DatabaseSidebarTree.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/build/assets/index.js`
- `docs/designs/2026-07-16-database-sidebar-navigation.md`
- 本变更记录。

## 验证

执行 `npm run build`，验证 Svelte 组件编译、前端打包和 JDBC agent 暂存均通过。

## 剩余风险

尚未连接真实 Oracle、SQL Server、达梦实例验证跨 catalog 的 Schema 枚举。现有代码库中存在与本次无关的 Svelte 可访问性告警，未在本次修改范围内处理。
