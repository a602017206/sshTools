# 修复上传历史和开发工具占用工作区的问题

## 背景

上传历史与开发工具虽然带有遮罩样式，但未复用全局设置的对话框门户，部分场景会作为主界面布局节点渲染，导致工作区被向下挤压。

## 范围

将上传任务和开发工具统一改为 `Dialog` 门户组件显示，保持上传任务状态、取消、清理历史和开发工具能力不变。

## 修改文件

- `frontend/src/components/DevToolsPanel.svelte`
- `frontend/src/components/UploadTaskDialog.svelte`
- `frontend/src/App.svelte`
- `frontend/test/overlayDialogStructure.test.js`
- `.github/workflows/release.yml`

## 验证

- 运行 `cd frontend && node --test test/overlayDialogStructure.test.js`。
- 运行前端 Vite 生产打包。

## 剩余风险

完整构建最后的 JDBC agent 暂存步骤受本机 Gradle 分发锁文件权限影响；前端 Vite 打包已成功完成。
