# 变更：终端 Command+C 无法复制选中内容

## 背景

在 macOS 上，应用菜单包含系统 `EditMenu`。`Command+C` 常由原生「拷贝」菜单项处理，按 WebView DOM 选区写入剪贴板。xterm.js 的选区是自绘图层，不在 DOM selection 中，因此控制台里选中文本后 `Command+C` 复制不到内容。此前仅在 keydown 里 `preventDefault` 并调用 `ClipboardSetText`，既可能拦掉浏览器 `copy` 事件，也无法覆盖菜单加速键先于前端处理的路径。

## 范围

- 监听终端容器的 `copy` 事件，用 `terminal.getSelection()` 写入 `clipboardData`，并同步到 Wails 剪贴板。
- `Command+C` / 复制快捷键路径不再 `preventDefault`，以便原生菜单或浏览器能触发 `copy` 事件；仍返回 `false` 阻止按键进入 PTY。
- macOS 无选区时的 `Cmd+C` 返回 `noop` 吞掉按键；`Ctrl+C` 无选区仍保持中断语义。
- 校验 `ClipboardSetText` 的布尔返回值，失败时回退到 `navigator.clipboard`。
- 快捷键识别同时支持 `event.key` 与 `event.code`。

## 修改文件

- `frontend/src/components/Terminal.svelte`
- `frontend/src/lib/terminalShortcuts.js`
- `frontend/test/terminalShortcuts.test.js`

## 验证

- 执行 `node --test test/terminalShortcuts.test.js`，验证快捷键识别规则。
- 手工：在 SSH 控制台选中文本后按 `Command+C`，粘贴到其他应用应得到选中内容；无选区时 `Ctrl+C` 仍应发送中断。

## 剩余风险

- 若终端未聚焦，原生菜单拷贝仍可能作用到其他可编辑控件；这是预期行为。
- Wails 剪贴板在个别 macOS 环境下仍可能异常；`copy` 事件的 `clipboardData` 路径是主要修复，Wails API 为兜底。
