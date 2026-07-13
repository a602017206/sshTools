# 原生 NoSQL 会话模型

## 背景

Redis、MongoDB、Elasticsearch 需要绕过 JDBC 连接，但仍需统一管理连接生命周期与资源浏览。

## 范围

新增原生 NoSQL 服务、会话、资源和协议 provider 契约。该任务不包含具体厂商客户端实现，不会调用 JDBC agent。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_database_test.go`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestNativeDatabaseService -v`，验证会话连接、资源浏览、关闭、类型校验和连接测试委托。

## 剩余风险

具体协议适配器尚未注册；在后续 Redis、MongoDB、Elasticsearch 任务完成前，生产应用不能创建原生 NoSQL 会话。
