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

  /// Add a new connection
  Future<ConnectionModel> addConnection(ConnectionModel connection) async {
    final response = await _apiClient.post(
      ApiConstants.connectionsEndpoint,
      data: connection.toJson(),
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
      data: connection.toJson(),
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
  Future<bool> testConnection({
    required String host,
    required int port,
    required String user,
    required String authType,
    required String authValue,
    String? passphrase,
  }) async {
    try {
      await _apiClient.post(
        '${ApiConstants.connectionsEndpoint}/test',
        data: {
          'host': host,
          'port': port,
          'user': user,
          'auth_type': authType,
          'auth_value': authValue,
          if (passphrase != null) 'passphrase': passphrase,
        },
      );
      return true;
    } catch (e) {
      rethrow;
    }
  }
}
