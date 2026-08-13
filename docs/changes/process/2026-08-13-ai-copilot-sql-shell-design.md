# AI Copilot（SQL / Shell）设计规格

## 背景

用户确认要在数据库工作区和 SSH 终端中，用自然语言生成 SQL / 命令，填入后手动执行。brainstorming 已选定「生成阶段只读工具调用」方案，需要把规格写入文档后再进入实现计划。

## 范围

- 新增设计规格与设计摘要。
- 不包含实现代码、前端组件或 Provider 接入。

## 修改文件

- `docs/superpowers/specs/2026-08-13-ai-copilot-sql-shell-design.md`
- `docs/designs/2026-08-13-ai-copilot-sql-shell.md`
- `docs/changes/process/2026-08-13-ai-copilot-sql-shell-design.md`

## 验证

- 规格覆盖架构、数据流、设置（手填模型名）、安全、测试与非目标。
- 无待填占位、无互斥选择、第一版范围可单独做实现计划。
- 正文为中文。

## 剩余风险

- 各家 OpenAI 兼容接口的 tool call 细节可能略有差异，实现时需用假客户端覆盖协议，真模型以手工回归验证。
- SSH 白名单探测能否覆盖常见发行版，要等实现时用真实会话确认。
