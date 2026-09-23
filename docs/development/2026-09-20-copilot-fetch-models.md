# 实现：按服务商获取模型列表

## 做了什么

- 后端对 OpenAI 兼容接口发 `GET /v1/models`，解析 `data[].id`，跳过空值和重复 ID。
- 设置页每个服务商增加「获取模型」，可用草稿密钥或已存密钥；Ollama 可不填密钥。
- 拉取结果先展示为待加入列表，用户选择加入。已有模型的自定义名称不会被覆盖。

## 关键实现

- `internal/service/copilot/list_models.go`：URL 推导、密钥策略、`ListModels`。
- `app.go`：导出 `ListCopilotProviderModels`，15 秒超时。
- 前端：`mergeFetchedModels` / `availableFetchedModels`，`CopilotSettingsSection` 的获取与加入按钮。

## 验证

```bash
go test ./internal/service/copilot -count=1 -run 'ListModels|ModelsURL|ResolveListModels|OpenAICompatible'
cd frontend && node --test test/copilotModels.test.js
```

未在 `wails dev` 里对真实 DeepSeek / Ollama 接口做手工点选。

## 未做

- 未过滤 embeddings / whisper 等非对话模型。
- 未接 Anthropic 原生 `/v1/models` 或 Ollama `/api/tags`。
- Flutter 未改。
