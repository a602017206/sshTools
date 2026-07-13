# MongoDB 原生连接

## 背景

MongoDB 使用原生二进制协议，不需要 JDBC 或 Java agent。

## 范围

新增 MongoDB 原生 provider：连接测试、会话连接、数据库和集合的只读浏览、客户端关闭。使用官方 MongoDB Go Driver v2。

## 修改文件

- `internal/service/native_mongodb.go`
- `internal/service/native_mongodb_test.go`
- `go.mod`
- `go.sum`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestMongoNative -v`，验证 Ping、数据库/集合浏览、关闭和错误传播。

## 剩余风险

首版不实现 SRV 连接串、副本集发现、Atlas Cloud 身份认证和文档查询；这些能力将在后续 provider 迭代中单独增加。
