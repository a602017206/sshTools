# 人大金仓表操作支持

## 背景

人大金仓连接已具备对象树、元数据和表数据浏览能力，但新建表、设计表保存、复制表和删除表被方言白名单禁用。

## 范围

- 人大金仓启用新建表与 DDL 预览。
- 人大金仓启用已有表结构修改与 `ALTER TABLE` 预览。
- 人大金仓启用复制表（仅结构、结构及数据）和删除表。
- 复用 PostgreSQL 的 Schema 限定与双引号 SQL 语法。

## 修改文件

- `frontend/src/lib/tableDefinitionSQL.js`
- `frontend/src/lib/tableAlterSQL.js`
- `frontend/src/lib/tableObjectMutations.js`
- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/test/tableDefinitionSQL.test.js`
- `frontend/test/tableAlterSQL.test.js`
- `frontend/test/tableObjectMutations.test.js`
- `docs/designs/2026-07-24-kingbase-table-operations.md`

## 验证

- 执行 `node --test test/tableDefinitionSQL.test.js test/tableAlterSQL.test.js test/tableObjectMutations.test.js test/tableStructureDesigner.test.js`。
- 执行 `npm run build`。

## 剩余风险

未连接真实 KingbaseES 实例。不同版本、不同兼容模式或非标准命名的主键约束可能使部分 DDL 执行失败，界面会显示数据库返回的错误信息。
