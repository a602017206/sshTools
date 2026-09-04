# 修复 SSH 终端粘贴与方向键历史

## 背景

Cmd+C 修复后 Cmd+V 粘贴失效；滚屏后连按上键翻 bash 历史时，光标往上走、当前行被清、历史输出被一点点擦掉。

## 范围

- 补全 `pasteFromClipboard`，走 Wails/浏览器剪贴板读文本后 `onData` 发送
- 滚屏状态下按上/下键前先 `scrollToBottom`，避免 xterm 把方向键当滚屏用
- 显式开启 `scrollOnUserInput`

## 修改文件

- `frontend/src/components/Terminal.svelte`
- `frontend/src/lib/terminalShortcuts.js`
- `frontend/test/terminalShortcuts.test.js`
- `.github/workflows/release.yml`
- `docs/changes/bugs/2026-09-02-terminal-paste-and-arrow-history.md`（本文）

## 验证

- `cd frontend && node --test test/terminalShortcuts.test.js`
- 手工：Cmd+V 粘贴；滚轮上滚后连按 ↑ 仍翻 shell 历史且不擦输出

## 剩余风险

- macOS 上 Ctrl+V 仍作为控制字符发给远端（Vim 等场景）；粘贴请用 Cmd+V 或右键
