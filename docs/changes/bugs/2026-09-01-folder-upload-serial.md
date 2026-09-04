# 修复文件夹上传拖垮 SSH 连接

## 背景

上传文件夹时应用卡死，同一服务器用其他工具也无法再连。SFTP 与终端共用一条 SSH 连接，文件夹实现按文件数无上限起 goroutine。

## 范围

- 目录树改为一次传输、后台串行 mkdir 与上传
- 进度按文件序号汇总，不再为每个文件开事件订阅风暴
- 上传流式拷贝时释放 SFTP 互斥锁，避免界面列举目录死锁

不改冲突对话框交互。不拆独立 SFTP 连接。

## 修改文件

- `internal/service/folder_upload.go`（新建）
- `internal/service/folder_upload_test.go`（新建）
- `internal/service/sftp_service.go`
- `internal/ssh/sftp.go`
- `frontend/test/fileDropUpload.test.js`
- `docs/designs/2026-09-01-folder-upload-serial.md`
- `docs/changes/bugs/2026-09-01-folder-upload-serial.md`（本文）

## 验证

- `go test -count=1 ./internal/service -run 'TestPartitionLocalUploadItems|TestRunFolderUpload'`
- `cd frontend && node --test test/fileDropUpload.test.js`
- 修复前 `runFolderUpload` 不存在，测试编译失败

无法在真实远端用超大目录做手工压测。

## 剩余风险

- 冲突阶段仍可能对已存在的同名目录逐层 `ListFiles`
- 超大本地树展开仍走一次 Wails JSON
- 单通道串行大流量仍可能让同一条 SSH 上的终端短暂卡顿
