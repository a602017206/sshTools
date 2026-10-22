import 'dart:io';

import 'package:dio/dio.dart';

import '../datasources/remote/api_client.dart';
import '../models/api_response.dart';
import '../models/sftp_file.dart';
import '../../core/constants/api_constants.dart';

/// SFTP 文件操作仓库：全部挂在某个已建立的 SSH 会话上
/// （接口路径 /sessions/:id/sftp/*，由内嵌 Go 的 API 提供）。
class SftpRepository {
  final ApiClient _apiClient;

  SftpRepository(this._apiClient);

  String _base(String sessionId) =>
      '${ApiConstants.sessionsEndpoint}/$sessionId/sftp';

  Future<ApiResponse<T>> _decode<T>(
    Response<dynamic> response,
    T Function(dynamic json) fromJson,
  ) async {
    return ApiResponse<T>.fromJson(
      response.data as Map<String, dynamic>,
      fromJson,
    );
  }

  /// 列目录；path 为空表示远端会话当前目录（家目录）。
  Future<List<SftpFile>> listFiles(String sessionId, String path) async {
    final response = await _apiClient.get(
      '${_base(sessionId)}/list',
      queryParameters: {if (path.isNotEmpty) 'path': path},
    );
    final apiResponse = await _decode<List<dynamic>>(
      response,
      (json) => json as List<dynamic>,
    );
    final files = (apiResponse.data ?? [])
        .map((e) => SftpFile.fromJson(e as Map<String, dynamic>))
        .toList();
    // 目录在前、同类按名字排序，与桌面文件管理器习惯一致
    files.sort((a, b) {
      if (a.isDir != b.isDir) return a.isDir ? -1 : 1;
      return a.name.toLowerCase().compareTo(b.name.toLowerCase());
    });
    return files;
  }

  Future<String> pwd(String sessionId) async {
    final response = await _apiClient.get('${_base(sessionId)}/pwd');
    final apiResponse = await _decode<Map<String, dynamic>>(
      response,
      (json) => json as Map<String, dynamic>,
    );
    return apiResponse.data?['path'] as String? ?? '';
  }

  Future<void> createDirectory(String sessionId, String path) async {
    await _apiClient.post(
      '${_base(sessionId)}/mkdir',
      data: {'path': path},
    );
  }

  Future<void> rename(String sessionId, String oldPath, String newPath) async {
    await _apiClient.post(
      '${_base(sessionId)}/rename',
      data: {'old_path': oldPath, 'new_path': newPath},
    );
  }

  Future<void> delete(String sessionId, String path) async {
    await _apiClient.delete(
      _base(sessionId),
      queryParameters: {'path': path},
    );
  }

  Future<String> readText(String sessionId, String path) async {
    final response = await _apiClient.get(
      '${_base(sessionId)}/read',
      queryParameters: {'path': path},
    );
    final apiResponse = await _decode<Map<String, dynamic>>(
      response,
      (json) => json as Map<String, dynamic>,
    );
    return apiResponse.data?['content'] as String? ?? '';
  }

  Future<void> writeText(String sessionId, String path, String content) async {
    await _apiClient.put(
      '${_base(sessionId)}/write',
      data: {'path': path, 'content': content},
    );
  }

  /// 上传单个本地文件到远端目录（请求内同步完成，返回远端路径）。
  Future<String> uploadFile(
    String sessionId,
    String remoteDir,
    String localPath,
    String fileName,
  ) async {
    final form = FormData.fromMap({
      'file': await MultipartFile.fromFile(localPath, filename: fileName),
    });
    final response = await _apiClient.dio.post(
      '${_base(sessionId)}/upload',
      queryParameters: {'dir': remoteDir},
      data: form,
    );
    final apiResponse = await _decode<Map<String, dynamic>>(
      response,
      (json) => json as Map<String, dynamic>,
    );
    return apiResponse.data?['remote_path'] as String? ?? '';
  }

  /// 下载远端文件并写入本地 [saveDir]，返回本地完整路径。
  Future<String> downloadFile(
    String sessionId,
    String remotePath,
    String saveDir,
  ) async {
    final response = await _apiClient.dio.get<List<int>>(
      '${_base(sessionId)}/download',
      queryParameters: {'path': remotePath},
      options: Options(responseType: ResponseType.bytes),
    );
    final bytes = response.data;
    if (bytes == null) {
      throw Exception('下载失败：服务器未返回数据');
    }
    final fileName = remotePath.split('/').where((s) => s.isNotEmpty).last;
    final dir = Directory(saveDir);
    await dir.create(recursive: true);
    final file = File('${dir.path}/$fileName');
    await file.writeAsBytes(bytes, flush: true);
    return file.path;
  }
}
