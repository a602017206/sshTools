# 文件管理在线编辑文本和配置

## 背景

配置和脚本以前只能下载到本地再打开。小改动需要来回传输，也不方便直接保存回服务器。

## 范围

- 右键「在线编辑」打开弹窗编辑器，带语法高亮
- 按文件类型和内容格式化，格式化可以撤销
- Ctrl+S / Cmd+S 把内容写回原文件
- 仅限 1MB 以内的 UTF-8 文本；已知二进制文件不提供该操作

## 修改文件

- `internal/ssh/text_file.go`
- `internal/ssh/text_file_test.go`
- `internal/service/sftp_service.go`
- `app.go`
- `frontend/src/lib/remoteTextFile.js`
- `frontend/src/lib/remoteTextFormat.js`
- `frontend/src/lib/fileManagerContextMenu.js`
- `frontend/src/components/RemoteTextEditorDialog.svelte`
- `frontend/src/components/FileManagerContextMenu.svelte`
- `frontend/src/components/FileManager.svelte`
- `frontend/test/remoteTextFile.test.js`
- `frontend/test/remoteTextFormat.test.js`
- `frontend/test/fileManagerContextMenu.test.js`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/models.ts`
- `frontend/package.json`
- `frontend/package-lock.json`
- `docs/designs/2026-09-23-remote-text-editor.md`
- `docs/development/2026-09-23-remote-text-editor.md`
- `docs/changes/features/2026-09-23-remote-text-editor.md`（本文）

## 验证

- `go test ./internal/ssh -count=1 -run TestValidateRemoteText`
- `cd frontend && node --test test/remoteTextFile.test.js test/remoteTextFormat.test.js test/fileManagerContextMenu.test.js`
- `npx vite build --outDir /tmp/sshtools-frontend-check`
- `go build -o /dev/null .`
- 未连接真实 SSH 做打开、格式化、Ctrl+S 保存的手工验证

## 剩余风险

- 服务器不支持 `posix-rename` 时会截断原文件再写入，写到一半失败可能留下不完整内容；临时文件会保留并在错误里给出路径
- XML 格式化按尖括号切分，属性值中的 `>` 或 CDATA 可能被拆开，可以用撤销恢复
- 未知扩展名会显示「在线编辑」，读到二进制或非 UTF-8 时由后端拒绝
