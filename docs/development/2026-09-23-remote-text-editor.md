# 实现：远程文本在线编辑

## 做了什么

- 文件管理右键增加「在线编辑」。已知二进制扩展名和超过 1MB 的文件不可用。
- 弹窗使用 CodeMirror 6，按文件名和内容选择高亮。
- 「格式化」把 JSON、YAML、XML/HTML、SQL 整理成一次编辑器修改，工具栏「撤销」和 Ctrl+Z 可以退回。其余文本只去掉行尾空白。
- Ctrl+S / Cmd+S 调用 `SaveRemoteTextFile` 写回原路径。关闭时如果还有未保存修改，会先确认。

## 关键实现

- `internal/ssh/text_file.go`：UTF-8 和大小校验；普通文件先写临时文件，再用 `posix-rename` 替换；符号链接直接写目标。
- `ReadRemoteTextFile` / `SaveRemoteTextFile` 从 `app.go` 导出。
- `frontend/src/lib/remoteTextFormat.js`：语言判断和格式化。YAML 用 `yaml` 的 `parseDocument`，保留注释。
- `RemoteTextEditorDialog.svelte`：编辑器、格式化、撤销和保存。文件管理在第一次打开时才加载这个组件，避免把 CodeMirror 打进首屏。打开时记录 LF 或 CRLF，保存时按原来的换行写回。

## 验证

```bash
go test ./internal/ssh -count=1 -run TestValidateRemoteText
cd frontend && node --test test/remoteTextFile.test.js test/remoteTextFormat.test.js test/fileManagerContextMenu.test.js
npx vite build --outDir /tmp/sshtools-frontend-check
go build -o /dev/null .
```

没有连上真实 SSH 做打开、格式化、保存的手工点选。

## 未做

- 双击仍然是下载到本地打开。
- 没有编辑锁，也没有和服务器上的新版本做合并。
- 大于 1MB、二进制和非 UTF-8 不能在线编辑。
