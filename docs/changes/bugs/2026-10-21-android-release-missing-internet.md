# 安卓正式包无网络：缺 INTERNET 权限与回环明文放行

## 背景

R8 闪退修复后 App 能打开，但真机表现为：进入即飘红 `No internet connection`，连接列表为空，添加连接报「添加失败」。本地 Go 服务的 API 走 `http://127.0.0.1:<随机端口>`，前端 Dio 把连接类错误统一映射为 `No internet connection`。

## 定位过程

解析正式包 `AndroidManifest.xml` 的字符串池（UTF-16，需按 AXML 格式解析）：权限只有 `android.permission.DUMP`，**没有 `android.permission.INTERNET`**；也没有任何明文流量相关属性。原因：

- Flutter 脚手架只在 debug/profile 清单里带 INTERNET 权限，正式包主清单没有；调试版（此前真机可用的 200MB 包）因此一切正常，正式包连本机回环也被系统直接拒绝。
- 另外安卓 9+ 默认禁止明文 HTTP，即使补上权限，回环的 http/ws 仍会被默认网络安全策略拦截，需要单独放行。

两点都只影响 release，与 R8 的方法裁剪相互独立，是瘦身/签名切换到正式构建后一并暴露的存量问题。

## 修复

`.github/workflows/mobile.yml` 的 Android 清单步骤扩展为三件事：

1. 显示名改为 `AHaSSHTools`（原改名步骤，合并至此）；
2. 在 `<application>` 前注入 `<uses-permission android:name="android.permission.INTERNET"/>`；
3. 写 `res/xml/network_security_config.xml`，仅对 `127.0.0.1` 与 `localhost` 放行明文（API 永远只走回环，不全局开放明文），并在清单 `application` 上挂 `android:networkSecurityConfig`。

iOS 侧同步在 Info.plist 增加 `NSAppTransportSecurity.NSAllowsLocalNetworking=true`。发布断言步骤新增清单 INTERNET 检查（二进制清单经 `tr -d '\000'` 后 grep），与 `Seq.getRef` 断言合并为一条。

## 验证

- 流水线（39583571084）全绿，CI 内两条断言通过。
- 下载 arm64 实物复查：清单字符串池含 `android.permission.INTERNET`、`networkSecurityConfig`、`AHaSSHTools`；`resources.arsc` 含 `network_security_config` 资源名，包内对应二进制 XML 含 `127.0.0.1`/`localhost`/`cleartextTrafficPermitted`；dex 中 `getRef` 保持。
- 待用户真机覆盖安装确认：红条消失、能添加并连接服务器。

## 剩余风险

- 若用户手机上旧连接"看似消失"，多半只是列表接口失败时的空状态展示，数据一般仍在；修好后应能重新读出，若确实不见需再查数据目录迁移。
- 以后脚手架若自带 INTERNET，本注入会产生重复 uses-permission（安卓允许重复声明，无实际影响）。
