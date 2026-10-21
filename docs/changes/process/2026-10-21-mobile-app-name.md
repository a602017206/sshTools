# 移动端应用显示名与桌面端统一

## 背景

移动端脚手架默认显示名为工程名 `ahasshtools_flutter`，用户要求与 PC 端保持一致（`wails.json` 的 `name`：`AHaSSHTools`）。

## 修改

- `.github/workflows/mobile.yml`：
  - Android job 在生成脚手架后用 `sed` 改 `android/app/src/main/AndroidManifest.xml` 的 `android:label` 为 `AHaSSHTools`，并 grep 输出便于日志核对。
  - iOS job 用 `plistlib` 写 `ios/Runner/Info.plist` 的 `CFBundleName` 与 `CFBundleDisplayName` 为 `AHaSSHTools`。
- 脚手架不入库，改名与签名注入一样在流水线里完成，避免模板漂移。

## 验证

- 流水线 39567110681 全绿；解析实物 APK 的清单字符串池（UTF-16，需按 AXML 格式解析，`strings` 读不到），确认含 `AHaSSHTools` 且 `label` 属性指向它。
- iOS 侧仅构建验证，显示名真机效果待有签名条件时确认。

## 剩余风险

- 包名（`com.ahasshtools.ahasshtools_flutter`）本次不变，改它会影响覆盖安装与数据保留，如需调整必须单独评估。
