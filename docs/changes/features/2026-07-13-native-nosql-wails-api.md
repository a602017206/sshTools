# 原生 NoSQL Wails API

## 背景

原生 provider 需要独立的前端调用入口，避免 Redis、MongoDB、Elasticsearch/OpenSearch 流入 JDBC gateway。

## 范围

应用启动时注册已实现的原生 provider，并导出测试连接、连接、一级资源、二级资源和关闭会话 API。每次调用使用十秒超时。

## 修改文件

- `app.go`
- `app_native_database_test.go`
- 本变更记录。

## 验证

执行 `go test . -run TestNativeDatabaseAPIs -v`，验证 Redis 原生 API 的连接、资源浏览与关闭。

## 剩余风险

Wails 前端 bindings 和原生资源面板尚未更新；后续 provider 需要在内置注册表任务完成后加入启动注册。
