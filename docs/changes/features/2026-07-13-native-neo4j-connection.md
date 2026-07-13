# Neo4j 原生连接

## 背景

Neo4j 使用 Bolt 协议，官方 Go driver v6 的最低 Go 版本为 1.24，符合项目跨平台构建基线，不需要 JDBC。

## 范围

新增 Neo4j provider，连接后使用 `VerifyConnectivity` 校验服务，并在 `system` 数据库执行只读 `SHOW DATABASES`。应用和前端静态注册 Neo4j，默认 Bolt 端口为 `7687`。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_neo4j.go`
- `internal/service/native_neo4j_test.go`
- `app.go`
- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- `go.mod`
- `go.sum`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestNeo4jNative -v`，验证连接、数据库浏览、错误传播和关闭；执行前端类型测试验证原生路由。

## 剩余风险

首版未暴露 `neo4j+s`、`neo4j+ssc`、Aura 证书和数据库内标签浏览选项。普通用户若没有查看系统数据库的权限，`SHOW DATABASES` 会返回服务端授权错误。
