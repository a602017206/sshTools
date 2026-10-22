# 修复关闭自动探活后进入状态墙仍探活一次

## 背景

用户在全局设置关闭「启用自动探活」后，进入状态墙页面仍发生一轮探活。原因是 `OpsDashboard` 挂载时无条件先执行一次 `refresh()` 再按开关决定是否启动轮询——首探不受开关约束。

## 范围

- `frontend/src/components/OpsDashboard.svelte`：挂载时先读设置，自动探活关闭则直接返回，不做首探也不启动轮询；「立即刷新」手动探活不受影响（用户主动操作）。

## 修改文件

- `frontend/src/components/OpsDashboard.svelte`
- `docs/changes/bugs/2026-10-22-wall-auto-probe-off-still-probes.md`（本文）

## 验证

- `vite build` 通过；逻辑为挂载分支调整，无新增单测覆盖（组件级行为，需 GUI 人工复核）。

## 剩余风险

- 无 GUI 环境未肉眼复核；复核方法：关自动探活后进墙，服务器 sshd 日志不应新增连接记录，点「立即刷新」才应出现一轮。
