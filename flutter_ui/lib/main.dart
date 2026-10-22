import 'dart:io' show Platform;

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';
import 'package:sshtools_go/sshtools_go.dart';

import 'core/constants/api_constants.dart';
import 'core/theme/theme_provider.dart';
import 'presentation/screens/app_shell.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await _startEmbeddedGoIfMobile();
  runApp(
    const ProviderScope(
      child: SSHToolsApp(),
    ),
  );
}

/// 移动端：在应用进程内启动 gomobile 内嵌的 Go API 服务器，
/// 并把 API 基地址切换到 127.0.0.1 + 实际端口。
///
/// 拿沙盒目录必须走 path_provider（应用私有目录），Go 侧会把
/// 配置与凭证写到该目录下的 .ahasshtools/ 中。
/// 桌面/Web 没有注册插件时保持默认地址，连独立运行的 cmd/apiserver，
/// 不影响现有桌面调试流程；启动失败也不阻断 UI，错误会在连接列表中
/// 以接口异常的形式暴露出来。
Future<void> _startEmbeddedGoIfMobile() async {
  if (kIsWeb || !(Platform.isAndroid || Platform.isIOS)) {
    return;
  }
  try {
    final dir = await getApplicationSupportDirectory();
    final port = await SshtoolsGo.start(dir.path);
    ApiConstants.configureEmbedded(port);
  } catch (e) {
    debugPrint('Failed to start embedded Go server: $e');
  }
}

class SSHToolsApp extends ConsumerWidget {
  const SSHToolsApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final themeMode = ref.watch(themeModeProvider);
    final accent = ref.watch(accentProvider);

    // 主题色与桌面端对齐：同一强调色在浅/深下各取桌面预设主色作种子
    final lightTheme = ThemeData(
      colorScheme: ColorScheme.fromSeed(
        seedColor: accent.lightSeed,
        brightness: Brightness.light,
      ),
      useMaterial3: true,
    );
    final darkTheme = ThemeData(
      colorScheme: ColorScheme.fromSeed(
        seedColor: accent.darkSeed,
        brightness: Brightness.dark,
      ),
      useMaterial3: true,
    );

    return MaterialApp(
      title: 'SSH Tools',
      theme: lightTheme,
      darkTheme: darkTheme,
      themeMode: themeMode,
      home: const AppShell(),
      debugShowCheckedModeBanner: false,
    );
  }
}
