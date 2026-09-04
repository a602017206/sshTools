# 变更：会话日志与命令建议 App API

## 背景

会话日志与命令历史服务已完成，需要在 App 层装配并向前端暴露 Wails API，同时在 SSH 输出路径挂钩日志写入。

## 范围

- 在 `App` 中初始化 `SessionLogService` / `CommandHistoryService`（根目录 `~/.ahasshtools/session-logs` 与 `~/.ahasshtools/commands`）
- 导出绑定、列表、搜索、导出、删除、清理、命令记录与建议方法
- `ConnectSSH` 输出回调写入会话日志（失败不阻断 SSH）；`CloseSSH` 清理 session→connection 映射
- 启动时按保留天数执行一次过期清理

## 修改文件

- `app.go`

## 验证

- `go build -o /dev/null .`

## 剩余风险

- `SessionLogService` 暂无按 session 关闭 writer；长时间多会话可能持有打开的日志文件句柄直至进程退出或删除日志
- 前端尚未调用本批 API，端到端行为依赖后续前端任务
