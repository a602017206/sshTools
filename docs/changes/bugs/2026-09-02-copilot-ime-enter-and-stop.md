# 变更：AI 输入框 IME 回车误发送与可停止生成

## 背景

中文输入法组字时按回车确认/切英文，会被当成发送；生成过程中也无法终止。

## 范围

- Enter 在 `isComposing` / keyCode 229 时不发送
- 暴露 `CopilotCancel`，面板提供「停止」按钮

## 修改文件

- `frontend/src/lib/composerKeys.js`
- `frontend/test/composerKeys.test.js`
- `frontend/src/components/AIPanel.svelte`
- `app.go`
- `frontend/wailsjs/go/main/App.js`、`App.d.ts`

## 验证

- `node --test test/composerKeys.test.js`
- `go test ./internal/service/copilot/ -run Cancel`

## 剩余风险

- 底层 HTTP 客户端若未及时响应 cancel，停止后服务端请求可能仍跑完但前端会丢弃结果
