import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/terminal_tabs_provider.dart';
import 'account/account_screen.dart';
import 'connections/connection_list_screen.dart';
import 'dashboard/dashboard_screen.dart';
import 'sessions/sessions_screen.dart';
import 'terminal/terminal_tabs_screen.dart';

/// 应用主壳：底部五栏导航（仪表盘 / 资产列表 / 终端 / 会话列表 / 账户）。
/// 当前栏下标由 appShellIndexProvider 持有，资产列表点「连接」后
/// 程序切到终端栏。各页用 IndexedStack 保活：监控轮询、
/// 终端标签与列表滚动位置不会因切栏丢失。
class AppShell extends ConsumerWidget {
  const AppShell({super.key});

  static const _pages = [
    DashboardScreen(),
    ConnectionListScreen(),
    TerminalTabsScreen(),
    SessionsScreen(),
    AccountScreen(),
  ];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final index = ref.watch(appShellIndexProvider);

    return Scaffold(
      body: IndexedStack(
        index: index,
        children: _pages,
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: index,
        onDestinationSelected: (i) =>
            ref.read(appShellIndexProvider.notifier).state = i,
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.dashboard_outlined),
            selectedIcon: Icon(Icons.dashboard),
            label: '仪表盘',
          ),
          NavigationDestination(
            icon: Icon(Icons.storage_outlined),
            selectedIcon: Icon(Icons.storage),
            label: '资产列表',
          ),
          NavigationDestination(
            icon: Icon(Icons.terminal_outlined),
            selectedIcon: Icon(Icons.terminal),
            label: '终端',
          ),
          NavigationDestination(
            icon: Icon(Icons.forum_outlined),
            selectedIcon: Icon(Icons.forum),
            label: '会话列表',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline),
            selectedIcon: Icon(Icons.person),
            label: '账户',
          ),
        ],
      ),
    );
  }
}
