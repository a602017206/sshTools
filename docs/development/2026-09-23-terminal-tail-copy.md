# 实现：tail -f 时复制日志与 Ctrl+V 粘贴

## 做法

- `terminalSelection.js` 维护选区快照，并负责粘贴前的换行与 bracketed paste。
- `Terminal.svelte` 在 `selectionchange`、按下和松开鼠标时更新快照。复制用解析后的文本，捕获阶段拦住 `copy`。
- `getTerminalShortcutAction` 把 Ctrl+V 算作粘贴。单独的 Control 键（含 `ControlLeft`）继续吞掉，避免 xterm 把修饰键当成输入滚到底部并清掉选区。

## 验证

见 `docs/changes/bugs/2026-09-23-terminal-tail-copy.md`。桌面端未能在真实 SSH 的 `tail -f` 里点选验证。
