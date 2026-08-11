# 变更：移除会话内容区冗余工具条

## 背景

会话标签栏已展示名称与连接状态；内容区上方再重复 `user@host:port`，以及未接线的复制/粘贴/最小化/最大化图标，占用垂直空间且无实际作用。

## 范围

- 删除 `TerminalPanel` 内每个会话内容顶部的 `session-toolbar`。
- 会话信息继续由标签栏承载；终端/数据库面板直接铺满内容区。
- 不改连接、会话保活与工具坞逻辑。

## 修改文件

- `frontend/src/components/TerminalPanel.svelte`
- `docs/changes/features/2026-08-10-remove-session-toolbar.md`（本文件）

## 验证

- 建议手工：`wails dev` 打开 SSH 会话，确认标签下直接是终端，无中间工具条；切换数据库面板同样无该行。

## 剩余风险

- 数据库子面板若依赖该工具条文案作为唯一上下文提示，需依赖标签名；当前标签已含会话/表相关命名，风险低。
