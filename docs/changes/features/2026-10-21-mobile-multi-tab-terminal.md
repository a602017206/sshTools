# 移动端多标签终端

## 背景

移动端此前只有一个终端页：从资产列表点连接推全屏页，退出即关会话，同一时刻只能开一个 shell。真机使用中需要在多台服务器（或同一台多 shell）之间来回切换，原路返回重连成本高。本次改为底部导航独立的「终端」栏，支持多标签保活。

## 范围

- 终端改为标签页架构：
  - 标签栏：状态点（连接中橙 / 已连接绿 / 断开或失败红）+ 连接名 + 关闭按钮，横向滚动；点标签切换，点 × 关闭标签（随之关闭该 SSH 会话）。
  - 内容区 `IndexedStack` 保活所有标签的 xterm 与会话，切标签、切底部导航都不中断后台会话；切到某标签时自动聚焦弹键盘。
  - 断开/失败状态条内提供「重新连接」，复用原会话收尾逻辑。
- 打开策略：
  - 资产列表点「连接」：同一资产已有标签则直接激活，否则新建，并自动切到终端栏。
  - 终端页右上角「+」：从已保存资产列表新建标签，同一资产允许多开 shell。
- 会话联动：终端标签把服务端会话 ID 与状态回报到 `terminalTabsProvider`；会话列表页点击终端会话可跳回对应标签，文件/监控会话仅展示。
- WebSocket 客户端改造：事件处理器由「同 type 单处理器覆盖」改为「同 type 多处理器并存、按处理器实例注销」，多标签各自过滤 `session_id`，互不覆盖、互不误删。
- 凭据弹窗提取为共享组件 `auth_input_dialog.dart`（`requestAuthValue`），资产列表的连接/测试/文件与新建标签统一走同一入口。
- 底部导航由四栏扩展为五栏（仪表盘 / 资产列表 / 终端 / 会话列表 / 账户），当前栏下标提升到 `appShellIndexProvider`，供跨页程序切换。

## 修改文件

- 新增：`flutter_ui/lib/presentation/providers/terminal_tabs_provider.dart`、`flutter_ui/lib/presentation/screens/terminal/terminal_tabs_screen.dart`、`flutter_ui/lib/presentation/widgets/auth_input_dialog.dart`
- 修改：`flutter_ui/lib/data/datasources/remote/websocket_client.dart`（多处理器）、`flutter_ui/lib/presentation/screens/terminal/terminal_tab_view.dart`（由原 `terminal_screen.dart` 重构）、`flutter_ui/lib/presentation/screens/app_shell.dart`（五栏 + provider 控制下标）、`flutter_ui/lib/presentation/screens/connections/connection_list_screen.dart`（开标签 + 共享凭据弹窗）、`flutter_ui/lib/presentation/screens/sessions/sessions_screen.dart`（点击跳标签）
- 删除：`flutter_ui/lib/presentation/screens/terminal/terminal_screen.dart`（已被标签页组件取代）

## 验证

- GitHub Actions 移动端流水线第 12 轮（39527274666）成功，Android release 签名包构建通过（Dart 编译即静态检查）。
- 尚待真机验证：同一资产连接两次的行为（资产列表激活已有标签、「+」新开第二个 shell）、多标签各开一台服务器互不串话（输出隔离）、切到底部其他栏再回来终端输出是否持续、关闭激活标签后焦点落到相邻标签、会话列表点终端会话跳转。

## 剩余风险

- 标签数量无上限，每标签一条 SSH 会话，极端多开会占用服务器会话配额；移动端使用场景下暂不限制。
- 关闭标签不做二次确认（与常见终端 App 一致），误触 × 会断开会话，需要时可加确认。
- xterm 的 `Terminal` 对象随标签保活，长时间挂很多标签的内存占用待真机观察。
- 应用切后台后系统可能回收进程，标签不做恢复（与之前单终端一致），重进应用需重新连接。
