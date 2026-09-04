# Oracle 表结构加载与修改

## 背景

设计表页对 Oracle 会同时打出「加载表结构失败」和「当前仅支持 MySQL、PostgreSQL 和人大金仓」。截图中的 `ORA-17008: 已关闭连接` 表示 JDBC 会话已死，但网关只在 agent 进程不可用时重连。表设计器方言白名单也不含 Oracle。

## 目标与范围

- 识别 Oracle 关闭连接（含 `ORA-17008`），在同一 agent 上关闭并重开该 JDBC session，然后重试元数据请求。
- 表设计器把 Oracle 纳入可查看、可预览、可执行的 `CREATE`/`ALTER` 方言；限定名使用 `schema.table`，不把服务名当 catalog。
- 顺序加载字段元数据和 DDL，避免对同一死连接并发两次 `GetTableSchema`。

不包含达梦、SQL Server 表设计，也不改 JDBC agent 连接池模型。

## 架构与取舍

关闭连接与 agent 崩溃分开处理：前者只需 `CloseSession` + `OpenSession`，不必重启 JVM。Oracle `ALTER` 使用 `ADD (...)` / `MODIFY (...)` / `RENAME COLUMN`，与 PostgreSQL 的 `ALTER COLUMN` 不同。主键变更走 `DROP PRIMARY KEY`，避免依赖 `IF EXISTS`（低版本 Oracle 不支持）。

执行 Oracle `CREATE TABLE` 时去掉 `DBMS_METADATA` 风格的列级 `ENABLE` 和表级存储子句，避免 JDBC 回放时出现 `ORA-00922`。设计器把 `VARCHAR`/`BIGINT` 映射为 `VARCHAR2`/`NUMBER`。

## 风险

重连成功后仍可能因权限或对象锁导致 `ALTER` 失败。远端会话被 DBA 杀掉后，重连会新建会话，未提交事务不会恢复。
