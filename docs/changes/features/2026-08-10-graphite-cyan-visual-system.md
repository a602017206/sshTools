# 变更：Graphite Cyan 视觉重构（Dark First）

## 背景

运维工作台界面布局拥挤、色彩混乱、状态反馈不直观。按 Graphite Cyan 设计方案落地第一阶段视觉与微交互升级。

## 范围

- 设计令牌与默认主题/字体/侧栏宽度
- 会话列表活跃态与连接状态呼吸灯
- SQL 工作区可拖拽分割与结果表空值展示
- 终端点阵氛围与侧栏品牌水印

## 修改文件

- `frontend/src/styles/app.css`
- `frontend/src/settings/appearance.js`
- `frontend/src/stores.js`
- `frontend/src/App.svelte`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/Terminal.svelte`
- `frontend/src/components/AboutDialog.svelte`
- `internal/config/config.go`
- `docs/designs/2026-08-10-graphite-cyan-visual-system.md`
- `docs/development/2026-08-10-graphite-cyan-visual-system.md`

## 验证

- `go test ./internal/config -count=1`
- 前端相关单测（`workspaceTabs` / `tableQueryBuilder` 等）抽测
- 手工核对暗色壳层、侧栏状态灯、SQL 分割与终端点阵

## 剩余风险

- 旧配置可能覆盖新默认值；部分页面仍有遗留硬编码色，未纳入本轮。
