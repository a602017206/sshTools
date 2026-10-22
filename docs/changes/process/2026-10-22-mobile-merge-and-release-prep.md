# 移动端合并 master 与发布流水线接入准备

## 背景

移动端（Flutter + 内嵌 Go）在 `feat/mobile-flutter-gomobile` 分支完成并经真机验证后合并回 master，并为下一次正式发版备好发布流水线改造：Android/iOS 成为与桌面并列的发布制品。经确认本次只合并、不打 tag，发版要等桌面在途改动收尾后一并处理（其发布说明已在维护者本地预填）。

## 本次变更

- 移动端全量经 PR #1 以合并提交进入 master（77 个文件，含 `flutter_ui/`、`mobile/` 子模块、移动 CI 与文档）。
- `.github/workflows/release.yml` 新增两个 job（本 PR）：
  - `android`：与 `mobile.yml` 同一套构建（gomobile 内嵌 Go、16KB 对齐、图标/清单/R8 保留、固定签名、分 ABI），产物改名 `AHaSSHTools-android-<abi>.apk`，构建后断言 Seq.getRef 存活且清单含 INTERNET 权限；
  - `ios`：无签名构建 `Runner.app` 后按 `Payload/` 结构打成 `AHaSSHTools-ios-unsigned.ipa`；
  - `release` job 依赖三组 job，制品清单增加 `*.apk`/`*.ipa`，更新说明模板在「新增功能」加入移动端条目，下载说明新增 Android/iOS 两节（iOS 显著标注未签名、需自签）。
- 桌面 job（Windows/macOS）不做任何改动。

## 验证

- PR #1 合并状态 CLEAN，已合并（81c41a2）。
- 本 PR 的 YAML 结构经本地解析校验；Android/iOS job 的每一步均源自 `mobile.yml` 已连续多轮全绿的同名步骤，未引入新命令（仅增加产物改名与 IPA 打包）。
- 实际发版验证留待下一次打 tag：预期 Release 附件出现 Windows exe、macOS 双 zip、Android 三 APK、iOS 未签名 IPA。

## 剩余风险

- 本 PR 暂不合并：维护者本地 `release.yml` 有未提交的发版说明编辑（与本 PR 修改的正文区域相邻），需等其随桌面改动提交后再合并本 PR，必要时解一次小冲突。
- iOS 无签名 IPA 不能直接安装，Release 说明中已标注需 AltStore 等自签；正式上架签名另行立项。
- Android 签名 Secrets 为仓库级，release 与 mobile 两条流水线共用同一密钥，覆盖安装兼容性不变。
