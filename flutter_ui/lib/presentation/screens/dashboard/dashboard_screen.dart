import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../data/models/connection_model.dart';
import '../../../data/models/monitoring_data.dart';
import '../../providers/api_providers.dart';
import '../../providers/connection_provider.dart' as conn_provider;
import '../../widgets/metric_gauge.dart';

/// 监控看板：选一台资产连上后轮询远端指标，
/// 总览四联环（负载/内存/CPU/磁盘）+ 系统信息 + 每核/分区明细。
/// 布局复刻 HexHub 移动端仪表盘的信息结构，数据走自家监控 API。
class DashboardScreen extends ConsumerStatefulWidget {
  const DashboardScreen({super.key});

  @override
  ConsumerState<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends ConsumerState<DashboardScreen> {
  ConnectionModel? _selected;
  String _sessionId = '';
  Timer? _timer;
  bool _connecting = false;
  bool _fetching = false;
  MonitoringData? _data;
  int? _latencyMs;
  String? _error;

  bool get _monitoring => _sessionId.isNotEmpty;

  @override
  void dispose() {
    _timer?.cancel();
    if (_sessionId.isNotEmpty) {
      ref.read(sessionRepositoryProvider).closeSession(_sessionId).catchError((_) {});
    }
    super.dispose();
  }

  Future<void> _startMonitoring(ConnectionModel conn) async {
    await _stopMonitoring();
    setState(() {
      _connecting = true;
      _error = null;
    });
    final sessionId = 'mobile-monitor-${const Uuid().v4()}';
    try {
      await ref.read(sessionRepositoryProvider).connectSSH(
            sessionId: sessionId,
            host: conn.host,
            port: conn.port,
            user: conn.user,
            authType: conn.authType,
            connectionId: conn.id,
            cols: 80,
            rows: 24,
          );
      _sessionId = sessionId;
      await _fetch();
      _timer = Timer.periodic(
        const Duration(seconds: 3),
        (_) => unawaited(_fetch()),
      );
    } catch (e) {
      if (mounted) setState(() => _error = '监控连接失败：$e');
    } finally {
      if (mounted) setState(() => _connecting = false);
    }
  }

  Future<void> _stopMonitoring() async {
    _timer?.cancel();
    _timer = null;
    final old = _sessionId;
    _sessionId = '';
    if (old.isNotEmpty) {
      await ref.read(sessionRepositoryProvider).closeSession(old).catchError((_) {});
    }
    if (mounted) {
      setState(() {
        _data = null;
        _latencyMs = null;
      });
    }
  }

  Future<void> _fetch() async {
    if (_fetching || _sessionId.isEmpty) return;
    _fetching = true;
    final sw = Stopwatch()..start();
    try {
      final data =
          await ref.read(monitorRepositoryProvider).getMetrics(_sessionId);
      sw.stop();
      if (!mounted) return;
      setState(() {
        _data = data;
        _latencyMs = sw.elapsedMilliseconds;
        _error = null;
      });
    } catch (e) {
      if (mounted) setState(() => _error = '指标拉取失败：$e');
    } finally {
      _fetching = false;
    }
  }

  @override
  Widget build(BuildContext context) {
    final connState = ref.watch(conn_provider.connectionProvider);
    final connections = connState.connections;
    _selected ??= connections.isNotEmpty ? connections.first : null;

    return Scaffold(
      appBar: AppBar(
        title: const Text('仪表盘'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: '刷新指标',
            onPressed: _monitoring ? () => unawaited(_fetch()) : null,
          ),
        ],
      ),
      body: connections.isEmpty
          ? _emptyAssetsHint(context)
          : ListView(
              padding: const EdgeInsets.all(12),
              children: [
                _assetPicker(context, connections),
                const SizedBox(height: 12),
                if (_error != null) _errorBanner(context),
                if (!_monitoring && !_connecting)
                  _connectPrompt(context)
                else if (_data == null)
                  const Padding(
                    padding: EdgeInsets.only(top: 64),
                    child: Center(child: CircularProgressIndicator()),
                  )
                else
                  ..._dashboardCards(context, _data!),
              ],
            ),
    );
  }

  Widget _emptyAssetsHint(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Text(
          '暂无资产。先到「资产列表」添加一台服务器，再回来打开监控看板。',
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.bodyLarge,
        ),
      ),
    );
  }

  Widget _assetPicker(BuildContext context, List<ConnectionModel> connections) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          children: [
            const Icon(Icons.dns),
            const SizedBox(width: 12),
            Expanded(
              child: DropdownButtonHideUnderline(
                child: DropdownButton<ConnectionModel>(
                  value: _selected,
                  isExpanded: true,
                  items: connections
                      .map(
                        (c) => DropdownMenuItem(
                          value: c,
                          child: Text(
                            '${c.name}（${c.user}@${c.host}）',
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      )
                      .toList(),
                  onChanged: _monitoring || _connecting
                      ? null
                      : (c) => setState(() => _selected = c),
                ),
              ),
            ),
            const SizedBox(width: 8),
            if (_monitoring)
              FilledButton.tonal(
                onPressed: () => unawaited(_stopMonitoring()),
                child: const Text('断开'),
              )
            else
              FilledButton(
                onPressed: _connecting || _selected == null
                    ? null
                    : () => unawaited(_startMonitoring(_selected!)),
                child: Text(_connecting ? '连接中...' : '连接监控'),
              ),
          ],
        ),
      ),
    );
  }

  Widget _connectPrompt(BuildContext context) {
    return const Card(
      child: Padding(
        padding: EdgeInsets.all(24),
        child: Center(child: Text('尚未连接。点「连接监控」开始实时查看这台服务器的状态。')),
      ),
    );
  }

  Widget _errorBanner(BuildContext context) {
    return Card(
      color: Colors.red.withValues(alpha: 0.12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Text(
          _error!,
          style: const TextStyle(color: Colors.red, fontSize: 12),
        ),
      ),
    );
  }

  List<Widget> _dashboardCards(BuildContext context, MonitoringData data) {
    final scheme = Theme.of(context).colorScheme;
    // 核心数只认每核采样：拿不到就显示 —，不再兜底成 1（与实际不符）
    final cores = data.cpu.perCore.isNotEmpty ? data.cpu.perCore.length : 0;
    final load1 =
        data.cpu.loadAverage.isNotEmpty ? data.cpu.loadAverage[0] : 0.0;
    final loadText = data.cpu.loadAverage.length >= 3
        ? data.cpu.loadAverage.map((v) => v.toStringAsFixed(2)).join(' / ')
        : '—';
    final disk = data.primaryPartition;

    return [
      Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(Icons.dashboard_customize, size: 18),
                  const SizedBox(width: 6),
                  Text('总览', style: Theme.of(context).textTheme.titleMedium),
                  const Spacer(),
                  Text(
                    // 取的是监控接口整轮往返（SSH 跑一串命令），并非网络 ping
                    '响应 ${_latencyMs != null ? '$_latencyMs ms' : '—'}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Theme.of(context)
                              .colorScheme
                              .onSurface
                              .withValues(alpha: 0.6),
                        ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 四环颜色取主题色板：切换主题色时看板跟着变，效果看得见
                  Expanded(
                    child: MetricGauge(
                      label: '负载 1/5/15m',
                      centerText: load1.toStringAsFixed(2),
                      subText: loadText,
                      fraction: cores == 0 ? 0 : load1 / cores,
                      color: scheme.primary,
                    ),
                  ),
                  Expanded(
                    child: MetricGauge(
                      label: '内存',
                      centerText: '${data.memory.usedPercent.round()}%',
                      subText:
                          '${_fmtBytes(data.memory.used)} / ${_fmtBytes(data.memory.total)}',
                      fraction: data.memory.usedPercent / 100,
                      color: scheme.tertiary,
                    ),
                  ),
                  Expanded(
                    child: MetricGauge(
                      label: 'CPU',
                      centerText: '${data.cpu.overall.toStringAsFixed(1)}%',
                      subText:
                          '用户 ${data.cpu.user.toStringAsFixed(1)}% · 系统 ${data.cpu.system.toStringAsFixed(1)}%',
                      fraction: data.cpu.overall / 100,
                      color: scheme.secondary,
                    ),
                  ),
                  Expanded(
                    child: MetricGauge(
                      label: '磁盘',
                      centerText:
                          disk == null ? '—' : '${disk.usedPercent.round()}%',
                      subText: disk == null
                          ? '无分区数据'
                          : '${_fmtBytes(disk.used)} / ${_fmtBytes(disk.total)}',
                      fraction: (disk?.usedPercent ?? 0) / 100,
                      color: Color.lerp(scheme.primary, scheme.tertiary, 0.55)!,
                    ),
                  ),
                ],
              ),
              const Divider(height: 28),
              _infoRow(context, '系统版本', data.system.os),
              _infoRow(context, '核心数', cores == 0 ? '—' : cores.toString()),
              _infoRow(context, '运行时间', _fmtUptime(data.system.uptime)),
              _infoRow(context, '主机', _selected?.host ?? data.system.hostname),
              _infoRow(
                context,
                '网络',
                '↓ ${_fmtRate(data.network.rxRate)}  ↑ ${_fmtRate(data.network.txRate)}',
              ),
              _infoRow(context, '用户', data.system.username),
              _infoRow(context, '内核', data.system.kernel),
              _infoRow(context, '进程数', data.system.processes.toString()),
            ],
          ),
        ),
      ),
      const SizedBox(height: 12),
      if (data.cpu.perCore.isNotEmpty)
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('每核 CPU', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 12),
                for (var i = 0; i < data.cpu.perCore.length; i++)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 3),
                    child: Row(
                      children: [
                        SizedBox(width: 52, child: Text('核心 $i')),
                        Expanded(
                          child: LinearProgressIndicator(
                            value: (data.cpu.perCore[i] / 100).clamp(0, 1),
                          ),
                        ),
                        const SizedBox(width: 8),
                        SizedBox(
                          width: 56,
                          child: Text(
                            '${data.cpu.perCore[i].toStringAsFixed(1)}%',
                            textAlign: TextAlign.right,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        ),
      if (data.partitions.isNotEmpty) ...[
        const SizedBox(height: 12),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('磁盘分区', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                for (final p in data.partitions)
                  ListTile(
                    dense: true,
                    contentPadding: EdgeInsets.zero,
                    title: Text(p.mountPoint),
                    subtitle: LinearProgressIndicator(
                      value: (p.usedPercent / 100).clamp(0, 1),
                    ),
                    trailing: Text(
                      '${_fmtBytes(p.used)} / ${_fmtBytes(p.total)} · ${p.usedPercent.toStringAsFixed(0)}%',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ),
              ],
            ),
          ),
        ),
      ],
      const SizedBox(height: 12),
      Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('内存与网络明细', style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              _infoRow(context, '可用内存', _fmtBytes(data.memory.available)),
              _infoRow(context, '空闲内存', _fmtBytes(data.memory.free)),
              _infoRow(
                context,
                'Swap',
                '${_fmtBytes(data.memory.swapUsed)} / ${_fmtBytes(data.memory.swapTotal)}',
              ),
              _infoRow(context, '累计下行', _fmtBytes(data.network.totalRxBytes)),
              _infoRow(context, '累计上行', _fmtBytes(data.network.totalTxBytes)),
            ],
          ),
        ),
      ),
    ];
  }

  Widget _infoRow(BuildContext context, String label, String value) {
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
          Expanded(child: Text(value.isEmpty ? '—' : value)),
        ],
      ),
    );
  }

  String _fmtBytes(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    if (bytes < 1024 * 1024 * 1024) {
      return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
    }
    return '${(bytes / (1024 * 1024 * 1024)).toStringAsFixed(1)} GB';
  }

  /// 把后端给的英文 uptime 原文（如 "up 32 weeks, 6 days, 15 hours, 21 minutes"）
  /// 压成紧凑中文："230天 15小时" / "15小时 21分" / "21分"。解析不了时原样返回。
  String _fmtUptime(String raw) {
    final text = raw.trim();
    if (text.isEmpty) return text;
    int pick(String unit) {
      final m = RegExp('(\\d+)\\s*$unit').firstMatch(text);
      return m == null ? 0 : int.parse(m.group(1)!);
    }

    final days = pick('weeks?') * 7 + pick('days?');
    final hours = pick('hours?');
    final minutes = pick('minutes?');
    if (days == 0 && hours == 0 && minutes == 0) {
      return text;
    }
    if (days > 0) return '$days天 $hours小时';
    if (hours > 0) return '$hours小时 $minutes分';
    return '$minutes分';
  }

  String _fmtRate(double rate) {
    if (rate < 1024) return '${rate.toStringAsFixed(0)} B/s';
    if (rate < 1024 * 1024) return '${(rate / 1024).toStringAsFixed(1)} KB/s';
    return '${(rate / (1024 * 1024)).toStringAsFixed(1)} MB/s';
  }
}
