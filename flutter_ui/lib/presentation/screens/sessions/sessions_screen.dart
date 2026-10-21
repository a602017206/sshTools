import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../providers/api_providers.dart';
import '../../providers/terminal_tabs_provider.dart';

/// 会话列表：展示服务端当前活跃的 SSH 会话（终端/文件/监控各自的连接），
/// 可逐个关闭。复刻 HexHub 底部「会话列表」的定位，做成可管理的清单。
class SessionsScreen extends ConsumerStatefulWidget {
  const SessionsScreen({super.key});

  @override
  ConsumerState<SessionsScreen> createState() => _SessionsScreenState();
}

class _SessionsScreenState extends ConsumerState<SessionsScreen> {
  List<String> _sessions = const [];
  bool _loading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final sessions =
          await ref.read(sessionRepositoryProvider).listSessions();
      if (!mounted) return;
      setState(() {
        _sessions = sessions;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  Future<void> _close(String sessionId) async {
    try {
      await ref.read(sessionRepositoryProvider).closeSession(sessionId);
      await _load();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('关闭失败：$e'), backgroundColor: Colors.red),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('会话列表'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: '刷新',
            onPressed: _load,
          ),
        ],
      ),
      body: _buildBody(context),
    );
  }

  Widget _buildBody(BuildContext context) {
    if (_loading && _sessions.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null && _sessions.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Text(_error!, style: const TextStyle(color: Colors.red)),
        ),
      );
    }
    if (_sessions.isEmpty) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(32),
          child: Text(
            '当前没有活跃会话。打开终端、文件或监控看板后，会话会出现在这里。',
            textAlign: TextAlign.center,
          ),
        ),
      );
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.builder(
        itemCount: _sessions.length,
        itemBuilder: (context, index) {
          final id = _sessions[index];
          return ListTile(
            leading: const Icon(Icons.terminal),
            title: Text(
              _friendlyName(id),
              overflow: TextOverflow.ellipsis,
            ),
            subtitle: Text(id, overflow: TextOverflow.ellipsis),
            // 终端会话点一下跳到对应标签；文件/监控会话无对应页，只作展示
            onTap: () => _openIfTerminalTab(id),
            trailing: IconButton(
              icon: const Icon(Icons.close, color: Colors.red),
              tooltip: '关闭会话',
              onPressed: () => _close(id),
            ),
          );
        },
      ),
    );
  }

  void _openIfTerminalTab(String sessionId) {
    final notifier = ref.read(terminalTabsProvider.notifier);
    if (notifier.activateBySessionId(sessionId)) {
      ref.read(appShellIndexProvider.notifier).state = AppShellTabs.terminal;
    }
  }

  String _friendlyName(String id) {
    if (id.startsWith('mobile-monitor-')) return '监控会话';
    if (id.startsWith('mobile-sftp-')) return '文件会话';
    if (id.startsWith('mobile-')) return '终端会话';
    return '会话';
  }
}
