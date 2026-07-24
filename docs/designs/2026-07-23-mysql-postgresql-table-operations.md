# MySQL 与 PostgreSQL 表操作设计

## 背景

数据库工作区需要先稳定支持 MySQL 和 PostgreSQL 的可执行建表、表数据更新和删除。两者的对象限定方式和字段类型语法不同，不能共用无差别 SQL。

## 决策

- 通过前端 `tableDefinitionSQL` 集中生成建表 SQL。
- MySQL 表名使用 `` `database`.`table` ``；PostgreSQL 表名使用 `"schema"."table"`，不拼接数据库名。
- 更新、删除 SQL 以主键为条件，字段与主键名均进行大小写无关匹配。
- 只有 MySQL、PostgreSQL 允许设计器直接创建表；其他数据库只保留后续扩展入口。

## 边界

- 支持字段、基础类型、长度、非空、主键和默认值。
- 不包含索引、外键、检查约束、表注释、存储引擎及既有表 `ALTER TABLE`。

## 验证策略

- 使用 Node 测试验证 MySQL 和 PostgreSQL 的 `CREATE TABLE`、`UPDATE`、`DELETE` 文本。
- 使用前端生产构建验证组件与绑定关系。
