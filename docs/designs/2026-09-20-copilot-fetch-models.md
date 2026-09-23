# 设计：按服务商获取模型列表

## 背景

多模型管理落地后，服务商下的模型仍需手填 ID。用户希望像 cc-switch 一样，用已配置的 Base URL 和 API Key 去拉 OpenAI 兼容的 `GET /v1/models`，再把返回的模型加进当前服务商。

此前多模型设计明确「不拉取 `/v1/models`」。本期改掉这条约束，仅补获取能力，不改对话协议，也不接 Anthropic / Codex OAuth。

## 决策

- 只走 OpenAI 兼容 `GET {base}/v1/models`。Base URL 没有 `/v1` 时自动补上；若误填了 `/v1/chat/completions` 则剥掉后再拼 `/models`。
- 密钥优先用设置页里尚未保存的草稿，其次用该服务商已存密钥。Ollama 允许空密钥；其他 kind 在请求前要求有密钥。
- 拉取结果**不覆盖**现有行：已有 `model_id` 保留自定义显示名；空行优先填入新 ID；其余按需追加。
- 界面提供「获取模型」，再「加入选中 / 加入全部」。OpenAI 一类接口可能返回上百个模型，默认不全部灌进表单。
- 错误信息脱敏，不回显 API Key。超时 15 秒。
- 仍支持手填模型。Flutter 本期不改。

## 接口

`ListCopilotProviderModels(baseURL, apiKey, providerID, providerKind)`

返回 `{id, name}` 数组。`name` 目前等于接口里的 `id`，用户加入后可再改显示名。

## 界面

每个服务商卡片的模型区：

- 「获取模型」：请求该服务商列表。
- 下拉列出尚未加入的 ID，「加入选中」或「加入全部」。
- 原有「+ 添加模型」保留。
