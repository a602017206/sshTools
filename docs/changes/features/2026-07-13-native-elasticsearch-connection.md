# Elasticsearch 与 OpenSearch 原生连接

## 背景

Elasticsearch 和 OpenSearch 通过 HTTP API 提供索引服务，不需要 JDBC 或 Java agent。

## 范围

新增兼容 provider：连接健康检查、索引列表、空的二级资源和关闭。使用官方 Elasticsearch Go Client v8，以保持项目现有 Go 1.24 构建基线；OpenSearch 使用兼容 HTTP API。

## 修改文件

- `internal/service/native_elasticsearch.go`
- `internal/service/native_elasticsearch_test.go`
- `go.mod`
- `go.sum`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestElasticsearchNative -v`，验证健康检查、索引浏览、关闭和错误传播。

## 剩余风险

首版默认使用 HTTP，尚未在连接表单暴露 HTTPS、API Key 或自定义 CA；OpenSearch 的非兼容扩展端点不在首版范围。官方 v9 客户端要求 Go 1.25，因此本次不升级项目的跨平台 Go 基线。
