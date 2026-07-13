# JDBC 驱动卸载引用保护

## 背景

JDBC 驱动管理器会直接删除所选 profile 的目录。已保存连接或正在运行的 JDBC 会话仍依赖该 profile 时，删除会导致后续连接或会话恢复失败。

## 范围

卸载前检查已保存数据库连接和活动 JDBC 会话；显式 profile 与旧连接的推荐 profile 都纳入检查。被引用时返回包含连接或会话标识的中文错误，不删除目录。

## 修改文件

- `app.go`
- `app_jdbc_test.go`
- `internal/service/jdbc_managed_gateway.go`
- 本变更记录。

## 验证

执行 `go test . -run TestRemoveJDBCDriver -v`，覆盖显式 profile、旧连接、活动会话和未引用 profile。

## 剩余风险

旧连接未保存 profile 时依赖当前推荐版本判定。用户修改推荐版本后，应重新保存连接以固定其实际使用的 profile。
