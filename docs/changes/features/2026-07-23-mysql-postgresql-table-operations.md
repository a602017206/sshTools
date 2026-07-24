# MySQL 与 PostgreSQL 表操作支持

## 背景

用户需要优先完成 MySQL 和 PostgreSQL 的建表、删除记录和更新数据能力，其他数据库暂不进入可执行方言范围。

## 范围

- MySQL、PostgreSQL 新建表分别生成正确的对象限定名和标识符引用。
- PostgreSQL 不再错误地将数据库名拼接进表名。
- 更新、删除数据继续以主键为条件执行，并覆盖 PostgreSQL 双引号语法。
- 非 MySQL、PostgreSQL 的设计器明确提示暂不支持可执行建表。

## 修改文件

- `frontend/src/lib/tableDefinitionSQL.js`
- `frontend/src/lib/tableDataMutations.js`
- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/test/tableDefinitionSQL.test.js`
- `frontend/test/tableDataMutations.test.js`

## 验证

- 执行 `node --test test/tableDataMutations.test.js test/tableDefinitionSQL.test.js test/tableStructureDesigner.test.js test/databaseTableWorkspace.test.js test/databaseHomeToolbar.test.js`，通过。
- 执行前端 `npm run build`，Vite 编译通过；JDBC agent Gradle 暂存继续运行时未返回常规结束摘要。

## 剩余风险

当前不支持索引、外键、表注释、MySQL 存储引擎、PostgreSQL Identity/Serial 以及已有表的 `ALTER TABLE`。
