import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../data/models/connection_model.dart';

/// 终端标签的连接状态（标签栏状态点与终端页状态条共用）
enum TerminalTabStatus { connecting, connected, disconnected, error }

/// 一个终端标签的元数据。xterm 的 Terminal 对象与会话生命周期
/// 由对应的 TerminalTabView 持有，这里只保存标签栏渲染
/// 与跨页联动（会话列表跳转）需要的轻量信息。
class TerminalTab {
  final String id;
  final ConnectionModel connection;

  /// 临时输入的凭据（尚未保存密码的连接）；为空时服务端回退已保存密码。
  final String authValue;

  /// 服务端会话 ID，会话建立后回填（会话列表页据此跳转到对应标签）
  final String? sessionId;
  final TerminalTabStatus status;

  const TerminalTab({
    required this.id,
    required this.connection,
    this.authValue = '',
    this.sessionId,
    this.status = TerminalTabStatus.connecting,
  });

  TerminalTab copyWith({
    String? sessionId,
    TerminalTabStatus? status,
  }) {
    return TerminalTab(
      id: id,
      connection: connection,
      authValue: authValue,
      sessionId: sessionId ?? this.sessionId,
      status: status ?? this.status,
    );
  }
}

class TerminalTabsState {
  final List<TerminalTab> tabs;
  final int activeIndex;

  const TerminalTabsState({this.tabs = const [], this.activeIndex = -1});

  TerminalTab? get activeTab =>
      (activeIndex >= 0 && activeIndex < tabs.length)
          ? tabs[activeIndex]
          : null;
}

/// 多标签终端的状态管理：开/关/切换标签。
///
/// 开标签策略：从资产列表连接时同一资产已有标签则直接激活
/// （避免误开一堆重复会话）；在终端页用「+」新建时允许多开
/// 同一资产（同机多 shell）。关闭标签后激活相邻标签；
/// 标签视图随之销毁并关闭其 SSH 会话。
class TerminalTabsNotifier extends StateNotifier<TerminalTabsState> {
  TerminalTabsNotifier() : super(const TerminalTabsState());

  /// 打开标签：已有该连接的标签则激活它，否则新建。返回标签 ID。
  String openOrActivate(ConnectionModel connection, {String authValue = ''}) {
    final existing =
        state.tabs.indexWhere((t) => t.connection.id == connection.id);
    if (existing >= 0) {
      state = TerminalTabsState(tabs: state.tabs, activeIndex: existing);
      return state.tabs[existing].id;
    }
    return openNew(connection, authValue: authValue);
  }

  /// 无条件新建标签（同一连接可多开）。返回标签 ID。
  String openNew(ConnectionModel connection, {String authValue = ''}) {
    final tab = TerminalTab(
      id: const Uuid().v4(),
      connection: connection,
      authValue: authValue,
    );
    state = TerminalTabsState(
      tabs: [...state.tabs, tab],
      activeIndex: state.tabs.length,
    );
    return tab.id;
  }

  /// 按服务端会话 ID 激活对应标签（会话列表页跳转用）。找不到返回 false。
  bool activateBySessionId(String sessionId) {
    final index = state.tabs.indexWhere((t) => t.sessionId == sessionId);
    if (index < 0) {
      return false;
    }
    state = TerminalTabsState(tabs: state.tabs, activeIndex: index);
    return true;
  }

  void activate(int index) {
    if (index < 0 || index >= state.tabs.length) {
      return;
    }
    state = TerminalTabsState(tabs: state.tabs, activeIndex: index);
  }

  /// 关闭标签：优先激活右邻，其次左邻；关掉最后一个则无激活标签。
  void closeTab(String tabId) {
    final index = state.tabs.indexWhere((t) => t.id == tabId);
    if (index < 0) {
      return;
    }
    final tabs = [...state.tabs]..removeAt(index);
    var active = state.activeIndex;
    if (tabs.isEmpty) {
      active = -1;
    } else if (index < active) {
      active -= 1;
    } else if (index == active && active >= tabs.length) {
      active = tabs.length - 1;
    }
    state = TerminalTabsState(tabs: tabs, activeIndex: active);
  }

  void setStatus(String tabId, TerminalTabStatus status) {
    _update(tabId, (t) => t.copyWith(status: status));
  }

  void setSessionId(String tabId, String sessionId) {
    _update(tabId, (t) => t.copyWith(sessionId: sessionId));
  }

  void _update(String tabId, TerminalTab Function(TerminalTab) fn) {
    final index = state.tabs.indexWhere((t) => t.id == tabId);
    if (index < 0) {
      return;
    }
    final tabs = [...state.tabs];
    tabs[index] = fn(tabs[index]);
    state = TerminalTabsState(tabs: tabs, activeIndex: state.activeIndex);
  }
}

/// 终端标签列表与激活下标
final terminalTabsProvider =
    StateNotifierProvider<TerminalTabsNotifier, TerminalTabsState>(
  (ref) => TerminalTabsNotifier(),
);

/// 底部导航当前栏下标，由 AppShell 消费；
/// 资产列表点「连接」后改此值切到终端栏。
final appShellIndexProvider = StateProvider<int>((ref) => 0);

/// 底部导航栏下标（顺序与 AppShell 的 destinations 一致）
abstract final class AppShellTabs {
  static const dashboard = 0;
  static const assets = 1;
  static const terminal = 2;
  static const sessions = 3;
  static const account = 4;
}
