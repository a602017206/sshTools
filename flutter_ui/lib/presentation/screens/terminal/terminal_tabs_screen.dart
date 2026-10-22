import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../data/models/connection_model.dart';
import '../../providers/connection_provider.dart' as conn_provider;
import '../../providers/terminal_tabs_provider.dart';
import '../../widgets/auth_input_dialog.dart';
import 'terminal_tab_view.dart';

/// 多标签终端页：顶部标签栏（状态点 + 名称 + 关闭），
/// 内容区用 IndexedStack 保活所有标签的终端与会话，
/// 切标签/切底部导航都不中断后台会话。标签栏末尾「+」
/// 从已保存资产新建标签（同一资产可多开一个 shell）。
class TerminalTabsScreen extends ConsumerWidget {
  const TerminalTabsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(terminalTabsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('终端'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            tooltip: '新建标签',
            onPressed: () => _showAssetPicker(context, ref),
          ),
        ],
      ),
      body: Column(
        children: [
          if (state.tabs.isNotEmpty) _buildTabBar(context, ref, state),
          Expanded(
            child: state.tabs.isEmpty
                ? _buildEmpty(context, ref)
                : IndexedStack(
                    index: state.activeIndex < 0 ? 0 : state.activeIndex,
                    children: [
                      for (var i = 0; i < state.tabs.length; i++)
                        TerminalTabView(
                          key: ValueKey(state.tabs[i].id),
                          tab: state.tabs[i],
                          isActive: i == state.activeIndex,
                        ),
                    ],
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildTabBar(
    BuildContext context,
    WidgetRef ref,
    TerminalTabsState state,
  ) {
    final scheme = Theme.of(context).colorScheme;
    return Material(
      color: scheme.surfaceContainerHighest,
      child: SizedBox(
        height: 42,
        child: ListView.builder(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 5),
          itemCount: state.tabs.length,
          itemBuilder: (context, index) {
            final tab = state.tabs[index];
            final active = index == state.activeIndex;
            return Padding(
              padding: const EdgeInsets.only(right: 6),
              child: InkWell(
                borderRadius: BorderRadius.circular(16),
                onTap: () =>
                    ref.read(terminalTabsProvider.notifier).activate(index),
                child: Container(
                  padding: const EdgeInsets.only(left: 10, right: 2),
                  decoration: BoxDecoration(
                    color: active
                        ? scheme.primaryContainer
                        : scheme.surfaceContainerHigh,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.circle,
                        size: 8,
                        color: switch (tab.status) {
                          TerminalTabStatus.connected => Colors.green,
                          TerminalTabStatus.connecting => Colors.orange,
                          _ => Colors.red,
                        },
                      ),
                      const SizedBox(width: 6),
                      ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 110),
                        child: Text(
                          tab.connection.name,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight:
                                active ? FontWeight.w600 : FontWeight.normal,
                          ),
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.close, size: 16),
                        tooltip: '关闭标签',
                        visualDensity: VisualDensity.compact,
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(
                          minWidth: 28,
                          minHeight: 28,
                        ),
                        onPressed: () => ref
                            .read(terminalTabsProvider.notifier)
                            .closeTab(tab.id),
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildEmpty(BuildContext context, WidgetRef ref) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.terminal,
              size: 80,
              color:
                  Theme.of(context).colorScheme.primary.withValues(alpha: 0.5),
            ),
            const SizedBox(height: 16),
            Text('暂无终端标签', style: Theme.of(context).textTheme.headlineSmall),
            const SizedBox(height: 8),
            Text(
              '从资产列表点「连接」，或点下方按钮选择一台服务器打开终端',
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurface
                        .withValues(alpha: 0.6),
                  ),
            ),
            const SizedBox(height: 20),
            FilledButton.icon(
              onPressed: () => _showAssetPicker(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('新建标签'),
            ),
          ],
        ),
      ),
    );
  }

  /// 新建标签选资产：点一台服务器即新开一个标签（允许多开同一资产）。
  Future<void> _showAssetPicker(BuildContext context, WidgetRef ref) async {
    final connections =
        ref.read(conn_provider.connectionProvider).connections;
    if (connections.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('暂无已保存连接，请先在资产列表添加')),
      );
      return;
    }

    final picked = await showModalBottomSheet<ConnectionModel>(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const ListTile(
              title: Text('选择服务器新建标签',
                  style: TextStyle(fontWeight: FontWeight.w600)),
            ),
            const Divider(height: 1),
            Flexible(
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: connections.length,
                itemBuilder: (context, index) {
                  final conn = connections[index];
                  return ListTile(
                    leading: const Icon(Icons.computer),
                    title: Text(conn.name),
                    subtitle: Text('${conn.user}@${conn.host}:${conn.port}'),
                    onTap: () => Navigator.pop(context, conn),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
    if (picked == null || !context.mounted) {
      return;
    }

    final authValue = await requestAuthValue(context, picked);
    if (authValue == null || !context.mounted) {
      return;
    }
    ref
        .read(terminalTabsProvider.notifier)
        .openNew(picked, authValue: authValue);
  }
}
