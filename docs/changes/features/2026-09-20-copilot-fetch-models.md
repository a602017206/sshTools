# 变更：按服务商获取模型列表

## 背景

设置里配置服务商后仍需手填模型 ID。cc-switch 已能用 API Key 请求 OpenAI 兼容的 `/v1/models`。用户希望 Copilot 也按服务商拉取模型列表。

## 范围

- 新增 `GET /v1/models` 拉取，设置页「获取模型」后按需加入。
- 合并策略：保留已有自定义名称，不整表替换。
- 不改对话调用协议，不接非 OpenAI 兼容接口，不改 Flutter。

## 修改文件

后端：

- 新增：`internal/service/copilot/list_models.go`、`internal/service/copilot/list_models_test.go`
- 修改：`internal/service/copilot/provider_openai.go`、`app.go`

前端：

- 修改：`frontend/src/lib/copilotModels.js`、`frontend/src/components/CopilotSettingsSection.svelte`、`frontend/test/copilotModels.test.js`、`frontend/wailsjs/go/main/App.js`、`frontend/wailsjs/go/main/App.d.ts`、`frontend/wailsjs/go/models.ts`

文档：

- 新增：`docs/designs/2026-09-20-copilot-fetch-models.md`、`docs/development/2026-09-20-copilot-fetch-models.md`、本文档

## 验证

```bash
go test ./internal/service/copilot -count=1 -run 'ListModels|ModelsURL|ResolveListModels|OpenAICompatible'
cd frontend && node --test test/copilotModels.test.js
```

上述命令已执行。未在桌面端对真实服务商做手工点选。

## 剩余风险

- OpenAI 等接口会返回非对话模型，用户若点「加入全部」可能一次加入很多条，需自行删除。
- Wails 绑定为手工同步，下次 `wails dev` 会再生成一遍。
- 部分兼容网关若不提供 `/v1/models`，仍需手填。
