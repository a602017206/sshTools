# 安卓正式包启动闪退真因：R8 裁掉 gomobile 桥接方法

## 背景

上一轮按 16KB 内存页对齐修复后，三星 S23+（安卓 16）仍一启动闪退，用户反馈最后一个能用的版本是签名瘦身前的 200MB 调试图包。本次拿到用户用无线调试抓取的 logcat，锁定真因。本文件结论取代 `2026-10-21-android16-16kb-page-crash.md` 的主因判断（16KB 对齐仍是真实的兼容性问题，修复保留）。

## 证据链

1. tombstone：`Fatal signal 6 (SIGABRT)`，Abort message 为 `failed to find method Seq.getRef`；调用栈经 `Java_go_Seq_init` → `go.Seq.<clinit>`，且上层帧类名已被混淆（如 `te.run`），是插件后台线程初始化 Go 核心时触发。
2. 解包验证：正式包 dex 字符串中只有 `getRefreshRate`，**没有 `getRef`/`incRef`**；`go.Seq` 类本身在，但方法被移除/改名。
3. 机制：gomobile 的 `Seq.getRef` 等方法只由 Go 运行时经 JNI 按名查找，Java 侧无静态引用，release 构建的 R8 将其判定为无用代码裁剪（或改名）。调试包不开 R8，故 200MB 的调试版一直正常——时间线上「正式包必崩、调试包可用」完全吻合。
4. 此前排除项复核：dex 类完整（只查了类名、没查方法，是上次漏检）、APK 内 `.so` 偏移对齐正常、图标与清单结构正常。

## 修复

- `.github/workflows/mobile.yml` 新增 "Keep gomobile bridge classes from R8" 步骤：写 `android/app/proguard-rules.pro`（`go.**`、`mobile.**`、`com.ahasshtools.sshtools_go.**` 整包 `-keep`，含全部成员），并用与签名注入相同的二次 `android {}` 块技巧把 proguardFiles 挂到 release 构建类型。
- 构建后新增 "Assert gomobile bridge survived R8" 断言：每个 APK 的全部 dex 里必须含 `getRef` 字符串，缺失直接让流水线变红，防止模板升级或规则被误删后问题回归。

## 修改文件

- 修改：`.github/workflows/mobile.yml`（R8 保留规则注入 + dex 回归断言；同一提交含应用改名，见同日 process 文档）

## 验证

- 流水线（39567110681）全绿，CI 内 dex 断言通过。
- 下载 arm64 实物复查：`getRef`、`incRef` 均已回归 dex；`libgojni.so` 16KB 对齐保持；清单字符串池含 `AHaSSHTools`。
- 待用户真机覆盖安装最终确认：能进入页面即关闭本问题。

## 剩余风险

- 若将来 Flutter 模板改为完全禁用 R8，keep 规则无害保留即可；若新增其它经 JNI/反射按名调用的库，需把对应包名补进保留规则。
- iOS 无 R8 机制，不受影响。
