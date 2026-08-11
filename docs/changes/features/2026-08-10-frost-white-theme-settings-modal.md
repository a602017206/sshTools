# 变更：霜白配色 + 设置弹窗居中修复

## 背景

用户反馈：整屏偏蓝难看、玻璃效果不明显且不如纯白干净；全局设置也应是居中弹窗，高大内容时会沉到窗口下方。

## 范围

- 亮色主题改为中性霜白/浅灰底，面板近乎不透明白；去掉大块青蓝氛围。
- 默认强调色改为 `teal`（海盐青），仅作点缀。
- `Dialog` 支持高内容滚动居中（设置不再沉底）。
- 设置侧栏样式对齐玻璃边框。
- 不改设置项业务逻辑。

## 修改文件

- `frontend/src/styles/app.css`
- `frontend/src/components/ui/Dialog.svelte`
- `frontend/src/components/GlobalSettingsDialog.svelte`
- `frontend/src/settings/appearance.js`
- `docs/changes/features/2026-08-10-frost-white-theme-settings-modal.md`（本文件）

## 验证

- `cd frontend && npx vite build`
- 手工：打开设置，确认居中遮罩弹窗；主界面应为浅灰白而非大片蓝。

## 剩余风险

- 若本地已保存 `accent_color: blue`，强调色仍可能偏蓝，但背景不再被染成蓝色。可在设置里改「海盐青」或恢复默认。
- 若仍希望强玻璃感，需在「干净白」与「彩色折射」之间再取舍；本期优先干净可读。
