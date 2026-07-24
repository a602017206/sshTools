# 表设计器 DDL 预览响应式修复

## 背景

新建表设计器在填充默认字段后，DDL 预览保持为空，导致“保存新表”和“复制 DDL”按钮被禁用。

## 范围

- 将 DDL 预览的响应式计算改为显式依赖数据库方言、库名、Schema、草稿表名和字段草稿。
- 增加回归断言，防止重新通过函数间接读取响应式输入。

## 修改文件

- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/test/tableStructureDesigner.test.js`

## 验证

- 执行 `node --test test/tableStructureDesigner.test.js`。
- 执行 `npm run build`。

## 剩余风险

该修复解决前端预览和按钮禁用问题。实际建表仍依赖已连接的 MySQL 或 PostgreSQL 数据库具备 `CREATE TABLE` 权限。
