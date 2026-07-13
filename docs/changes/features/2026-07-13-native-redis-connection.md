# Redis 与 KeyDB 原生连接

## 背景

Redis 和 KeyDB 使用 RESP 协议，不需要也不应使用 JDBC 驱动或 Java agent。

## 范围

新增 Redis 原生 provider：连接测试、连接生命周期、逻辑数据库发现和最多一千个键的只读扫描。KeyDB 复用该 provider。

## 修改文件

- `internal/service/native_redis.go`
- `internal/service/native_redis_test.go`
- `go.mod`
- `go.sum`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestRedisNative -v`，验证 Ping、逻辑数据库、键扫描、逻辑数据库校验和错误传播。

## 剩余风险

首版未实现 Redis Cluster、Sentinel 和键详情读取；大键空间采用受限 `SCAN`，不会保证一次列出全部键。
