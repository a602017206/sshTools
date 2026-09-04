# Oracle 建表 ORA-00922

## 背景

从 SQL 查询或对象 DDL 执行 Oracle `CREATE TABLE` 时出现 `ORA-00922: 选项缺失或无效`。常见来源是 `DBMS_METADATA` / Navicat 导出的 `NOT NULL ENABLE` 以及 `SEGMENT CREATION`、`PCTFREE` 等存储子句；JDBC 不能像 SQL*Plus 那样原样回放这些选项。设计器里的通用类型（`VARCHAR`、`BIGINT`）也不是 Oracle 建表的惯用写法。

## 范围

- 执行 Oracle SQL 前去掉列级 `ENABLE`/`DISABLE`、SQL*Plus `/`，以及表级物理存储子句
- 设计器生成 Oracle 建表语句时把 `VARCHAR` 映射为 `VARCHAR2`，整数/布尔映射为 `NUMBER`

不改查询浏览、不引入独立的 DDL 美化编辑器。

## 修改文件

- `internal/service/jdbc_gateway.go`
- `internal/service/jdbc_gateway_test.go`
- `internal/service/jdbc_managed_gateway.go`
- `frontend/src/lib/tableDefinitionSQL.js`
- `frontend/test/tableDefinitionSQL.test.js`
- `docs/changes/bugs/2026-08-31-oracle-create-table-ora-00922.md`（本文）

## 验证

- `go test ./internal/service -count=1 -run TestSanitizeOracleExecutableSQLStripsEnableAndStorageClauses`
- `cd frontend && node --test test/tableDefinitionSQL.test.js`
- 手工：把带 `NOT NULL ENABLE` 的 `CREATE TABLE` 在 SQL 查询中再跑一次；或用新建表保存默认字段

## 剩余风险

- 去掉 `TABLESPACE` 后表会建在用户默认表空间，与导出脚本可能不一致
- `USING INDEX TABLESPACE "..."` 若仍保留且表空间不存在，可能改报其他 Oracle 错误
