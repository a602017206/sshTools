# PC 端多资产状态墙设计

## 背景

`frontend/src/components/OpsDashboard.svelte` 是早期统一运维工作区留下的概览页：布局完整，但无人引用，且「会话容量 64%、CPU 18.6%」等是写死占位值。本次把它做成真实功能：PC 端多资产状态墙。

## 已定口径（用户拍板）

- 入口：顶栏模式切换新增第三项「状态墙」，与「SSH 会话 / 数据」并列；`TerminalPanel` 常驻挂载，切墙不中断已有会话。
- 数据：全部已保存资产做 TCP 探活（可达性 + 延迟，不登录、不需要凭据）；已建立 SSH 会话的资产叠加真实监控指标（CPU / 内存，复用 `GetMonitoringData`）。
- 刷新：墙打开期间自动轮询（探活 + 已连接会话指标同轮采样），另有手动刷新；点击卡片连接资产或切到其已有会话。
- 灵活控制（后续细化）：全局开关可停自动探活、间隔 5–300 秒可调、指标采样可单独关；每个资产可在卡上单独关闭探活（标记存连接 metadata 的 `status_wall_excluded`）。探活本身只是 TCP 握手，对服务器性能影响可忽略，设开关主要是避开 sshd/数据库的连接日志噪音与敏感环境限制。

## 方案

### 后端（根包 `asset_probe.go`）

- `App.ProbeAllAssets() []AssetProbeResult`：读取配置中全部连接（SSH 与数据库资产都只有 host/port 参与探活），并发 `net.DialTimeout`（超时 2 秒、并发上限 16），返回每资产的 `reachable / latency_ms / error / checked_at`。
- 纯函数与 App 方法分离，纯函数可用本机监听端口做单测。

### 数据合并（前端纯逻辑 `frontend/src/lib/assetWall.js`）

- `buildAssetCards(assets, sessions, probes, metrics)` 把资产、会话、探活、指标合并成卡片并派生状态：已连接 > 探活在线 > 探活离线 > 未知（未探）。
- `summarizeCards` 出顶部统计；`groupCards` 按资产分组分段。纯函数配套 node 单测。

### 界面（`OpsDashboard.svelte` 重写）

- 复用既有视觉语言（metric 卡、信号色），全部换真数据；卡片网格按分组排列，卡片含状态点、名称、host:port、类型、延迟，已连接卡显示 CPU/内存小条。
- 删除原假数据（64% 会话容量、18.6% CPU 等）与无落点的快捷动作占位。
- 墙以覆盖层形式盖在主舞台（参照数据库空态覆盖层），组件卸载即停轮询。

## 不做

- 不对未连接资产做 SSH 登录级健康检查（慢且需要凭据）；不做告警、历史曲线与阈值配置；移动端不涉及（本次仅 PC）。

## 验证

- Go：探活纯函数单测（可达 / 不可达 / 空主机）；`go test ./...`。
- 前端：`assetWall` 纯逻辑单测；剪贴板契约三件套回归（动了 App/TerminalPanel 挂载需确认终端不受影响）；`npm run build`。
- 实际界面效果需在 `wails dev` 或打包后人工确认（本环境无 GUI）。
