# Oracle 末尾分号导致 ORA-00933 修复说明

## 原因

Oracle JDBC 不接受 SQL 末尾的 `;`。UI 生成的浏览 SQL 带分号，触发 `ORA-00933`。

## 实现

- `sanitizeJDBCSQL`：去掉尾部分号后发给 JDBC agent
- 前端 `executeQuery` 同步剥离，热更新即可生效
- `ORA-*` 映射为 `QUERY_FAILED`

## 验证

```bash
go test ./internal/service -run 'TestJdbcGatewayExecuteQueryStripsTrailingSemicolons|TestSanitizeJDBCSQL|TestJDBCErrorMapsOracleSyntaxToQueryFailed'
```
