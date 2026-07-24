# 人大金仓表操作兼容设计

## 背景

人大金仓已接入 JDBC、对象树、Schema 限定表数据浏览和 JDBC 元数据，但建表、表结构修改及表右键操作仍被 MySQL/PostgreSQL 的前端方言白名单禁用。

## 设计

- 将 `kingbase` 归入 PostgreSQL 兼容 DDL 分支，使用双引号标识符与 `schema.table` 限定名。
- 复用 PostgreSQL 的 `CREATE TABLE`、`ALTER TABLE`、`COMMENT ON COLUMN`、`CREATE TABLE ... (LIKE ... INCLUDING ALL)` 和 `DROP TABLE` 语法。
- 表数据新增、更新和删除继续复用已有双引号与 `IS NOT DISTINCT FROM` 语义。

## 依据

KingbaseES 官方 `ALTER TABLE` 手册明确支持表定义变更；官方兼容性资料说明其提供 PostgreSQL 兼容模式。实际执行仍受实例兼容模式、版本及对象权限影响。
