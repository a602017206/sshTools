# PC 端多资产状态墙

## 背景

`frontend/src/components/OpsDashboard.svelte` 是早期留下的概览页：无人引用进主界面，且「会话容量 64%、CPU 18.6%」等为写死占位。此前盘点确认它是半成品孤儿组件。本次按用户拍板口径做成真实功能（设计见 `docs/designs/2026-10-22-ops-status-wall-design.md`）。

## 范围

- 入口：顶栏模式切换新增「状态墙」（`workspaceTabs.js` 的 `APP_MODES`），与「SSH 会话 / 数据」并列。墙以覆盖层盖主舞台，`TerminalPanel` 常驻不卸载，切墙不中断已有会话（给 TerminalPanel 传的模式把 overview 映射回 ssh，因其内部有 `mode === 'ssh'` 渲染分支）。
- 探活：后端新增 `App.ProbeAllAssets()`，对全部已保存连接并发做 TCP 探活（超时 2 秒、并发 16），返回可达性、延迟与错误；不登录、不需要凭据，SSH 与数据库资产一视同仁。
- 真实指标：已连接的 SSH 会话同轮采样 `GetMonitoringData`，卡片显示 CPU / 内存；删除原组件全部占位假数据与无落点快捷动作。
- 刷新与交互：墙打开期间每 15 秒自动一轮 + 手动刷新；点击卡片复用 `handleConnect` 连接或切回已有会话。

## 修改文件

- `asset_probe.go`、`asset_probe_test.go`（新增，后端探活与单测）
- `frontend/wailsjs/go/main/App.js`、`App.d.ts`、`frontend/wailsjs/go/models.ts`（手工补齐绑定与 `main.AssetProbeResult` 模型，与 wails 生成格式一致）
- `frontend/src/lib/assetWall.js`（新增，资产/会话/探活/指标合并纯逻辑）、`frontend/test/assetWall.test.js`（新增）
- `frontend/src/components/OpsDashboard.svelte`（重写为真数据状态墙）
- `frontend/src/lib/workspaceTabs.js`、`frontend/src/App.svelte`（模式接入与覆盖层）
- `docs/designs/2026-10-22-ops-status-wall-design.md`（设计）
- `.github/workflows/release.yml`（下一版本发布说明预填本次新增功能一条）

## 验证

- `go build ./...` 通过；`go test . ./internal/...` 全绿，含探活 4 项新单测（可达 / 不可达 / 空主机 / 保序）。
- 前端 `node --test test/*.test.js`：285 项 283 过；2 项失败为既存的 `nativeDatabasePanelLayout.test.js`（Redis/Kafka 面板断言，与本次文件无关）；新增 `assetWall` 5 项全过；剪贴板契约三件套在全量中通过。
- `npm run build` 成功（含 vite 构建与 JDBC agent 暂存）。

## 剩余风险

- 无 GUI 环境，未实际打开窗口确认墙面渲染与卡片点击跳转，需在真机（wails dev 或打包）人工看一眼。
- 探活只到 TCP 层：端口可达不等于 SSH 服务可用，卡片文案用「在线」表示可达，已连接状态才是登录级确认。
- 资产很多时每 15 秒一轮全量探活，后端已做并发上限与 2 秒超时；百台以上规模未实测。
