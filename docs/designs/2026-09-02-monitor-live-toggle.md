# 服务器性能实时查询开关

## 背景

性能页默认每 2 秒经 SSH 采集 CPU/内存等。弱机或负载高的服务器会因此额外承压；切到「文件」后面板 keep-alive，轮询仍会继续。

## 方案

- 按会话保存「实时」开关，默认关闭
- 关闭时可手动「刷新」一次
- 开启后约每 2 秒轮询；切到非性能页或会话不可用时自动停轮询

## 修改文件

- `frontend/src/lib/monitorLiveControl.js`
- `frontend/test/monitorLiveControl.test.js`
- `frontend/src/components/ServerMonitor.svelte`
- `frontend/src/components/SessionToolDock.svelte`
