# Couchbase 原生连接

## 背景

Couchbase 的 bucket、scope 与 collection 通过官方 Go SDK 管理，不需要 JDBC 或 Java agent。当前 SDK `gocb/v2 v2.12.4` 声明兼容 Go 1.24，可保持项目既有的跨平台构建基线。

## 范围

新增 Couchbase provider，支持连接测试、bucket 列表、bucket 内 `scope.collection` 列表与会话关闭。应用启动时静态注册 provider，前端增加 Couchbase 类型及默认管理端口 `8091`。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_couchbase.go`
- `internal/service/native_couchbase_test.go`
- `app.go`
- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- `go.mod`
- `go.sum`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestCouchbaseNative -v`，验证连接、bucket 与 collection 浏览、错误传播和关闭；执行前端类型测试，验证 Couchbase 使用原生路由。

## 剩余风险

首版连接表单只支持用户名密码和明文 `couchbase://`，未暴露 TLS、Capella 证书、DNS SRV 或 `couchbase2://` 选项。SDK 引入了较新的 gRPC 传递依赖，完整测试将验证其与现有 JDBC agent 的兼容性。
