import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:logger/logger.dart';
import '../../../core/constants/api_constants.dart';

/// WebSocket client for real-time communication
class WebSocketClient {
  WebSocketChannel? _channel;
  final String url;
  final Logger _logger = Logger();

  // Event handlers mapped by message type
  final Map<String, Function(Map<String, dynamic>)> _handlers = {};

  // 期望保持的订阅集合：未连接时 subscribe 只登记，
  // 连接（含重连）成功后统一补发，避免订阅抢跑丢失早期输出
  final Set<String> _subscriptions = {};

  // Connection state
  bool _isConnected = false;
  bool _isConnecting = false;

  // Reconnection
  Timer? _reconnectTimer;
  int _reconnectAttempts = 0;
  static const int maxReconnectAttempts = 5;
  static const Duration reconnectDelay = Duration(seconds: 3);

  // Stream controller for connection state
  final _connectionStateController = StreamController<bool>.broadcast();

  WebSocketClient({String? url})
      : url = url ?? ApiConstants.wsUrl;

  /// Get connection state stream
  Stream<bool> get connectionState => _connectionStateController.stream;

  /// Check if connected
  bool get isConnected => _isConnected;

  /// 等待 WebSocket 真正就绪（连接中会顺带发起连接）。
  /// 终端在订阅会话、发起 SSH 连接前调用，避免输出在订阅生效前丢失。
  Future<void> ensureConnected({
    Duration timeout = const Duration(seconds: 10),
  }) async {
    if (_isConnected) {
      return;
    }
    final ready = _connectionStateController.stream
        .firstWhere((state) => state)
        .timeout(
          timeout,
          onTimeout: () => throw Exception('WebSocket 连接超时'),
        );
    await connect();
    await ready;
  }

  /// Connect to WebSocket server
  Future<void> connect() async {
    if (_isConnected || _isConnecting) {
      _logger.w('WebSocket already connected or connecting');
      return;
    }

    _isConnecting = true;

    try {
      _logger.i('Connecting to WebSocket: $url');
      _channel = WebSocketChannel.connect(Uri.parse(url));

      // Listen to messages
      _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
        cancelOnError: false,
      );

      // 等待握手真正完成（3.x 的 ready），失败会走 catch 安排重连
      await _channel!.ready;

      _isConnected = true;
      _isConnecting = false;
      _reconnectAttempts = 0;
      _connectionStateController.add(true);

      // 补发断线期间登记的订阅
      for (final target in _subscriptions) {
        _sendSubscribe(target);
      }

      _logger.i('WebSocket connected successfully');
    } catch (e) {
      _isConnecting = false;
      _isConnected = false;
      _logger.e('WebSocket connection error: $e');
      _scheduleReconnect();
    }
  }

  /// Handle incoming messages
  void _onMessage(dynamic message) {
    try {
      final data = jsonDecode(message as String) as Map<String, dynamic>;
      final type = data['type'] as String?;

      if (type != null && _handlers.containsKey(type)) {
        _handlers[type]!(data);
      } else {
        _logger.w('No handler for message type: $type');
      }
    } catch (e) {
      _logger.e('Error parsing WebSocket message: $e');
    }
  }

  /// Handle errors
  void _onError(error) {
    _logger.e('WebSocket error: $error');
    _isConnected = false;
    _connectionStateController.add(false);
  }

  /// Handle connection close
  void _onDone() {
    _logger.w('WebSocket connection closed');
    _isConnected = false;
    _connectionStateController.add(false);
    _scheduleReconnect();
  }

  /// Schedule reconnection
  void _scheduleReconnect() {
    if (_reconnectAttempts >= maxReconnectAttempts) {
      _logger.e('Max reconnection attempts reached');
      return;
    }

    _reconnectAttempts++;
    _logger.i('Scheduling reconnect attempt $_reconnectAttempts');

    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(reconnectDelay, () {
      connect();
    });
  }

  /// Subscribe to a session or transfer.
  /// 未连接时只登记，等连接（含重连）成功后自动补发。
  void subscribe(String target) {
    _subscriptions.add(target);
    if (_isConnected) {
      _sendSubscribe(target);
    }
    _logger.d('Subscribed to: $target');
  }

  void _sendSubscribe(String target) {
    send({
      'action': ApiConstants.wsActionSubscribe,
      'target': target,
    });
  }

  /// Unsubscribe from a session or transfer
  void unsubscribe(String target) {
    _subscriptions.remove(target);
    if (!_isConnected) {
      return;
    }

    send({
      'action': ApiConstants.wsActionUnsubscribe,
      'target': target,
    });
    _logger.d('Unsubscribed from: $target');
  }

  /// Register event handler for a message type
  void onEvent(String type, Function(Map<String, dynamic>) handler) {
    _handlers[type] = handler;
  }

  /// Remove event handler
  void removeHandler(String type) {
    _handlers.remove(type);
  }

  /// Send message to server
  void send(Map<String, dynamic> message) {
    if (!_isConnected) {
      _logger.w('Cannot send: WebSocket not connected');
      return;
    }

    try {
      _channel?.sink.add(jsonEncode(message));
    } catch (e) {
      _logger.e('Error sending message: $e');
    }
  }

  /// Close WebSocket connection
  void close() {
    _reconnectTimer?.cancel();
    _subscriptions.clear();
    _channel?.sink.close();
    _isConnected = false;
    _connectionStateController.add(false);
    _logger.i('WebSocket closed');
  }

  /// Dispose resources
  void dispose() {
    close();
    _connectionStateController.close();
  }
}
