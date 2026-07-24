# 表结构设计器与新建表页面

## 背景

原表结构面板只展示字段和 DDL，新建表只能跳转到 SQL 模板，缺少可视化字段设计体验。

## 范围

- 表结构面板统一为设计器布局，展示字段名、类型、长度、非空、主键、默认值和注释。
- 新建表进入创建模式，可添加和删除字段，并实时预览 DDL。
- 创建模式执行 `CREATE TABLE` 后显示创建结果。
- 现有表设计模式当前只读展示其结构和 DDL，不执行 `ALTER TABLE`。

## 修改文件

- `frontend/src/components/TableStructurePanel.svelte`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/App.svelte`
- `frontend/src/stores.js`
- `frontend/test/tableStructureDesigner.test.js`

## 验证

- 执行 `node --test test/databaseTableWorkspace.test.js test/tableStructureDesigner.test.js test/tableDataMutations.test.js test/databaseHomeToolbar.test.js`，通过。
- 执行前端 `npm run build`，Vite 编译通过。

## 剩余风险

创建 SQL 采用通用字段类型映射。不同数据库特有的存储引擎、字符集、索引和外键仍需后续扩展；既有表的结构修改尚未生成数据库方言对应的 `ALTER TABLE` 语句。
