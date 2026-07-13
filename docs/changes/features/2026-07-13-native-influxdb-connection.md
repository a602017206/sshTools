# InfluxDB 原生连接

## 背景

InfluxDB v2 提供稳定的 HTTP API，可读取健康状态和 bucket，不需要 JDBC 或 Java agent。

## 范围

新增 InfluxDB provider，使用 `/health` 测试连接，使用 `/api/v2/buckets` 只读浏览 bucket。密码字段作为 API Token，未填写 Token 时使用基础认证。应用与前端静态注册 InfluxDB，默认端口为 `8086`。

## 修改文件

- `internal/service/native_database.go`
- `internal/service/native_influxdb.go`
- `internal/service/native_influxdb_test.go`
- `app.go`
- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestInfluxDBNative -v`，验证连接、bucket 浏览、错误传播和关闭；执行前端类型测试验证原生路由。

## 剩余风险

首版未实现 Flux 查询以列出 measurement，也未暴露 HTTPS、自定义 CA 和组织筛选参数。InfluxDB v1 的 `SHOW DATABASES` 兼容接口不在本任务范围。
