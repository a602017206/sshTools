# SFTP 上传同名冲突

## 背景

上传与远程同名的文件会直接报错，文件夹也无法选择全部覆盖或逐个处理。

设计文档：`docs/designs/2026-08-31-sftp-upload-conflict.md`。

## 范围

- 上传前检测同名文件/文件夹并弹出处理对话框
- 文件：覆盖、重命名、跳过
- 文件夹：全部覆盖、逐个选择、重命名整棵树
- 覆盖时删除并重建远程文件，类型冲突先删后传

不包含自定义输入新文件名。

## 修改文件

- `frontend/src/lib/uploadConflict.js`（新建）
- `frontend/test/uploadConflict.test.js`（新建）
- `frontend/src/components/UploadConflictDialog.svelte`（新建）
- `frontend/src/components/FileManager.svelte`
- `frontend/test/fileDropUpload.test.js`
- `internal/service/upload_paths.go`
- `internal/service/sftp_service.go`
- `internal/ssh/sftp.go`
- `app.go`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/models.ts`
- `docs/designs/2026-08-31-sftp-upload-conflict.md`
- `docs/changes/features/2026-08-31-sftp-upload-conflict.md`（本文）

## 验证

- `cd frontend && node --test test/uploadConflict.test.js test/fileDropUpload.test.js`
- `go test ./internal/service -run 'TestExpandLocalUploadPaths|TestJoinRemoteUploadPath'`
- 手工：向已有同名文件的目录上传文件，选择覆盖/重命名；上传已存在的文件夹，分别试全部覆盖与逐个选择

## 剩余风险

- 本次无法在 Wails 窗口内自动点选验证
- 超大目录列出嵌套冲突可能较慢
- 覆盖类型冲突会删除远程原对象
