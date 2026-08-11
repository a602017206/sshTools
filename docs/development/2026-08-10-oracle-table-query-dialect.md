# Oracle 表查询方言修复实现说明

## 实现

- `buildTableBrowseSQL` 增加按 `databaseType` 分支的分页子句
- 新增导出 `buildQualifiedTableName`，供表数据面板与查询面板共用
- `DatabaseTablePanel` 默认查询 / 翻页改为调用 `buildTableBrowseSQL`，不再字符串替换 `LIMIT 100`
- 查询失败、删除、保存错误统一经 `formatConnectionError` 展示

## 验证

```bash
cd frontend
node --test test/tableQueryBuilder.test.js src/lib/formatConnectionError.test.js
```

全部通过。未跑完整前端构建；变更限于 SQL 拼装与错误文案，风险较低。
