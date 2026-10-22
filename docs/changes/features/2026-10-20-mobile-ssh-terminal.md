# 移动端 SSH 终端页

## 背景

真机复测反馈：测试连接已能通过、确认可连，但点连接仍进不了 SSH 控制台。此前移动端明确只做了连接管理，终端页未实现，卡片点击停在「测试连接」。本次补齐终端页，打通「点连接 → 开会话 → 收发终端数据」的完整链路。

关于前端形态一并澄清：移动端 `flutter_ui/` 是独立重写的 Flutter 界面，不复用桌面端 Svelte 前端；复用的是 Go 后端服务层（SSH/SFTP/配置/凭证）与 HTTP/WebSocket API，前端重写不等于后端重写。

## 范围

包含：SSH 终端页（渲染、输入、尺寸同步、断开/重连）、特殊按键条、会话凭据回退、列表进入终端的接线。不包含：多标签会话、SFTP、后台保活与断线自动重连、终端选区复制的专门适配（沿用 xterm 默认行为，待真机反馈再打磨）。

## 修改文件

- `flutter_ui/lib/presentation/screens/terminal/terminal_screen.dart`（新增）：xterm 终端页。先经 WebSocket `ensureConnected` 并订阅会话，再调开会话接口，避免登录横幅等早期输出丢失；输出经 `ssh:output` 渲染、断开经 `ssh:closed` 提示；键入与窗口尺寸分别走 send/resize 接口；顶部状态点（连接中/已连接/已断开）与重连按钮；退出页面关会话。特殊按键条提供 Esc、Tab、方向键、Home/End、PgUp/PgDn、常用符号与 Ctrl 粘滞键（点亮后下一个字母转控制字符，覆盖 ^C 组合键场景）；终端键盘类型设为 visiblePassword 以关掉 Android 输入联想对命令的改写。
- `flutter_ui/lib/presentation/screens/connections/connection_list_screen.dart`：点卡片或「连接」按钮进入终端；未保存凭据时先弹窗索要（开会话成功后服务端回存）；从终端返回后刷新列表状态。
- `flutter_ui/lib/data/repositories/session_repository.dart`：开会话支持只带 `connection_id`（服务端回退已保存密码）。
- `flutter_ui/lib/data/datasources/remote/websocket_client.dart`：`web_socket_channel` 升 3.x 并改用 `channel.ready` 确认握手完成；订阅改为「登记 + 连接成功后补发」，断线重连自动恢复订阅；新增 `ensureConnected`。
- `internal/api/handlers/session.go`：开会话请求新增 `connection_id`，`auth_value` 为空时按它回退已保存密码/私钥路径（与测试连接口径一致）；临时密码开会话成功后回存。
- `internal/api/server.go`：会话 handler 注入 ConnectionService 以支持凭据回退。
- `flutter_ui/pubspec.yaml`：`xterm` 3.5 → 4.x（3.5 使用已被新版 Flutter 移除的 `RawKeyEvent`，此前因无人引用未被编译而未暴露，终端页引用后流水线首跑即报错）；`web_socket_channel` 2.4 → 3.x。

## 验证

- 本机：`go build ./...`、`go vet ./internal/api/...` 通过；`cd mobile && go test ./...` 通过。会话 handler 的凭据回退逻辑与连接测试共用同一口径，已有 handler 测试覆盖回退路径。
- GitHub 流水线：终端版首跑（[39391681481](https://github.com/a602017206/sshTools/actions/runs/39391681481)）因 xterm 3.5 的 `RawKeyEvent` 编译失败；升级后第 7 轮（[39394400066](https://github.com/a602017206/sshTools/actions/runs/39394400066)）Android 与 iOS 双双成功，产物 `sshtools-android-debug-apk` 已更新。
- 仍需真机复测：点「连接」应进入终端并出现远端提示符，可执行命令；特殊按键条的方向键/Ctrl 粘滞键可用；切后台回来连接断开后能点重连恢复。

## 剩余风险

- WebSocket 客户端的事件处理器按类型全局单例注册，多终端页同时存在会互相覆盖；当前一次只打开一个终端页，后续做多标签需改为按会话分发。
- iOS/安卓切后台后长连接会被系统回收，目前靠手动重连，未做保活与自动重连。
- 终端文本选区复制、长按菜单等沿用 xterm 4 默认实现，未针对真机手势专门验证。
- 终端页未做自动化 UI 测试，交互正确性依赖真机复测。
