/// API configuration constants
class ApiConstants {
  // Base URLs
  //
  // 桌面调试时默认连独立运行的 Go API 服务器（cmd/apiserver）；
  // 移动端 App 启动时会通过 [configureEmbedded] 改为内嵌 Go 服务器的
  // 本机回环地址（127.0.0.1 + 随机端口）。
  static String defaultBaseUrl = 'http://localhost:8080';
  static const String apiVersion = 'v1';

  /// 把 API 基地址切换为内嵌 Go 服务器（移动端启动成功后调用）。
  static void configureEmbedded(int port) {
    defaultBaseUrl = 'http://127.0.0.1:$port';
  }

  /// 当前是否走内嵌 Go 服务器。
  static bool get isEmbedded => defaultBaseUrl.startsWith('http://127.0.0.1:');

  // Full API base URL
  static String get apiBaseUrl => '$defaultBaseUrl/api/$apiVersion';

  // WebSocket URL（从当前基地址推导，保证内嵌模式端口一致）
  static String get wsUrl =>
      '${defaultBaseUrl.replaceFirst('http', 'ws')}/api/$apiVersion/ws';

  // Timeouts
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
  static const Duration sendTimeout = Duration(seconds: 30);

  // Endpoints
  static const String healthEndpoint = '/health';
  static const String connectionsEndpoint = '/connections';
  static const String sessionsEndpoint = '/sessions';
  static const String sftpEndpoint = '/sftp';
  static const String monitorEndpoint = '/monitor';
  static const String settingsEndpoint = '/settings';
  static const String credentialsEndpoint = '/credentials';

  // WebSocket Actions
  static const String wsActionSubscribe = 'subscribe';
  static const String wsActionUnsubscribe = 'unsubscribe';

  // WebSocket Message Types
  static const String wsTypeSSHOutput = 'ssh:output';
  static const String wsTypeTransferProgress = 'transfer:progress';
}
