# macOS Flutter应用设置指南

## 问题总结

浏览器版本可以正常访问后端API，但macOS原生应用无法连接。

## 已解决的问题

### 1. CocoaPods环境问题
**问题**: Ruby版本冲突导致CocoaPods无法正常工作
- 系统有多个Ruby版本（系统2.6.10、RVM 3.3.4、Homebrew 3.4.8）
- CocoaPods依赖缺失

**解决方案**:
- 使用RVM的Ruby 3.3.4环境
- 重新安装CocoaPods及其依赖
- 修复Xcode配置文件以包含Pods配置

### 2. macOS网络权限问题
**问题**: macOS沙盒应用缺少网络客户端权限
- 应用无法作为客户端发起HTTP请求
- 浏览器版本正常是因为浏览器本身有网络权限

**解决方案**:
在entitlements文件中添加网络客户端权限：
```xml
<key>com.apple.security.network.client</key>
<true/>
```

修改的文件：
- `macos/Runner/DebugProfile.entitlements`
- `macos/Runner/Release.entitlements`

### 3. HTTP代理问题
**问题**: 系统配置了HTTP代理(127.0.0.1:7890)，导致localhost请求被代理
- 通过代理访问localhost会失败

**解决方案**:
在`lib/data/datasources/remote/api_client.dart`中配置Dio绕过localhost代理：
```dart
(_dio.httpClientAdapter as DefaultHttpClientAdapter).onHttpClientCreate = (client) {
  client.findProxy = (uri) {
    if (uri.host == 'localhost' || uri.host == '127.0.0.1') {
      return 'DIRECT';
    }
    return 'DIRECT';
  };
  return client;
};
```

## 运行应用

### 方法1: 使用便捷脚本（推荐）
```bash
cd flutter_ui
./run_macos.sh
```

### 方法2: 手动运行
```bash
cd flutter_ui

# 设置Ruby环境
source ~/.rvm/scripts/rvm
rvm use 3.3.4
export PATH="$HOME/.rvm/gems/ruby-3.3.4/bin:$HOME/.rvm/rubies/ruby-3.3.4/bin:$PATH"

# 运行Flutter应用
/Users/dingwei/development/flutter/bin/flutter run -d macos
```

## 前置条件

1. **Go后端服务器必须运行**
   ```bash
   cd /Users/dingwei/go/sshTools
   wails dev
   # 或
   go run .
   ```
   后端会在http://localhost:8080运行

2. **Flutter SDK已安装**
   - 路径: `/Users/dingwei/development/flutter`

3. **RVM和Ruby 3.3.4已安装**

## 文件修改清单

1. **网络权限** (已修改)
   - `macos/Runner/DebugProfile.entitlements`
   - `macos/Runner/Release.entitlements`

2. **CocoaPods配置** (已修改)
   - `macos/Runner/Configs/Debug.xcconfig`
   - `macos/Runner/Configs/Release.xcconfig`

3. **代理配置** (已修改)
   - `lib/data/datasources/remote/api_client.dart`

4. **启动脚本** (新建)
   - `run_macos.sh`

## 验证后端连接

测试后端API是否可访问：
```bash
# 绕过代理测试
curl --noproxy "*" http://localhost:8080/api/v1/connections

# 应该返回类似这样的JSON:
# {"data":[{"id":"conn_xxx","name":"xxx",...}]}
```

## 常见问题

### Q: 应用启动但无法连接后端
A: 检查Go后端服务器是否在运行（端口8080）

### Q: CocoaPods错误
A: 确保使用RVM的Ruby环境运行Flutter命令

### Q: 构建失败
A: 尝试清理并重新构建：
```bash
flutter clean
cd macos
pod install
cd ..
flutter run -d macos
```

## 技术细节

### 为什么需要网络客户端权限？
macOS沙盒应用默认没有网络访问权限。必须在entitlements中显式声明：
- `com.apple.security.network.client` - 允许应用作为客户端发起网络请求
- `com.apple.security.network.server` - 允许应用作为服务器接收请求

### 为什么浏览器版本正常？
Web应用运行在浏览器中，浏览器已经有完整的网络权限。但原生macOS应用在沙盒中运行，需要明确声明每个权限。

### 代理配置的作用？
系统配置的HTTP代理会影响所有网络请求，包括localhost。通过配置Dio的`findProxy`，我们让应用直接访问localhost，不经过代理。
