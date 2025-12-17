/// API configuration constants
class ApiConstants {
  // Base URLs
  static const String defaultBaseUrl = 'http://localhost:8080';
  static const String apiVersion = 'v1';

  // Full API base URL
  static String get apiBaseUrl => '$defaultBaseUrl/api/$apiVersion';

  // WebSocket URL
  static String get wsUrl => 'ws://localhost:8080/api/$apiVersion/ws';

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
