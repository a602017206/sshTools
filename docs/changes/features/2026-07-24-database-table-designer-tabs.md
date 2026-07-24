# 表设计器独立标签页

## 背景

新建表和表结构设计显示在右侧面板，空间不足且操作方式不符合常见数据库客户端的工作区模式。

## 范围

- 将新建表和设计表改为主工作区的独立标签页。
- 支持同一已有表复用设计标签，以及多个新建表标签并存。
- 关闭表设计、数据和查询子标签时保持数据库连接不受影响。
- 移除右侧面板中的表设计器渲染入口。

## 修改文件

- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/App.svelte`
- `frontend/test/databaseHomeToolbar.test.js`
- `frontend/test/tableStructureDesigner.test.js`
- `docs/designs/2026-07-24-database-table-designer-tabs.md`

## 验证

- 执行 `node --test test/databaseHomeToolbar.test.js test/tableStructureDesigner.test.js`，通过。
- 执行前端 Vite 编译，确认 Svelte 组件可编译。

## 剩余风险

表设计器当前复用既有结构编辑组件；新建表完成后不会自动关闭标签或跳转到新表数据页，后续可按交互需要补充。
