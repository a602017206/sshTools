import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../presentation/screens/connections/connection_list_screen.dart';
import '../../presentation/screens/home_screen.dart';

/// Application routes
class AppRoutes {
  static const String home = '/';
  static const String connections = '/connections';
  static const String terminal = '/terminal';
  static const String fileManager = '/files';
  static const String monitor = '/monitor';
  static const String settings = '/settings';
}

/// App router configuration
final appRouter = GoRouter(
  initialLocation: AppRoutes.connections,
  routes: [
    GoRoute(
      path: AppRoutes.home,
      builder: (context, state) => const HomeScreen(),
    ),
    GoRoute(
      path: AppRoutes.connections,
      builder: (context, state) => const ConnectionListScreen(),
    ),
    // TODO: Add other routes as screens are implemented
    // GoRoute(
    //   path: AppRoutes.terminal,
    //   builder: (context, state) => const TerminalScreen(),
    // ),
    // GoRoute(
    //   path: AppRoutes.fileManager,
    //   builder: (context, state) => const FileManagerScreen(),
    // ),
    // GoRoute(
    //   path: AppRoutes.monitor,
    //   builder: (context, state) => const MonitorScreen(),
    // ),
    // GoRoute(
    //   path: AppRoutes.settings,
    //   builder: (context, state) => const SettingsScreen(),
    // ),
  ],
);
