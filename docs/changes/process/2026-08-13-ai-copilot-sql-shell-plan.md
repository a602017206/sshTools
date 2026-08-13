# AI Copilot（SQL / Shell）实现计划

## 背景

设计规格已确认，需要一份按任务拆分的实现计划，供后续按 TDD 落地。

## 范围

- 新增实现计划文档。
- 更新规格状态，指向该计划。
- 不包含功能代码。

## 修改文件

- `docs/plans/2026-08-13-ai-copilot-sql-shell.md`
- `docs/superpowers/specs/2026-08-13-ai-copilot-sql-shell-design.md`
- `docs/changes/process/2026-08-13-ai-copilot-sql-shell-plan.md`

## 验证

- 计划覆盖规格中的入口、设置、只读工具、填入/执行、危险确认、隐私与测试。
- 任务含明确文件路径、接口名和验证命令。
- 正文为中文。

## 剩余风险

- 各家 OpenAI 兼容 tool call 细节可能不同，计划用假 HTTP 覆盖协议，真模型依赖手工回归。
- 前端 Wails 绑定需在 `wails dev` 时再生，计划中的 `go test` 不能代替 GUI 验证。
