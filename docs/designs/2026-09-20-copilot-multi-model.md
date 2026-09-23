# 设计：Copilot 多模型管理

## 背景

AI Copilot 目前只能配置一套 OpenAI 兼容接口：一个 Base URL、一个模型名、一把 API Key。对话时后端强制使用该模型，无法在 DeepSeek / OpenAI / 本地 Ollama 之间切换，也不能给模型起显示名。

用户需要接近 Cursor / Trae 的模型管理：先配服务商，再在服务商下挂多个模型，使用时在输入框旁选择，并记住上次选用的模型。

## 决策

- 数据按 **服务商嵌套模型** 存储，不做成扁平模型表。
- 对话输入框旁下拉选择，**全局记住** `copilot_active_model_id`。
- 添加服务商时提供 DeepSeek / OpenAI / Ollama 预设，并允许自定义；设置页可用「获取模型」拉取 OpenAI 兼容的 `/v1/models`，再按需加入，不整表覆盖已有自定义名称。
- 每个服务商一把密钥，存凭据库；`config.json` 不写 API Key。
- 旧的单模型配置在加载时迁成「一个自定义服务商 + 一条模型」。
- Ollama 允许空密钥；其余服务商必须有密钥。
- 协议仍只走 OpenAI 兼容 Chat Completions。Flutter 客户端本期不改。

## 数据模型

`config.json` 的 `settings`：

```json
{
  "copilot_providers": [
    {
      "id": "uuid",
      "kind": "deepseek",
      "name": "DeepSeek",
      "base_url": "https://api.deepseek.com/v1",
      "models": [
        { "id": "uuid", "name": "DeepSeek Chat", "model_id": "deepseek-chat" }
      ]
    }
  ],
  "copilot_active_model_id": "uuid",
  "copilot_max_tool_rounds": 4,
  "copilot_max_tool_result_chars": 8000
}
```

`kind` 取值：`deepseek` / `openai` / `ollama` / `custom`。

兼容字段 `copilot_base_url` / `copilot_model` 在保存服务商时回写成当前选用模型的镜像，便于旧配置与局部更新共存。`copilot_provider` 仍为 `openai_compatible`。

密钥键：

- 新：`copilot:provider:{providerID}:api_key`
- 旧：`copilot:api_key`。若加载后恰好只有一个服务商且该服务商还没有密钥，则把旧密钥拷过去。

## 解析与调用

`CopilotChat` 不再无条件覆盖 `req.Model`。

1. 优先用请求里的 `ModelProfileID`，否则用 `copilot_active_model_id`。
2. 在 `copilot_providers` 中找到对应模型与服务商。
3. 读该服务商密钥，缺失时回退旧全局密钥。
4. 校验 Base URL、模型 ID；非 Ollama 还要校验密钥。
5. 把真实 `model_id` 发给现有 OpenAI 兼容 Provider。

找不到所选模型时，回退到该配置里的第一条模型；若服务商列表为空，再回退旧的 `copilot_base_url` + `copilot_model`。

## 界面

设置「AI Copilot」：

- 添加入口：DeepSeek / OpenAI / Ollama / 自定义。
- 每个服务商卡片：显示名、Base URL、API Key、模型列表（显示名 + 模型 ID）、添加/删除模型、删除服务商。
- 全局仍保留最大工具轮次与工具结果上限。

对话面板：

- 无模型时提示去设置。
- 有模型时在输入框上方显示选择器，选项为「显示名 · 服务商名」。
- 切换立即写入 `copilot_active_model_id`，历史消息保留，下一轮用新模型。
- 当前服务商缺密钥且不是 Ollama 时，禁止发送并提示去设置。

## 错误处理

- 未配置模型：`请先在设置中添加模型`
- 缺 Base URL / 模型 ID：`请先在设置中填写 Base URL 和模型`
- 非 Ollama 缺密钥：沿用 `请先在设置中填写 Base URL 和 API Key`
- 删除当前选用模型后，自动改选剩余第一条。

## 测试

- 旧配置迁成一个自定义服务商 + 一条模型，且不把 API Key 写入 `config.json`。
- `ResolveEndpoint`：按 profile ID 选中、ID 失效时回退第一条、无服务商时走旧字段。
- Ollama 允许空密钥，其他 kind 拒绝空密钥。
- 前端：预设生成、模型列表摊平、当前模型 ID 解析、payload 带 `ModelProfileID`。
