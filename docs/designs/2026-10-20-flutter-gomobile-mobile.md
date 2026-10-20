# Flutter + 内嵌 Go 移动端适配设计

## 背景

桌面端使用 Wails v2，仅支持 Windows/macOS/Linux，无法构建 Android/iOS。需要在不重写 Go 核心的前提下验证移动端路线，并按仓库维护者的要求：本地不做打包，全部通过 GitHub 流水线验证。

探测结论（2026-10-20 本机验证）：

- `GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./cmd/apiserver` 可通过，Go 服务层（`internal/ssh`、`internal/service`、`internal/config`、`internal/store`、`internal/api`）的纯 Go 依赖在 Android 上可编译；
- iOS 需 cgo 外部链接，由 gomobile 工具链（Clang/Xcode）处理，正是 gomobile 的标准场景。

## 决策

采用 **Flutter UI + gomobile 内嵌 Go**：

- Go 侧新增 `mobile` 包，用 gomobile 绑定把现有 API 服务器装配（与 `cmd/apiserver` 相同的服务组合）跑在应用进程内，仅监听 `127.0.0.1` 随机端口。选择「进程内库」而不是桌面式独立进程，是因为 iOS 禁止子进程、Android 也不适合以裸二进制常驻。
- Flutter 侧新增本地插件 `flutter_ui/plugins/sshtools_go`（MethodChannel `sshtools_go`），Android 链接 gomobile 产出的 AAR，iOS 链接 `Mobile.xcframework`，负责把沙盒目录传入 Go 并拿回端口。
- Flutter 现有数据层（Dio + WebSocket，走 `/api/v1`）保持不变，只是基地址在移动端改为内嵌服务器地址；桌面调试仍连独立运行的 `cmd/apiserver`（`localhost:8080`）。
- Flutter 平台工程脚手架（`android/`、`ios/`）暂不入库，统一由 CI 用 `flutter create` 按当前稳定版 Flutter 生成，避免在无 Flutter SDK 的开发机上手工维护易漂移的模板文件；后续真机调试常态化后再把脚手架入库固化。

## 不在本次范围

- SSH 终端 UI（xterm 集成）、特殊按键条、软键盘适配；
- SFTP/监控等路由在 API 层的补齐（当前 `internal/api` 仅注册了连接与会话路由）；
- JDBC 功能：依赖拉起 Java 子进程，移动端明确不支持；
- 本地终端（`creack/pty` 拉起本机 Shell）：移动沙盒无此概念，明确不做；
- 后台保活（iOS 挂起、Android 前台服务）与断线重连策略；
- 凭证存储从机器特征派生密钥迁移到 Keychain/Keystore（先沿用现有实现，靠应用沙盒隔离；已在风险中记录）。

## 数据目录

现有 `config`/`store` 均按 `$HOME/.ahasshtools` 解析。移动内嵌时由 Flutter 经 `path_provider` 取应用私有目录，Go 绑定在装配服务前把进程 `HOME` 指向该目录，使配置与凭证落入沙盒。这是最小侵入方案；若后续给 `config` 增加显式 DataDir 构造参数，应改为显式注入并去掉该处理。

## 安全

- 内嵌服务器只绑 `127.0.0.1`（`api.NewServerWithHost`），不暴露局域网；桌面 `cmd/apiserver` 行为不变，仍监听全网卡。
- 当前 API 层无鉴权，回环隔离是移动端的主要边界；路线若改为远程服务器模式，必须先补 TLS 与鉴权，不能直接复用这一假设。

## 验证方式

GitHub Actions 工作流 `.github/workflows/mobile.yml`：

- Android（ubuntu）：gomobile 生成 AAR → `flutter create` 生成脚手架 → `flutter build apk --debug`，上传 APK 产物；
- iOS（macos）：gomobile 生成 XCFramework → `flutter create` 生成脚手架 → `flutter build ios --no-codesign --debug`，上传 Runner.app 产物。

本机只跑 `go test ./mobile/` 与交叉编译检查，不在本地打包。
