# 修复：Copilot 无法识别当前目录脚本与会话上下文缺失

## 背景

用户要求重启当前目录下的服务时，Copilot 无法发现目录中的脚本；本地终端的当前目录和近期输出也没有进入模型请求，导致多轮交流缺少环境关联。

## 范围

仅修复 Shell Copilot 的只读环境发现与会话上下文传递。不改变最终命令执行流程，不开放生成阶段的任意命令执行，也不持久化对话。

## 修改文件

- `internal/service/copilot/probe.go`
- `internal/service/copilot/tools.go`
- `internal/service/copilot/service.go`
- `internal/service/copilot/probe_test.go`
- `internal/service/copilot/service_test.go`
- `internal/ssh/manager.go`
- `internal/ssh/manager_test.go`
- `frontend/src/lib/copilotContext.js`
- `frontend/src/stores/copilot.js`
- `frontend/src/components/AIPanel.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/test/copilotContext.test.js`

## 验证

执行 Go 的 Copilot 和 SSH 回归测试，以及前端 Copilot 上下文测试和构建验证。测试命令与结果以本次提交前的实际输出为准。

## 剩余风险

终端是否提供 OSC 7 取决于 Shell 配置；本地会话在没有 OSC 7 时保持启动目录。目录工具只显示文件名与元数据，不能理解脚本内部逻辑，用户仍需检查生成的命令后再执行。
