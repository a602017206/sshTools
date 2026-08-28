# 原生数据库工作区 Phase 2 拆分

## 背景

`NativeDatabasePanel` 将 Redis / Elasticsearch / Kafka / 其他原生库混在同一组件，能力开关无法表达各域真实交互。按设计将面板拆为独立工作区，并由薄路由按类型分发。

## 范围

- 新增 `RedisWorkspace`、`ElasticsearchWorkspace`、`KafkaWorkspace`、`GenericNativeWorkspace`
- `NativeDatabasePanel` 改为基于 `resolveNativeWorkspaceKind` 的薄路由
- 新建连接对话框数据库类型按域分组（`optgroup`）
- 更新布局、域分组与 `resolveNativeWorkspaceKind` 相关测试
- 删除临时源副本 `_panel_source.svelte`

## 修改文件

- `frontend/src/components/NativeDatabasePanel.svelte`
- `frontend/src/components/workspaces/RedisWorkspace.svelte`
- `frontend/src/components/workspaces/ElasticsearchWorkspace.svelte`
- `frontend/src/components/workspaces/KafkaWorkspace.svelte`
- `frontend/src/components/workspaces/GenericNativeWorkspace.svelte`
- `frontend/src/lib/nativeDatabaseWorkspace.js`
- `frontend/src/lib/assetDomain.js`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/test/nativeDatabasePanelLayout.test.js`
- `frontend/test/nativeDatabaseWorkspace.test.js`
- `frontend/test/assetDomain.test.js`
- `docs/plans/2026-08-28-multi-domain-phase2.md`

## 验证

- `cd frontend && node --test test/nativeDatabasePanelLayout.test.js test/nativeDatabaseWorkspace.test.js test/assetDomain.test.js`

## 剩余风险

- 各工作区仍各自复制了 `native-database-panel__*` 样式，后续可抽公共样式减少漂移
- Kafka 工作区仅为 Topic 元数据只读骨架，尚未覆盖消费/生产
- 未做完整 Wails 端到端手工联调（依赖真实会话）
- `.git/worktrees/es-panel-metadata` 残留元数据可能因权限无法自动清理；分支本身已不存在
