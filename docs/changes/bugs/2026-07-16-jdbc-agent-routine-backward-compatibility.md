# JDBC agent 例程接口向后兼容

## 背景

运行旧版 JDBC agent 时，新前端调用 `ListRoutines` 会得到 gRPC `Unimplemented`。错误此前被映射为数据库连接失败，导致存储过程和函数区域出现误导性报错。

## 范围

对 `ListRoutines` 的 `Unimplemented` 响应降级为空列表。新版 agent 仍正常返回例程；旧 agent 下表和视图不受影响，过程和函数显示为空。

## 修改文件

- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_gateway_test.go`
- `frontend/build/assets/index.js`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run 'TestJdbcGatewayListRoutines(TreatsOldAgentAsEmpty|PassesFunctionKindToAgent)' -v`，新版与旧版 agent 路径均通过；执行 `npm run build`，前端构建和 JDBC agent 暂存通过。

## 剩余风险

旧 agent 无法提供例程数据，只能显示为空。关闭并重新打开应用后，启动流程会安装嵌入的新 agent jar 并启动新版 agent。
