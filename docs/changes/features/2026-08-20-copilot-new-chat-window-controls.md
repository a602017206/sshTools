# 功能：Copilot 新对话与窗口最大化

## 背景

Copilot 会无上限携带同一会话的全部历史，增加 token 消耗；应用缺少明确的最大化控制入口。

## 范围

- AI 面板新增“新对话”，仅清除该标签的 AI 历史和终端上下文。
- 请求只保留最近 12 轮、总计最多 12000 字符的对话历史。
- 应用默认最大化，并提供最大化/还原按钮。

## 修改文件

- `frontend/src/lib/copilotContext.js`
- `frontend/src/components/AIPanel.svelte`
- `frontend/src/App.svelte`
- `frontend/test/copilotContext.test.js`
- `main.go`

## 验证

- `cd frontend && node --test test/copilotContext.test.js`

## 剩余风险

窗口控件依赖 Wails Runtime；浏览器预览环境不会执行原生窗口操作。
