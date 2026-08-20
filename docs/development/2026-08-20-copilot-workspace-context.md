# 开发记录：Copilot 工作目录与会话上下文

## 实现内容

- `internal/service/copilot/` 增加 `list_working_directory` 工具和工作目录命令构造函数。命令只列目录，目录经单引号安全转义。
- `SessionManager.ExecuteCommand` 对本地会话返回明确错误，避免访问空 SSH session。
- 本地 `ManagedSession` 创建时记录 `os.Getwd()`，使 `fillCopilotRequest` 可复用现有的当前目录读取链路。
- 前端增加 `copilotContext` 辅助模块和按会话终端 tail 缓冲；SSH 与本地输出都会记录，AIPanel 发话时传入对应 tail。
- 前端不再跳过本地输出中的 OSC 7 工作目录更新。

## 验证

- 新增 Go 回归测试：目录命令的转义与非法路径拒绝、目录工具使用请求工作目录、本地会话拒绝 SSH 独立命令、本地初始工作目录。
- 新增 Node 回归测试：终端 tail 按会话隔离、超长时保留最新文本。
- 完整验证结果见本次缺陷变更记录。

## 剩余风险

- 当前目录列举通过独立 SSH 会话执行；若服务端的 SSH 非交互命令策略与交互 Shell 不同，目录读取可能被服务端拒绝。
- 终端输出 tail 仍可能包含无法由当前规则识别的敏感文本；发送前会执行既有脱敏，后续可增加更严格的输出过滤。
