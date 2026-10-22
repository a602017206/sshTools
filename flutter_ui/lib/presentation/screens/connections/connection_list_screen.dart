import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../data/models/connection_model.dart';
import '../../providers/connection_provider.dart' as conn_provider;
import '../../providers/terminal_tabs_provider.dart';
import '../../widgets/auth_input_dialog.dart';
import '../sftp/sftp_screen.dart';
import 'widgets/connection_form_page.dart';

/// 连接列表页（移动端）。
///
/// 针对真机反馈重做了交互：每张卡片直接给出「连接 / 测试 / 编辑 / 删除」
/// 按钮，不再把操作全埋进右上角菜单；点卡片或「连接」进入 SSH 终端，
/// 「测试」只验证连通性；两者都优先用已保存密码，不再每次弹窗索要。
class ConnectionListScreen extends ConsumerStatefulWidget {
  const ConnectionListScreen({super.key});

  @override
  ConsumerState<ConnectionListScreen> createState() =>
      _ConnectionListScreenState();
}

class _ConnectionListScreenState extends ConsumerState<ConnectionListScreen> {
  final _searchController = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final connectionState = ref.watch(conn_provider.connectionProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('SSH 连接'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              ref.read(conn_provider.connectionProvider.notifier).loadConnections();
            },
            tooltip: '刷新',
          ),
        ],
      ),
      body: Column(
        children: [
          if (connectionState.connections.isNotEmpty)
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 10, 12, 2),
              child: TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  hintText: '搜索名称 / 主机 / 用户 / 分组',
                  prefixIcon: const Icon(Icons.search),
                  suffixIcon: _query.isEmpty
                      ? null
                      : IconButton(
                          icon: const Icon(Icons.clear),
                          tooltip: '清空搜索',
                          onPressed: () {
                            _searchController.clear();
                            setState(() => _query = '');
                          },
                        ),
                  border: const OutlineInputBorder(),
                  isDense: true,
                ),
                textCapitalization: TextCapitalization.none,
                autocorrect: false,
                onChanged: (v) => setState(() => _query = v),
              ),
            ),
          Expanded(child: _buildBody(context, ref, connectionState)),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showConnectionForm(context, ref),
        icon: const Icon(Icons.add),
        label: const Text('添加连接'),
      ),
    );
  }

  Widget _buildBody(
    BuildContext context,
    WidgetRef ref,
    conn_provider.ConnectionState state,
  ) {
    if (state.error != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(state.error!),
            backgroundColor: Colors.red,
            duration: const Duration(seconds: 6),
            action: SnackBarAction(
              label: '关闭',
              textColor: Colors.white,
              onPressed: () {
                ref.read(conn_provider.connectionProvider.notifier).clearError();
              },
            ),
          ),
        );
        ref.read(conn_provider.connectionProvider.notifier).clearError();
      });
    }

    if (state.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.connections.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.storage_outlined,
                size: 80,
                color: Theme.of(context).colorScheme.primary.withValues(alpha: 0.5),
              ),
              const SizedBox(height: 16),
              Text(
                '暂无连接',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              Text(
                '点击右下角「添加连接」，填写主机、用户名和密码即可保存',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Theme.of(context)
                          .colorScheme
                          .onSurface
                          .withValues(alpha: 0.6),
                    ),
              ),
            ],
          ),
        ),
      );
    }

    // 搜索过滤（名称/主机/用户/任意标签，含分组）
    final query = _query.trim().toLowerCase();
    bool matches(ConnectionModel c) =>
        query.isEmpty ||
        c.name.toLowerCase().contains(query) ||
        c.host.toLowerCase().contains(query) ||
        c.user.toLowerCase().contains(query) ||
        (c.tags ?? []).any((t) => t.toLowerCase().contains(query));
    final filtered = state.connections.where(matches).toList();

    if (filtered.isEmpty) {
      return Center(
        child: Text('没有匹配「${_query.trim()}」的连接'),
      );
    }

    // 按分组分段：分组 = tags 首个（桌面端约定），无标签归「默认分组」并垫底
    final groups = <String, List<ConnectionModel>>{};
    for (final c in filtered) {
      final g = (c.tags != null && c.tags!.isNotEmpty) ? c.tags!.first : '默认分组';
      groups.putIfAbsent(g, () => []).add(c);
    }
    final groupNames = groups.keys.toList()
      ..sort((a, b) {
        if (a == '默认分组') return 1;
        if (b == '默认分组') return -1;
        return a.compareTo(b);
      });

    return ListView(
      padding: const EdgeInsets.fromLTRB(8, 4, 8, 88),
      children: [
        for (final g in groupNames) ...[
          _groupHeader(context, g, groups[g]!.length),
          for (final connection in groups[g]!)
            _ConnectionCard(
              connection: connection,
              onTap: () => _openTerminal(context, ref, connection),
              onConnect: () => _openTerminal(context, ref, connection),
              onFiles: () => _openSftp(context, ref, connection),
              onEdit: () => _showConnectionForm(context, ref, connection),
              onClone: () =>
                  _showConnectionForm(context, ref, connection, true),
              onDelete: () => _deleteConnection(context, ref, connection),
              onTest: () => _testConnection(context, ref, connection),
            ),
        ],
      ],
    );
  }

  Widget _groupHeader(BuildContext context, String name, int count) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 2),
      child: Row(
        children: [
          Icon(Icons.folder_open,
              size: 18, color: Theme.of(context).colorScheme.primary),
          const SizedBox(width: 6),
          Text(name, style: Theme.of(context).textTheme.titleSmall),
          const SizedBox(width: 6),
          Text(
            '$count 台',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context)
                      .colorScheme
                      .onSurface
                      .withValues(alpha: 0.55),
                ),
          ),
        ],
      ),
    );
  }

  Future<void> _showConnectionForm(
    BuildContext context,
    WidgetRef ref, [
    ConnectionModel? connection,
    bool clone = false,
  ]) async {
    final existingGroups = <String>{
      for (final c in ref.read(conn_provider.connectionProvider).connections)
        if (c.tags != null && c.tags!.isNotEmpty) c.tags!.first,
    }.toList()
      ..sort();
    final result = await Navigator.push<ConnectionModel>(
      context,
      MaterialPageRoute(
        builder: (context) => ConnectionFormPage(
          connection: connection,
          clone: clone,
          existingGroups: existingGroups,
        ),
      ),
    );

    if (result != null) {
      final notifier = ref.read(conn_provider.connectionProvider.notifier);
      // 新增与克隆都走新增接口（克隆的 ID 已在表单里清空）
      final isCreate = connection == null || clone;
      final success = isCreate
          ? await notifier.addConnection(result)
          : await notifier.updateConnection(result);
      final verb = connection == null ? '添加' : (clone ? '克隆' : '更新');

      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              success ? '已$verb ${result.name}' : '$verb失败，请查看错误提示',
            ),
            backgroundColor: success ? Colors.green : Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _deleteConnection(
    BuildContext context,
    WidgetRef ref,
    ConnectionModel connection,
  ) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('删除连接'),
        content: Text('确定要删除「${connection.name}」吗？已保存的密码会一并删除。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('删除'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      final success =
          await ref.read(conn_provider.connectionProvider.notifier).deleteConnection(
                connection.id,
              );

      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(success ? '已删除 ${connection.name}' : '删除失败，请查看错误提示'),
            backgroundColor: success ? Colors.green : Colors.red,
          ),
        );
      }
    }
  }

  /// 打开 SSH 终端标签：同一资产已有标签时直接激活，
  /// 否则新建；随后切到终端栏。没有已保存凭据时先弹窗索要，
  /// 连接成功后服务端会回存，后续直接开标签。
  Future<void> _openTerminal(
    BuildContext context,
    WidgetRef ref,
    ConnectionModel connection,
  ) async {
    final authValue = await requestAuthValue(context, connection);
    if (authValue == null || !context.mounted) {
      return;
    }

    ref
        .read(terminalTabsProvider.notifier)
        .openOrActivate(connection, authValue: authValue);
    ref.read(appShellIndexProvider.notifier).state = AppShellTabs.terminal;
    // 开标签后刷新一次，让 has_password 等状态与服务端对齐
    await ref.read(conn_provider.connectionProvider.notifier).loadConnections();
  }

  /// 进入 SFTP 文件管理。凭据处理与终端一致：
  /// 已保存直接进，未保存先弹窗索要一次。
  Future<void> _openSftp(
    BuildContext context,
    WidgetRef ref,
    ConnectionModel connection,
  ) async {
    final authValue = await requestAuthValue(context, connection);
    if (authValue == null || !context.mounted) {
      return;
    }

    if (!context.mounted) return;
    await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => SftpScreen(
          connection: connection,
          authValue: authValue,
        ),
      ),
    );
    if (context.mounted) {
      await ref.read(conn_provider.connectionProvider.notifier).loadConnections();
    }
  }

  Future<void> _testConnection(
    BuildContext context,
    WidgetRef ref,
    ConnectionModel connection,
  ) async {
    // 没有已保存密码时才向用户索要；测试成功后服务端会顺手保存，
    // 下次点测试就不用再输了。
    final authInput = await requestAuthValue(context, connection);
    if (authInput == null || !context.mounted) {
      return;
    }
    final authValue = authInput;

    if (!context.mounted) return;

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(
        child: Card(
          child: Padding(
            padding: EdgeInsets.all(24.0),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                CircularProgressIndicator(),
                SizedBox(height: 16),
                Text('正在测试连接...'),
              ],
            ),
          ),
        ),
      ),
    );

    final notifier = ref.read(conn_provider.connectionProvider.notifier);
    final success = await notifier.testConnection(
      host: connection.host,
      port: connection.port,
      user: connection.user,
      authType: connection.authType,
      authValue: authValue,
      connectionId: connection.id,
    );

    if (context.mounted) {
      Navigator.pop(context); // 关闭 loading

      if (success) {
        // 测试成功后刷新一次，让 has_password 等状态与服务端对齐
        await notifier.loadConnections();
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('连接成功'),
              backgroundColor: Colors.green,
            ),
          );
        }
      } else {
        final error = ref.read(conn_provider.connectionProvider).error;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(error ?? '连接失败，请检查地址与凭据'),
            backgroundColor: Colors.red,
            duration: const Duration(seconds: 6),
          ),
        );
        notifier.clearError();
      }
    }
  }
}

/// 连接卡片：名称、账号地址、已保存密码标记，以及一排可见的操作按钮。
class _ConnectionCard extends StatelessWidget {
  final ConnectionModel connection;
  final VoidCallback onTap;
  final VoidCallback onConnect;
  final VoidCallback onFiles;
  final VoidCallback onEdit;
  final VoidCallback onClone;
  final VoidCallback onDelete;
  final VoidCallback onTest;

  const _ConnectionCard({
    required this.connection,
    required this.onTap,
    required this.onConnect,
    required this.onFiles,
    required this.onEdit,
    required this.onClone,
    required this.onDelete,
    required this.onTest,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 8, 4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    Icons.computer,
                    size: 32,
                    color: Theme.of(context).colorScheme.primary,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          connection.name,
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '${connection.user}@${connection.host}:${connection.port}',
                          style:
                              Theme.of(context).textTheme.bodyMedium?.copyWith(
                                    color: Theme.of(context)
                                        .colorScheme
                                        .onSurface
                                        .withValues(alpha: 0.7),
                                  ),
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            Icon(
                              connection.hasPassword ||
                                      connection.authType == 'key'
                                  ? Icons.lock
                                  : Icons.lock_open,
                              size: 14,
                              color: Theme.of(context)
                                  .colorScheme
                                  .onSurface
                                  .withValues(alpha: 0.6),
                            ),
                            const SizedBox(width: 4),
                            Text(
                              connection.authType == 'key'
                                  ? '私钥认证'
                                  : (connection.hasPassword ? '密码已保存' : '未保存密码'),
                              style: Theme.of(context)
                                  .textTheme
                                  .bodySmall
                                  ?.copyWith(
                                    color: Theme.of(context)
                                        .colorScheme
                                        .onSurface
                                        .withValues(alpha: 0.6),
                                  ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (connection.tags != null && connection.tags!.isNotEmpty) ...[
                const SizedBox(height: 8),
                Wrap(
                  spacing: 8,
                  children: connection.tags!.map((tag) {
                    return Chip(
                      label: Text(tag),
                      labelStyle: const TextStyle(fontSize: 12),
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      visualDensity: VisualDensity.compact,
                    );
                  }).toList(),
                ),
              ],
              // 五个操作按钮横排在窄屏上会溢出卡片（删除曾被挤到块外），
              // 改 Wrap 自动换行，右对齐保持不变
              Wrap(
                alignment: WrapAlignment.end,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 4,
                runSpacing: 4,
                children: [
                  FilledButton.icon(
                    onPressed: onConnect,
                    icon: const Icon(Icons.terminal, size: 20),
                    label: const Text('连接'),
                  ),
                  TextButton.icon(
                    onPressed: onFiles,
                    icon: const Icon(Icons.folder, size: 20),
                    label: const Text('文件'),
                  ),
                  TextButton.icon(
                    onPressed: onTest,
                    icon: const Icon(Icons.play_arrow, size: 20),
                    label: const Text('测试'),
                  ),
                  TextButton.icon(
                    onPressed: onEdit,
                    icon: const Icon(Icons.edit, size: 20),
                    label: const Text('编辑'),
                  ),
                  TextButton.icon(
                    onPressed: onClone,
                    icon: const Icon(Icons.copy, size: 20),
                    label: const Text('克隆'),
                  ),
                  TextButton.icon(
                    onPressed: onDelete,
                    icon: const Icon(Icons.delete, size: 20, color: Colors.red),
                    label: const Text('删除', style: TextStyle(color: Colors.red)),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
