# 状态墙探活与采样控制开关

## 背景

状态墙上线后用户问：探活和 CPU/内存采样是否影响服务器性能。评估结论：TCP 探活只是握手即断，指标采样走已有会话读 `/proc`，性能影响都可忽略；真实成本是 sshd/数据库的连接日志噪音与敏感环境限制。因此补齐灵活控制，让用户按环境决定开多大、开多勤、开哪几台。

## 范围

- 全局设置新增「多资产状态墙」分区：`asset_wall_enabled`（自动探活总开关）、`asset_wall_metrics_enabled`（已连接会话指标采样开关）、`asset_wall_probe_interval_seconds`（间隔 5–300 秒，默认 15；前后端双侧钳制）。
- 按资产开关：后端新增 `SetAssetProbeExcluded(id, excluded)`，标记持久化在连接 metadata 的 `status_wall_excluded`；`ProbeAllAssets` 跳过被标记资产。状态墙卡片右下角有「关闭探活 / 开启探活」切换，状态单列「已关闭探活」；被关资产的会话指标也不再采样（仍可正常连接使用）。
- 墙面消费设置：关自动探活时不启动轮询（手动刷新仍可用），间隔按设置生效；文案说明探活的影响口径。

## 修改文件

- `internal/config/config.go`（设置字段、解析与钳制、默认值）、`internal/config/config_test.go`
- `asset_probe.go`（跳过排除资产 + `SetAssetProbeExcluded`）、`asset_probe_test.go`
- `frontend/wailsjs/go/main/App.js`、`App.d.ts`、`frontend/wailsjs/go/models.ts`（绑定与设置模型）
- `frontend/src/settings/appearance.js`、`frontend/test/appearanceDefaults.test.js`
- `frontend/src/components/GlobalSettingsDialog.svelte`（状态墙设置分区）、`frontend/src/App.svelte`（保存透传）
- `frontend/src/lib/assetWall.js`（excluded 状态派生）、`frontend/test/assetWall.test.js`
- `frontend/src/components/OpsDashboard.svelte`（消费设置、卡上单资产开关、卡片结构改为主按钮 + 独立开关按钮避免嵌套）
- `docs/designs/2026-10-22-ops-status-wall-design.md`（设计补记）

## 验证

- `go build ./...`、`go test ./...` 全绿（含配置钳制与排除过滤新断言）。
- 前端 `node --test test/*.test.js`：286 项 284 过，2 项既存失败（`nativeDatabasePanelLayout.test.js`，无关）；剪贴板契约三件套 23 项全过；`assetWall`/`appearanceDefaults` 新断言全过。
- `vite build` 编译通过。

## 剩余风险

- 无 GUI 环境，未实际打开设置对话框与状态墙确认交互，需人工看一眼。
- 单资产标记存于连接 metadata：克隆连接会带上该标记，符合「这台机器不探活」的预期。
