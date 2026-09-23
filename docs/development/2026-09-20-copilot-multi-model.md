# 实现：Copilot 多模型管理

## 做了什么

- 设置改为服务商嵌套模型：DeepSeek / OpenAI / Ollama 预设加自定义，显示名可改。
- 对话输入框旁选择模型，并持久化 `copilot_active_model_id`。
- 每个服务商单独存 API Key；旧的单模型配置加载时迁成一个自定义服务商。
- Ollama 允许空密钥。删除全部服务商时会清掉旧字段，避免重启后又被迁回来。

## 关键实现

- `internal/config/copilot.go`：服务商结构、旧配置迁移、Wails 数组解析。
- `internal/service/copilot/models.go`：按模型档案解析 Base URL / 模型 ID / 密钥策略。
- `app.go`：`CopilotChat` 按选中模型建 Provider；新增 `Has/Set/ClearCopilotProviderAPIKey`。
- 前端：`copilotModels.js`、`CopilotSettingsSection.svelte`、`AIPanel` 选择器。

## 验证

```bash
go test ./internal/service/copilot ./internal/config . -count=1 -run 'Copilot|SetCopilot|ClearCopilot|MigrateCopilot|UpdateSettingsPersistsCopilot|UpdateSettingsClearsProviders'
cd frontend && node --test test/copilotModels.test.js test/copilotContext.test.js test/appearanceDefaults.test.js
```

## 未做

- 未拉 `/v1/models`。
- 未接 Anthropic 等非兼容协议。
- Flutter 客户端未改。
- 桌面应用未在本次用 `wails dev` 做手工点选。
