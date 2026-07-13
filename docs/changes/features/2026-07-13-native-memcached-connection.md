# Memcached 原生连接

## 背景

Memcached 使用简单 TCP 文本协议，不需要 JDBC 或 Java agent。

## 范围

新增 Memcached provider，通过只读 `stats` 命令测试连接并浏览服务统计项；不支持键扫描或写命令。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_memcached.go`
- `internal/service/native_memcached_test.go`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestMemcachedNative -v`，验证统计读取、关闭和错误传播。

## 剩余风险

Memcached 协议不支持遍历全部键；首版不实现 SASL 认证或多节点分片。
