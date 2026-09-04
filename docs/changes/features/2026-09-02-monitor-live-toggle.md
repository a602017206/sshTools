# 服务器性能实时查询开关

## 背景

部分服务器性能较弱，性能页持续实时采集会加重压力。需要可按会话关闭实时查询。

## 范围

- 性能页增加「实时」开关（按会话，默认关）与「刷新」一次
- 面板隐藏或会话不可用时停止轮询

## 修改文件

- `frontend/src/lib/monitorLiveControl.js`
- `frontend/test/monitorLiveControl.test.js`
- `frontend/src/components/ServerMonitor.svelte`
- `frontend/src/components/SessionToolDock.svelte`
- `docs/designs/2026-09-02-monitor-live-toggle.md`
- `.github/workflows/release.yml`
- `docs/changes/features/2026-09-02-monitor-live-toggle.md`（本文）

## 验证

- `cd frontend && node --test test/monitorLiveControl.test.js`
- 手工：默认不轮询；开实时后约 2 秒更新；关实时后停；切「文件」后停；再开「性能」且实时开则恢复

## 剩余风险

- 开关仅会话内存态，重连后需重新开启
- 手动刷新仍会走完整采集命令，极弱机也可能短暂卡顿
