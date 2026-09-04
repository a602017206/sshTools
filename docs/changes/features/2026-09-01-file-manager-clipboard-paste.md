# 文件管理支持粘贴本地文件和文件夹

## 背景

文件管理只能通过按钮或拖放上传。用户需要 Ctrl+V / Cmd+V 把系统剪贴板中的文件或文件夹传到当前目录；选中文件夹时传到该文件夹内。必须串行上传，避免再次打满 SSH。

## 范围

- 原生读取系统剪贴板中的本地路径
- 快捷键与右键粘贴：有本地路径则上传，否则走原来的远程粘贴
- 恰好选中一个文件夹时作为上传目标
- 同一会话上传加锁，同时只跑一棵目录树

## 修改文件

- `internal/service/clipboard_files.go` 及平台实现
- `internal/service/clipboard_files_test.go`
- `internal/service/upload_gate_test.go`
- `internal/service/sftp_service.go`
- `app.go`
- `frontend/src/lib/uploadDestination.js`
- `frontend/test/uploadDestination.test.js`
- `frontend/src/components/FileManager.svelte`
- `frontend/src/lib/fileManagerContextMenu.js`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `docs/designs/2026-09-01-file-manager-clipboard-paste.md`
- `docs/changes/features/2026-09-01-file-manager-clipboard-paste.md`（本文）

## 验证

- `go test -count=1 ./internal/service -run 'TestParseFileURIList|TestNormalizeClipboardFilePaths|TestSessionUploadLock|TestRunFolderUpload'`
- `cd frontend && node --test test/uploadDestination.test.js test/fileDropUpload.test.js test/fileManagerContextMenu.test.js`
- 无法在此环境用真实 Finder/资源管理器复制后点选验证

## 剩余风险

- 部分 Linux 桌面未安装 `wl-paste`/`xclip` 时读不到文件剪贴板
- 超大目录展开仍在 Wails 调用中完成，界面可能短暂停顿，但 SSH 保持串行
