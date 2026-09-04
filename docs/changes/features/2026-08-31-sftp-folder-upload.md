# SFTP 文件夹上传

## 背景

文件管理器无法上传文件夹：选择器只能选文件，拖入目录也不会展开成远程目录树。

设计文档：`docs/designs/2026-08-31-sftp-folder-upload.md`。

## 范围

- 后端展开本地目录并 `MkdirAll` 后按相对路径上传
- 工具栏「上传文件夹」、右键「选择文件夹上传」
- 拖放文件夹复用 `UploadFiles`

不包含远程文件夹下载和符号链接跟随。

## 修改文件

- `internal/service/upload_paths.go`（新建）
- `internal/service/upload_paths_test.go`（新建）
- `internal/service/sftp_service.go`
- `internal/ssh/sftp.go`
- `app.go`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/src/components/FileManager.svelte`
- `frontend/src/components/FileManagerContextMenu.svelte`
- `frontend/src/lib/fileManagerContextMenu.js`
- `frontend/test/fileDropUpload.test.js`
- `docs/designs/2026-08-31-sftp-folder-upload.md`
- `docs/changes/features/2026-08-31-sftp-folder-upload.md`（本文）

## 验证

- `go test ./internal/service -run 'TestExpandLocalUploadPaths|TestJoinRemoteUploadPath' -v`
- `cd frontend && node --test test/fileDropUpload.test.js`
- 手工：工具栏上传文件夹、右键选择文件夹上传、从 Finder 拖入文件夹；确认远程出现同名目录、嵌套文件和空子目录

## 剩余风险

- 本次无法在 Wails 窗口内自动拖放验证
- 超大目录展开可能短暂卡住界面
- 符号链接会被跳过
