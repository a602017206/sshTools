# 移动端 SFTP 文件管理

## 背景

终端页打通后，按与 PC 端的差距清单继续补齐：SFTP 是 PC 端核心能力之一。此前 `internal/api` 只注册了连接与会话路由，SFTP 服务（`SFTPService`）虽已装配但没有 HTTP 出口，移动端用不上。本次给它补齐 API 并实现移动端文件管理页。

## 范围

包含：目录浏览与导航、新建文件夹、重命名、删除（文件与递归删目录）、单文件上传/下载、远端文本查看与编辑保存。不包含：文件夹上传/下载、传输进度条与可取消传输（当前单文件同步传输）、下载文件的系统分享/导出（先落应用文档目录并提示路径）、断点续传。

## 修改文件

- `internal/api/handlers/sftp.go`（新增）：在已建立的 SSH 会话上挂 SFTP 接口 `GET list/pwd/stat/read/download`、`POST mkdir/rename/upload`、`PUT write`、`DELETE`。上传先落沙盒临时文件（文件名经 `filepath.Base` 防穿越）再同步上传；下载先同步落临时目录再以附件回传并清理。
- `internal/api/server.go`：会话路由下注册 `/:id/sftp` 路由组，`TODO: SFTP` 注释移除。
- `internal/service/sftp_sync.go`（新增）：同步版单文件上传/下载（`UploadFileSync`/`DownloadFileSync`），供移动端在请求内完成传输；桌面原有的异步进度回调模型不受影响。
- `flutter_ui/lib/data/models/sftp_file.dart`、`data/repositories/sftp_repository.dart`（新增）：文件条目模型与仓库（列表排序目录优先、建夹、重命名、删除、读/写文本、multipart 上传、下载落盘）。
- `flutter_ui/lib/presentation/providers/api_providers.dart`：注册 SFTP 仓库。
- `flutter_ui/lib/presentation/screens/sftp/sftp_screen.dart`（新增）：文件管理页。自己开一条专用 SSH 会话（凭据回退与终端一致），退出即关；路径栏 + 上级/新建/上传/刷新，下拉刷新，条目操作面板（下载、查看/编辑文本、重命名、删除），文本编辑页保存即回写远端。首次进入从条目反推当前目录、空目录时用 pwd 接口定位。
- `flutter_ui/lib/presentation/screens/connections/connection_list_screen.dart`：卡片增加「文件」入口，未保存凭据时先弹窗索要（与终端同口径）。
- `flutter_ui/pubspec.yaml`：重新引入 `file_picker`——旧 6.x 已删，本次用 ^10.0.0 新版（旧版使用已移除的嵌入 API 才被删；新版供上传选文件）。

## 验证

- 本机：`go build ./...`、`go vet ./internal/api/...` 通过（API 层编译与路由注册无误）；服务层同步方法是对既有 SFTP 客户端同步原语的薄封装。
- GitHub 流水线第 8 轮（[39404929768](https://github.com/a602017206/sshTools/actions/runs/39404929768)）：Android 与 iOS 双双成功，产物已更新。
- 仍需真机复测：进「文件」应列出家目录；新建/重命名/删除即时生效；上传一个小文件后远端可见；下载后按提示路径在本机文档目录；打开文本文件编辑保存后远端内容变更。

## 剩余风险

- 下载走整文件内存中转（Dio bytes），超大文件可能内存吃紧；后续应改流式落盘并接传输进度。
- 上传/下载为同步请求，长耗时依赖 HTTP 超时（当前客户端 30s），大文件可能超时失败，需后续异步化 + 进度 UI。
- SFTP 会话与终端会话各自独立连接同一服务器，服务器侧连接数翻倍；后续多标签时可考虑同一连接复用终端与文件页。
- 删除目录为递归删除，已加二次确认，但无回收站，真机操作需谨慎。
