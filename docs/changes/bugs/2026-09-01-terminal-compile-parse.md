# 修复终端面板生产构建解析失败

## 背景

`wails build -platform darwin/arm64` 在编译前端时失败：`TerminalPanel.svelte` 出现 `'return' outside of function`。原因是热切换编码时插入 `handleLiveCharsetChange` / `persistConnectionCharset`，误删了 `handleTerminalData` 函数头，循环体留在 script 顶层。

## 范围

- 恢复 `handleTerminalData(sessionId, data)` 包裹
- 去掉 `Terminal.svelte` 编码下拉改版后残留的多余 `</div>`

不改终端编码与复制粘贴行为。

## 修改文件

- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/Terminal.svelte`
- `docs/changes/bugs/2026-09-01-terminal-compile-parse.md`（本文）

## 验证

- `cd frontend && npm run build`：Vite 产出成功（`✓ built in 4.92s`）
- 沙箱内 JDBC agent Gradle 下载失败，与本次 Svelte 解析无关

## 剩余风险

- 本地完整 `wails build` 仍需能访问 Gradle 分发
