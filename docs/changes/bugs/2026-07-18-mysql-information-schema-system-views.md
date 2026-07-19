# MySQL information_schema 系统视图展示修复

## 背景

数据库模式统一改用 JDBC 元数据接口后，默认表分类仅请求 `TABLE` 与 `SYSTEM TABLE`。MySQL 将 `information_schema` 中的对象报告为 `SYSTEM VIEW` 时，用户进入数据库后默认看到“暂无表”。

## 范围

调整 MySQL 数据库对象浏览器的 JDBC 元数据类型筛选，并同步前端构建产物；不改变查询、权限或其他数据库类型的对象分类。

## 修改文件

- `frontend/src/lib/databaseObjectTree.js`：MySQL 的表与视图分类均额外包含 `SYSTEM VIEW`，使 `information_schema` 在默认表页签可见。
- `frontend/src/components/SelectedDatabaseObjects.svelte`：按当前数据库类型生成对象分类。
- `frontend/test/databaseObjectTree.test.js`：增加 MySQL 系统视图分类回归测试。
- `frontend/build/assets/index.js`：同步前端生产构建产物。

## 验证

执行 `cd frontend && node --test test/databaseObjectTree.test.js`，确认 MySQL 默认表分类包含 `TABLE`、`SYSTEM TABLE` 与 `SYSTEM VIEW`。

## 剩余风险

不同 MySQL JDBC 驱动对元数据类型的报告可能存在差异。本次保留既有 `VIEW` 类型，并新增标准的 `SYSTEM VIEW`，未在真实目标驱动连接上进行端到端验证。
