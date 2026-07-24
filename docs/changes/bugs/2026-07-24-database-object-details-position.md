# 数据库对象详情栏位置修复

## 背景

对象页从 grid 布局改为 flex 布局后，主内容区未声明弹性增长，导致主区按内容宽度收缩，详情栏未贴靠对象页右侧并留下空白区域。

## 范围

- 为对象页主内容区增加弹性增长规则，使右侧详情栏始终位于对象页最右侧。

## 修改文件

- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/test/objectBrowserLayout.test.js`

## 验证

- 执行 `node --test test/objectBrowserLayout.test.js`。
- 执行 `npm run build`。
- 执行 `git diff --check`。

## 剩余风险

极窄窗口下详情栏仍按既有响应式规则隐藏，避免挤压对象列表。
