import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';
import 'package:xterm/xterm.dart';

import '../../../data/models/connection_model.dart';
import '../../providers/api_providers.dart';

enum _SessionStatus { connecting, connected, disconnected, error }

/// SSH 终端页：xterm 渲染 + WebSocket 收输出 + HTTP 发输入/尺寸。
///
/// 数据通路与桌面端一致地复用既有 API：
/// 先经 WebSocket 订阅会话，再 POST /sessions/connect（凭 connection_id
/// 回退已保存密码开会话），远端 PTY 输出经 ws 的 ssh:output 回推渲染；
/// 键入经 POST /sessions/:id/send 回传。退出页面即关会话。
class TerminalScreen extends ConsumerStatefulWidget {
  final ConnectionModel connection;

  /// 临时输入的凭据（尚未保存密码的连接）；为空时服务端回退已保存密码。
  final String authValue;

  const TerminalScreen({
    super.key,
    required this.connection,
    this.authValue = '',
  });

  @override
  ConsumerState<TerminalScreen> createState() => _TerminalScreenState();
}

class _TerminalScreenState extends ConsumerState<TerminalScreen> {
  final Terminal _terminal = Terminal(maxLines: 10000);
  final FocusNode _focusNode = FocusNode();

  String _sessionId = '';
  _SessionStatus _status = _SessionStatus.connecting;
  String? _errorText;
  bool _ctrlActive = false;
  bool _disposed = false;

  @override
  void initState() {
    super.initState();

    _terminal.onOutput = (data) {
      var out = data;
      // Ctrl 粘滞键：激活后下一个字母转成对应控制字符（a->^A ... z->^Z）
      if (_ctrlActive && data.length == 1) {
        final code = data.toLowerCase().codeUnitAt(0);
        if (code >= 0x61 && code <= 0x7a) {
          out = String.fromCharCode(code - 0x60);
        }
        setState(() => _ctrlActive = false);
      }
      unawaited(_send(out));
    };
    _terminal.onResize = (width, height, pixelWidth, pixelHeight) {
      if (_status == _SessionStatus.connected) {
        unawaited(
          ref
              .read(sessionRepositoryProvider)
              .resizeTerminal(_sessionId, width, height)
              .catchError((_) {}),
        );
      }
    };

    unawaited(_startSession());
  }

  Future<void> _startSession() async {
    final ws = ref.read(webSocketClientProvider);
    final sessionRepo = ref.read(sessionRepositoryProvider);

    // 重连时先收尾旧会话
    if (_sessionId.isNotEmpty) {
      ws.unsubscribe(_sessionId);
      unawaited(sessionRepo.closeSession(_sessionId).catchError((_) {}));
    }

    final sessionId = 'mobile-${const Uuid().v4()}';
    _sessionId = sessionId;

    if (mounted) {
      setState(() {
        _status = _SessionStatus.connecting;
        _errorText = null;
      });
    }
    _terminal.write('正在连接 ${widget.connection.user}@${widget.connection.host}:${widget.connection.port} ...\r\n');

    ws.onEvent('ssh:output', _onWsEvent);
    ws.onEvent('ssh:closed', _onWsEvent);

    try {
      // 先把 WS 订上再开会话，避免登录横幅等早期输出丢失
      await ws.ensureConnected();
      ws.subscribe(sessionId);

      await sessionRepo.connectSSH(
        sessionId: sessionId,
        host: widget.connection.host,
        port: widget.connection.port,
        user: widget.connection.user,
        authType: widget.connection.authType,
        authValue: widget.authValue,
        connectionId: widget.connection.id,
        cols: _terminal.viewWidth,
        rows: _terminal.viewHeight,
      );

      if (!mounted || _disposed) return;
      setState(() => _status = _SessionStatus.connected);
      _focusNode.requestFocus();
    } catch (e) {
      if (!mounted || _disposed) return;
      setState(() {
        _status = _SessionStatus.error;
        _errorText = e.toString();
      });
      _terminal.write('\r\n\x1b[31m连接失败：$e\x1b[0m\r\n');
    }
  }

  void _onWsEvent(Map<String, dynamic> message) {
    if (message['session_id'] != _sessionId || !mounted) {
      return;
    }
    switch (message['type']) {
      case 'ssh:output':
        final data = message['data'];
        if (data is String && data.isNotEmpty) {
          _terminal.write(data);
        }
        break;
      case 'ssh:closed':
        _terminal.write('\r\n\x1b[33m[连接已断开]\x1b[0m\r\n');
        setState(() => _status = _SessionStatus.disconnected);
        break;
    }
  }

  Future<void> _send(String data) async {
    if (_status != _SessionStatus.connected || data.isEmpty) {
      return;
    }
    try {
      await ref.read(sessionRepositoryProvider).sendData(_sessionId, data);
    } catch (_) {
      // 发送失败通常伴随 ssh:closed 事件，统一在那里提示
    }
  }

  void _sendKey(String data) {
    unawaited(_send(data));
    _focusNode.requestFocus();
  }

  @override
  void dispose() {
    _disposed = true;
    final ws = ref.read(webSocketClientProvider);
    ws.removeHandler('ssh:output');
    ws.removeHandler('ssh:closed');
    if (_sessionId.isNotEmpty) {
      ws.unsubscribe(_sessionId);
      unawaited(
        ref.read(sessionRepositoryProvider).closeSession(_sessionId).catchError((_) {}),
      );
    }
    _focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final conn = widget.connection;
    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(conn.name, style: const TextStyle(fontSize: 16)),
            Text(
              '${conn.user}@${conn.host}:${conn.port} · ${_statusLabel()}',
              style: const TextStyle(fontSize: 12),
            ),
          ],
        ),
        actions: [
          Icon(
            Icons.circle,
            size: 12,
            color: switch (_status) {
              _SessionStatus.connected => Colors.green,
              _SessionStatus.connecting => Colors.orange,
              _ => Colors.red,
            },
          ),
          const SizedBox(width: 8),
          if (_status == _SessionStatus.disconnected ||
              _status == _SessionStatus.error)
            IconButton(
              icon: const Icon(Icons.refresh),
              tooltip: '重新连接',
              onPressed: () => unawaited(_startSession()),
            ),
          IconButton(
            icon: const Icon(Icons.close),
            tooltip: '断开并返回',
            onPressed: () => Navigator.pop(context),
          ),
        ],
      ),
      body: Column(
        children: [
          if (_status == _SessionStatus.connecting)
            const LinearProgressIndicator(),
          if (_status == _SessionStatus.error && _errorText != null)
            Container(
              width: double.infinity,
              color: Colors.red.withValues(alpha: 0.15),
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              child: Text(
                _errorText!,
                style: const TextStyle(color: Colors.red, fontSize: 12),
              ),
            ),
          Expanded(
            child: TerminalView(
              _terminal,
              focusNode: _focusNode,
              autofocus: true,
              // visiblePassword 让 Android 键盘关闭联想/自动纠正，
              // 避免命令被改写（xterm 未单独暴露 enableSuggestions）
              keyboardType: TextInputType.visiblePassword,
              textStyle: const TerminalStyle(fontSize: 13),
              padding: const EdgeInsets.all(4),
            ),
          ),
          _buildKeyBar(),
        ],
      ),
    );
  }

  String _statusLabel() {
    return switch (_status) {
      _SessionStatus.connecting => '连接中',
      _SessionStatus.connected => '已连接',
      _SessionStatus.disconnected => '已断开',
      _SessionStatus.error => '连接失败',
    };
  }

  /// 移动端软键盘没有 Esc/Ctrl/方向键，补一条特殊按键栏；
  /// Ctrl 为粘滞键：点亮后下一个字母键转控制字符。
  Widget _buildKeyBar() {
    final keys = <(String, String)>[
      ('Esc', '\x1b'),
      ('Tab', '\t'),
      ('↑', '\x1b[A'),
      ('↓', '\x1b[B'),
      ('→', '\x1b[C'),
      ('←', '\x1b[D'),
      ('Home', '\x1b[H'),
      ('End', '\x1b[F'),
      ('PgUp', '\x1b[5~'),
      ('PgDn', '\x1b[6~'),
      ('/', '/'),
      ('|', '|'),
      ('~', '~'),
      ('-', '-'),
      ('^C', '\x03'),
      ('^Z', '\x1a'),
      ('^D', '\x04'),
      ('^L', '\x0c'),
    ];

    return Material(
      color: Theme.of(context).colorScheme.surfaceContainerHighest,
      child: SafeArea(
        top: false,
        child: SizedBox(
          height: 44,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 6),
            children: [
              Padding(
                padding: const EdgeInsets.only(right: 6),
                child: FilledButton.tonal(
                  style: FilledButton.styleFrom(
                    padding: const EdgeInsets.symmetric(horizontal: 12),
                    backgroundColor: _ctrlActive
                        ? Theme.of(context).colorScheme.primary
                        : null,
                    foregroundColor: _ctrlActive
                        ? Theme.of(context).colorScheme.onPrimary
                        : null,
                  ),
                  onPressed: () {
                    setState(() => _ctrlActive = !_ctrlActive);
                    _focusNode.requestFocus();
                  },
                  child: const Text('Ctrl'),
                ),
              ),
              for (final (label, data) in keys)
                Padding(
                  padding: const EdgeInsets.only(right: 6),
                  child: FilledButton.tonal(
                    style: FilledButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                    ),
                    onPressed: () => _sendKey(data),
                    child: Text(label),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
