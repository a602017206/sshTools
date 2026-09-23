# 缺陷：SSH 里 vim 画面被拆散

## 背景

SSH 终端打开 vim 后，文件前几行还能看见，下面却堆满 `,15`、`All` 和右侧一列数字。这些是状态栏被拆开后留在屏幕上的碎片。

## 范围

- xterm 关闭 `convertEol`。该选项会在每次 LF 时把光标拉回第 0 列。
- vim、less、htop 用裸 LF 表示「下移一行、列不变」，被拉回行首后状态栏就会散开。
- 普通 shell 换行仍由远端 PTY 的 ONLCR 转成 CRLF，不依赖这个选项。

## 修改文件

- `frontend/src/lib/xtermOptions.js`
- `frontend/src/components/Terminal.svelte`
- `frontend/test/xtermOptions.test.js`
- 本文档

## 验证

```bash
cd frontend && node --test test/xtermOptions.test.js
```

未在桌面端对真实 SSH 的 vim 做手工复现。请重新打开会话后进 vim，状态栏应留在最后一行。

## 剩余风险

- 若远端关掉了 ONLCR（`stty -onlcr`），shell 的裸 LF 会呈阶梯状。这是终端应有行为，不再在本地强行转成 CRLF。
