import 'package:flutter/material.dart';

import 'account/account_screen.dart';
import 'connections/connection_list_screen.dart';
import 'dashboard/dashboard_screen.dart';
import 'sessions/sessions_screen.dart';

/// 应用主壳：底部四栏导航（仪表盘 / 资产列表 / 会话列表 / 账户），
/// 结构对齐 HexHub 移动端。各页用 IndexedStack 保活，
/// 监控轮询与列表滚动位置不会因切栏丢失。
class AppShell extends StatefulWidget {
  const AppShell({super.key});

  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> {
  int _index = 0;

  static const _pages = [
    DashboardScreen(),
    ConnectionListScreen(),
    SessionsScreen(),
    AccountScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _index,
        children: _pages,
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
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
