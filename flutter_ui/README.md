# SSH Tools - Flutter Frontend

Flutter 全平台前端，支持 iOS、Android、macOS、Windows、Linux 和 Web。

## 架构

### 分层架构
```
lib/
├── core/                    # 核心基础设施
│   ├── constants/           # API 和应用常量
│   ├── theme/               # 主题配置
│   └── utils/               # 工具函数
│
├── data/                    # 数据层
│   ├── models/              # 数据模型（JSON 序列化）
│   ├── datasources/         # 数据源
│   │   ├── remote/          # 远程 API (HTTP + WebSocket)
│   │   └── local/           # 本地存储
│   └── repositories/        # 数据仓库（业务逻辑抽象）
│
└── presentation/            # 表现层
    ├── providers/           # Riverpod 状态管理
    ├── screens/             # 页面
    └── widgets/             # UI 组件
```

### 技术栈
- **状态管理**: Riverpod 2.x
- **HTTP 客户端**: Dio 5.x
- **WebSocket**: web_socket_channel 2.x
- **终端**: xterm 3.x
- **路由**: go_router 12.x
- **本地存储**: shared_preferences + flutter_secure_storage
- **图表**: fl_chart 0.65.x

## 快速开始

### 前提条件
1. **Flutter SDK** >= 3.0.0
2. **Go API Server** 运行在 `http://localhost:8080`

### 安装依赖

```bash
cd flutter_ui
flutter pub get
```

### 运行代码生成（如果需要）

```bash
# 生成 JSON 序列化代码
flutter pub run build_runner build --delete-conflicting-outputs
```

### 运行应用

#### 移动端（iOS/Android）
```bash
# iOS (需要 macOS)
flutter run -d ios

# Android
flutter run -d android
```

#### 桌面端
```bash
# macOS
flutter run -d macos

# Windows
flutter run -d windows

# Linux
flutter run -d linux
```

#### Web
```bash
flutter run -d chrome
```

## 配置

### API 端点配置

编辑 `lib/core/constants/api_constants.dart`：

```dart
class ApiConstants {
  static const String defaultBaseUrl = 'http://localhost:8080';
  static String get wsUrl => 'ws://localhost:8080/api/v1/ws';
}
```

**生产环境**：将 `localhost` 替换为实际的服务器地址。

## 项目状态

### 已完成 ✅
- [x] 项目结构搭建
- [x] 核心常量和配置
- [x] 数据模型（Connection, Session, ApiResponse）
- [x] HTTP 客户端（Dio 封装）
- [x] WebSocket 客户端（自动重连）
- [x] Repository 层（Connection, Session）
- [x] Riverpod Providers
- [x] 基础 UI 框架
- [x] 路由配置（go_router）
- [x] **连接管理 UI（添加/编辑/删除/测试）**

### 待实现 🚧
- [ ] SSH 终端 UI（xterm 集成）
- [ ] SFTP 文件管理器
- [ ] 系统监控面板
- [ ] 设置页面
- [ ] 凭证存储集成（密码输入）
- [ ] 完整的错误处理和加载状态

## 使用示例

### 1. 连接管理

```dart
// 获取所有连接
final connections = await ref.read(connectionRepositoryProvider).getConnections();

// 添加连接
final newConn = ConnectionModel(
  id: '',
  name: 'My Server',
  host: '192.168.1.100',
  port: 22,
  user: 'root',
  authType: 'password',
);
await ref.read(connectionRepositoryProvider).addConnection(newConn);
```

### 2. SSH 会话

```dart
// 建立 SSH 连接
final sessionId = 'session_${DateTime.now().millisecondsSinceEpoch}';
await ref.read(sessionRepositoryProvider).connectSSH(
  sessionId: sessionId,
  host: '192.168.1.100',
  port: 22,
  user: 'root',
  authType: 'password',
  authValue: 'password123',
  cols: 80,
  rows: 24,
);

// 通过 WebSocket 接收 SSH 输出
final ws = ref.read(webSocketClientProvider);
ws.onEvent('ssh:output', (data) {
  if (data['session_id'] == sessionId) {
    print('SSH Output: ${data['data']}');
  }
});
ws.subscribe(sessionId);

// 发送命令
await ref.read(sessionRepositoryProvider).sendData(sessionId, 'ls -la\n');
```

### 3. WebSocket 实时通信

```dart
final ws = ref.read(webSocketClientProvider);

// 监听连接状态
ws.connectionState.listen((isConnected) {
  print('WebSocket connected: $isConnected');
});

// 订阅会话输出
ws.onEvent('ssh:output', (data) {
  final sessionId = data['session_id'];
  final output = data['data'];
  // 显示在终端
});

// 订阅传输进度
ws.onEvent('transfer:progress', (data) {
  final progress = data['data'];
  // 更新进度条
});
```

### 4. 连接管理 UI

```dart
// 使用连接列表
final connectionState = ref.watch(connectionProvider);

// 加载连接
if (connectionState.isLoading) {
  return CircularProgressIndicator();
}

// 显示连接
final connections = connectionState.connections;

// 添加连接
final notifier = ref.read(connectionProvider.notifier);
await notifier.addConnection(newConnection);

// 测试连接
final success = await notifier.testConnection(
  host: '192.168.1.100',
  port: 22,
  user: 'root',
  authType: 'password',
  authValue: 'password123',
);
```

**查看详细文档**: [CONNECTION_UI_IMPLEMENTATION.md](CONNECTION_UI_IMPLEMENTATION.md)

## 平台特定配置

### iOS
需要在 `Info.plist` 中添加网络权限（已在模板中）。

### Android
需要在 `AndroidManifest.xml` 中添加网络权限（已在模板中）。

### macOS
需要在 `DebugProfile.entitlements` 和 `Release.entitlements` 中启用网络权限（已在模板中）。

## 开发指南

### 添加新页面
1. 在 `lib/presentation/screens/` 创建新目录
2. 创建 `screen_name_screen.dart`
3. 在 `widgets/` 子目录添加组件

### 添加新数据模型
1. 在 `lib/data/models/` 创建 `model_name.dart`
2. 使用 `@JsonSerializable()` 注解
3. 运行代码生成：`flutter pub run build_runner build`

### 添加新 Repository
1. 在 `lib/data/repositories/` 创建 `repository_name.dart`
2. 注入 `ApiClient` 依赖
3. 在 `api_providers.dart` 中添加 Provider

## 故障排查

### 无法连接到 API 服务器
- 确认 Go API 服务器已启动：`./build/bin/sshtools-api`
- 检查 `api_constants.dart` 中的 URL 配置
- 检查防火墙设置

### WebSocket 连接失败
- 确认 WebSocket URL 正确（`ws://` 而不是 `wss://`）
- 检查服务器是否支持 WebSocket
- 查看浏览器/应用控制台日志

### 代码生成错误
```bash
# 清除旧的生成文件
flutter pub run build_runner clean

# 重新生成
flutter pub run build_runner build --delete-conflicting-outputs
```

## 测试

```bash
# 运行所有测试
flutter test

# 运行特定测试
flutter test test/data/repositories/connection_repository_test.dart

# 生成测试覆盖率
flutter test --coverage
```

## 构建发布版本

### Android APK
```bash
flutter build apk --release
# 输出: build/app/outputs/flutter-apk/app-release.apk
```

### iOS IPA (需要 Apple Developer 账号)
```bash
flutter build ios --release
# 然后在 Xcode 中 Archive
```

### macOS App
```bash
flutter build macos --release
# 输出: build/macos/Build/Products/Release/sshtools_flutter.app
```

### Windows
```bash
flutter build windows --release
# 输出: build/windows/runner/Release/
```

### Web
```bash
flutter build web --release
# 输出: build/web/
```

## 许可证

与主项目相同。
