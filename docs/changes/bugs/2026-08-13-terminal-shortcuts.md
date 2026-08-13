# 变更：终端常用复制与粘贴快捷键

## 背景

终端原先仅依赖浏览器 `navigator.clipboard`。在 Wails 桌面应用中，该接口可能受权限或运行时环境限制，使 `Ctrl+C`、`Ctrl+V` 等快捷键无法正确完成复制和粘贴。

## 范围

- 为终端接入 Wails 原生剪贴板，并在不可用时降级到浏览器剪贴板。
- 识别 `Ctrl/Cmd+C`（有选区时）、`Ctrl/Cmd+V`、`Ctrl+Shift+C`、`Ctrl+Shift+V`、`Ctrl+Insert` 和 `Shift+Insert`。
- 接收系统粘贴事件，支持菜单粘贴等非键盘入口。
- 无选区的 `Ctrl+C` 仍交给终端，保持向远端发送中断信号的语义。

## 修改文件

- `frontend/src/components/Terminal.svelte`
- `frontend/src/lib/terminalShortcuts.js`
- `frontend/test/terminalShortcuts.test.js`

## 验证

- 执行 `node --test test/terminalShortcuts.test.js`，验证快捷键识别规则。
- 执行 `npm run build`，验证前端可构建。

## 剩余风险

- 剪贴板的最终可用性仍受操作系统权限控制；若 Wails 原生接口不可用，应用会自动尝试浏览器剪贴板接口。
