# MySQL/PostgreSQL 表结构修改

## 背景

设计表标签只能查看已有表的字段和 DDL，字段控件被禁用，不能保存结构变更。

## 范围

- 允许在已有表设计页编辑字段、新增字段和删除字段。
- 生成并展示 MySQL、PostgreSQL 的 `ALTER TABLE` 预览。
- 支持字段重命名、类型/长度、空值、默认值、注释和主键变更。
- 保存时逐条执行变更并重新加载表结构。

## 修改文件

- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/src/lib/tableAlterSQL.js`
- `frontend/test/tableAlterSQL.test.js`
- `frontend/test/tableStructureDesigner.test.js`
- `docs/designs/2026-07-24-mysql-postgresql-table-alter.md`

## 验证

- 执行 `node --test test/tableAlterSQL.test.js test/tableStructureDesigner.test.js test/tableDefinitionSQL.test.js`。
- 执行 `npm run build`。

## 剩余风险

多条结构语句不具备跨数据库的原子性；其中一条执行失败时，已成功执行的前序变更不会自动回滚。PostgreSQL 非标准命名的主键约束需后续从元数据读取后再删除。
