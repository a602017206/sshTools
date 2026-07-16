# PostgreSQL 兼容数据库对象树设计

## 背景

PostgreSQL、人大金仓和 openGauss 的数据库对象位于 Schema 命名空间内。仅按数据库展示表会丢失对象归属，且同名表可能读取到错误的字段定义。

## 目标

在不改变 MySQL 等现有浏览入口的前提下，为 PostgreSQL 兼容数据库提供接近 Navicat 的 `数据库 -> Schema -> 对象分类 -> 对象` 浏览层级，并让表结构读取使用用户选择的 Schema。

## 方案

前端对象树通过 PostgreSQL 系统目录按需读取 Schema 与对象名称。展开 Schema 后展示表、视图、物化视图、存储过程、函数和扩展分类；只有展开的分类会发起查询，避免一次性加载大量对象。

表节点派发包含 `schemaName` 的表结构事件。Wails 新增 `GetTableDDLInSchema`，服务层和 JDBC gateway 将该名称写入既有 gRPC `ListColumnsRequest.schema` 字段。协议字段已存在，因此不变更 `.proto`、不重新生成 gRPC 协议代码，也不改变旧 `GetTableDDL` 调用。

## 兼容性与边界

当前对象树适用于 PostgreSQL、人大金仓和 openGauss。不同厂商系统目录和对象种类差异较大，MySQL、Oracle、SQL Server、达梦等仍使用既有数据库浏览视图；后续如要统一，应由 JDBC agent 暴露结构化元数据接口，而不是在前端累计厂商 SQL。

## 风险

系统目录查询依赖当前用户的元数据权限，且部分兼容数据库可能不完整支持物化视图或扩展目录。查询失败时界面保留错误信息，不应影响已连接会话。
