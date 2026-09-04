# 变更：ES AI 回复解析漏掉 artifact 导致无填入/执行按钮

## 背景

模型常先贴一份裸 DSL JSON，再贴 `{"type":"native_query",...}`。旧解析从第一个 `{` 起读，把 DSL 当 artifact 失败，按钮不出现；包装 JSON 还整段显示在聊天里。

## 范围

- 扫描全部 `{`，优先识别带 type 的 artifact
- 原生库无 typed artifact 时，把含 `query` 的裸 JSON 提升为 `native_query`
- 展示文案去掉 artifact 信封；提示词要求 DSL 放进 content

## 验证

- `go test ./internal/service/copilot/`

## 剩余风险

- 旧历史消息仍无按钮，需重新提问
- `now-7d` 对 long+epoch_millis 为合法 ES date math，并非杂质
