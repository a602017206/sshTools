# 缺陷：上传任务没有文件名、大小和进度

## 背景

上传任务的历史记录标题全是「上传完成」，大小是 0 B。进行中一直是 0，也看不到字节进度。

## 范围

- 上传进度使用相对路径作为文件名，并累计已传字节和总大小。
- 完成、失败、取消不再把状态文案写成文件名。
- 文件管理在会话上提前订阅 `sftp:session-progress`，避免上传开始后才订阅而丢掉进行中的事件。
- 多个文件的历史标题保留各文件名；名字过长时显示第一个文件和数量。

## 修改文件

- `internal/service/upload_paths.go`
- `internal/service/upload_paths_test.go`
- `internal/service/sftp_service.go`
- `app.go`
- `frontend/src/components/FileManager.svelte`
- `frontend/test/fileDropUpload.test.js`
- `.github/workflows/release.yml`
- 本文档

## 验证

```bash
go test ./internal/service -count=1 -run 'TestUploadBatchLabelUsesFileNames|TestTransferPercentageUsesBytes|TestPartitionLocalUploadItems|TestRunFolderUpload'
cd frontend && node --test test/fileDropUpload.test.js
```

未在桌面端对真实 SFTP 上传做手工验证。

## 剩余风险

同一次上传里的多个文件仍共用一条任务。进行中显示当前文件名和整批字节进度，历史里用文件名列表作为标题。下载进度仍按传输 ID 事后订阅。
