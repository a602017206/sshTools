/**
 * 多资产状态墙的数据合并逻辑（纯函数，便于单测）。
 *
 * 输入：资产列表、会话 Map 的值数组、探活结果 Map（id -> probe）、
 * 指标 Map（sessionId -> MonitoringData）。输出：可直接渲染的卡片数组。
 */

export const ASSET_STATE_CONNECTED = 'connected';
export const ASSET_STATE_ONLINE = 'online';
export const ASSET_STATE_OFFLINE = 'offline';
export const ASSET_STATE_UNKNOWN = 'unknown';

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

    let state = ASSET_STATE_UNKNOWN;
    if (session) state = ASSET_STATE_CONNECTED;
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
  const summary = { total: 0, connected: 0, online: 0, offline: 0, unknown: 0 };
  for (const card of cards || []) {
    summary.total += 1;
    if (card.state === ASSET_STATE_CONNECTED) summary.connected += 1;
    else if (card.state === ASSET_STATE_ONLINE) summary.online += 1;
    else if (card.state === ASSET_STATE_OFFLINE) summary.offline += 1;
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
    default:
      return '未探测';
  }
}
