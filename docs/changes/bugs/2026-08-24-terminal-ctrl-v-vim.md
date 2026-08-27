# 变更：Vim 中 Ctrl+V 被错误作为粘贴处理

## 背景

SSH 控制台将 `Ctrl+V` 识别为本地剪贴板粘贴。Vim 将该按键作为字面量输入控制字符；例如在 `/` 搜索输入中需要用它转义下一个字符。前端截获该按键后改发剪贴板内容，导致搜索行出现异常文本或乱码。

## 范围

- 保留裸 `Ctrl+V` 给 xterm.js 和远端 PTY，使其发送终端控制字符。
- macOS 继续以 `Cmd+V` 粘贴；同时保留 `Ctrl+Shift+V` 与 `Shift+Insert` 粘贴。
- 为 `event.key` 和 `event.code` 两条按键识别路径增加回归覆盖。

## 修改文件

- `frontend/src/lib/terminalShortcuts.js`
- `frontend/test/terminalShortcuts.test.js`
- `docs/changes/bugs/2026-08-24-terminal-ctrl-v-vim.md`

## 验证

- 执行 `node --test test/terminalShortcuts.test.js`。
- 手工：SSH 后在 Vim 中输入 `/`、`Ctrl+V` 和任意字符，应由 Vim 接收字面量输入；`Cmd+V`、`Ctrl+Shift+V` 仍可粘贴。

## 剩余风险

- 不同远端程序可能对 `Ctrl+V` 有不同语义；本次修复恢复标准终端透传行为。
