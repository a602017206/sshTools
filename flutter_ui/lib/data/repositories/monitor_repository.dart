import '../datasources/remote/api_client.dart';
import '../models/api_response.dart';
import '../models/monitoring_data.dart';
import '../../core/constants/api_constants.dart';

/// 系统监控仓库：拉取某个已建立会话的远端指标快照。
class MonitorRepository {
  final ApiClient _apiClient;

  MonitorRepository(this._apiClient);

  Future<MonitoringData> getMetrics(String sessionId) async {
    final response = await _apiClient.get(
      '${ApiConstants.sessionsEndpoint}/$sessionId/monitor',
    );
    final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(
      response.data as Map<String, dynamic>,
      (json) => json as Map<String, dynamic>,
    );
    return MonitoringData.fromJson(apiResponse.data ?? const {});
  }
}
