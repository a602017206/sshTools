# 修复 DatabaseTypeIcon PostgreSQL 图标重复 fill 属性

## 背景

`wails build` 前端编译失败：`DatabaseTypeIcon.svelte` 中 PostgreSQL 微笑路径同时写了 `fill="#fff"` 与 `fill="none"`，Svelte 解析报错 `Attributes need to be unique`。

## 范围

- 去掉重复的 `fill="#fff"`，保留描边微笑弧（`fill="none"` + `stroke`）

## 修改文件

- `frontend/src/components/icons/DatabaseTypeIcon.svelte`

## 验证

- `cd frontend && npm run build`

## 剩余风险

- 日志中的 A11y 警告（`AssetList` / `SelectedDatabaseObjects`）不阻断构建，未在本次处理
