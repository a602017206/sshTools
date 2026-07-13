# JDBC 驱动卸载确认无响应

## 背景

驱动管理器的卸载按钮依赖浏览器原生 `window.confirm`。在 Wails 嵌入式 WebView 中，该原生对话框可能不显示，用户点击后无法继续确认或看到错误提示。

## 范围

将卸载操作改为应用内 `ConfirmDialog`，确认时使用打开对话框时锁定的 driver/profile 调用后端。后端仍会拒绝卸载被已保存连接或活动 JDBC 会话引用的 profile，并显示可操作错误。

## 修改文件

- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend/src/lib/jdbcDriverRemovalState.js`
- `frontend/test/jdbcDriverRemovalState.test.js`
- 本变更记录。

## 验证

执行 `node --test test/jdbcDriverRemovalState.test.js`，验证确认信息和空选择处理；执行前端构建，验证 Svelte 组件与 Wails 资源可编译。

## 剩余风险

本地尚未在正在运行的桌面应用中手工点击确认对话框。后端引用保护由 Go 单元测试覆盖。
