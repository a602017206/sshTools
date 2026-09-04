# 数据库表数据批量操作与筛选保持

## 背景

用户反馈：表数据设置筛选/排序后，删除一条记录筛选就消失；无法批量选中删除或批量修改某一字段。需覆盖全部 JDBC 表数据面板，而非仅 Oracle。

## 范围

- 删除、保存、翻页、重载后保留已应用的筛选与排序
- Navicat 式单元格选中（已回退）
- 复选框多选当前页行；表头全选；选中后工具栏胶囊操作条
- 批量复制、复制为 INSERT、改字段、删除
- 右键保留：复制行、复制为 INSERT、删除记录；多选时额外显示批量复制/删除
- Oracle 删除含 NULL 字段改用 `IS NULL`，修复 ORA-00908

## 修改文件

- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/lib/tableDataMutations.js`
- `frontend/src/lib/tableRowSelection.js`
- `frontend/test/tableDataMutations.test.js`
- `frontend/test/tableRowSelection.test.js`
- `docs/designs/2026-09-02-database-table-batch-operations.md`
- `.github/workflows/release.yml`
- `docs/changes/features/2026-09-02-database-table-batch-operations.md`（本文）

## 验证

- `cd frontend && node --test test/tableDataMutations.test.js test/tableRowSelection.test.js`
- 手工：MySQL/Oracle 表应用筛选 → 删一行 → 条件仍在；多选 → 批量改字段 / 批量删除

## 剩余风险

- 无主键表批量操作仍可能误伤重复行
- 批量逐条 UPDATE/DELETE，大行数时较慢
