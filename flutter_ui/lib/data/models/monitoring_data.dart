/// 远端系统监控数据（对应 Go 侧 ssh.MonitoringData 的 JSON）。
class MonitoringData {
  final int timestamp;
  final SystemInfo system;
  final CpuMetrics cpu;
  final MemoryMetrics memory;
  final NetworkMetrics network;
  final List<PartitionInfo> partitions;

  const MonitoringData({
    required this.timestamp,
    required this.system,
    required this.cpu,
    required this.memory,
    required this.network,
    required this.partitions,
  });

  factory MonitoringData.fromJson(Map<String, dynamic> json) {
    final disk = json['disk'] as Map<String, dynamic>? ?? const {};
    return MonitoringData(
      timestamp: (json['timestamp'] as num?)?.toInt() ?? 0,
      system: SystemInfo.fromJson(
        json['system'] as Map<String, dynamic>? ?? const {},
      ),
      cpu: CpuMetrics.fromJson(
        json['cpu'] as Map<String, dynamic>? ?? const {},
      ),
      memory: MemoryMetrics.fromJson(
        json['memory'] as Map<String, dynamic>? ?? const {},
      ),
      network: NetworkMetrics.fromJson(
        json['network'] as Map<String, dynamic>? ?? const {},
      ),
      partitions: ((disk['partitions'] as List<dynamic>?) ?? const [])
          .map((e) => PartitionInfo.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  /// 磁盘总览取根分区（无根分区时取容量最大的一个）。
  PartitionInfo? get primaryPartition {
    if (partitions.isEmpty) return null;
    for (final p in partitions) {
      if (p.mountPoint == '/') return p;
    }
    return partitions.reduce((a, b) => a.total >= b.total ? a : b);
  }
}

class SystemInfo {
  final String hostname;
  final String uptime;
  final String os;
  final String kernel;
  final String username;
  final int processes;

  const SystemInfo({
    required this.hostname,
    required this.uptime,
    required this.os,
    required this.kernel,
    required this.username,
    required this.processes,
  });

  factory SystemInfo.fromJson(Map<String, dynamic> json) {
    return SystemInfo(
      hostname: json['hostname'] as String? ?? '',
      uptime: json['uptime'] as String? ?? '',
      os: json['os'] as String? ?? '',
      kernel: json['kernel'] as String? ?? '',
      username: json['username'] as String? ?? '',
      processes: (json['processes'] as num?)?.toInt() ?? 0,
    );
  }
}

class CpuMetrics {
  final double overall;
  final double user;
  final double system;
  final double ioWait;
  final List<double> perCore;
  final List<double> loadAverage; // 1 / 5 / 15 分钟

  const CpuMetrics({
    required this.overall,
    required this.user,
    required this.system,
    required this.ioWait,
    required this.perCore,
    required this.loadAverage,
  });

  factory CpuMetrics.fromJson(Map<String, dynamic> json) {
    List<double> nums(String key) =>
        ((json[key] as List<dynamic>?) ?? const [])
            .map((e) => (e as num).toDouble())
            .toList();
    double val(String key) => ((json[key] as num?) ?? 0).toDouble();
    return CpuMetrics(
      overall: val('overall'),
      user: val('user'),
      system: val('system'),
      ioWait: val('iowait'),
      perCore: nums('per_core'),
      loadAverage: nums('load_average'),
    );
  }
}

class MemoryMetrics {
  final int total;
  final int used;
  final int free;
  final int available;
  final double usedPercent;
  final int swapTotal;
  final int swapUsed;

  const MemoryMetrics({
    required this.total,
    required this.used,
    required this.free,
    required this.available,
    required this.usedPercent,
    required this.swapTotal,
    required this.swapUsed,
  });

  factory MemoryMetrics.fromJson(Map<String, dynamic> json) {
    int val(String key) => ((json[key] as num?) ?? 0).toInt();
    return MemoryMetrics(
      total: val('total'),
      used: val('used'),
      free: val('free'),
      available: val('available'),
      usedPercent: ((json['used_percent'] as num?) ?? 0).toDouble(),
      swapTotal: val('swap_total'),
      swapUsed: val('swap_used'),
    );
  }
}

class NetworkMetrics {
  final int totalRxBytes;
  final int totalTxBytes;
  final double rxRate;
  final double txRate;

  const NetworkMetrics({
    required this.totalRxBytes,
    required this.totalTxBytes,
    required this.rxRate,
    required this.txRate,
  });

  factory NetworkMetrics.fromJson(Map<String, dynamic> json) {
    return NetworkMetrics(
      totalRxBytes: ((json['total_rx_bytes'] as num?) ?? 0).toInt(),
      totalTxBytes: ((json['total_tx_bytes'] as num?) ?? 0).toInt(),
      rxRate: ((json['rx_rate'] as num?) ?? 0).toDouble(),
      txRate: ((json['tx_rate'] as num?) ?? 0).toDouble(),
    );
  }
}

class PartitionInfo {
  final String mountPoint;
  final int total;
  final int used;
  final int free;
  final double usedPercent;

  const PartitionInfo({
    required this.mountPoint,
    required this.total,
    required this.used,
    required this.free,
    required this.usedPercent,
  });

  factory PartitionInfo.fromJson(Map<String, dynamic> json) {
    return PartitionInfo(
      mountPoint: json['mount_point'] as String? ?? '',
      total: ((json['total'] as num?) ?? 0).toInt(),
      used: ((json['used'] as num?) ?? 0).toInt(),
      free: ((json['free'] as num?) ?? 0).toInt(),
      usedPercent: ((json['used_percent'] as num?) ?? 0).toDouble(),
    );
  }
}
