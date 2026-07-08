import 'dart:io';

import 'package:dio/dio.dart';
import 'package:dio/io.dart';
import 'package:logger/logger.dart';
import '../../../core/constants/api_constants.dart';

/// HTTP API client using Dio
class ApiClient {
  late final Dio _dio;
  final Logger _logger = Logger();

  ApiClient({String? baseUrl}) {
    _dio = Dio(
      BaseOptions(
        baseUrl: baseUrl ?? ApiConstants.apiBaseUrl,
        connectTimeout: ApiConstants.connectTimeout,
        receiveTimeout: ApiConstants.receiveTimeout,
        sendTimeout: ApiConstants.sendTimeout,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ),
    );

    // Configure HTTP client adapter to bypass proxy for localhost
    // This fixes issues when system has a proxy configured (e.g., 127.0.0.1:7890)
    (_dio.httpClientAdapter as DefaultHttpClientAdapter).onHttpClientCreate =
        (client) {
      client.findProxy = (uri) {
        // Always bypass proxy for localhost and 127.0.0.1
        if (uri.host == 'localhost' || uri.host == '127.0.0.1') {
          return 'DIRECT';
        }
        // For other hosts, use DIRECT (no proxy) as well
        // Can be changed to use system proxy if needed
        return 'DIRECT';
      };
      return client;
    };

    // Add interceptors
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          _logger.d('REQUEST[${options.method}] => ${options.uri}');
          return handler.next(options);
        },
        onResponse: (response, handler) {
          _logger.d(
            'RESPONSE[${response.statusCode}] => ${response.requestOptions.uri}',
          );
          return handler.next(response);
        },
        onError: (error, handler) {
          _logger.e(
            'ERROR[${error.response?.statusCode}] => ${error.requestOptions.uri}',
            error: error.message,
          );
          return handler.next(error);
        },
      ),
    );
  }

  /// GET request
  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      return await _dio.get<T>(
        path,
        queryParameters: queryParameters,
        options: options,
      );
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  /// POST request
  Future<Response<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      return await _dio.post<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  /// PUT request
  Future<Response<T>> put<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      return await _dio.put<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  /// DELETE request
  Future<Response<T>> delete<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      return await _dio.delete<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  /// Handle Dio errors
  Exception _handleError(DioException error) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return TimeoutException('Connection timeout');
      case DioExceptionType.badResponse:
        final statusCode = error.response?.statusCode;
        final message = error.response?.data?['error'] ?? 'Unknown error';
        return ApiException(statusCode ?? 500, message);
      case DioExceptionType.cancel:
        return RequestCancelledException('Request cancelled');
      case DioExceptionType.connectionError:
        return ConnectionException('No internet connection');
      default:
        return UnknownException('Unknown error occurred');
    }
  }

  /// Get the underlying Dio instance
  Dio get dio => _dio;
}

/// Base exception for API errors
abstract class ApiClientException implements Exception {
  final String message;
  ApiClientException(this.message);

  @override
  String toString() => message;
}

class ApiException extends ApiClientException {
  final int statusCode;
  ApiException(this.statusCode, String message) : super(message);

  @override
  String toString() => '[$statusCode] $message';
}

class TimeoutException extends ApiClientException {
  TimeoutException(String message) : super(message);
}

class ConnectionException extends ApiClientException {
  ConnectionException(String message) : super(message);
}

class RequestCancelledException extends ApiClientException {
  RequestCancelledException(String message) : super(message);
}

class UnknownException extends ApiClientException {
  UnknownException(String message) : super(message);
}
