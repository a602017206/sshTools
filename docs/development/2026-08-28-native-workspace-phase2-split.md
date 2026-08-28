# 原生数据库工作区 Phase 2 实现说明

## 实现摘要

将原单体 `NativeDatabasePanel` 拆为四个工作区组件，并由面板按 `resolveNativeWorkspaceKind(db_type)` 路由：

| kind | 组件 | 能力 |
|------|------|------|
| `redis` | `RedisWorkspace` | 逻辑库下拉、键列表、`RedisKeyEditor`、保存/删除 |
| `elasticsearch` | `ElasticsearchWorkspace` | 集群概览、资源/查询页签、索引搜索、DSL、文档写删、mapping、可调 inspector |
| `kafka` | `KafkaWorkspace` | Topic 扁平列表 + Describe JSON；仅展示分区元数据 |
| `generic` | `GenericNativeWorkspace` | 可展开树或扁平列表；可选 Describe |

## 执行细节

1. 从 `_panel_source.svelte` 按域裁剪逻辑与样式，修正 `workspaces/` 下的相对导入路径。
2. `ConfirmDialog` 继续使用 `onConfirm` / `onCancel` props（非事件）。
3. Redis 检查器仅走编辑器路径，不再回退展示通用 JSON。
4. 删除临时源文件 `_panel_source.svelte`。

## 验证命令

```bash
cd frontend
node --test test/nativeDatabasePanelLayout.test.js test/nativeDatabaseWorkspace.test.js
```
