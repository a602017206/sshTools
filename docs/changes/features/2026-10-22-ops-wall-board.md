# 状态墙升级为可自选看板（机器与参数逐卡可配）

## 背景

用户希望参考 HexHub 移动端看板：看板默认空、手动指定加哪些机器、每台机器单独指定显示哪些参数，而不是全部资产一股脑上墙。经选项确认：默认空看板、逐卡选参数、墙内「编辑看板」模式管理。

## 范围

- 看板成员：连接 metadata 新增 `wall_on_board`。墙只渲染已加入的资产；后端 `ProbeAllAssets` 探活范围同步收敛为「在看板且未关闭探活」的资产，墙外机器零打扰。指标采样同样只采看板内会话。
- 参数目录：`cpu / memory / disk / network / load / uptime / processes`（与后端白名单一致），逐卡存在 metadata 的 `wall_metrics`（逗号分隔），未配置默认 CPU + 内存。磁盘取最忙分区，网络显示上下行速率，负载取 1 分钟 load。
- 编辑模式：墙头「编辑看板」开关。编辑态下每卡出现参数 chip 组（逐个勾选即时保存）与「移出看板」；底部列出未上看板机器，可一键加入。新增资产默认拥有上次保存的参数组合。
- 后端新增 `SetAssetWallPrefs(id, onBoard, metrics)`，参数 id 服务端白名单清洗；wails 绑定手工补齐。

## 修改文件

- `asset_probe.go`、`asset_probe_test.go`
- `frontend/wailsjs/go/main/App.js`、`App.d.ts`
- `frontend/src/lib/assetWall.js`（参数目录、解析、指标行拼装）、`frontend/test/assetWall.test.js`
- `frontend/src/components/OpsDashboard.svelte`（看板过滤汇总、编辑模式、添加列表、参数 chip）
- `docs/designs/2026-10-22-ops-status-wall-design.md`、`docs/changes/bugs/2026-10-22-wall-auto-probe-off-still-probes.md`（同批 bug 修复，单独成文）
- `.github/workflows/release.yml`（发布说明新增功能条目补看板自选描述）

## 验证

- `go build ./...`、`go test .`（探活过滤与参数清洗）通过。
- 前端 `node --test test/*.test.js`：289 项 287 过，2 项既存失败（Redis/Kafka 面板断言，无关）；`assetWall` 新增参数解析/指标行/看板标记 3 项通过。
- `vite build` 编译通过。

## 剩余风险

- 无 GUI 环境，编辑模式交互（chip 勾选即时保存、加入/移出）未肉眼验证，需人工过一遍。
- 已连接但未上看板的会话不再被状态墙采样指标（属预期收敛）；单会话性能页不受影响。
