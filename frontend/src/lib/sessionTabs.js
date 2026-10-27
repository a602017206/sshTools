/**
 * 终端页签的可分辨性与批量执行目标（纯函数，便于单测）。
 * 同一服务器开多个页签时：自动序号 + 用户@主机副标题 + 按会话稳定配色。
 */

const ACCENT_PALETTE = [
  '#f59e0b', '#06b6d4', '#a78bfa', '#f472b6',
  '#34d399', '#facc15', '#60a5fa', '#fb7185'
];

/** 按 sessionId 哈希取稳定配色：同一页签每次渲染颜色不变。 */
export function sessionAccentColor(sessionId) {
  const text = String(sessionId || '');
  let hash = 0;
  for (let i = 0; i < text.length; i += 1) {
    hash = (hash * 31 + text.charCodeAt(i)) >>> 0;
  }
  return ACCENT_PALETTE[hash % ACCENT_PALETTE.length];
}

function tabGroupKey(session) {
  if (session?.type === 'database') return `db:${session?.connection?.id || session?.sessionId || ''}`;
  if (session?.type === 'docker') return `docker:${session?.connection?.id || session?.sessionId || ''}`;
  if (session?.type === 'docker-exec') return `docker-exec:${session?.containerId || session?.sessionId || ''}`;
  if (session?.type === 'local') return `local:${session?.sessionId || ''}`;
  return `ssh:${session?.connection?.id || session?.sessionId || ''}`;
}

/** 返回 Map<sessionId, 序号(1 起)>：同一资产的多个页签按列表顺序编号。 */
export function decorateSessionTabs(sessions) {
  const counters = new Map();
  const ordinals = new Map();
  for (const session of sessions || []) {
    const key = tabGroupKey(session);
    const next = (counters.get(key) || 0) + 1;
    counters.set(key, next);
    ordinals.set(session.sessionId, next);
  }
  return ordinals;
}

/** Docker 容器短 ID（前 12 位）。 */
export function dockerShortId(containerId) {
  return String(containerId || '').trim().slice(0, 12);
}

/** 页签标题：自定义名优先；未改名且同资产有多个页签时自动带 #序号。 */
export function sessionTabTitle(session, ordinal = 1, groupSize = 1) {
  if (session?.type === 'docker-exec' && !session?.tabName) {
    const label = session?.containerName || dockerShortId(session?.containerId) || '容器';
    const base = `${label} · exec`;
    return groupSize > 1 ? `${base} #${ordinal}` : base;
  }
  const base = session?.tabName || session?.connection?.name || '会话';
  if (session?.tabName) return base;
  return groupSize > 1 ? `${base} #${ordinal}` : base;
}

/** 页签副标题：SSH 显示 用户@主机:端口，断开时标注。 */
export function sessionTabSubtitle(session) {
  if (session?.type === 'database') return '数据库';
  if (session?.type === 'docker') return session?.connected ? 'Docker' : 'Docker · 未连接';
  if (session?.type === 'docker-exec') return dockerShortId(session?.containerId) || 'Docker Exec';
  if (session?.type === 'local') return '本地终端';
  const conn = session?.connection || {};
  const identity = conn.host
    ? `${conn.username ? `${conn.username}@` : ''}${conn.host}${conn.port ? `:${conn.port}` : ''}`
    : '';
  if (!identity) return session?.connected ? '已连接' : '已断开';
  return session?.connected ? identity : `${identity} · 已断开`;
}

/** 可批量下发的目标：已连接的远程 SSH 会话（不含数据库与本地终端）。 */
export function isBatchableSession(session) {
  return Boolean(
    session?.connected &&
      session?.type !== 'database' &&
      session?.type !== 'docker' &&
      session?.type !== 'docker-exec' &&
      session?.type !== 'local' &&
      session?.connection?.host
  );
}

/** 主视口是否为控制台（终端）样式：无会话、SSH、本地、docker-exec 为 true。 */
export function viewportIsConsole(session) {
  if (!session) return true;
  if (session.type === 'docker-exec') return true;
  return session.type !== 'database' && session.type !== 'docker';
}

export function listBatchTargets(sessions) {
  const list = sessions || [];
  const ordinals = decorateSessionTabs(list);
  const sizes = new Map();
  for (const session of list) {
    const key = tabGroupKey(session);
    sizes.set(key, (sizes.get(key) || 0) + 1);
  }
  return list.filter(isBatchableSession).map((session) => ({
    sessionId: session.sessionId,
    title: sessionTabTitle(session, ordinals.get(session.sessionId), sizes.get(tabGroupKey(session))),
    subtitle: sessionTabSubtitle(session),
    color: sessionAccentColor(session.sessionId),
    connectionId: session?.connection?.id || ''
  }));
}

/** 把多行命令规整为终端可执行的字节序：行间与结尾按回车发送。 */
export function buildBatchPayload(command, appendEnter) {
  const lines = String(command || '').replace(/\r\n/g, '\n').replace(/\n+$/, '');
  if (!lines.trim()) return '';
  const body = lines.split('\n').join('\r');
  return appendEnter ? `${body}\r` : body;
}
