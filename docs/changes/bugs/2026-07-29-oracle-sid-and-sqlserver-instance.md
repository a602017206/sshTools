# Oracle 与 SQL Server 连接地址修复

## 背景

Oracle 连接只使用通用数据库名字段时，未指定目标名称会生成不完整的轻松连接字符串并触发 ORA-12261。SQL Server 命名实例也缺少表单输入和 JDBC 属性传递路径。

## 范围

新增 Oracle 服务名/SID 选择、SQL Server 实例名输入、连接元数据持久化和带连接属性的 Wails 调用。Go 网关按 Oracle 模式选择 URL 模板，并过滤内部模式标记；SQL Server 实例名按驱动标准传递为 `instanceName`。

## 修改文件

- `app.go`
- `internal/service/database_service.go`
- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_gateway_test.go`
- `frontend/src/App.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/lib/jdbcConnectionOptions.js`
- `frontend/test/jdbcConnectionOptions.test.js`
- `docs/designs/2026-07-29-enterprise-jdbc-connection-addresses.md`

## 验证

前端选项构造器测试覆盖 Oracle 服务名、SID 与 SQL Server 实例名。Go 网关测试覆盖 Oracle SID URL 选择和内部标记过滤。前端构建与 Wails 打包用于验证新增绑定和界面。

## 剩余风险

用户需要根据实际 Oracle 部署选择服务名或 SID；若填写错误，数据库服务端仍会拒绝连接。已有连接没有模式元数据时会默认按服务名连接。
