# 数据库表右键操作

## 背景

数据库对象页面缺少右键快捷菜单，无法快速对指定表执行打开、设计、复制或删除操作。

## 范围

- 为表行添加右键菜单：打开表、设计表、新建表、复制表和删除表。
- 支持 MySQL、PostgreSQL 仅结构复制与结构及数据复制。
- 删除表提供二次确认，并在操作成功后刷新列表。

## 修改文件

- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/lib/tableObjectMutations.js`
- `frontend/test/tableObjectMutations.test.js`
- `docs/designs/2026-07-24-database-table-context-actions.md`

## 验证

- 执行 `node --test test/tableObjectMutations.test.js`。
- 执行 `npm run build`。

## 剩余风险

复制“结构及数据”由两条语句组成。若第二条插入失败，新建的空表会保留，便于用户检查或手动删除。
