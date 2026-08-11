# Oracle 表数据查询失败（LIMIT / 表名限定）

## 背景

打开 Oracle 表数据时，默认 SQL 形如：

```sql
SELECT * FROM "pdb"."DW_INTERFACE_TYPE" LIMIT 100;
```

界面报错「查询执行失败: 未知错误」。Oracle 不支持 `LIMIT`/`OFFSET`，且表名应使用 schema 限定，而不是连接服务名（如 `pdb`）。同时 Wails 可能以字符串抛出错误，`error.message` 为空时被前端吞成「未知错误」。

## 范围

- 表浏览 SQL 按数据库方言生成分页子句（Oracle/达梦：`FETCH FIRST` / `OFFSET … FETCH NEXT`；SQL Server：`OFFSET … FETCH`）
- Oracle / SQL Server / 达梦等使用 `schema.table` 限定，不再把服务名当 catalog 拼进 SQL
- 表数据面板与查询面板统一走方言 SQL 构建，并改进错误信息提取

## 修改文件

- `frontend/src/lib/tableQueryBuilder.js`
- `frontend/test/tableQueryBuilder.test.js`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `docs/changes/bugs/2026-08-10-oracle-table-query-limit.md`
- `docs/designs/2026-08-10-oracle-table-query-dialect.md`
- `docs/development/2026-08-10-oracle-table-query-dialect.md`

## 验证

- `cd frontend && node --test test/tableQueryBuilder.test.js src/lib/formatConnectionError.test.js`（全部通过）
- 手工：连接 Oracle → 打开表数据，默认 SQL 应为 `… FETCH FIRST 100 ROWS ONLY`，限定名为 `"schema"."table"`；失败时应显示后端原始错误而非「未知错误」

## 剩余风险

- Oracle 12c 之前不支持 `FETCH FIRST`；更老版本需改用 `ROWNUM` 包装（当前未覆盖）
- SQL Server 无排序时使用 `ORDER BY (SELECT NULL)`，仅用于满足 OFFSET/FETCH 语法要求
- 若侧栏未选中 schema（`schemaName` 为空），将只生成未限定表名，依赖当前登录用户默认 schema
