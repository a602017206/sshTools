# 移动端流水线：固定发布签名与 APK 瘦身

## 背景

前几轮的 debug APK 由 GitHub runner 的临时调试密钥签名，每次打包签名都不同，真机必须卸载重装、应用数据（已保存连接与密码）全丢；同时单个 APK 内嵌三个 ABI 的 Go 库，未压缩约 250MB。本次把签名与体积这两个体验问题一并解决。

## 范围

仅 CI/打包流程：Android 固定发布签名、release 构建、按 ABI 拆分、gomobile 去符号。不改任何业务代码。iOS 仍为无签名验证构建（上架签名另行处理）。

## 修改文件

- `.github/workflows/mobile.yml`：
  - 签名注入步骤：从 GitHub Secrets 解出密钥库写 `key.properties`，并向 `flutter create` 生成的 `app/build.gradle.kts` 追加 `signingConfigs.release` 配置。Secrets 经 `env:` 传入而非插值进 `run:`，避免命令注入面；Secrets 缺失时跳过注入并回退 debug 构建，fork 不受影响。
  - 追加脚本踩到的坑：新版 Gradle 中 `java` 被扩展访问器占用，`java.util.Properties` 无法直接解析，必须先给文件头 `sed` 插入 `import java.util.Properties`/`java.io.FileInputStream`。
  - Android 构建由 `--debug` 改为 `--release --split-per-abi`，产物名改为 `sshtools-android-apk`（含三个 ABI 的包）。
  - gomobile 绑定（Android/iOS）加 `-ldflags "-s -w"`。
- 仓库外变更（不入库）：本机生成的发布密钥库备份在维护者主目录 `~/.sshtools-signing/`（含口令文件，权限 600），base64 与口令写入仓库 Secrets：`ANDROID_KEYSTORE_BASE64`、`ANDROID_KEYSTORE_PASSWORD`、`ANDROID_KEY_ALIAS`、`ANDROID_KEY_PASSWORD`。密钥库与口令严禁入库。

## 验证

- 流水线第 9 轮因 Gradle import 问题失败，修复后第 10 轮（[39410081933](https://github.com/a602017206/sshTools/actions/runs/39410081933)）Android 与 iOS 双双成功；日志确认 `release signing injected` 且产出 release 包。
- 本机下载复核：产物为 `app-arm64-v8a-release.apk`（51MB）、`app-armeabi-v7a-release.apk`（48MB）、`app-x86_64-release.apk`（55MB），各自只含对应架构的 `libgojni.so`；APK 内存在 v2 签名块（`APK Sig Block 42`）。对比旧 debug fat APK 252MB，单机型体积降到约 1/5。
- 签名连续性：本版签名与之前 debug 包仍不同，真机需最后卸载一次；此后同密钥签名的版本可覆盖安装、数据保留（待下轮更新时验证）。

## 剩余风险

- `libgojni.so` 单架构未压缩仍约 34MB：`-s -w` 对 gomobile 的 c-shared 产物压缩有限，体积大头是整个 `internal/service` 依赖图（数据库驱动、消息队列客户端等）。进一步瘦身须按移动 MVP 裁剪依赖（构建标签拆分），不在本次范围。
- 密钥单点在维护者本机与 GitHub Secrets：丢失密钥将无法再覆盖更新已装应用，应另行备份（建议密码管理器存一份 base64 与口令）。
- release 构建目前未做真机回归，R8/资源收缩未启用；后续启用混淆需为 gomobile 的 Java 类补 keep 规则。
- iOS 上架签名、图标与商店元数据仍未处理。
