# 变更：SFTP 拖放上传

## 背景

右侧文件管理器仅支持通过选择器上传本地文件，桌面使用场景下无法直接从 Finder 等文件管理器拖入文件。

## 范围

- 启用 Wails 原生文件拖放能力。
- 将右侧远程文件管理区域作为拖放目标，拖入后上传到当前远程目录。
- 复用既有上传队列、进度和目录刷新逻辑。
- 仅允许已连接的远程 SSH 会话接收拖入文件。

## 修改文件

- `main.go`
- `frontend/src/components/FileManager.svelte`
- `frontend/src/lib/fileDropUpload.js`
- `frontend/test/fileDropUpload.test.js`

## 验证

- 执行 `node --test test/fileDropUpload.test.js`。
- 执行 `npm run build`。

## 剩余风险

- 浏览器预览不提供桌面文件的绝对路径，拖放上传仅在 Wails 桌面应用中可用；文件选择器上传不受影响。
