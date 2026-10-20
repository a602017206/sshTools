import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/datasources/remote/api_client.dart';
import '../../data/datasources/remote/websocket_client.dart';
import '../../data/repositories/connection_repository.dart';
import '../../data/repositories/session_repository.dart';
import '../../data/repositories/sftp_repository.dart';

/// API Client Provider
final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient();
});

/// WebSocket Client Provider
final webSocketClientProvider = Provider<WebSocketClient>((ref) {
  final client = WebSocketClient();
  // Auto-connect when provider is created
  client.connect();

  // Cleanup on dispose
  ref.onDispose(() {
    client.dispose();
  });

  return client;
});

/// Connection Repository Provider
final connectionRepositoryProvider = Provider<ConnectionRepository>((ref) {
  return ConnectionRepository(ref.watch(apiClientProvider));
});

/// Session Repository Provider
final sessionRepositoryProvider = Provider<SessionRepository>((ref) {
  return SessionRepository(ref.watch(apiClientProvider));
});

/// SFTP Repository Provider
final sftpRepositoryProvider = Provider<SftpRepository>((ref) {
  return SftpRepository(ref.watch(apiClientProvider));
});
