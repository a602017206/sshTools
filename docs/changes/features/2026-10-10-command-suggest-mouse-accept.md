# 命令提示默认仅鼠标采纳，浮层贴近光标

## 背景

SSH 终端出现常用命令提示时，Tab 会直接填入第一条建议，抢占了 Shell 原生 Tab 补全，误触频繁；同时浮层固定在终端左下角，离当前输入行太远。

## 范围

- 新增设置 `command_suggest_tab_accept`（默认 `false`）：
  - 关闭（默认）：只能鼠标点击建议填入；Tab、↑/↓ 全部交给 Shell，Esc 仅关闭浮层。
  - 开启：恢复原行为，Tab 填入高亮项，↑/↓ 切换。
- 设置入口：全局设置 → 常用命令提示 → “Tab 键采纳提示”。
- 浮层改为跟随光标：左侧对齐当前命令起点，默认出现在光标行下方，下方空间不足时翻到上方。

## 修改文件

- `internal/config/config.go`、`internal/config/config_test.go`
- `frontend/wailsjs/go/models.ts`
- `frontend/src/settings/appearance.js`、`frontend/test/appearanceDefaults.test.js`
- `frontend/src/App.svelte`
- `frontend/src/components/GlobalSettingsDialog.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/Terminal.svelte`

## 验证

- `go test ./internal/config/` 通过。
- `node --test test/appearanceDefaults.test.js` 通过。
- `npm run build` 通过。
- 未在真实 SSH 会话中做手工 UI 回归。

## 剩余风险

- 浮层位置按 `cursorX - 本地行缓冲长度` 估算命令起点，含宽字符或命令折行时可能略有偏移（已限制在终端可视范围内）。
- 旧配置文件没有该字段时按 `false` 处理，即升级后默认变为仅鼠标采纳。
