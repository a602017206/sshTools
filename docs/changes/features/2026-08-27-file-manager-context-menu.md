# 变更：文件管理右键菜单

## 背景

文件管理列表只有精简右键菜单，缺少刷新、上传、新建、改权限和剪切复制等常用操作。需要按常见 SSH 客户端的菜单结构补齐，并接通现有 SFTP 能力。

## 范围

- 文件行和空白处右键弹出分组菜单，含顶部图标和「更多」子菜单。
- 接通打开、刷新、收藏路径、下载、上传文件、重命名、删除、复制路径、新建文件夹/文件、修改权限、剪切/复制/粘贴。
- 后端增加空文件创建、远程文件复制和 chmod。
- 不包含在线编辑、文件夹上传、SCP、压缩和解压。

## 修改文件

- `frontend/src/lib/fileManagerContextMenu.js`
- `frontend/test/fileManagerContextMenu.test.js`
- `frontend/src/components/FileManagerContextMenu.svelte`
- `frontend/src/components/FileManager.svelte`
- `internal/ssh/sftp.go`
- `internal/ssh/sftp_mode_test.go`
- `internal/service/sftp_service.go`
- `app.go`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `docs/designs/2026-08-27-file-manager-context-menu.md`
- `docs/development/2026-08-27-file-manager-context-menu.md`

## 验证

- `cd frontend && node --test test/fileManagerContextMenu.test.js`
- `go test ./internal/ssh -run TestParseOctalFileMode`
- 手工：在 SSH 会话的文件管理中右键文件、文件夹和空白处，检查菜单项启用状态；新建文件夹/文件后列表刷新；剪切后到其他目录粘贴应移动成功。

## 剩余风险

- 复制目录仍不支持。
- 个别 SFTP 服务器可能拒绝独占创建或 chmod，失败信息会显示在面板错误区。
- 快捷键仅在文件管理面板聚焦时生效，避免打断终端输入。
