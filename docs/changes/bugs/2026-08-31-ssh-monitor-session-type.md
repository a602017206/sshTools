# 修复性能页监控数据采集

## 背景

性能页绑定了 SSH 会话后仍一直显示「实时数据获取中…」，CPU/内存/磁盘为 0，系统信息为 Unknown。文件管理仍可用，因为 SFTP 不走 `ExecuteCommand`。

设计文档：`docs/designs/2026-08-31-ssh-monitor-session-type.md`。

## 范围

- 创建远程会话时写入 `Type: ssh`
- 独立命令允许「非本地且存在 SSH Session」的会话，避免旧连接因类型零值被拒
- 本地终端仍禁止远程命令

## 修改文件

- `internal/ssh/manager.go`
- `internal/ssh/manager_test.go`
- `docs/designs/2026-08-31-ssh-monitor-session-type.md`
- `docs/changes/bugs/2026-08-31-ssh-monitor-session-type.md`（本文）

## 验证

- `go test ./internal/ssh -count=1`
- 手工：SSH 连上后打开「性能」，应出现非零 CPU/内存或真实系统信息，而不是一直 Unknown

## 剩余风险

- 采集命令本身仍可能因目标机无 `top`/`free` 等工具而部分为空
- 本次无法在 Wails 窗口内连真实主机验证，需本地确认性能页刷新
