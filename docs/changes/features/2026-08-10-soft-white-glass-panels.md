# 变更：轻微玻璃（白面板 + 极淡阴影）

## 背景

用户要求面板呈现「轻微玻璃」：白底、极淡阴影，并且不再被强调色（尤其蓝色）染屏。

## 范围

- 亮色面板改为近不透明白（约 92%–98%），弱化模糊与饱和度。
- 阴影改为中性灰的极淡层叠；去掉弹窗上的强调色描边晕。
- 氛围光与强调色解耦：`--bg-glow-*` 固定中性 slate，切换主题色不再染背景。
- 海盐青次色去掉天空蓝 `#38bdf8`，改用同系青绿。
- 工作区模式切换 pill 选中态改为白底中性描边，不再青/蓝填充。

## 修改文件

- `frontend/src/styles/app.css`
- `frontend/src/settings/appearance.js`
- `frontend/src/components/WorkspaceNavigation.svelte`
- `docs/changes/features/2026-08-10-soft-white-glass-panels.md`（本文件）

## 验证

- `cd frontend && npx vite build`
- 手工：亮色下侧栏/顶栏/浮动面板应为近白 + 轻阴影；切换「晴空蓝」强调色时背景不应发蓝。

## 剩余风险

- 若本地仍存旧 inline 样式或缓存，需硬刷新/`wails dev` 重载。
- 深色主题仍为 slate 玻璃，未改成纯白策略（仅亮色走白面板）。
