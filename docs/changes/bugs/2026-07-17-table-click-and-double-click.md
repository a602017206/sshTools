# 表对象单击与双击行为修复

## 背景

对象浏览器改造后，表名单击直接进入数据浏览标签，无法快速查看 DDL，且不符合数据库客户端常见的单击查看定义、双击浏览数据的交互习惯。

## 范围

表名单击打开表结构 DDL；双击在单击计时器触发前取消该操作并打开数据浏览标签。组件销毁时会清理未触发的计时器。

## 修改文件

- `frontend/src/lib/databaseObjectActions.js`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/test/databaseObjectActions.test.js`

## 验证

执行 `node --test frontend/test/databaseObjectActions.test.js`、前端构建和桌面应用构建。手工确认单击只显示 DDL，双击只进入数据浏览。

## 剩余风险

双击识别使用 220 毫秒延迟，极慢的双击会按两次单击处理，符合系统常规双击时间阈值之外的行为。
