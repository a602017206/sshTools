# 数据表工作区快捷操作

## 背景

数据库主页缺少常用的查询、结构设计、建表和刷新入口，用户需要在多个页面或手动修改 SQL 后才能完成这些操作。

## 范围

- 在数据库主页工具栏新增“新增查询”“设计表”“新建表”“刷新”操作。
- 选择表后，“设计表”打开该表的结构设计视图；双击表打开数据浏览页。
- “新增查询”和“新建表”都会新建独立 SQL 查询标签页，不依赖当前表；新建表仅预填通用 `CREATE TABLE` SQL 模板，不会自动执行。
- “刷新”重新加载当前对象类型的列表。

## 修改文件

- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/test/databaseTableWorkspace.test.js`
- `frontend/test/databaseHomeToolbar.test.js`

## 验证

- 执行 `node --test test/databaseTableWorkspace.test.js`，通过。
- 执行 `npm run build`，通过。

## 剩余风险

新建表 SQL 为通用模板，实际字段、存储引擎和数据库方言选项仍需用户按目标库调整后执行。
