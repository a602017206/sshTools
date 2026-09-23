# 变更：Wails 绑定生成报 Not found: time.Time

## 背景

`wails generate` / `wails dev` 在生成绑定时两次打印 `Not found: time.Time`。Wails 把 `time.Time` 当成嵌套结构体去解析，而 `KnownStructs` 里没有这个标准库类型。

## 范围

- `CommandHistoryEntry.LastUsed`、`SessionLogInfo.ModTime` 增加 `ts_type:"string"`，与已有的 `ssh.FileInfo.ModTime` 一致。
- 同步 `frontend/wailsjs/go/models.ts` 把这两处从 `any` 改成 `string`。

## 修改文件

- `internal/service/command_history_service.go`
- `internal/service/session_log_service.go`
- `frontend/wailsjs/go/models.ts`
- 本文档

## 验证

```bash
go test ./internal/service -count=1 -run 'CommandHistory|SessionLog'
```

请再跑一次 `wails dev` 或绑定生成，日志里不应再出现 `Not found: time.Time`。

## 剩余风险

- Wails 运行时仍会把 `time.Time` 序列化成 RFC3339 字符串，前端 `new Date(raw)` 的展示逻辑保持兼容。
