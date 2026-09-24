# 缺陷：tail -f 后复制日志没有内容，Ctrl+V 不能粘贴

## 背景

SSH 控制台选中 `tail -f` 日志后按 Ctrl+C，复制动作发生了，剪贴板里没有文本。Ctrl+V 也不能把日志贴出去。

## 范围

- 选区还在但当前文本已空时，使用拖动时记下的快照。
- 复制时阻止空的 DOM 选区覆盖剪贴板。
- Ctrl+V 在控制台中粘贴。无选区的 Ctrl+C 仍是中断。

## 修改文件

- `docs/designs/2026-09-23-terminal-tail-copy.md`
- `docs/development/2026-09-23-terminal-tail-copy.md`
- `frontend/src/lib/terminalSelection.js`
- `frontend/src/lib/terminalShortcuts.js`
- `frontend/src/components/Terminal.svelte`
- `frontend/test/terminalSelection.test.js`
- `frontend/test/terminalShortcuts.test.js`
- `.github/workflows/release.yml`
- 本文档

## 验证

```bash
cd frontend && node --test test/terminalSelection.test.js test/terminalShortcuts.test.js
```

未在桌面应用里对真实 SSH 会话做 `tail -f` 后的选中、Ctrl+C、Ctrl+V 手工验证。

## 剩余风险

vim 等程序里的 Ctrl+V 会在本地被粘贴占用，不再发给远端。快照只覆盖「选区还在但文本变空」；若输出把选区整个清掉，Ctrl+C 仍会发送中断。
