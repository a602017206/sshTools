# Cassandra 原生连接

## 背景

Cassandra 通过 CQL 原生协议提供服务，不应通过 JDBC 或 Java agent 连接。其 Go Modules 兼容驱动使用实际模块路径 `github.com/gocql/gocql`；上游 ScyllaDB 的 fork 保留该模块声明路径，因此不能直接作为独立模块依赖。

## 范围

新增 Cassandra/ScyllaDB provider，支持连接测试、keyspace 列表、指定 keyspace 的表列表和会话关闭。应用启动时静态注册 Cassandra 和已实现的 Memcached provider，前端提供对应的连接类型和默认端口。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_cassandra.go`
- `internal/service/native_cassandra_test.go`
- `app.go`
- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/App.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- `go.mod`
- `go.sum`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestCassandraNative -v`，验证连接、keyspace 与表浏览、错误传播和关闭；执行前端类型测试，验证 Cassandra 与 Memcached 不会进入 JDBC 路由。

## 剩余风险

首版未暴露 TLS、证书、数据中心感知和多节点高级负载策略。服务端不可达或认证失败时，由 CQL 驱动返回原始连接错误。
