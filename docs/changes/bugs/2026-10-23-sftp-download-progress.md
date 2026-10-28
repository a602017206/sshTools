# 变更：SFTP 下载缺少进度条

## 背景

文件管理器上传有进度面板，从服务器下载时无进度条。后端已按块回调进度并发送 `sftp:progress:{id}`，前端订阅后却把 `kind === 'download'` 直接丢弃。

## 范围

- 前端：`subscribeToTransfer` 对 upload/download 均写入 `uploadStore`。
- 后端：`DownloadFile` / `DownloadFiles` 改为复用 `emitSFTPProgress`（同时发 transfer 与 session 进度），与上传一致，减轻订阅稍晚丢首包的问题。
- 不改动终端剪贴板受保护区域。

## 修改文件

- `frontend/src/components/FileManager.svelte`
- `frontend/src/lib/transferProgress.js`
- `frontend/test/transferProgress.test.js`
- `frontend/src/components/UploadTaskDialog.svelte`（取消按钮文案改为「取消传输」）
- `app.go`
- `docs/changes/bugs/2026-10-23-sftp-download-progress.md`

## 验证

```bash
cd frontend && node --test test/transferProgress.test.js
```

手工：SSH 会话中下载较大文件，传输面板「进行中」应出现文件名与百分比；上传回归仍正常。

## 剩余风险

- 极快的小文件可能在 UI 订阅前已完成；session 级事件可覆盖大部分情况，极短传输仍可能一闪而过只进历史。
- 进度面板标题仍偏「上传任务」命名（`UploadTaskDialog`），功能已通用，后续可改名。
