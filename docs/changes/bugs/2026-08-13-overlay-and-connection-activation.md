# 变更：上传历史、开发工具与连接激活交互

## 背景

上传历史以高侧栏形式占用工作区，开发工具弹层没有稳定居中，且单击连接行会直接建立会话，容易误操作。

## 范围

- 上传任务历史调整为居中弹窗。
- 开发工具调整为固定居中的独立弹窗。
- 连接行改为双击连接，保留明确的连接按钮和键盘 Enter。
- 修复克隆连接弹窗因状态更新顺序而未应用预填配置的问题，并保留已保存密码。
- 修复编辑不同连接时旧异步请求覆盖新连接表单的问题。
- 修复从 A 切换编辑 B 时首次打开仅显示空表单的问题。
- 同步更新自动发布说明。

## 修改文件

- `frontend/src/App.svelte`
- `frontend/src/components/DevToolsPanel.svelte`
- `frontend/src/components/AssetList.svelte`
- `frontend/src/lib/assetActivation.js`
- `frontend/src/lib/cloneDialogState.js`
- `frontend/src/lib/editConnectionLoadState.js`
- `.github/workflows/release.yml`

## 验证

- 执行 `node --test test/assetActivation.test.js`、`node --test test/cloneDialogState.test.js` 与 `node --test test/editConnectionLoadState.test.js`。
- 执行 `npm run build`。
- 手工确认：单击连接行不连接，双击连接；上传任务与开发工具均居中显示。

## 剩余风险

- 双击连接增加了一次确认操作；用户仍可点击连接图标或按 Enter 快速连接。
