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

      _isConnected = true;
      _isConnecting = false;
      _reconnectAttempts = 0;
      _connectionStateController.add(true);

      _logger.i('WebSocket connected successfully');
    } catch (e) {
      _isConnecting = false;
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

  /// Subscribe to a session or transfer
  void subscribe(String target) {
    if (!_isConnected) {
      _logger.w('Cannot subscribe: WebSocket not connected');
      return;
    }

    final message = {
      'action': ApiConstants.wsActionSubscribe,
      'target': target,
    };

    send(message);
    _logger.d('Subscribed to: $target');
  }

  /// Unsubscribe from a session or transfer
  void unsubscribe(String target) {
    if (!_isConnected) {
      _logger.w('Cannot unsubscribe: WebSocket not connected');
      return;
    }

    final message = {
      'action': ApiConstants.wsActionUnsubscribe,
      'target': target,
    };

    send(message);
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
