# 移动端主题切换、底部四栏与监控看板

## 背景

移动端此前全应用强制深色主题（`themeMode: ThemeMode.dark`），用户希望能切换外观。参考 HexHub 移动端的结构（底部四栏「仪表盘 / 资产列表 / 会话列表 / 账户」，仪表盘为服务器总览卡：延迟、四联环、系统信息），为移动端复刻同类功能。说明：仅参考其信息结构与交互形态，HexHub 的具体界面实现无公开资料可复制，视觉与代码均为自研。

## 范围

- 主题切换：跟随系统 / 浅色 / 深色三态，持久化到 SharedPreferences，账户页切换。
- 底部四栏导航壳 `AppShell`（IndexedStack 保活），对齐 HexHub：仪表盘、资产列表、会话列表、账户。
- 监控看板：选择一台已保存资产，自动打开专用监控会话并 3 秒轮询；总览卡四联环（负载、内存、CPU、主分区磁盘）+ 系统信息行 + 延迟；明细卡含每核 CPU、磁盘分区、内存与网络。
- 会话列表：展示服务端活跃会话并逐个关闭。
- 账户页（本地模式）：主题切换、连接状态与应用信息（无云账号体系，对应 HexHub 账户栏的位置）。
- Go API 新增 `GET /sessions/:id/monitor`（监控服务未装配时 501）。监控会话凭据走已保存连接回退，复用连接管理文档中的机制。
- 移除 go_router 占位路由与 `HomeScreen` 占位页，`main.dart` 改为 `MaterialApp` + `AppShell`。

## 修改文件

- 新增：`internal/api/handlers/monitor.go`；`flutter_ui/lib/core/theme/theme_provider.dart`、`flutter_ui/lib/data/models/monitoring_data.dart`、`flutter_ui/lib/data/repositories/monitor_repository.dart`、`flutter_ui/lib/presentation/widgets/metric_gauge.dart`、`flutter_ui/lib/presentation/screens/app_shell.dart`、`flutter_ui/lib/presentation/screens/dashboard/dashboard_screen.dart`、`flutter_ui/lib/presentation/screens/sessions/sessions_screen.dart`、`flutter_ui/lib/presentation/screens/account/account_screen.dart`
- 修改：`internal/api/server.go`（注册 monitor 路由）、`flutter_ui/lib/presentation/providers/api_providers.dart`（monitorRepositoryProvider）、`flutter_ui/lib/main.dart`（三态主题 + AppShell）
- 删除：`flutter_ui/lib/core/router/app_router.dart`、`flutter_ui/lib/presentation/screens/home_screen.dart`

## 验证

- `go build ./...`、`gofmt -l` 通过；流水线第 11 轮（39425297146）成功，Android release 签名包构建通过。
- 尚待真机验证：账户页切浅色/深色重启仍生效；仪表盘连接已存连接后四环与信息每 3 秒刷新、延迟数值变化；会话列表能看到终端/文件/监控会话且可关闭；深浅色下各页观感。

## 剩余风险

- 监控会话 ID 固定前缀 `mobile-monitor-`，同设备双开仪表盘会互相挤占（会话 ID 全局命名，Manager 直接替换），移动端单前台场景下可接受；需要时可改为随机后缀。
- 仪表盘切换资产前必须先断开监控（下拉在监控中禁用），与 HexHub 的单机看板形态一致。
- 负载环图以 `负载 / 核心数` 归一化，超过 1 时画满，与 HexHub 的小环一致；数值本身以文本展示在环下。
- 浅色主题下终端页保持深色（终端自身配色固定），是刻意选择，账户页里已注明。
- 流水线全绿仅证明编译与后端逻辑，视觉与交互细节仍待真机反馈后迭代。
