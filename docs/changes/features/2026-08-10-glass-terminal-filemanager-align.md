# 变更：终端与文件管理对齐玻璃拟态高保真

## 背景

对照 `docs/designs/assets/aether-glass/` 高保真稿，SSH 模式下差异最大：亮色主题下终端变成白底控制台，文件管理仍是实心白/紫色列表，主区缺少浮动玻璃外框与深色圆角视口。

## 范围

- 终端视口固定深色（不随壳层亮色变白），圆角嵌入玻璃工作区。
- 左资源树 / 右会话工具坞改为圆角浮动毛玻璃面板，主区留出间距。
- 文件管理去实心白底与紫色强调，改用 glass token 与信号色。
- 不改连接、SFTP、会话保活逻辑；不恢复顶栏「新建连接」。

## 修改文件

- `frontend/src/components/Terminal.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/FileManager.svelte`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/App.svelte`
- `frontend/src/styles/app.css`
- `docs/changes/features/2026-08-10-glass-terminal-filemanager-align.md`（本文件）

## 验证

- `cd frontend && npx vite build`
- 建议手工：`wails dev` 亮色主题下打开 SSH，确认终端为深色圆角视口；右侧文件路径条与列表为玻璃风格；侧栏/工具坞有圆角留白。

## 剩余风险

- 高保真稿中的「快速命令输入条 / 底栏状态」尚未实现，属产品能力增量，本期只做视觉对齐。
- FileManager 设置弹层内部仍有部分旧紫色装饰，不影响主工作面观感。
- `:has` 已避免；数据库面板与终端视口通过活动会话类型切换样式。
