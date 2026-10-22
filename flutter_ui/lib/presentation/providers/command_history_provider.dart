import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// 终端命令历史：用户在终端回车提交过的命令（新→旧排序），持久化到本地。
///
/// 供终端输入时的建议条使用（前缀/包含匹配）。上限 200 条，
/// 重复命令去重并置顶。不分服务器：移动端多在同类机器间切换，
/// 通用历史比按机隔离更顺手；设置页可一键清空。
class CommandHistoryNotifier extends StateNotifier<List<String>> {
  static const _prefsKey = 'terminal_command_history';
  static const _maxEntries = 200;

  CommandHistoryNotifier() : super(const []) {
    _load();
  }

  Future<void> _load() async {
    final prefs = await SharedPreferences.getInstance();
    state = prefs.getStringList(_prefsKey) ?? const [];
  }

  /// 记录一条刚提交的命令：去重后置顶并落盘。太短的输入不记。
  Future<void> record(String command) async {
    final cmd = command.trim();
    if (cmd.length < 2) {
      return;
    }
    final next = [cmd, ...state.where((c) => c != cmd)];
    state = next.length > _maxEntries ? next.sublist(0, _maxEntries) : next;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setStringList(_prefsKey, state);
  }

  Future<void> clear() async {
    state = const [];
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_prefsKey);
  }
}

final commandHistoryProvider =
    StateNotifierProvider<CommandHistoryNotifier, List<String>>(
  (ref) => CommandHistoryNotifier(),
);
