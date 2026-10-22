# 移动端应用图标与桌面端统一

## 背景

桌面端已有自有图标 `build/appicon.png`（终端窗 + `>_ aha` + 网线插画，1024×1024），而移动端流水线生成的脚手架一直用 Flutter 默认图标。用户要求安卓与 iOS 引用同一张图标。

## 范围

- 图标源图从 `build/appicon.png` 拷贝到 `flutter_ui/assets/icon/icon.png` 入库，作为移动端图标的单一来源（桌面图更新时同步替换此文件即可）。
- `pubspec.yaml` 增加 `flutter_launcher_icons` 配置：Android 生成各密度 mipmap 与自适应图标（源图为白底满幅插画，背景取纯白、整图作前景，系统安全区缩放后四周是同色白边）；iOS 生成全套 `AppIcon.appiconset` 且 `remove_alpha_ios: true`（iOS 不接受透明通道，源图本身为不透明 RGB，正合要求）。
- 流水线 Android/iOS 两个 job 在 `flutter create` 生成脚手架后运行 `dart run flutter_launcher_icons` 刷图标。

## 修改文件

- 新增：`flutter_ui/assets/icon/icon.png`
- 修改：`flutter_ui/pubspec.yaml`（dev 依赖与图标配置）、`.github/workflows/mobile.yml`（两 job 各加 Apply app icon 步骤；脚手架改为 `--platforms=android,ios` 双平台生成）

## 验证

- 第一轮流水线（39534668341）两 job 均红在 Apply app icon：图标工具同时写两端目录，而各 job 只生成了单平台脚手架，缺目录直接 `PathNotFoundException`。改为双平台脚手架后（39536187379）全绿，多出的另一端模板只是文件、不参与各自构建。
- APK 解包确认图标经 `resources.arsc` 索引引用（release 会缩短 `res/` 文件名，按文件名找不到属正常）。
- iOS 侧仅构建验证；图标真机效果（圆角遮罩裁切白底四角）待有签名条件时确认。

## 剩余风险

- Android 自适应图标前景用整图插画，个别厂商桌面（三星 One UI 的圆角方形遮罩）会裁掉插画四周少量白边，主体（终端窗）居中不受影响；后续若要更精细，可单独出一个去掉外圈光晕的前景版图再配置。
- Android 12+ 的单色主题图标（themed icon）未配置，需要时再补 monochrome 前景。
