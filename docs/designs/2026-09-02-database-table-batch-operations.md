# 数据库表数据批量操作与筛选保持

## 背景

JDBC 表数据面板（MySQL、PostgreSQL、Oracle、SQL Server、达梦等）已有筛选/排序构建器，但删除或保存后会调用无条件的默认查询，导致筛选丢失。同时缺少 Navicat 风格的行多选、批量删除与批量改字段。

## 方案

- 统一 `buildBrowseQuery` / `refreshTableData`：分页、重载、删改后都按当前 `filterRules` 与 `sortRules` 重新查询
- 网格首列增加复选框与全选；工具栏显示批量删除、批量改字段
- 批量 SQL 复用 `tableDataMutations.js` 的 DELETE/UPDATE 生成，逐条执行（与现有单行删改一致）

## 取舍

- 批量更新为同一字段写入同一新值，不支持每行不同值
- 仍按页读取，全选仅作用于当前页可见行

## 修改文件

- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/lib/tableDataMutations.js`
- `frontend/src/lib/tableRowSelection.js`（新建）
- `frontend/test/tableDataMutations.test.js`
- `frontend/test/tableRowSelection.test.js`（新建）
