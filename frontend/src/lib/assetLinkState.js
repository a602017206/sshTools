export const ASSET_LINK_STATE_LABELS = {
  online: '已连接',
  idle: '未连接',
  connecting: '连接中',
  error: '连接失败'
};

export function getAssetLinkState(asset, sessions = []) {
  if (!asset) return 'idle';

  const relatedSessions = (Array.isArray(sessions) ? sessions : []).filter(
    (session) => session?.connection?.id === asset.id
  );

  if (relatedSessions.some((session) => session && session.connected === false && !session.panelType)) {
    return 'connecting';
  }

  if (asset.type === 'database') {
    if (asset.dbConnected || relatedSessions.some((session) => session?.connected)) {
      return 'online';
    }
    if (asset.status === 'error') {
      return 'error';
    }
    return 'idle';
  }

  if (relatedSessions.some((session) => session?.connected)) {
    return 'online';
  }
  if (asset.status === 'error') {
    return 'error';
  }
  return 'idle';
}

export function assetLinkStateLabel(state) {
  return ASSET_LINK_STATE_LABELS[state] || ASSET_LINK_STATE_LABELS.idle;
}
