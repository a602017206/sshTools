# Flutter + 内嵌 Go 移动端适配（流水线打包验证）

## 背景

桌面端基于 Wails v2，仅支持 Windows/macOS/Linux，无法构建 Android/iOS。经评估，Go 服务层（`internal/ssh`、`internal/service`、`internal/config`、`internal/store`、`internal/api`）以纯 Go 网络/加密代码为主，具备内嵌进移动应用的条件；仓库已有独立进程版 API 服务器（`cmd/apiserver`）与 Flutter 客户端脚手架（`flutter_ui/`），本次在此基础上落地「Flutter UI + gomobile 内嵌 Go」路线。按维护者要求：本地不做移动端打包，全部通过 GitHub 流水线（本机 `gh` 触发/查看）验证。设计与取舍见 `docs/designs/2026-10-20-flutter-gomobile-mobile.md`。

## 范围

包含：

- Go 侧 gomobile 绑定（进程内启动 API 服务器、沙盒数据目录、仅回环监听）；
- Flutter 侧原生桥接插件与 App 启动接线（移动端自动拉起内嵌 Go 并切换 API 基地址）；
- GitHub Actions 移动端打包流水线（Android APK、iOS 无签名构建）与产物上传。

不包含（后续工作）：SSH 终端 UI 与特殊按键条、SFTP/监控路由在 API 层的补齐、JDBC 与本地终端（移动端明确不支持）、后台保活与断线重连、Keychain/Keystore 凭证迁移、平台工程脚手架（`android/`、`ios/`）入库固化（现由流水线用 `flutter create` 生成）。

## 修改文件

- `mobile/mobile.go`（新增）：gomobile 导出 `Start`/`Stop`/`Port`/`IsRunning`/`DataDir`；装配与 `cmd/apiserver` 相同的服务组合，在进程内启动服务器，`Start` 内等健康检查通过才返回端口。
- `mobile/bind_runtime.go`（新增）：空导入 `golang.org/x/mobile/bind`，防止 `go mod tidy` 删除 gomobile 运行时依赖。
- `mobile/go.mod`、`mobile/go.sum`（新增）：移动绑定拆为独立子模块（`replace` 引用主模块）。原因：流水线首跑发现 CI 里动态 `go get golang.org/x/mobile/bind@latest` 会把 2019 年旧版 `genproto` 写进主模块、与拆分后的 `genproto/googleapis/rpc` 冲突（gobind 全图加载报 ambiguous import），且会抬升 grpc 等桌面依赖；子模块把版本变动隔离在移动端，桌面 `go.mod` 不动，gomobile 工具版本由子模块锁定。
- `mobile/mobile_test.go`（新增）：内嵌服务器启停、健康检查、重复 Start 幂等、空目录拒绝等测试。
- `internal/api/server.go`：新增 `NewServerWithHost`，`NewServer` 改为委托它；桌面 `cmd/apiserver` 行为不变（仍监听全网卡），内嵌模式传 `127.0.0.1` 仅回环。
- `flutter_ui/plugins/sshtools_go/`（新增）：Flutter 桥接插件。Dart 侧 MethodChannel `sshtools_go`；Android 为 Kotlin 插件（后台线程调起 Go、主线程回结果）；iOS 为 Swift 插件（调用 gomobile 的 `MobileStart` 等 C 接口），podspec 声明链接 `resolv`（Go cgo DNS 解析需要，否则链接报 `_res_9_*` 未定义）。gomobile 产物（jar/so/xcframework）由流水线生成，插件目录 `.gitignore` 排除。
- `flutter_ui/lib/main.dart`：移动端启动时经 `path_provider` 取沙盒目录、拉起内嵌 Go 并把 API 基地址切到 `127.0.0.1:<port>`；桌面/Web 保持默认地址（独立 `apiserver`），启动失败不阻断 UI。
- `flutter_ui/lib/core/constants/api_constants.dart`：`defaultBaseUrl` 改为可配置，`wsUrl` 从当前基地址推导（保证内嵌模式端口一致），新增 `configureEmbedded`/`isEmbedded`。
- `flutter_ui/pubspec.yaml`：新增 `path_provider` 与本地插件 `sshtools_go`；删除 `file_picker`——其 6.x 版本使用已被 Flutter 移除的 `PluginRegistry.Registrar`，且 `lib/` 内无任何引用。
- `.github/workflows/mobile.yml`（新增）：两条 job。Android（ubuntu）：按锁定版本安装 gomobile → `cd mobile && go test` → gomobile 生成 AAR 并解包为 `classes.jar`+`jniLibs/*.so`（库工程不能内嵌本地 AAR，AGP 会在 `bundleDebugAar` 报错）→ `flutter create` 生成脚手架 → `flutter build apk --debug` → 上传 APK；iOS（macos）：gomobile 生成 `Mobile.xcframework` → `flutter create` → `flutter build ios --no-codesign --debug` → 上传 Runner.app。
- `docs/designs/2026-10-20-flutter-gomobile-mobile.md`（新增）：设计与取舍记录。
- `flutter_ui/README.md`：项目状态补充内嵌 Go 与流水线条目。

## 验证

本地（不做移动端打包，仅代码级检查）：

- `GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/apiserver`：通过（服务层依赖图在 Android 可编译）；iOS 需 cgo，交由 gomobile 工具链处理。
- `cd mobile && go vet ./... && go test ./...`：通过。
- 主模块回归：`go build ./...` 通过；`go test ./internal/api/...` 通过（该包无测试文件，仅编译性验证）；桌面入口未改行为。

GitHub 流水线（`gh` 触发/查看，分支 `feat/mobile-flutter-gomobile`）：

- 第 1 轮（39355876824）：gomobile 绑定失败——CI 动态 `go get` 引入旧版 `genproto` 造成 ambiguous import；改为子模块锁定版本后解决。
- 第 2 轮（39358115643）：双平台 gomobile 产物生成成功、Go 模块测试通过；iOS 链接缺 `_res_9_*`（补 `resolv`）；Android 被 `file_picker` 旧 API 卡住（删依赖）。
- 第 3 轮（39363054700）：iOS 全绿；iOS 之外 Android 在 `:sshtools_go:bundleDebugAar` 报错——库工程不能内嵌本地 AAR，改解包消费。
- 第 4 轮（39367818524）：**Android 与 iOS 双双成功**，产物 `sshtools-android-debug-apk`、`sshtools-ios-runner-app` 已上传。本机下载 APK 复核：内含 `lib/arm64-v8a|armeabi-v7a|x86_64/libgojni.so`（每个约 34MB），Go 核心确实内嵌进包。

未做真机/模拟器运行验证：本次只验证到「能打包且产物结构正确」，App 在真机上启动内嵌 Go、完成 SSH 连接的实际行为需后续装机验证。

## 剩余风险

- **包体积**：debug APK 约 252MB。内嵌的是整个 `internal/service` 依赖图（含数据库驱动、grpc、消息队列客户端等），且 gomobile 默认带符号、三 ABI 各一份 `libgojni.so`。后续应按移动 MVP 裁剪依赖（如给重型服务加构建标签、只绑 SSH/SFTP/连接子集）、发版用 `-ldflags "-s -w"` 与按 ABI 拆分，可大幅瘦身。
- **构建时长**：iOS 的 gomobile 绑定冷构建约 40 分钟（cgo 全量编译依赖图），与上一条同源；裁剪依赖图后会显著改善。流水线目前无缓存优化。
- **插件 Kotlin 模型**：Flutter 新版对「插件自行 apply Kotlin Gradle Plugin」给出 Built-in Kotlin 迁移警告（当前仅警告，本插件在列）。后续需按官方指南迁移插件的 Android 构建脚本。
- **数据目录手法**：Go 侧靠绑定前设置进程 `HOME` 让 `config`/`store` 落入沙盒，属最小侵入做法；若后续给 `config` 增加显式 DataDir 构造参数，应改为显式注入。
- **凭证加密**：移动端沿用机器特征（hostname+HOME）派生密钥，未迁 Keychain/Keystore，安全性依赖系统沙盒隔离；上架前应迁移。
- **进程与连接生命周期**：iOS 切后台会被挂起/回收，SSH 长连接保活与断线重连策略尚未实现；Android 前台服务亦未接入。
- **平台脚手架未入库**：`flutter create` 生成结果随 Flutter 版本漂移，真机调试常态化后应把 `android/`、`ios/` 入库固化并锁定 Flutter 版本。
- API 层目前仅注册连接与会话路由，SFTP/监控/设置路由待补齐后移动端才能用上对应功能；终端 UI 未实现前，移动端可完成连接管理与连接测试，尚不能开终端会话。
