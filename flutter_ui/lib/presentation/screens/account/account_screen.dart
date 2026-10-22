import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/api_constants.dart';
import '../../../core/theme/theme_provider.dart';
import '../../providers/command_history_provider.dart';

/// 设置页（本地模式，原「账户」）：主题切换、连接状态与应用信息。
/// 本应用无云账号体系，这一栏实际承载的是设置，故按用户反馈改名。
/// 承载外观设置与本地数据说明。
class AccountScreen extends ConsumerWidget {
  const AccountScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final themeMode = ref.watch(themeModeProvider);
    final accent = ref.watch(accentProvider);
    final terminalThemeMode = ref.watch(terminalThemeModeProvider);
    final terminalFontSize = ref.watch(terminalFontSizeProvider);
    final historyCount = ref.watch(commandHistoryProvider).length;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      appBar: AppBar(title: const Text('设置')),
      body: ListView(
        padding: const EdgeInsets.all(12),
        children: [
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  const CircleAvatar(
                    radius: 26,
                    child: Icon(Icons.person, size: 28),
                  ),
                  const SizedBox(width: 16),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '本地模式',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '连接与密码只保存在本机，不经任何云端',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
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
          ),
          const SizedBox(height: 12),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.palette, size: 20),
                      const SizedBox(width: 8),
                      Text('外观主题',
                          style: Theme.of(context).textTheme.titleMedium),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '明暗、主题色与终端配色均与桌面版同一套预设，选择即时生效并保存',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Theme.of(context)
                              .colorScheme
                              .onSurface
                              .withValues(alpha: 0.6),
                        ),
                  ),
                  const SizedBox(height: 12),
                  SegmentedButton<ThemeMode>(
                    segments: const [
                      ButtonSegment(
                        value: ThemeMode.system,
                        icon: Icon(Icons.brightness_auto),
                        label: Text('跟随系统'),
                      ),
                      ButtonSegment(
                        value: ThemeMode.light,
                        icon: Icon(Icons.light_mode),
                        label: Text('浅色'),
                      ),
                      ButtonSegment(
                        value: ThemeMode.dark,
                        icon: Icon(Icons.dark_mode),
                        label: Text('深色'),
                      ),
                    ],
                    selected: {themeMode},
                    onSelectionChanged: (selection) {
                      ref
                          .read(themeModeProvider.notifier)
                          .setMode(selection.first);
                    },
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Text('主题色',
                          style: Theme.of(context).textTheme.titleSmall),
                      const SizedBox(width: 8),
                      Text(
                        accent.label,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Theme.of(context).colorScheme.primary,
                              fontWeight: FontWeight.w600,
                            ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  Wrap(
                    spacing: 10,
                    runSpacing: 10,
                    children: [
                      for (final item in appAccents)
                        _AccentDot(
                          accent: item,
                          dark: isDark,
                          selected: item.id == accent.id,
                          onTap: () => ref
                              .read(accentProvider.notifier)
                              .setAccent(item),
                        ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  // 主题色实时预览：染色面积小的时候用户感知不到切换效果，
                  // 这里给一组跟随当前配色的示例控件，点色点立刻能对比
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 12, vertical: 10),
                    decoration: BoxDecoration(
                      color: Theme.of(context)
                          .colorScheme
                          .surfaceContainerHighest
                          .withValues(alpha: 0.5),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Row(
                      children: [
                        FilledButton(
                          onPressed: () {},
                          child: const Text('示例按钮'),
                        ),
                        const SizedBox(width: 10),
                        ChoiceChip(
                          label: const Text('已选中'),
                          selected: true,
                          onSelected: (_) {},
                        ),
                        const SizedBox(width: 10),
                        Icon(Icons.favorite,
                            color: Theme.of(context).colorScheme.primary),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text('终端主题',
                      style: Theme.of(context).textTheme.titleSmall),
                  const SizedBox(height: 10),
                  SegmentedButton<TerminalThemeMode>(
                    segments: const [
                      ButtonSegment(
                        value: TerminalThemeMode.dark,
                        label: Text('深色'),
                      ),
                      ButtonSegment(
                        value: TerminalThemeMode.light,
                        label: Text('浅色'),
                      ),
                      ButtonSegment(
                        value: TerminalThemeMode.follow,
                        label: Text('跟随界面'),
                      ),
                    ],
                    selected: {terminalThemeMode},
                    onSelectionChanged: (selection) {
                      ref
                          .read(terminalThemeModeProvider.notifier)
                          .setMode(selection.first);
                    },
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Text('终端字号',
                          style: Theme.of(context).textTheme.titleSmall),
                      const SizedBox(width: 8),
                      Text(
                        '${terminalFontSize.round()} 磅',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Theme.of(context).colorScheme.primary,
                              fontWeight: FontWeight.w600,
                            ),
                      ),
                    ],
                  ),
                  Slider(
                    value: terminalFontSize,
                    min: TerminalFontSizeNotifier.minSize,
                    max: TerminalFontSizeNotifier.maxSize,
                    divisions: 10,
                    label: '${terminalFontSize.round()}',
                    onChanged: (v) =>
                        ref.read(terminalFontSizeProvider.notifier).setSize(v),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  const Icon(Icons.history, size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('命令历史',
                            style: Theme.of(context).textTheme.titleMedium),
                        const SizedBox(height: 4),
                        Text(
                          historyCount == 0
                              ? '暂无记录。在终端里执行过的命令会出现在这里，并用于输入建议'
                              : '已记录 $historyCount 条，用于终端输入时的命令建议',
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: Theme.of(context)
                                        .colorScheme
                                        .onSurface
                                        .withValues(alpha: 0.6),
                                  ),
                        ),
                      ],
                    ),
                  ),
                  TextButton(
                    onPressed: historyCount == 0
                        ? null
                        : () async {
                            await ref
                                .read(commandHistoryProvider.notifier)
                                .clear();
                            if (context.mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(content: Text('命令历史已清除')),
                              );
                            }
                          },
                    child: const Text('清除'),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.info_outline, size: 20),
                      const SizedBox(width: 8),
                      Text('关于', style: Theme.of(context).textTheme.titleMedium),
                    ],
                  ),
                  const SizedBox(height: 12),
                  _row(context, '应用', 'AHaSSHTools 移动端'),
                  _row(context, '版本', '1.0.0'),
                  _row(
                    context,
                    '后端模式',
                    ApiConstants.isEmbedded ? '内嵌 Go（进程内）' : '外部 API 服务',
                  ),
                  _row(context, 'API 地址', ApiConstants.defaultBaseUrl),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _row(BuildContext context, String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 72,
            child: Text(
              label,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurface
                        .withValues(alpha: 0.6),
                  ),
            ),
          ),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}

/// 主题色选择圆点：桌面预设的主→辅渐变，选中时带描边与对勾。
class _AccentDot extends StatelessWidget {
  final AppAccent accent;
  final bool dark;
  final bool selected;
  final VoidCallback onTap;

  const _AccentDot({
    required this.accent,
    required this.dark,
    required this.selected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final main = dark ? accent.darkSeed : accent.lightSeed;
    final secondary = dark ? accent.darkSecondary : accent.lightSecondary;
    return Tooltip(
      message: accent.label,
      child: InkWell(
        onTap: onTap,
        customBorder: const CircleBorder(),
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [main, secondary],
            ),
            border: selected
                ? Border.all(
                    color: Theme.of(context).colorScheme.onSurface,
                    width: 2.5,
                  )
                : Border.all(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurface
                        .withValues(alpha: 0.15),
                  ),
          ),
          child: selected
              ? const Icon(Icons.check, size: 20, color: Colors.white)
              : null,
        ),
      ),
    );
  }
}
