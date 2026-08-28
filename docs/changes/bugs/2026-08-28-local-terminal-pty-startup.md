# 修复：macOS 本地终端无法创建

## 背景

创建本地终端时，界面短暂出现后立即关闭。后端创建 PTY 的过程中启动 `/bin/zsh` 失败，并返回 `operation not permitted`。

## 范围

仅修复 macOS/Linux 本地终端的 PTY 启动参数冲突；不修改 SSH 连接、终端配置文件或用户的 Shell 启动脚本。

## 修改文件

- `internal/ssh/local_session_unix.go`：移除与 `creack/pty` 的 `Setsid`、`Setctty` 初始化冲突的 `Setpgid` 配置；关闭时仍按 PTY 创建的进程组发送终止信号。
- `internal/ssh/local_session_unix_test.go`：新增本地 Shell 启动后仍保持可用的回归测试。

## 验证

- `go test ./internal/ssh -run TestLocalShellStaysAvailableAfterStartup -count=1 -v`
- `go test ./internal/ssh -count=1`

## 剩余风险

测试依赖 macOS/Linux 能创建 PTY；Windows 使用单独的本地 Shell 实现，不受本次改动影响。
