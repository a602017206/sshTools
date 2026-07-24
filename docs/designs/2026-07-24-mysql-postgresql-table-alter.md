# MySQL 与 PostgreSQL 表结构修改设计

## 背景

表设计器此前仅能在新建表模式编辑字段；打开已有表时所有字段控件为只读，无法完成实际的结构维护。

## 设计

- 读取表结构时保留一份原始字段快照，界面编辑的字段草稿与快照进行差异比较。
- 字段新增、删除、重命名、类型/长度、非空、默认值、注释及主键变化都生成预览 SQL。
- 保存时按顺序逐条调用数据库执行接口，避免依赖 JDBC 驱动的多语句执行策略；成功后重新读取结构。
- MySQL 使用 `MODIFY COLUMN`、`CHANGE COLUMN`、`ADD/DROP COLUMN` 与主键操作。
- PostgreSQL 使用 `RENAME COLUMN`、`ALTER COLUMN`、`ADD/DROP COLUMN`、`COMMENT ON COLUMN` 与主键操作。

## 取舍

表名改名、索引、外键、触发器及数据库特有的列属性不在本次范围内。PostgreSQL 主键约束按常规 `<table>_pkey` 名称处理；非标准约束名需要后续从元数据中读取。
