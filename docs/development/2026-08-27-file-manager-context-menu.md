# 文件管理右键菜单实现说明

## 实现

前端用 `FileManagerContextMenu.svelte` 渲染菜单和「更多」子菜单，动作由 `FileManager.svelte` 统一处理。空白处右键打开同一菜单，但针对文件的项（重命名、删除、下载、复制）禁用。

后端在 SFTP 客户端增加 `CreateFile`、`CopyFile`、`ChmodFile`，经 `SFTPService` 和 `App` 导出给 Wails。

## 交互

- 文件夹双击仍进入目录；文件双击仍下载。
- `Enter` 打开目录或下载文件；`F2` 重命名；`Backspace`/`Delete` 删除；`⌘/Ctrl+R` 刷新。快捷键只在文件管理面板聚焦且没有对话框时生效。
- 收藏当前路径写入文件管理历史并持久化。

## 未做

在线编辑、文件夹上传、SCP、压缩和解压未实现，菜单中不展示，避免出现不可用项。
