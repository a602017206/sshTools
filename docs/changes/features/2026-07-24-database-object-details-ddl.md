# 数据库对象页详情栏 DDL 与可调宽度

## 背景

对象页右侧只能查看固定宽度的字段摘要，无法查看表 DDL，也不能根据内容调整宽度或暂时隐藏详情栏。

## 范围

- 增加对象信息与 DDL 两种详情模式。
- 为右侧详情栏增加拖拽调整宽度能力和显示/隐藏开关。
- 为 DDL 加载、空状态和错误状态提供反馈。

## 修改文件

- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/test/objectBrowserLayout.test.js`
- `docs/designs/2026-07-24-database-object-details-ddl.md`

## 验证

- 执行 `node --test test/objectBrowserLayout.test.js test/tableStructureMetadata.test.js`。
- 执行 `npm run build`。
- 执行 `git diff --check`。

## 剩余风险

DDL 由 JDBC 元数据和各数据库驱动返回；个别驱动权限不足或不支持提取完整 DDL 时，界面将显示加载错误或空状态。移动端宽度下详情栏维持隐藏，仍可通过独立表设计标签页操作。
