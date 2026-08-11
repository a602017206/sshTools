# Oracle / 多方言表浏览 SQL

## 问题

表数据默认查询与筛选器统一生成 `LIMIT/OFFSET`，对 Oracle、SQL Server 等方言不兼容。Oracle 场景还把连接服务名当成 catalog 拼进 `FROM`，导致非法对象引用。

## 决策

1. 在 `tableQueryBuilder` 集中处理分页方言：
   - MySQL / PostgreSQL 系：`LIMIT … OFFSET …`
   - Oracle / 达梦：`FETCH FIRST …` / `OFFSET … FETCH NEXT …`
   - SQL Server：`OFFSET … FETCH NEXT …`，无显式排序时补 `ORDER BY (SELECT NULL)`
2. 限定表名规则：`oracle` / `sqlserver` / `dm` / PostgreSQL 兼容库使用 `schema.table`；MySQL 继续使用 `database.table`。
3. 错误展示复用 `formatConnectionError`，兼容字符串与 `Error` 对象，避免 UI 只显示「未知错误」。

## 非目标

- 不为 Oracle 11g 及更早提供 `ROWNUM` 兼容层
- 不改 JDBC agent 执行路径或错误码体系
