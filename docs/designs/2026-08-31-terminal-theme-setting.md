# 终端主题与整体外观分离

## 背景

整体外观已支持浅色、深色和跟随系统，但 xterm 视口被写死为深色控制台配色。浅色界面下终端仍是黑底，无法满足「亮壳暗终端」或「全浅色」两类习惯。

## 目标与范围

设置里把两项拆开：

- 整体外观：浅色 / 深色 / 跟随系统（沿用 `theme_mode`）
- 终端主题：深色 / 浅色 / 跟随界面（新用 `terminal_theme`）

终端颜色只由 `terminal_theme` 决定；选「跟随界面」时才跟随已解析的外观明暗。后端 `AppSettings.TerminalTheme` 已存在，补齐设置页、持久化和 xterm 应用。旧值 `default` 按深色处理，保持现有用户观感。

不改 SSH 协议，不引入多套可自定义配色编辑器。

## 架构与取舍

解析与 xterm 色板放在 `terminalTheme.js`，与设置弹窗、Terminal 组件分离，便于单测。外观应用时同步写入 `--ops-terminal-bg`，让视口铬层与 xterm 背景一致。默认仍为深色终端，避免这次改动把现有会话突然变成浅色。

## 风险

已打开的会话依赖 `app:appearance-updated` 刷新 xterm 主题；若事件未带完整 settings，会回落到 `data-terminal-theme`。浅色终端对部分远端彩色输出对比度可能偏低。
