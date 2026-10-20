import '../datasources/remote/api_client.dart';
import '../models/connection_model.dart';
import '../models/api_response.dart';
import '../../core/constants/api_constants.dart';

/// Repository for managing SSH connections
class ConnectionRepository {
  final ApiClient _apiClient;

  ConnectionRepository(this._apiClient);

  /// Get all connections
  Future<List<ConnectionModel>> getConnections() async {
    final response = await _apiClient.get(ApiConstants.connectionsEndpoint);

    final apiResponse = ApiResponse<List<dynamic>>.fromJson(
      response.data,
      (json) => json as List<dynamic>,
    );

    if (apiResponse.data == null) {
      return [];
    }

    return apiResponse.data!
        .map((json) => ConnectionModel.fromJson(json as Map<String, dynamic>))
        .toList();
  }

  /// 把模型转成请求体：表单带了密码才拼 password 字段，
  /// 服务端收到后加密保存（不落 config.json），编辑留空即不修改。
  Map<String, dynamic> _toRequestBody(ConnectionModel connection) {
    return {
      ...connection.toJson(),
      if (connection.password != null && connection.password!.isNotEmpty)
        'password': connection.password,
    };
  }

  /// Add a new connection
  Future<ConnectionModel> addConnection(ConnectionModel connection) async {
    final response = await _apiClient.post(
      ApiConstants.connectionsEndpoint,
      data: _toRequestBody(connection),
    );

    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      response.data,
      (json) => json as Map<String, dynamic>,
    );

    return ConnectionModel.fromJson(apiResponse.data!);
  }

  /// Update an existing connection
  Future<ConnectionModel> updateConnection(ConnectionModel connection) async {
    final response = await _apiClient.put(
      '${ApiConstants.connectionsEndpoint}/${connection.id}',
      data: _toRequestBody(connection),
    );

    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      response.data,
      (json) => json as Map<String, dynamic>,
    );

    return ConnectionModel.fromJson(apiResponse.data!);
  }

  /// Delete a connection
  Future<void> deleteConnection(String id) async {
    await _apiClient.delete('${ApiConstants.connectionsEndpoint}/$id');
  }

  /// Test a connection
  ///
  /// 已保存密码的连接可省略 [authValue]、只带 [connectionId]，
  /// 服务端会回退使用加密存储的密码；临时输入新密码测试成功后
  /// 服务端也会顺手保存（与桌面端行为一致）。
  Future<bool> testConnection({
    required String host,
    required int port,
    required String user,
    required String authType,
    String authValue = '',
    String? connectionId,
    String? passphrase,
  }) async {
    await _apiClient.post(
      '${ApiConstants.connectionsEndpoint}/test',
      data: {
        'host': host,
        'port': port,
        'user': user,
        'auth_type': authType,
        'auth_value': authValue,
        if (connectionId != null && connectionId.isNotEmpty)
          'connection_id': connectionId,
        if (passphrase != null) 'passphrase': passphrase,
      },
    );
    return true;
  }
}
