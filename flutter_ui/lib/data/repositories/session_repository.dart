import '../datasources/remote/api_client.dart';
import '../models/api_response.dart';
import '../../core/constants/api_constants.dart';

/// Repository for managing SSH sessions
class SessionRepository {
  final ApiClient _apiClient;

  SessionRepository(this._apiClient);

  /// Connect to SSH server
  Future<String> connectSSH({
    required String sessionId,
    required String host,
    required int port,
    required String user,
    required String authType,
    required String authValue,
    String? passphrase,
    required int cols,
    required int rows,
  }) async {
    final response = await _apiClient.post(
      '${ApiConstants.sessionsEndpoint}/connect',
      data: {
        'session_id': sessionId,
        'host': host,
        'port': port,
        'user': user,
        'auth_type': authType,
        'auth_value': authValue,
        if (passphrase != null) 'passphrase': passphrase,
        'cols': cols,
        'rows': rows,
      },
    );

    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      response.data,
      (json) => json as Map<String, dynamic>,
    );

    return apiResponse.data?['session_id'] as String? ?? sessionId;
  }

  /// Send data to SSH session
  Future<void> sendData(String sessionId, String data) async {
    await _apiClient.post(
      '${ApiConstants.sessionsEndpoint}/$sessionId/send',
      data: {'data': data},
    );
  }

  /// Resize terminal
  Future<void> resizeTerminal(String sessionId, int cols, int rows) async {
    await _apiClient.post(
      '${ApiConstants.sessionsEndpoint}/$sessionId/resize',
      data: {
        'cols': cols,
        'rows': rows,
      },
    );
  }

  /// Close SSH session
  Future<void> closeSession(String sessionId) async {
    await _apiClient.delete('${ApiConstants.sessionsEndpoint}/$sessionId');
  }

  /// List all active sessions
  Future<List<String>> listSessions() async {
    final response = await _apiClient.get(ApiConstants.sessionsEndpoint);

    final apiResponse = ApiResponse<List<dynamic>>.fromJson(
      response.data,
      (json) => json as List<dynamic>,
    );

    if (apiResponse.data == null) {
      return [];
    }

    return apiResponse.data!.map((e) => e.toString()).toList();
  }
}
