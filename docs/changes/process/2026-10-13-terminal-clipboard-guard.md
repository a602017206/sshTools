# 变更：SSH 终端复制粘贴设为受保护区域

## 背景

SSH 终端的复制粘贴已多次因修改终端其他功能而回归：

- 复制含 `{}` 的日志得到空白剪贴板；
- `tail -f` 刷屏时复制为空、Ctrl+V 不能粘贴；
- Cmd+V 粘贴失效（`pasteFromClipboard` 未定义）；
- 开启 `convertEol` 导致 vim 等全屏程序画面被拆散。

纯逻辑已抽到 `terminalSelection.js`、`terminalShortcuts.js` 并有单元测试，但把它们接到 xterm 上的代码位于 `Terminal.svelte`，没有任何测试兜底，改终端其他功能时容易顺手改坏。

## 范围

不改任何运行逻辑，只增加三层约束：

1. **规则**：新增 `.cursor/rules/terminal-clipboard.mdc`（修改终端相关文件时自动生效），并在 `AGENTS.md`、`CLAUDE.md` 写明：未经用户明确要求不得改动复制粘贴，新增按键处理不得拦截 Ctrl/Cmd+C/V，改终端后必须跑契约测试。
2. **契约测试**：新增 `frontend/test/terminalClipboardContract.test.js`，读取 `Terminal.svelte` 源码锁定关键写法：
   - 复制只走 Wails `ClipboardSetText` 且在 `setTimeout` 中执行；
   - 不出现 `execCommand(`、`clipboardData.setData`、`navigator.clipboard.writeText`；
   - 粘贴先读 `ClipboardGetText` 再回退浏览器，并经 `prepareTerminalPaste`；
   - `copy` 监听在捕获阶段且 `preventDefault` + `stopPropagation`；
   - 选区快照（`onSelectionChange`、捕获阶段 `mousedown`、`mouseup`）仍在；
   - 快捷键处理中复制、粘贴分支 `preventDefault` 并返回 `false`；
   - 复制粘贴快捷键语义、`convertEol: false` 不变。
3. **代码标记**：在 `Terminal.svelte` 两段剪贴板代码首尾加 `剪贴板受保护区域 开始/结束` 注释，契约测试校验标记成对存在。

## 修改文件

- `.cursor/rules/terminal-clipboard.mdc`
- `AGENTS.md`
- `CLAUDE.md`
- `frontend/src/components/Terminal.svelte`（仅新增注释）
- `frontend/test/terminalClipboardContract.test.js`
- 本文档

## 验证

```bash
cd frontend && node --test test/terminalClipboardContract.test.js test/terminalSelection.test.js test/terminalShortcuts.test.js
```

- 23 个测试全部通过。
- 变异测试：逐一模拟 6 种回归改法（copy 监听改为冒泡阶段、去掉 `setTimeout`、改用 `execCommand`、快捷键复制不再返回 `false`、粘贴先读浏览器剪贴板、去掉 `mousedown` 捕获），契约测试均失败；还原后全部通过。
- 首版契约测试用正则懒惰匹配检查 `}, true);`，会误匹配到后面的 `mousedown` 监听，变异测试未能拦住，已改为精确截取监听代码块后校验。

## 剩余风险

- 契约测试基于源码文本匹配，只能拦住对受保护写法的改动，无法替代桌面端真实复制粘贴验证；合理的重构（如改变量名、缩进）也会触发失败，此时需经用户同意后同步更新测试。
- 规则只约束遵守 `AGENTS.md` / `CLAUDE.md` / Cursor 规则的工具；人工修改仍依赖代码评审与测试。
- 测试未接入 CI，需要在本地手动运行。
