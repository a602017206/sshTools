import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:xterm/xterm.dart';

/// 主题模式（跟随系统 / 浅色 / 深色），持久化到本地。
/// 早期版本写死了深色，真机反馈「一直是黑的」，这里补齐切换。
class ThemeModeNotifier extends StateNotifier<ThemeMode> {
  static const _prefsKey = 'app_theme_mode';

  ThemeModeNotifier() : super(ThemeMode.system) {
    _load();
  }

  Future<void> _load() async {
    final prefs = await SharedPreferences.getInstance();
    final value = prefs.getString(_prefsKey);
    state = switch (value) {
      'light' => ThemeMode.light,
      'dark' => ThemeMode.dark,
      _ => ThemeMode.system,
    };
  }

  Future<void> setMode(ThemeMode mode) async {
    state = mode;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_prefsKey, mode.name);
  }
}

final themeModeProvider =
    StateNotifierProvider<ThemeModeNotifier, ThemeMode>((ref) {
  return ThemeModeNotifier();
});

/// 主题色（强调色）预设，与桌面端 ACCENT_PRESETS 对齐：
/// 同名同色，浅色/深色各取桌面的 accentPrimary 作为种子色生成
/// Material 3 配色，辅色用于选择器色点的渐变展示。
class AppAccent {
  final String id;
  final String label;
  final Color lightSeed;
  final Color darkSeed;
  final Color lightSecondary;
  final Color darkSecondary;

  const AppAccent({
    required this.id,
    required this.label,
    required this.lightSeed,
    required this.darkSeed,
    required this.lightSecondary,
    required this.darkSecondary,
  });
}

const appAccents = <AppAccent>[
  AppAccent(
    id: 'teal',
    label: '海盐青',
    lightSeed: Color(0xFF0F9F9A),
    darkSeed: Color(0xFF5EEAD4),
    lightSecondary: Color(0xFF14B8A6),
    darkSecondary: Color(0xFF2DD4BF),
  ),
  AppAccent(
    id: 'blue',
    label: '晴空蓝',
    lightSeed: Color(0xFF2F6DF6),
    darkSeed: Color(0xFF8AB7FF),
    lightSecondary: Color(0xFF60A5FA),
    darkSecondary: Color(0xFF93C5FD),
  ),
  AppAccent(
    id: 'emerald',
    label: '薄荷绿',
    lightSeed: Color(0xFF10B981),
    darkSeed: Color(0xFF6EE7B7),
    lightSecondary: Color(0xFF84CC16),
    darkSecondary: Color(0xFFBEF264),
  ),
  AppAccent(
    id: 'amber',
    label: '柚子橙',
    lightSeed: Color(0xFFF59E0B),
    darkSeed: Color(0xFFFBBF24),
    lightSecondary: Color(0xFFFB7185),
    darkSecondary: Color(0xFFFDA4AF),
  ),
  AppAccent(
    id: 'purple',
    label: '莓果紫',
    lightSeed: Color(0xFF8B5CF6),
    darkSeed: Color(0xFFC4B5FD),
    lightSecondary: Color(0xFFEC4899),
    darkSecondary: Color(0xFFF0ABFC),
  ),
  AppAccent(
    id: 'sky',
    label: '湖水蓝',
    lightSeed: Color(0xFF0EA5E9),
    darkSeed: Color(0xFF7DD3FC),
    lightSecondary: Color(0xFF22D3EE),
    darkSecondary: Color(0xFF67E8F9),
  ),
  AppAccent(
    id: 'rose',
    label: '樱桃粉',
    lightSeed: Color(0xFFFB7185),
    darkSeed: Color(0xFFFDA4AF),
    lightSecondary: Color(0xFFF472B6),
    darkSecondary: Color(0xFFF9A8D4),
  ),
  AppAccent(
    id: 'coral',
    label: '珊瑚红',
    lightSeed: Color(0xFFF97316),
    darkSeed: Color(0xFFFDBA74),
    lightSecondary: Color(0xFFFB7185),
    darkSecondary: Color(0xFFFDA4AF),
  ),
  AppAccent(
    id: 'lime',
    label: '青柠绿',
    lightSeed: Color(0xFF84CC16),
    darkSeed: Color(0xFFBEF264),
    lightSecondary: Color(0xFF22C55E),
    darkSecondary: Color(0xFF86EFAC),
  ),
];

class AccentNotifier extends StateNotifier<AppAccent> {
  static const _prefsKey = 'app_accent_color';

  AccentNotifier() : super(appAccents.first) {
    _load();
  }

  Future<void> _load() async {
    final prefs = await SharedPreferences.getInstance();
    final id = prefs.getString(_prefsKey);
    state = appAccents.firstWhere(
      (a) => a.id == id,
      orElse: () => appAccents.first,
    );
  }

  Future<void> setAccent(AppAccent accent) async {
    state = accent;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_prefsKey, accent.id);
  }
}

final accentProvider =
    StateNotifierProvider<AccentNotifier, AppAccent>((ref) {
  return AccentNotifier();
});

/// 终端主题：深色 / 浅色 / 跟随界面（与桌面端 terminal_theme 对齐）
enum TerminalThemeMode { dark, light, follow }

class TerminalThemeModeNotifier extends StateNotifier<TerminalThemeMode> {
  static const _prefsKey = 'app_terminal_theme';

  TerminalThemeModeNotifier() : super(TerminalThemeMode.dark) {
    _load();
  }

  Future<void> _load() async {
    final prefs = await SharedPreferences.getInstance();
    final value = prefs.getString(_prefsKey);
    state = switch (value) {
      'light' => TerminalThemeMode.light,
      'follow' => TerminalThemeMode.follow,
      _ => TerminalThemeMode.dark,
    };
  }

  Future<void> setMode(TerminalThemeMode mode) async {
    state = mode;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_prefsKey, mode.name);
  }
}

final terminalThemeModeProvider =
    StateNotifierProvider<TerminalThemeModeNotifier, TerminalThemeMode>(
  (ref) => TerminalThemeModeNotifier(),
);

/// 终端字号（10~20 磅），持久化保存，桌面版默认 14、移动端此前写死 13。
class TerminalFontSizeNotifier extends StateNotifier<double> {
  static const _prefsKey = 'app_terminal_font_size';
  static const minSize = 10.0;
  static const maxSize = 20.0;

  TerminalFontSizeNotifier() : super(13) {
    _load();
  }

  Future<void> _load() async {
    final prefs = await SharedPreferences.getInstance();
    final value = prefs.getDouble(_prefsKey);
    if (value != null) {
      state = value.clamp(minSize, maxSize);
    }
  }

  Future<void> setSize(double size) async {
    state = size.clamp(minSize, maxSize);
    final prefs = await SharedPreferences.getInstance();
    await prefs.setDouble(_prefsKey, state);
  }
}

final terminalFontSizeProvider =
    StateNotifierProvider<TerminalFontSizeNotifier, double>(
  (ref) => TerminalFontSizeNotifier(),
);

/// 按终端主题设置与当前界面亮度解析出 xterm 配色。
/// 两套 16 色与桌面端 terminalTheme.js 的 DARK/LIGHT_XTERM_THEME 一一对应。
TerminalTheme resolveXtermTheme(TerminalThemeMode mode, Brightness brightness) {
  final useLight = switch (mode) {
    TerminalThemeMode.light => true,
    TerminalThemeMode.dark => false,
    TerminalThemeMode.follow => brightness == Brightness.light,
  };
  return useLight ? _lightXtermTheme : _darkXtermTheme;
}

const _darkXtermTheme = TerminalTheme(
  cursor: Color(0xFF39D5E8),
  selection: Color(0x4739D5E8),
  foreground: Color(0xFFD8E1EF),
  background: Color(0xFF050914),
  black: Color(0xFF000000),
  red: Color(0xFFCD3131),
  green: Color(0xFF0DBC79),
  yellow: Color(0xFFE5E510),
  blue: Color(0xFF2472C8),
  magenta: Color(0xFFBC3FBC),
  cyan: Color(0xFF11A8CD),
  white: Color(0xFFE5E5E5),
  brightBlack: Color(0xFF666666),
  brightRed: Color(0xFFF14C4C),
  brightGreen: Color(0xFF23D18B),
  brightYellow: Color(0xFFF5F543),
  brightBlue: Color(0xFF3B8EEA),
  brightMagenta: Color(0xFFD670D6),
  brightCyan: Color(0xFF29B8DB),
  brightWhite: Color(0xFFE5E5E5),
  searchHitBackground: Color(0xFFFFFF2B),
  searchHitBackgroundCurrent: Color(0xFF31FF26),
  searchHitForeground: Color(0xFF000000),
);

const _lightXtermTheme = TerminalTheme(
  cursor: Color(0xFF0F766E),
  selection: Color(0x380D9488),
  foreground: Color(0xFF1E293B),
  background: Color(0xFFF7F8FB),
  black: Color(0xFF1E293B),
  red: Color(0xFFB91C1C),
  green: Color(0xFF15803D),
  yellow: Color(0xFFA16207),
  blue: Color(0xFF1D4ED8),
  magenta: Color(0xFF7E22CE),
  cyan: Color(0xFF0E7490),
  white: Color(0xFF334155),
  brightBlack: Color(0xFF64748B),
  brightRed: Color(0xFFDC2626),
  brightGreen: Color(0xFF16A34A),
  brightYellow: Color(0xFFCA8A04),
  brightBlue: Color(0xFF2563EB),
  brightMagenta: Color(0xFF9333EA),
  brightCyan: Color(0xFF0891B2),
  brightWhite: Color(0xFF0F172A),
  searchHitBackground: Color(0xFFFFFF2B),
  searchHitBackgroundCurrent: Color(0xFF31FF26),
  searchHitForeground: Color(0xFF000000),
);
