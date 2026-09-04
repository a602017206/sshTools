# Oracle 表结构 ORA-17027

## 背景

对象页选中 Oracle 表后，详情加载失败：`ORA-17027: 流已被关闭`。这不是会话断开，而是 Oracle JDBC 把 `getColumns` 的 `REMARKS`、`COLUMN_DEF` 当成 LONG 流：必须按 ResultSet 列序各读一次；先读默认值再读注释、或对 `COLUMN_DEF` 调用两次 `getObject`/`getString`，就会关掉流。

## 范围

- JDBC agent 按 JDBC 列序读取字段元数据，默认值只读一次
- `ORA-17027` 视为可恢复的失效会话，必要时重开会话再试

不改对象树列表、不改查询执行。

## 修改文件

- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/MetadataServiceImplTest.java`
- `internal/service/jdbc_errors.go`
- `internal/service/jdbc_errors_test.go`
- `docs/changes/bugs/2026-08-31-oracle-ora-17027-stream-closed.md`（本文）

## 验证

- `cd jdbc-agent && ./gradlew test --tests 'com.sshtools.jdbcagent.MetadataServiceImplTest'`
- `go test ./internal/service -count=1 -run TestJDBCClosedOracleConnectionIsStaleSession`
- 重新打包 agent 后，在 `pdb / PEMS` 下点选表，右侧详情应能出字段

## 剩余风险

- 已运行的桌面进程仍使用旧 agent jar，需重启应用才会加载新 jar
- 个别驱动若在 `getPrimaryKeys` 后弄脏连接，仍可能要靠会话重开
