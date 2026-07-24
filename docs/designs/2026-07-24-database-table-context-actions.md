# 数据库表右键操作设计

## 背景

数据库对象页只能通过工具栏打开或设计已选表，缺少针对单个表的快捷操作入口。

## 设计

- 在表行右键菜单提供打开表、设计表、新建表、复制表和删除表。
- 复制表分为“仅结构”和“结构及数据”，先实现 MySQL、PostgreSQL。
- 删除表使用应用内确认对话框，避免桌面容器原生确认框兼容性问题。
- SQL 标识符由共享生成器按方言引用，避免将表名直接拼入 SQL。

## 方言

- MySQL：`CREATE TABLE ... LIKE ...`，可选 `INSERT INTO ... SELECT * FROM ...`。
- PostgreSQL：`CREATE TABLE ... (LIKE ... INCLUDING ALL)`，可选 `INSERT INTO ... SELECT * FROM ...`。
