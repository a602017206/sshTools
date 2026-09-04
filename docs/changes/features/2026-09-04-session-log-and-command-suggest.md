# 变更：会话日志记录与常用命令提示

## 背景

SSH 终端此前仅有实时 PTY 输入输出，缺少会话落盘与按连接的命令统计/输入提示。本期交付会话日志（自动记录、搜索导出、保留策略、敏感过滤）、常用命令提示，并将文件管理器目录跟踪默认改为开启；同步更新 GitHub Release 发布说明。

## 范围

- **后端**：`SessionLogService`（`~/.ahasshtools/session-logs/`）、`CommandHistoryService`（`~/.ahasshtools/commands/`）；`ConnectSSH` 输出回调写入日志（失败不阻断 SSH）；`RedactSessionLog` 写入前脱敏；启动时按保留天数清理过期日志
- **配置**：`AppSettings` 扩展会话日志与命令提示开关/保留天数/条数上限；`DefaultFileManagerSettings.DirectoryTracking` 默认 `true`
- **前端**：终端行缓冲与命令建议浮层（Tab/点击填入、回车执行）；`SessionLogPanel` 列表/搜索/导出；全局设置分区；文件管理器 `directory_tracking` fallback 改为 `?? true`
- **发布**：`.github/workflows/release.yml` 新增功能说明

## 修改文件

- `internal/config/config.go` — 默认设置与 FM 目录跟踪
- `internal/service/session_log*.go`、`command_history*.go`、`session_log_redact.go`
- `app.go` — 服务装配、Wails API、输出挂钩
- `frontend/src/lib/commandLineBuffer.js`、`commandSuggest.js`
- `frontend/src/components/Terminal.svelte`、`SessionLogPanel.svelte`、`GlobalSettingsDialog.svelte`、`SessionToolDock.svelte`、`TerminalPanel.svelte`
- `frontend/src/components/FileManager.svelte`、`frontend/src/stores.js` — 目录跟踪 fallback
- `.github/workflows/release.yml`
- `docs/designs/`、`docs/plans/`、`docs/superpowers/` 相关设计/计划文档

## 验证

```bash
go test ./internal/config ./internal/service -count=1
cd frontend && node --test test/commandLineBuffer.test.js test/commandSuggest.test.js
```

## 剩余风险

- `SessionLogService` 长时间多会话可能持有打开的日志文件句柄直至进程退出或删除日志
- 脱敏规则基于正则启发式，无法覆盖所有敏感信息格式
- 命令建议依赖前端回车时提取当前行，复杂 shell 语法（多行、管道组合）可能统计不完整
- 已持久化为 `directory_tracking: false` 的连接保持用户选择，仅新连接或未保存过该字段的连接受新默认影响
