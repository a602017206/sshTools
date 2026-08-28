# 多域导航轨与品牌风格资产图标（Phase 1）

## 背景

Redis / Elasticsearch / Kafka 等资产此前与关系型库混在同一「数据库」列表，侧栏图标同构换色，难以区分类型；后续 Docker 等模块也不适合继续塞进同一抽象。

设计文档：`docs/designs/2026-08-28-multi-domain-nav-and-brand-icons.md`。

## 范围

Phase 1 仅实现：

1. 资产域推导与按域过滤（`assetDomain`）
2. 左侧 `DomainRail` 域切换轨（全部 / SSH / 数据库 / 缓存 / 搜索 / 消息队列 / Docker）
3. 品牌识别色与简化图形的 `DatabaseTypeIcon`（含 SSH / Docker / NoSQL）
4. 域选择持久化；新建连接时按当前域预选类型；保存时写入 `metadata.domain`

不包含：独立 Redis/ES/Kafka 工作区拆分、新建对话框按域分组、Docker 运行时能力。

## 修改文件

- `frontend/src/lib/assetDomain.js`（新建）
- `frontend/test/assetDomain.test.js`（新建）
- `frontend/src/components/DomainRail.svelte`（新建）
- `frontend/src/components/icons/DatabaseTypeIcon.svelte`
- `frontend/src/lib/databaseTypeIcon.js`（既有解析逻辑沿用）
- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/App.svelte`
- `frontend/src/stores.js`
- `docs/designs/2026-08-28-multi-domain-nav-and-brand-icons.md`
- `docs/changes/features/2026-08-28-multi-domain-nav-and-brand-icons.md`（本文）

## 验证

- `cd frontend && node --test test/assetDomain.test.js test/databaseTypeIcon.test.js`
- 手工：切换「缓存」仅见 Redis/Memcached；「搜索」仅见 ES；折叠侧栏再展开后域选择保持；侧栏 Redis / ES / MySQL 图标可区分

## 剩余风险

- 旧连接无 `metadata.domain` 时依赖 `type` + `db_type` 推导；未知 `db_type` 默认归入 `database`
- 域轨与顶部 `WorkspaceNavigation` 仍并存：本阶段域轨只过滤资产，不强制切换主舞台模式
- Docker 域入口可见但无完整工作区，可能造成「空列表」预期差
