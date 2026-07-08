# macOS 应用签名与分发指南

## 问题说明

在其他 Mac 上运行应用时出现"文件已损坏"错误，是因为 macOS Gatekeeper 安全机制。

## 解决方案

### 方案一：Ad-hoc 签名（无需开发者账号）✅ 推荐快速使用

#### 1. 使用构建脚本

```bash
# 使用提供的构建脚本
./scripts/build-mac.sh
```

#### 2. 分发应用

打包成 zip：
```bash
cd build/bin
zip -r sshTools.zip sshTools.app
```

#### 3. 用户端操作

用户下载后需要执行：
```bash
# 解压后执行
xattr -cr sshTools.app

# 然后可以正常打开
open sshTools.app
```

或者用户可以：
1. 尝试打开应用（会失败）
2. 打开 **系统偏好设置 → 安全性与隐私 → 通用**
3. 点击"仍要打开"按钮

---

### 方案二：正式签名（需要 Apple Developer 账号）

如果你有 Apple Developer 账号（¥688/年），可以进行正式签名和公证。

#### 1. 获取开发者证书

```bash
# 查看可用的签名身份
security find-identity -v -p codesigning
```

输出示例：
```
1) XXXXXX "Developer ID Application: Your Name (TEAM_ID)"
```

#### 2. 创建签名构建脚本

创建 `scripts/build-mac-signed.sh`：

```bash
#!/bin/bash

set -e

# 配置你的签名身份
SIGNING_IDENTITY="Developer ID Application: Your Name (TEAM_ID)"
BUNDLE_ID="com.yourcompany.sshtools"

echo "Building with signing..."
wails build -clean

APP_PATH="./build/bin/sshTools.app"

# 签名
echo "Signing application..."
codesign --force --deep \
  --sign "$SIGNING_IDENTITY" \
  --timestamp \
  --options runtime \
  --entitlements build/darwin/entitlements.plist \
  "$APP_PATH"

# 验证签名
codesign -dv --verbose=4 "$APP_PATH"

# 创建 DMG
echo "Creating DMG..."
create-dmg \
  --volname "sshTools" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --app-drop-link 400 200 \
  "build/bin/sshTools.dmg" \
  "$APP_PATH"

echo "✅ Signed build complete!"
```

#### 3. 创建 Entitlements 文件

创建 `build/darwin/entitlements.plist`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
    <key>com.apple.security.network.server</key>
    <true/>
</dict>
</plist>
```

#### 4. 公证（Notarization）

```bash
# 1. 上传到 Apple 进行公证
xcrun notarytool submit build/bin/sshTools.dmg \
  --apple-id "your@email.com" \
  --team-id "TEAM_ID" \
  --password "app-specific-password" \
  --wait

# 2. 验证公证
xcrun stapler staple build/bin/sshTools.dmg

# 3. 验证结果
spctl -a -vv -t install build/bin/sshTools.dmg
```

---

## 快速参考

### 当前推荐流程（无开发者账号）

```bash
# 1. 构建
./scripts/build-mac.sh

# 2. 打包
cd build/bin && zip -r sshTools.zip sshTools.app

# 3. 分发时告知用户执行
xattr -cr sshTools.app
```

### 如果有开发者账号

1. 配置签名身份
2. 使用 `codesign` 签名
3. 使用 `notarytool` 公证
4. 分发 DMG 文件

---

## 常见问题

### Q: 为什么需要 xattr -cr？
A: macOS 会给从网络下载的文件添加"隔离"标记，xattr -cr 可以移除这个标记。

### Q: Ad-hoc 签名和正式签名的区别？
A:
- Ad-hoc 签名（`-`）：免费，但用户需要手动允许
- 正式签名：需要开发者账号，用户可以直接运行

### Q: 如何获取 app-specific-password？
A: 访问 https://appleid.apple.com → 安全 → App 专用密码

### Q: 能否跳过公证？
A: 签名后可以运行，但仍会提示"来自未知开发者"，公证后才能完全消除提示。

---

## 相关链接

- [Wails 打包文档](https://wails.io/docs/guides/building/)
- [Apple 代码签名指南](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [create-dmg 工具](https://github.com/create-dmg/create-dmg)
