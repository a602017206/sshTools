/**
 * 多资产状态墙的数据合并逻辑（纯函数，便于单测）。
 *
 * 输入：资产列表、会话 Map 的值数组、探活结果 Map（id -> probe）、
 * 指标 Map（sessionId -> MonitoringData）。输出：可直接渲染的卡片数组。
 */

/** 看板可显示的会话参数目录（与后端白名单保持一致）。 */
export const WALL_METRIC_CATALOG = [
  { id: 'cpu', label: 'CPU' },
  { id: 'memory', label: '内存' },
  { id: 'disk', label: '磁盘' },
  { id: 'network', label: '网络' },
  { id: 'load', label: '负载' },
  { id: 'uptime', label: '运行时间' },
  { id: 'processes', label: '进程数' }
];

export const DEFAULT_WALL_METRICS = ['cpu', 'memory'];

export function parseWallMetrics(csv) {
  const valid = new Set(WALL_METRIC_CATALOG.map((item) => item.id));
  const picked = String(csv || '')
    .split(',')
    .map((part) => part.trim())
    .filter((id) => valid.has(id));
  const unique = [...new Set(picked)];
  return unique.length ? unique : [...DEFAULT_WALL_METRICS];
}

export function formatRate(bytesPerSec) {
  const value = Number(bytesPerSec) || 0;
  if (value < 1024) return `${Math.round(value)} B/s`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB/s`;
  if (value < 1024 * 1024 * 1024) return `${(value / 1024 / 1024).toFixed(1)} MB/s`;
  return `${(value / 1024 / 1024 / 1024).toFixed(2)} GB/s`;
}

/** 把一台已连接会话的 MonitoringData 按所选参数拼成展示行。 */
export function buildMetricRows(metricIds, data) {
  if (!data) return [];
  const rows = [];
  for (const id of metricIds || []) {
    if (id === 'cpu') {
      const percent = roundPercent(data.cpu?.overall);
      if (percent !== null) rows.push({ id, label: 'CPU', percent, text: `${percent}%` });
    } else if (id === 'memory') {
      const percent = roundPercent(data.memory?.used_percent);
      if (percent !== null) rows.push({ id, label: '内存', percent, text: `${percent}%` });
    } else if (id === 'disk') {
      const partitions = data.disk?.partitions || [];
      if (partitions.length) {
        const busiest = partitions.reduce((a, b) => ((b.used_percent || 0) > (a.used_percent || 0) ? b : a));
        const percent = roundPercent(busiest.used_percent);
        rows.push({ id, label: '磁盘', percent, text: `${percent}%`, hint: busiest.mount_point || '' });
      }
    } else if (id === 'network') {
      rows.push({ id, label: '网络', percent: null, text: `↓${formatRate(data.network?.rx_rate)} ↑${formatRate(data.network?.tx_rate)}` });
    } else if (id === 'load') {
      const load = data.cpu?.load_average?.[0];
      if (Number.isFinite(load)) rows.push({ id, label: '负载', percent: null, text: load.toFixed(2) });
    } else if (id === 'uptime') {
      if (data.system?.uptime) rows.push({ id, label: '运行时间', percent: null, text: String(data.system.uptime) });
    } else if (id === 'processes') {
      const count = data.system?.processes;
      if (Number.isFinite(count)) rows.push({ id, label: '进程数', percent: null, text: String(count) });
    }
  }
  return rows;
}

export const ASSET_STATE_CONNECTED = 'connected';
export const ASSET_STATE_ONLINE = 'online';
export const ASSET_STATE_OFFLINE = 'offline';
export const ASSET_STATE_UNKNOWN = 'unknown';
export const ASSET_STATE_EXCLUDED = 'excluded';

function findConnectedSession(asset, sessions) {
  if (!asset) return null;
  if (asset.type === 'database') {
    if (!asset.dbConnected) return null;
    return (
      sessions.find(
        (session) =>
          session?.type === 'database' &&
          session?.connected &&
          (session.sessionId === asset.dbSessionId || session?.connection?.id === asset.id)
      ) || { sessionId: asset.dbSessionId, type: 'database', connected: true }
    );
  }
  return (
    sessions.find(
      (session) => session?.type !== 'database' && session?.connected && session?.connection?.id === asset.id
    ) || null
  );
}

/**
 * @param {Array} assets assetsStore 中的资产
 * @param {Array} sessions connectionsStore 的会话值数组
 * @param {Map<string, object>} probes id -> AssetProbeResult
 * @param {Map<string, object>} metrics sessionId -> MonitoringData
 */
export function buildAssetCards(assets, sessions, probes, metrics) {
  const sessionList = Array.isArray(sessions) ? sessions : [];
  return (assets || []).map((asset) => {
    const session = findConnectedSession(asset, sessionList);
    const probe = probes?.get?.(asset.id) || null;
    const metric = session ? metrics?.get?.(session.sessionId) || null : null;
    const probeExcluded = asset.metadata?.status_wall_excluded === 'true';

    let state = ASSET_STATE_UNKNOWN;
    if (session) state = ASSET_STATE_CONNECTED;
    else if (probeExcluded) state = ASSET_STATE_EXCLUDED;
    else if (probe) state = probe.reachable ? ASSET_STATE_ONLINE : ASSET_STATE_OFFLINE;

    return {
      id: asset.id,
      name: asset.name || asset.host || '未命名资产',
      host: asset.host || '',
      port: asset.port || 0,
      username: asset.username || '',
      group: asset.group || '默认分组',
      type: asset.type === 'database' ? 'database' : 'ssh',
      dbType: asset.metadata?.db_type || asset.dbType || '',
      asset,
      state,
      probeExcluded,
      onBoard: asset.metadata?.wall_on_board === 'true',
      wallMetrics: parseWallMetrics(asset.metadata?.wall_metrics),
      metricRows: buildMetricRows(parseWallMetrics(asset.metadata?.wall_metrics), metric),
      connected: Boolean(session),
      sessionId: session?.sessionId || null,
      latencyMs: probe?.reachable ? probe.latency_ms : null,
      probeError: probe && !probe.reachable ? probe.error || '' : '',
      cpuPercent: metric ? roundPercent(metric.cpu?.overall) : null,
      memPercent: metric ? roundPercent(metric.memory?.used_percent) : null
    };
  });
}

function roundPercent(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) return null;
  return Math.round(num * 10) / 10;
}

export function summarizeCards(cards) {
  const summary = { total: 0, connected: 0, online: 0, offline: 0, unknown: 0, excluded: 0 };
  for (const card of cards || []) {
    summary.total += 1;
    if (card.state === ASSET_STATE_CONNECTED) summary.connected += 1;
    else if (card.state === ASSET_STATE_ONLINE) summary.online += 1;
    else if (card.state === ASSET_STATE_OFFLINE) summary.offline += 1;
    else if (card.state === ASSET_STATE_EXCLUDED) summary.excluded += 1;
    else summary.unknown += 1;
  }
  // 「可用」口径：已连接 + 探活在线
  summary.available = summary.connected + summary.online;
  return summary;
}

export function groupCards(cards) {
  const groups = [];
  const byName = new Map();
  for (const card of cards || []) {
    const name = card.group || '默认分组';
    if (!byName.has(name)) {
      const group = { name, cards: [] };
      byName.set(name, group);
      groups.push(group);
    }
    byName.get(name).cards.push(card);
  }
  return groups;
}

export function stateLabel(state) {
  switch (state) {
    case ASSET_STATE_CONNECTED:
      return '已连接';
    case ASSET_STATE_ONLINE:
      return '在线';
    case ASSET_STATE_OFFLINE:
      return '离线';
    case ASSET_STATE_EXCLUDED:
      return '已关闭探活';
    default:
      return '未探测';
  }
}
