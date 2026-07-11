# JDBC Agent 失败状态刷新

## 背景

真实桌面验收中，点击“重启 agent”后若启动失败，错误面板会显示 `AGENT_UNAVAILABLE`，但顶部状态仍停留在重启前的“已停止”，没有展示 supervisor 已记录的失败状态和最后错误。

## 范围

- Agent 重启操作无论成功或失败，结束前都重新读取 `GetJDBCAgentStatus`。
- 错误仍由现有 `runTask` 统一分类，顶部状态与最后错误改为使用重启后的 supervisor 快照。

## 修改文件

- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend/build/assets/index.js`

## 验证

- `cd frontend && npm run build`
- 启动 Wails 桌面应用，在不可用 Java 环境下点击“重启 agent”，确认顶部显示“启动失败”及最后错误，错误面板显示 `AGENT_UNAVAILABLE`。

## 剩余风险

- Agent 状态只在页面加载、刷新和重启操作结束时更新，尚未增加后台轮询。
