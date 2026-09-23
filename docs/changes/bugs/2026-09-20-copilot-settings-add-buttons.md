# 变更：AI 设置添加服务商/模型按钮不刷新

## 背景

设置页点「+ DeepSeek / OpenAI / Ollama / 自定义」或「+ 添加模型」当时看不到新卡片，保存后再打开才出现。数据已经写进 draft，只是列表没有重绘。

## 范围

- `CopilotSettingsSection` 的 `{#each}` 改为订阅响应式的 `providerList`（依赖 `draft.copilot_providers`）。
- 添加/删除只通过重新赋值 `draft` 提交，避免模板里调用 `providers()` 导致 Svelte 跟踪不到 `draft`。

## 修改文件

- `frontend/src/components/CopilotSettingsSection.svelte`
- 本文档

## 验证

无法在浏览器里打开 Wails 设置弹层。请在 `wails dev` 中：打开全局设置 → AI Copilot → 点添加服务商、添加模型，卡片应立刻出现，无需先保存。

## 剩余风险

- 若其它设置分段也用「函数调用 + 内部读 props」做 `{#each}`，会有同类问题。
