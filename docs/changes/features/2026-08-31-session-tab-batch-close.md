# 会话标签批量关闭

## 背景

顶部会话标签打开过多时只能逐个关闭，缺少浏览器式的批量关闭入口。

设计文档：`docs/designs/2026-08-31-session-tab-batch-close.md`。

## 范围

- 在会话标签上右键提供：全部关闭、关闭左侧、关闭右侧、关闭其它
- 已连接 SSH 的批量关闭合并为一次确认
- 关闭范围按右键目标标签计算，并禁用当前不可用的项

不包含标签拖拽排序、固定标签、关闭按钮常显。

## 修改文件

- `frontend/src/lib/sessionTabClose.js`（新建）
- `frontend/test/sessionTabClose.test.js`（新建）
- `frontend/src/components/TerminalPanel.svelte`
- `docs/designs/2026-08-31-session-tab-batch-close.md`
- `docs/changes/features/2026-08-31-session-tab-batch-close.md`（本文）

## 验证

- `cd frontend && node --test test/sessionTabClose.test.js`
- 手工：打开多个 SSH 标签后右键中间标签，分别验证关闭左侧、右侧、其它、全部；首个标签上「关闭左侧」应禁用

## 剩余风险

- 桌面窗口边缘处菜单可能被裁切
- 本次无法在 Wails 窗口内自动点选验证，需本地确认右键菜单与确认框
