# App 退出清理子进程实现说明

## 实现要点

退出顺序固定为：

1. `databaseService.CloseAllSessions()`
2. `nativeDatabaseService.CloseAll()`
3. `sessionService.CloseAllSessions()`（含本地终端）
4. `jdbcAgentSupervisor.Close()` → `AgentProcessManager.Stop()`

Unix agent / 本地 shell：

- 启动时 `SysProcAttr.Setpgid = true`
- 停止时先对进程组发 `SIGTERM`，等待最多 3 秒，再 `SIGKILL` 并短等

## 说明

本次为缺陷修复，未单独新增设计文档；行为对齐既有 Wails `OnShutdown` 生命周期，不改变对外 API。
