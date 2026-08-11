# Oracle ORA-00933 由末尾分号引起

## 背景

用户打开 Oracle 表数据时出现：

```text
ORA-00933: SQL 命令未正确结束
```

SQL 形如：

```sql
SELECT * FROM "DW_CP_CONTROL_TYPE" FETCH FIRST 100 ROWS ONLY;
```

这不是数据库慢。目标为 PDB（Oracle 12c+），`FETCH FIRST` 语法本身可用；Oracle JDBC 的 `Statement.execute` **不能携带末尾分号**，否则会报 ORA-00933。

## 范围

- JDBC 查询发送前剥离末尾 `;`
- ORA-/SQL 语法类错误映射为 `QUERY_FAILED`，避免误标为连接失败

## 修改文件

- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_gateway_test.go`
- `internal/service/database_service.go`
- `internal/service/jdbc_errors.go`
- `internal/service/jdbc_errors_test.go`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `docs/changes/bugs/2026-08-10-oracle-trailing-semicolon-ora-00933.md`
- `docs/development/2026-08-10-oracle-trailing-semicolon-ora-00933.md`

## 验证

- `go test ./internal/service -run 'TestJdbcGatewayExecuteQueryStripsTrailingSemicolons|TestSanitizeJDBCSQL|TestJDBCErrorMapsOracleSyntaxToQueryFailed'`
- 手工：Oracle 表数据默认查询应成功；编辑器中带 `;` 的 SQL 也能执行

## 剩余风险

- 多语句脚本（以分号分隔）仍不支持，会按单语句剥离末尾分号后执行
- Go 侧剥离需重启/重载后端后生效；前端剥离在热更新后即可覆盖主路径
