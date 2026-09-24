# 缺陷：复制含 `{}` 的日志得到空白

## 背景

SSH 控制台里，选中的日志只要同时包含 `{` 和 `}`，Ctrl+C 的结果就是空白。没有这对括号，或者只选中一边，复制是正常的。

## 范围

- 终端复制不再走浏览器 `copy` 事件的 `clipboardData`，也不再用隐藏文本框执行 `execCommand('copy')`。
- 复制事件被拦住后，等事件结束，再通过 Wails 的 `ClipboardSetText`（macOS 上是 `pbcopy`）写入。

## 修改文件

- `frontend/src/components/Terminal.svelte`
- `.github/workflows/release.yml`
- 本文档

## 验证

```bash
cd frontend && node --test test/terminalSelection.test.js test/terminalShortcuts.test.js
```

未在桌面端对含 `{}` 的真实日志做手工复制。

## 剩余风险

系统剪贴板写入是异步的。如果 `ClipboardSetText` 本身失败，没有再回退到浏览器剪贴板，避免再次把成对括号写成空白。
