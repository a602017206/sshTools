# 数据行删除确认与结果反馈

## 背景

表数据页使用浏览器原生 `confirm` 确认删除。在桌面容器中该确认可能没有可见反馈；同时删除调用没有显示受影响行数，用户无法区分未执行、未匹配与删除成功。

## 范围

- 使用项目内 `ConfirmDialog` 代替原生确认框。
- 删除完成后读取 `ExecuteDatabaseQuery` 返回的 `affected` 并显示结果。
- 未匹配任何记录时明确提示数据没有变化。

## 修改文件

- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/test/databaseTableWorkspace.test.js`

## 验证

- 执行 `node --test test/databaseTableWorkspace.test.js test/tableDataMutations.test.js`，通过。
- 执行前端 `npm run build`，Vite 编译通过。

## 剩余风险

无主键表仍使用整行值匹配。存在完全重复行时，删除可能影响多条记录，确认框会保留相应风险提示。
