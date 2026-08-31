import { isNativeDatabaseType } from './nativeDatabaseTypes.js';

export function resolveDatabaseSessionId(asset) {
  if (!asset) return '';
  return asset.dbSessionId || (asset.id ? `db-${asset.id}` : '');
}

/** 返回应关闭的会话 ID：父会话优先，其余为同连接子面板 */
export function listDatabaseSessionsToClose(sessions, asset) {
  const sessionId = resolveDatabaseSessionId(asset);
  if (!sessionId) return [];

  const list = Array.isArray(sessions) ? sessions : [];
  const related = list.filter(
    (session) =>
      session?.type === 'database' &&
      (session.sessionId === sessionId || session.dbSessionId === sessionId)
  );

  const parent = related.find(
    (session) =>
      session.sessionId === sessionId ||
      session.panelType === 'database-list' ||
      session.panelType === 'native-database'
  );

  if (parent?.sessionId) {
    return [parent.sessionId];
  }

  return related.map((session) => session.sessionId).filter(Boolean);
}

export function resolveDatabaseCloseBinding(asset, bindings = {}) {
  const dbType = asset?.metadata?.db_type || asset?.dbType || '';
  if (isNativeDatabaseType(dbType)) {
    return typeof bindings.CloseNativeDatabase === 'function' ? bindings.CloseNativeDatabase : null;
  }
  return typeof bindings.CloseDatabase === 'function' ? bindings.CloseDatabase : null;
}
