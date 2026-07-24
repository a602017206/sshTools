# 数据库对象页表结构预览

## 背景

表设计器迁移到独立标签页后，数据库对象页右侧详情栏只保留数据库信息。选中表时无法快速看到原有的字段结构预览。

## 范围

- 选中表后，在对象页右侧详情栏加载并展示只读字段预览。
- 展示字段名、类型、长度、主键和非空标记。
- 保持完整表设计在独立标签页，右侧预览不提供编辑操作。
- 使用请求键避免快速切换表时异步响应错位。

## 修改文件

- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/test/objectBrowserLayout.test.js`

## 验证

- 执行 `node --test test/objectBrowserLayout.test.js test/tableStructureMetadata.test.js`。
- 执行 `npm run build`。

## 剩余风险

字段数很多的表会使右侧详情栏滚动长度增加；完整字段编辑和 DDL 操作仍应在独立设计表标签页完成。
