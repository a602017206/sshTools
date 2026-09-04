# 终端主题与整体外观分离

## 背景

无论整体外观选浅色还是深色，终端视口都保持黑底。需要把终端配色做成独立设置，适配亮壳暗终端和全浅色两类需求。

设计文档：`docs/designs/2026-08-31-terminal-theme-setting.md`。

## 范围

- 设置页新增「终端主题」：深色、浅色、跟随界面
- 「主题模式」文案改为「整体外观」，与终端主题并列
- 持久化已有字段 `terminal_theme`，并应用到 xterm 与终端视口背景
- 旧配置 `default` 视为深色

不包含自定义色板编辑、按会话覆盖主题。

## 修改文件

- `frontend/src/lib/terminalTheme.js`（新建）
- `frontend/test/terminalTheme.test.js`（新建）
- `frontend/src/settings/appearance.js`
- `frontend/src/components/GlobalSettingsDialog.svelte`
- `frontend/src/components/Terminal.svelte`
- `frontend/src/App.svelte`
- `frontend/src/styles/app.css`
- `internal/config/config.go`
- `docs/designs/2026-08-31-terminal-theme-setting.md`
- `docs/changes/features/2026-08-31-terminal-theme-setting.md`（本文）

## 验证

- `cd frontend && node --test test/terminalTheme.test.js`
- `cd frontend && node node_modules/vite/bin/vite.js build`
- 手工：浅色外观 + 深色终端应保持黑底；浅色外观 + 浅色终端应变白底；跟随界面应随整体外观切换

## 剩余风险

- 本次无法在 Wails 窗口内自动点选验证，需本地打开设置预览确认
- 浅色终端下部分 ANSI 亮色对比度可能不足
