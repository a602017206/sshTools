# 修复 SSH 连接后资产树状态灯不刷新

## 背景

SSH 连接成功后，标签页显示「已连接」且为绿灯，但资产树行内状态灯仍为灰色 idle。

## 根因

`AssetList` 通过函数闭包读取 `connectionSessions` 计算灯色：

```svelte
class={`ops-pulse is-${resolveAssetLinkState(asset)}`}
```

Svelte 编译器只把模板参数 `asset` 标为依赖，看不到函数内部的 `connectionSessions`。会话 `connected` 变为 `true` 后，行内 class 不会重算。

## 范围

- 增加响应式 `assetLinkStateById`，显式依赖 `connectionSessions` 与 `$assetsStore`
- 模板改为读取 `assetLinkStateById[asset.id]`

## 修改文件

- `frontend/src/components/AssetList.svelte`
- `docs/changes/bugs/2026-08-28-asset-link-state-not-refreshing.md`（本文）

## 验证

- `node --test frontend/test/assetLinkState.test.js`
- 手工：连接 SSH 后资产树对应行状态灯应变绿；断开后恢复灰色

## 剩余风险

- 其他模板若用「闭包读 store」的函数算 class，仍可能有同类问题
