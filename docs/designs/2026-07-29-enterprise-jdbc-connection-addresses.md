# 企业 JDBC 连接地址设计

## 背景

Oracle 的网络地址除了主机和端口，还必须指定服务名或 SID。SQL Server 的命名实例也需要通过 JDBC `instanceName` 属性表达。将这些内容混入通用“数据库名”字段会生成不完整或错误的 JDBC URL。

## 方案

- Oracle 保留单一目标输入框，并通过“服务名 / SID”分段控件选择连接语义。
- 服务名使用 `jdbc:oracle:thin:@//{host}:{port}/{database}`；SID 使用 `jdbc:oracle:thin:@{host}:{port}:{database}`。
- 连接模式仅在 Go 网关中用于选择 URL 模板，不能透传给 JDBC 驱动。
- SQL Server 将可选实例名作为 JDBC `instanceName` 连接属性传递，仍使用现有 `databaseName` URL 模板。
- 连接配置持久化到 metadata；旧配置缺少这些元数据时，Oracle 默认按服务名处理，SQL Server 不设置实例名。
- 新增带 options 的 Wails 方法，旧方法继续保留，以便旧前端资源和已有调用兼容。

## 风险

该设计只覆盖 Oracle Thin 与 Microsoft SQL Server JDBC 驱动的标准网络格式。Oracle TNS 别名、完整 DESCRIPTION 描述符和 SQL Server 的自定义协议参数仍通过后续高级 URL 配置支持，不在本次表单范围内。
