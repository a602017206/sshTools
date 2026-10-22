# README 下载区补齐移动端

## 背景

v0.0.24 起 Release 同时发布 Android APK 与 iOS 未签名 IPA，但仓库 README 的下载表格仍只有 Windows 与 macOS 两类，用户红框标注要求一并补上。

## 范围

- `README.md` 下载表格新增四行：Android ARM64（主流真机）、Android ARMv7、Android x86_64、iOS（未签名，需自签侧载），链接与 shields.io 下载量徽章沿用既有写法，均指向最新 Release 对应附件名。
- 表格下方补充安装说明：Android 直接安装 APK 并允许未知来源、固定签名可覆盖升级；iOS 需 AltStore/Sideloadly 自签侧载。
- 简介一行由「桌面客户端」改为涵盖移动端，并注明移动端由同一 Go 后端驱动（Flutter 界面）。

## 修改文件

- `README.md`
- `docs/changes/process/2026-10-22-readme-mobile-downloads.md`（本文）

## 验证

- 附件名与 v0.0.24 Release 实际资产逐一核对：`AHaSSHTools-android-arm64-v8a.apk`、`AHaSSHTools-android-armeabi-v7a.apk`、`AHaSSHTools-android-x86_64.apk`、`AHaSSHTools-ios-unsigned.ipa` 均存在。
- 徽章 URL 规则与现有桌面三行一致（`github/downloads/.../latest/<附件名>`）。

## 剩余风险

- shields.io 徽章对刚发布的附件可能有短暂缓存延迟，下载量显示为 0 属正常。
- iOS 自签流程随所用工具版本变化，README 仅作指引，以工具官方说明为准。
