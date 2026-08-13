# 设计：AI Copilot（SQL / Shell）

## 背景

运维工作区已能执行 SQL 与 SSH 命令，但缺少自然语言到可执行产物的一层。用户需要在当前数据库或 SSH 会话里生成脚本，确认后再跑。

## 决策摘要

- 共用一套 Go Copilot 内核；数据库生成 SQL，SSH 生成命令。
- 侧边对话跟随当前标签；先填入，再点执行；危险写操作二次确认。
- 生成阶段可调只读工具（schema / 主机白名单探测），不自动执行用户产物。
- 模型接入为 OpenAI 兼容 API；Base URL 与模型名称均手填，原样发送，不猜测、不拉列表。
- API Key 进加密凭据库；本地 Ollama 只预留接口。
- 完整规格：`docs/superpowers/specs/2026-08-13-ai-copilot-sql-shell-design.md`。
