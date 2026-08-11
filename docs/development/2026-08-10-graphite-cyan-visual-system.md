# 开发：Graphite Cyan 视觉系统落地

## 实现说明

1. 在 `app.css` 收口 Graphite Cyan Token（中性色阶、强调色、功能色、圆角、阴影、布局度量），并补充：
   - `.ops-asset-row.is-active` 左侧发光竖条
   - `.ops-pulse` 呼吸灯动画（含 `prefers-reduced-motion`）
   - `.ops-split-handle` 1px 分割线
   - `.ops-terminal-canvas` 点阵网格
   - `.brand-watermark` / `.result-table` / `.cell-null`
2. 默认 Dark First：前端 `getDefaultAppSettings` 与 Go `DefaultSettings` 同步为 dark / 侧栏 260 / JetBrains Mono。
3. `AssetList`：按会话映射活跃行与连接状态灯，侧栏底部加入品牌水印。
4. `App`：侧栏拖拽范围 200–420，双击复位 260；分割线改用 token 化样式。
5. `DatabasePanel`：编辑器/结果可拖拽 58/42 分割；结果表 sticky 表头 + NULL/空值斜体占位；执行中左侧信息色条。
6. `DatabaseTablePanel`：行悬停改 `--bg-hover`；空/Null 用 placeholder。
7. `Terminal`：挂载 `.ops-terminal-canvas`；`AboutDialog` Logo 改强调色渐变。

## 验证

- `go test ./internal/config -count=1` — 通过
- `node --test src/lib/workspaceTabs.test.js src/lib/formatConnectionError.test.js test/tableQueryBuilder.test.js` — 19 通过
- 手工：暗色启动 → 侧栏活跃条/呼吸灯 → SQL 分割线拖拽 → 终端点阵可见

## 剩余风险

- 已有用户本地 `localStorage` / `config.json` 会覆盖默认侧栏宽度与主题，需手动切暗色或清配置才能看到默认效果。
- 部分历史组件仍含 Tailwind 紫/灰硬编码，视觉统一需后续分批清理。
