# JDBC Agent 日志尾部读取

## 背景

JDBC 错误界面的“查看日志”此前只显示约定路径，用户不能在应用内读取诊断内容。直接暴露文件路径参数又会形成任意文件读取风险。

## 范围

- 增加只读取固定 `jdbc-agent.log` 的日志尾部服务。
- 默认读取 64 KiB，请求范围限制为 1 KiB 至 1 MiB。
- 使用文件定位从尾部读取，不先加载完整日志。
- 日志不存在时返回空结果；目录、符号链接和其他非普通文件被拒绝。
- 打开文件后再次核对文件身份，降低检查与读取间被替换的风险。
- 非法 UTF-8 字节使用替换字符显示。
- App API 只接收读取上限，不接收调用方路径。

## 修改文件

- `internal/service/jdbc_log_tail.go`
- `internal/service/jdbc_log_tail_test.go`
- `internal/service/jdbc_api_models.go`
- `app.go`
- `app_jdbc_test.go`

## 验证

- `go test ./internal/service -run TestJDBCLogTail -v`
- `go test . -run TestGetJDBCAgentLogTailUsesConfiguredPath -v`
- `go test ./internal/service -v`

## 剩余风险

- 当前按字节截断，首个 UTF-8 字符可能从中间开始并显示替换字符，这是尾部读取的预期降级。
- 本任务不提供日志清空、下载、搜索或长期归档。
