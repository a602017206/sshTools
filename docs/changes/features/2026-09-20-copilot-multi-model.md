# 变更：Copilot 多模型管理

## 背景

AI 模块原来只能配一个 Base URL、一个模型和一把密钥，对话时无法切换。需要按 Cursor / Trae 的方式管理多个服务商和模型，使用时选择，并支持自定义显示名。

## 范围

- 配置从单模型改为服务商 → 模型两层；设置页可添加 DeepSeek / OpenAI / Ollama / 自定义。
- 对话面板增加模型下拉，全局记住上次选用的模型。
- API Key 按服务商加密存储；旧配置自动迁移。
- 不拉取模型列表，不新增非 OpenAI 兼容协议，不改 Flutter。

## 修改文件

后端：

- 新增：`internal/config/copilot.go`、`internal/service/copilot/models.go`、`internal/service/copilot/models_test.go`
- 修改：`internal/config/config.go`、`internal/config/config_test.go`、`internal/service/copilot/service.go`、`app.go`、`app_copilot_key_test.go`

前端：

- 新增：`frontend/src/lib/copilotModels.js`、`frontend/src/components/CopilotSettingsSection.svelte`、`frontend/test/copilotModels.test.js`
- 修改：`frontend/src/components/GlobalSettingsDialog.svelte`、`frontend/src/components/AIPanel.svelte`、`frontend/src/App.svelte`、`frontend/src/lib/copilotContext.js`、`frontend/src/stores/copilot.js`、`frontend/src/settings/appearance.js`、`frontend/test/copilotContext.test.js`、`frontend/wailsjs/go/main/App.js`、`frontend/wailsjs/go/main/App.d.ts`、`frontend/wailsjs/go/models.ts`

文档：

- 新增：`docs/designs/2026-09-20-copilot-multi-model.md`、`docs/development/2026-09-20-copilot-multi-model.md`、本文档

## 验证

```bash
go test ./internal/service/copilot ./internal/config . -count=1 -run 'Copilot|SetCopilot|ClearCopilot|MigrateCopilot|UpdateSettingsPersistsCopilot|UpdateSettingsClearsProviders'
cd frontend && node --test test/copilotModels.test.js test/copilotContext.test.js test/appearanceDefaults.test.js
```

上述命令已通过。未在 `wails dev` 中做桌面端手工点选。

## 剩余风险

- Wails 绑定文件为手工同步，下次 `wails dev` 会再生成一遍。
- 旧密钥只在「恰好一个服务商」时拷到新键；用户若先加第二个服务商再开对话，第二个服务商不会自动继承旧密钥。
- 对话中途换模型会继续同一段历史，不同模型对工具调用格式的兼容性取决于服务商。
