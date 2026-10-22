# 页签可分辨性增强与批量执行命令

## 背景

用户反馈：同一服务器开多个页签时标题全是同一个 IP，副标题只有「已连接」，分不清当前在哪个页签下操作；同时提出运维批量场景需要把一条命令同时发到多个页签执行。

## 范围

### 页签可分辨（`frontend/src/lib/sessionTabs.js` + TerminalPanel）
- 同一资产的多个页签自动加序号（`192.168.121.110 #1 / #2`，双击改过的名不再加）。
- 副标题改为 `用户@主机:端口`（断开标注），不再只有「已连接」。
- 每个页签按 sessionId 哈希获得稳定配色：页签左侧色条、活动页签同色下划线；终端区顶部有一条与活动页签同色的锚点带，当前在哪个会话一眼可辨。

### 批量执行（新增 `BatchExecuteDialog.svelte`）
- SSH 页签栏新增「批量执行」按钮（带可执行页签数）。弹窗内输入命令（多行逐行发送）、勾选目标页签（序号标题 + 用户@主机 + 同色方点，全选/清空），默认全选已连接远程 SSH 会话。
- 「发送后自动回车执行」默认开启；不勾选则只把命令打到各页签输入行不执行，适合先检查再手动回车。
- 下发逐个调用 `SendSSHData` 并汇总成功/失败数回显。后端无新增接口。

## 修改文件

- `frontend/src/lib/sessionTabs.js`、`frontend/test/sessionTabs.test.js`（新增）
- `frontend/src/components/BatchExecuteDialog.svelte`（新增）
- `frontend/src/components/TerminalPanel.svelte`（页签渲染、色带、批量按钮与下发）
- `.github/workflows/release.yml`（发布说明预填对应条目）

## 验证

- `node --test test/*.test.js`：294 项 292 过，2 项既存失败（Redis/Kafka 面板断言，无关）；新增 sessionTabs 5 项全过。
- 剪贴板契约三件套 23 项全过（改动未触 `Terminal.svelte` 受保护区）。
- `vite build` 与完整 `npm run build` 通过。

## 剩余风险

- 无 GUI 环境未肉眼验证页签配色与弹窗交互，需人工打开多个同机页签复核。
- 批量执行不做危险命令拦截（运维工具惯例），靠目标勾选与「不回车」选项兜底；误发风险由操作者确认清单承担。
