import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';
import 'package:xterm/xterm.dart';

import '../../providers/api_providers.dart';
import '../../providers/terminal_tabs_provider.dart';

/// 单个终端标签页：xterm 渲染 + WebSocket 收输出 + HTTP 发输入/尺寸。
///
/// 数据通路与桌面端一致地复用既有 API：先经 WebSocket 订阅会话，
/// 再 POST /sessions/connect（凭 connection_id 回退已保存密码开会话），
/// 远端 PTY 输出经 ws 的 ssh:output 回推渲染；键入经
/// POST /sessions/:id/send 回传。
///
/// 生命周期绑定标签：组件在 IndexedStack 里保活，切栏/切标签不销毁；
/// 只有关闭标签（组件真正 dispose）时才退订并关闭 SSH 会话。
/// 连接状态与会话 ID 通过 terminalTabsProvider 回报给标签栏。
class TerminalTabView extends ConsumerStatefulWidget {
  final TerminalTab tab;

  /// 是否为当前激活标签（切到本标签时抢焦点弹键盘）
  final bool isActive;

  const TerminalTabView({
    super.key,
    required this.tab,
    required this.isActive,
  });

  @override
  ConsumerState<TerminalTabView> createState() => _TerminalTabViewState();
}

class _TerminalTabViewState extends ConsumerState<TerminalTabView> {
  final Terminal _terminal = Terminal(maxLines: 10000);
  final FocusNode _focusNode = FocusNode();

  String _sessionId = '';
  TerminalTabStatus _status = TerminalTabStatus.connecting;
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
      if (_status == TerminalTabStatus.connected) {
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

  @override
  void didUpdateWidget(TerminalTabView oldWidget) {
    super.didUpdateWidget(oldWidget);
    // 切到本标签时把焦点交给终端，软键盘随之弹起
    if (widget.isActive && !oldWidget.isActive) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          _focusNode.requestFocus();
        }
      });
    }
  }

  void _reportStatus(TerminalTabStatus status) {
    if (!mounted || _disposed) {
      return;
    }
    setState(() => _status = status);
    ref.read(terminalTabsProvider.notifier).setStatus(widget.tab.id, status);
  }

  Future<void> _startSession() async {
    final ws = ref.read(webSocketClientProvider);
    final sessionRepo = ref.read(sessionRepositoryProvider);

    // 重连时先收尾旧会话（处理器只注册一次，靠 _sessionId 过滤）
    if (_sessionId.isNotEmpty) {
      ws.unsubscribe(_sessionId);
      unawaited(sessionRepo.closeSession(_sessionId).catchError((_) {}));
    } else {
      ws.onEvent('ssh:output', _onWsEvent);
      ws.onEvent('ssh:closed', _onWsEvent);
    }

    final sessionId = 'mobile-${const Uuid().v4()}';
    _sessionId = sessionId;
    ref
        .read(terminalTabsProvider.notifier)
        .setSessionId(widget.tab.id, sessionId);

    if (mounted) {
      setState(() {
        _status = TerminalTabStatus.connecting;
        _errorText = null;
      });
    }
    ref
        .read(terminalTabsProvider.notifier)
        .setStatus(widget.tab.id, TerminalTabStatus.connecting);
    final conn = widget.tab.connection;
    _terminal.write('正在连接 ${conn.user}@${conn.host}:${conn.port} ...\r\n');

    try {
      // 先把 WS 订上再开会话，避免登录横幅等早期输出丢失
      await ws.ensureConnected();
      ws.subscribe(sessionId);

      await sessionRepo.connectSSH(
        sessionId: sessionId,
        host: conn.host,
        port: conn.port,
        user: conn.user,
        authType: conn.authType,
        authValue: widget.tab.authValue,
        connectionId: conn.id,
        cols: _terminal.viewWidth,
        rows: _terminal.viewHeight,
      );

      if (!mounted || _disposed) return;
      _reportStatus(TerminalTabStatus.connected);
      if (widget.isActive) {
        _focusNode.requestFocus();
      }
    } catch (e) {
      if (!mounted || _disposed) return;
      setState(() => _errorText = e.toString());
      _reportStatus(TerminalTabStatus.error);
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
        _reportStatus(TerminalTabStatus.disconnected);
        break;
    }
  }

  Future<void> _send(String data) async {
    if (_status != TerminalTabStatus.connected || data.isEmpty) {
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
    // 只摘自己的处理器，不影响其它标签
    ws.removeHandler('ssh:output', _onWsEvent);
    ws.removeHandler('ssh:closed', _onWsEvent);
    if (_sessionId.isNotEmpty) {
      ws.unsubscribe(_sessionId);
      unawaited(
        ref
            .read(sessionRepositoryProvider)
            .closeSession(_sessionId)
            .catchError((_) {}),
      );
    }
    _focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        if (_status == TerminalTabStatus.connecting)
          const LinearProgressIndicator(),
        if (_status == TerminalTabStatus.error && _errorText != null)
          Container(
            width: double.infinity,
            color: Colors.red.withValues(alpha: 0.15),
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    _errorText!,
                    style: const TextStyle(color: Colors.red, fontSize: 12),
                  ),
                ),
                TextButton(
                  onPressed: () => unawaited(_startSession()),
                  child: const Text('重试'),
                ),
              ],
            ),
          ),
        if (_status == TerminalTabStatus.disconnected)
          Container(
            width: double.infinity,
            color: Colors.orange.withValues(alpha: 0.15),
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
            child: Row(
              children: [
                const Expanded(
                  child: Text(
                    '连接已断开',
                    style: TextStyle(color: Colors.orange, fontSize: 12),
                  ),
                ),
                TextButton(
                  onPressed: () => unawaited(_startSession()),
                  child: const Text('重新连接'),
                ),
              ],
            ),
          ),
        Expanded(
          child: TerminalView(
            _terminal,
            focusNode: _focusNode,
            autofocus: widget.isActive,
            // visiblePassword 让 Android 键盘关闭联想/自动纠正，
            // 避免命令被改写（xterm 未单独暴露 enableSuggestions）
            keyboardType: TextInputType.visiblePassword,
            textStyle: const TerminalStyle(fontSize: 13),
            padding: const EdgeInsets.all(4),
          ),
        ),
        _buildKeyBar(),
      ],
    );
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
