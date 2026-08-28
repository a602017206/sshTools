# 资产树数据库类型图标区分

## 背景

左侧资产树中，所有数据库连接都使用相同的蓝色圆柱图标，无法快速区分 MySQL、Redis、Elasticsearch 等不同类型。

## 范围

- 新增 `DatabaseTypeIcon` 组件，为各数据库类型提供专属品牌色图标
- 资产列表根据连接的 `db_type` 渲染对应图标
- 鼠标悬停时显示数据库类型名称

## 修改文件

- `frontend/src/lib/databaseTypeIcon.js`（新增）
- `frontend/src/components/icons/DatabaseTypeIcon.svelte`（新增）
- `frontend/src/components/AssetList.svelte`
- `frontend/test/databaseTypeIcon.test.js`（新增）

## 验证

- [x] `node --test frontend/test/databaseTypeIcon.test.js`
- [ ] 手工检查：资产树中 Redis、Elasticsearch、MySQL 等连接显示不同图标
- [ ] 手工检查：未知 JDBC 类型仍显示通用数据库图标

## 剩余风险

- 图标为简化品牌色图形，与官方 Logo 并非完全一致，但足以区分常见类型
