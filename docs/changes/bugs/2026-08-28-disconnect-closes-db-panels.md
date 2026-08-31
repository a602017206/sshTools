# 右键断开数据库时同步关闭右侧相关窗口

## 背景

数据库 / 缓存 / 搜索类资产右键「断开」时，若右侧仍有该连接的工作区标签，会提示断开失败。原因是侧栏只调用 `CloseDatabase`，未关闭 UI 会话，且 Redis/ES 等原生连接应使用 `CloseNativeDatabase`。

## 范围

- 断开时先关闭该连接在右侧的全部相关标签（父工作区 + 查询/表等子面板）
- 再按类型调用 `CloseNativeDatabase` 或 `CloseDatabase`
- 清理资产 `dbConnected` / `dbSessionId` 与侧栏展开状态

## 修改文件

- `frontend/src/lib/databaseSessionClose.js`
- `frontend/test/databaseSessionClose.test.js`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/App.svelte`
- `docs/changes/bugs/2026-08-28-disconnect-closes-db-panels.md`（本文）

## 验证

- `cd frontend && node --test test/databaseSessionClose.test.js`
- 手工：打开 Redis/ES/MySQL 工作区后，资产树右键断开，右侧标签应全部关闭且无「断开失败」提示

## 剩余风险

- `database:disconnect` 为异步处理；极短窗口内连点可能重复触发关闭（后端应幂等）
