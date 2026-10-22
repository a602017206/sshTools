# 移动端实机反馈四问题修复

## 背景

正式版在三星 S23+（安卓 16）实机验证通过后，用户从截图反馈四个问题：①「账户」应叫「设置」，且主题色切换看不出效果；②已连接终端但会话列表为空；③资产卡片按钮溢出（删除跑到卡片外）；④顺带复查其它隐患。同时提出数据库能否在移动端使用的问题（结论附后）。

## 问题与修复

1. **账户改名设置**：底部导航第五栏与页面标题改为「设置」，图标换齿轮。主题色"没效果"经核截图属染色面积太小（生效但只染了少量控件）：仪表盘四环改取主题色板（primary/secondary/tertiary 及混合色），切换主题色时整个看板跟着变；设置页外观卡新增一组实时预览控件（示例按钮/已选 Chip/主题色图标），点色点立刻可对比。
2. **会话列表不刷新**：本页在 `AppShell` 的 `IndexedStack` 中自启动即保活，`initState` 只执行一次，启动时的空结果被一直展示。改为监听 `appShellIndexProvider`，每次切到会话栏重新拉取；手动刷新与下拉刷新保留。
3. **卡片按钮溢出**：五个操作按钮（连接/文件/测试/编辑/删除）横排超过窄屏卡片宽度，`Row` 改 `Wrap` 自动换行、右对齐不变，间距由 Wrap 的 spacing/runSpacing 接管。
4. **顺带查出的仪表盘数据问题**：后端 `collectCPUMetrics` 从未给 `PerCore` 赋值，前端核心数被兜底成 1，负载环比例失真（实机同一台机 HexHub 显示 2 核、我们显示 1 核）。后端新增每核采样：一次远端命令内读两次 `/proc/stat`（间隔 0.3s），Go 侧解析差值算每核占用率（新增 `parsePerCoreUsage`，含单测，本地已通过）；前端核心数取不到时显示 `—` 不再造假。附带：运行时间由英文原文格式化为「X天 X小时」，环下子文本允许两行防截断，「延迟」改称「响应」（实为监控接口整轮往返，非 ping）。

## 修改文件

- 修改：`internal/ssh/monitor.go`（每核采样与解析）、`flutter_ui/lib/presentation/screens/app_shell.dart`、`flutter_ui/lib/presentation/screens/account/account_screen.dart`、`flutter_ui/lib/presentation/screens/sessions/sessions_screen.dart`、`flutter_ui/lib/presentation/screens/connections/connection_list_screen.dart`、`flutter_ui/lib/presentation/screens/dashboard/dashboard_screen.dart`、`flutter_ui/lib/presentation/widgets/metric_gauge.dart`
- 新增：`internal/ssh/monitor_test.go`

## 数据库结论（用户第 5 问，本次只判断未实现）

- PC 端是双通道：MySQL/PostgreSQL 默认走 Go 原生 `database/sql` + 纯 Go 驱动直连（`internal/service/database_service.go` 的 `GetDSN/GetDriverName`），**不依赖 JDBC**；Oracle/SQL Server/自定义驱动配置走 JDBC 网关（拉起 Java agent 子进程 + gRPC）。
- 移动端判定：Go 原生通道与移动端后端是同一套 Go 代码，**MySQL/PostgreSQL 可以做**（需在移动装配中加 DatabaseService、数据库 API 与 Flutter 页面）；**JDBC 通道在移动端是硬墙**——手机上没有桌面 JVM，iOS 亦不允许子进程，无法照搬。Redis/Mongo 等原生 provider 理论可下沉，驱动逐个评估。
- 建议第一期：MySQL/PostgreSQL 连接管理 + 库表浏览 + SQL 查询结果表；生产库多在内网，后续再叠加经 SSH 会话的隧道连接。

## 验证

- `go test ./internal/ssh/ -run TestParsePerCoreUsage` 本地通过（20%/60% 两核差值算例 + 异常输入）。
- GitHub Actions 移动端流水线（39710119031）成功，含发布断言。
- 待真机复测：设置页切主题色看仪表盘四环与预览是否变色；连终端后切到会话列表应立即出现会话；窄屏卡片按钮完整换行；仪表盘核心数显示 2。

## 剩余风险

- 每核采样在极老旧或裁剪过的远端系统上若无 `/proc/stat` 的 cpuN 行会返回空，前端显示 `—`，不影响其余指标。
- 主题色经 M3 种子生成，浅色模式下各色差异主要体现在主色与容器色，仍有用户觉得淡时可再考虑深色点缀区。
