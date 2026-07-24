# 数据表删除记录主键大小写修复

## 背景

部分 JDBC 驱动返回的主键元数据列名与查询结果列名大小写不同，例如主键为 `ID` 而结果列为 `id`。删除语句构建器使用精确匹配时无法定位列索引，最终生成空 SQL，导致删除操作没有响应。

## 范围

- JDBC agent 读取列与主键元数据时，以不区分大小写的方式匹配主键列名。
- 删除和更新语句构建时，以不区分大小写的方式匹配主键元数据与结果列名。
- 未识别到主键时，MySQL 使用 `<=>`、PostgreSQL 使用 `IS NOT DISTINCT FROM` 按原始整行值匹配记录，避免操作入口长期禁用。
- SQL 中仍使用查询结果实际返回的列名进行引用。
- 增加该场景的回归测试。

## 修改文件

- `frontend/src/lib/tableDataMutations.js`
- `frontend/test/tableDataMutations.test.js`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- `jdbc-agent/src/test/java/com/sshtools/jdbcagent/MetadataServiceImplTest.java`

## 验证

- 执行 `node --test test/tableDataMutations.test.js test/databaseTableWorkspace.test.js test/databaseHomeToolbar.test.js`，通过。
- 执行前端 `npm run build`，Vite 编译通过；JDBC agent Gradle 测试已启动，但本地 Gradle daemon 未返回常规结束摘要。

## 剩余风险

无主键或元数据接口未返回主键时，完全重复的数据行可能被同时更新或删除。界面会在删除前提示该风险；建议为业务表补充主键或唯一约束。
