# 人大金仓左右表列表不一致

## 背景

人大金仓连接的左侧展开树通过 `ListDatabaseTablesInDatabase` 能正确显示表名，但右侧表信息列表使用另一套自定义 SQL。该查询只覆盖固定 schema，导致同一数据库在左侧有表、右侧显示暂无表。

## 范围

右侧表信息列表改为复用 `ListDatabaseTablesInDatabase`，与左侧共享 JDBC 元数据来源。表名正常显示；JDBC 表列表接口未提供行数、数据大小等统计信息时，右侧对应列保持空占位。

## 修改文件

- `frontend/src/components/DatabaseListPanel.svelte`
- `frontend/src/lib/tableMetadataQuery.js`
- `frontend/test/tableMetadataQuery.test.js`
- 本变更记录。

## 验证

执行 `node --test test/tableMetadataQuery.test.js`，验证 API 返回的表名可转换为右侧表信息项；执行前端生产构建验证组件编译。

## 剩余风险

右侧不再展示数据库厂商专有的表统计字段。后续若 JDBC agent 增加跨厂商表统计元数据接口，可在不改变表名数据源的前提下补充这些字段。
