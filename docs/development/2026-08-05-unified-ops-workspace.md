# 统一运维工作区实现说明

## 实现内容

- `workspaceTabs.js` 提供七个顶层工作区与输入兜底，避免未知导航值破坏主界面。
- `WorkspaceNavigation.svelte` 提供可键盘访问的工作区标签，并以单一青色高亮当前页。
- `OpsDashboard.svelte` 从现有资产和连接 store 计算资源数量、活跃会话和连接列表；没有实时会话时使用明确空状态，而非伪造监控数据。
- `App.svelte` 保持现有 Wails 生命周期、连接、终端、SFTP、监控和数据库逻辑，在顶部加入工作区导航。概览为默认页，其他页继续使用既有工作台布局；文件和性能页会展开右侧工具面板。
- 文件、性能和数据库工作区改为独立全屏工作面，直接复用 `FileManager`、`ServerMonitor` 与 `TerminalPanel`；终端保留资源树和上下文工具侧栏。
- `OpsUnavailableWorkspace.svelte` 为尚未接入后端的 Docker、日志提供明确的接入状态和下一步，不伪装成已可执行功能。
- `app.css` 新增工作区级 token 和背景层，不改变既有弹窗与工具组件的颜色规则。

## 验证

- `node --test src/lib/workspaceTabs.test.js`：3 个工作区状态测试通过。
- `npm run build`：Vite 前端打包与 JDBC Agent Gradle 打包均完成。

## 已知限制

- Docker、Redis 和云资源树仅作为统一导航信息架构中的占位分类；当前后端尚未接入相应控制能力。
- 现有项目仍会在构建时报告多个历史 Svelte 无障碍警告，本次未修改这些既有组件。
