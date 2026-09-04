import { resolveAssetDomain } from './assetDomain.js';
import { isNativeDatabaseType } from './nativeDatabaseTypes.js';

export const DEFAULT_TERMINAL_TAIL_LIMIT = 6000;

export function appendTerminalTail(tails, sessionId, output, limit = DEFAULT_TERMINAL_TAIL_LIMIT) {
  if (!sessionId || !output) return tails || {};
  const current = String((tails || {})[sessionId] || '');
  const max = Math.max(1, Number(limit) || DEFAULT_TERMINAL_TAIL_LIMIT);
  const next = (current + String(output)).slice(-max);
  return { ...(tails || {}), [sessionId]: next };
}

export function getTerminalTail(tails, sessionId) {
  return String((tails || {})[sessionId] || '');
}

export function nativeCopilotObjectKind(databaseType) {
  const type = String(databaseType || '').toLowerCase();
  if (type === 'redis') return 'key';
  if (type === 'elasticsearch' || type === 'opensearch') return 'index';
  if (type === 'kafka' || type === 'rocketmq') return 'topic';
  if (type === 'rabbitmq') return 'queue';
  if (type === 'mongodb') return 'collection';
  if (type === 'cassandra') return 'table';
  return 'resource';
}

function isCopilotNativeType(databaseType) {
  const type = String(databaseType || '').toLowerCase();
  return isNativeDatabaseType(type) || type === 'opensearch';
}

export function resolveWorkspaceFocus(focusBySession, sessionId, backendSessionId) {
  const map = focusBySession || {};
  if (sessionId && map[sessionId]) return map[sessionId];
  if (backendSessionId && map[backendSessionId]) return map[backendSessionId];
  return null;
}

function isForcedSshContext({ mode = '', sessionType = '', panelType = '' } = {}) {
  if (String(mode || '').toLowerCase() === 'ssh') return true;
  if (sessionType === 'ssh' || sessionType === 'local') return true;
  if (panelType.startsWith('database-')) return false;
  return Boolean(sessionType && sessionType !== 'database');
}

export function buildCopilotWorkspaceContext({ session = null, navigation = null, focus = null, mode = '' } = {}) {
  const connection = session?.connection || {};
  const metadata = connection.metadata || {};
  const dbType = String(metadata.db_type || connection.dbType || session?.dbType || '').toLowerCase();
  const host = String(connection.host || '');
  const user = String(connection.user || connection.username || '');
  const panelType = String(session?.panelType || '');
  const editorContent = String(focus?.editorContent || '');
  const sessionType = String(session?.type || '');

  if (panelType === 'database-table' || panelType === 'database-table-designer') {
    return {
      host,
      user,
      dbType,
      database: String(session.databaseName || metadata.database || ''),
      schema: String(session.schemaName || ''),
      objectKind: 'table',
      objectName: String(session.tableName || ''),
      objectParent: '',
      editorContent,
      workspaceKind: 'jdbc'
    };
  }

  if (panelType === 'database-query') {
    return {
      host,
      user,
      dbType,
      database: String(session.databaseName || navigation?.databaseName || metadata.database || ''),
      schema: String(session.schemaName || navigation?.schemaName || ''),
      objectKind: 'query',
      objectName: '',
      objectParent: '',
      editorContent: editorContent || String(session.initialQuery || ''),
      workspaceKind: 'jdbc'
    };
  }

  // SSH / 本地终端：即使资产误存了 db_type（表单默认 mysql），也不当成数据库上下文。
  if (isForcedSshContext({ mode, sessionType, panelType })) {
    return {
      host,
      user,
      dbType: '',
      database: '',
      schema: '',
      objectKind: '',
      objectName: '',
      objectParent: '',
      editorContent: '',
      workspaceKind: 'ssh'
    };
  }

  if (isCopilotNativeType(dbType)) {
    return {
      host,
      user,
      dbType,
      database: String(navigation?.databaseName || focus?.database || metadata.database || ''),
      schema: '',
      objectKind: String(focus?.objectKind || nativeCopilotObjectKind(dbType)),
      objectName: String(focus?.objectName || ''),
      objectParent: String(focus?.objectParent || ''),
      editorContent,
      workspaceKind: 'native'
    };
  }

  if (sessionType === 'database' || panelType.startsWith('database-')) {
    return {
      host,
      user,
      dbType,
      database: String(navigation?.databaseName || focus?.database || metadata.database || ''),
      schema: String(navigation?.schemaName || focus?.schema || ''),
      objectKind: String(focus?.objectKind || (focus?.objectName ? 'table' : '')),
      objectName: String(focus?.objectName || ''),
      objectParent: '',
      editorContent,
      workspaceKind: 'jdbc'
    };
  }

  return {
    host,
    user,
    dbType: '',
    database: '',
    schema: '',
    objectKind: '',
    objectName: '',
    objectParent: '',
    editorContent,
    workspaceKind: 'ssh'
  };
}

export function formatCopilotWorkspaceLabel(context) {
  if (!context) return '';
  if (context.workspaceKind === 'ssh') {
    const host = String(context.host || '').trim();
    const user = String(context.user || '').trim();
    if (user && host) return `${user}@${host}`;
    return host || user || 'SSH';
  }
  const dbType = String(context.dbType || '').trim();
  const parent = String(context.objectParent || '').trim();
  const name = String(context.objectName || '').trim();
  if (parent && name && parent !== name) {
    return [dbType, `${parent} / ${name}`].filter(Boolean).join(' · ');
  }
  const path = [...new Set([context.database, context.schema, name].filter(Boolean))].join('.');
  return [dbType, path].filter(Boolean).join(' · ');
}

/** AI 面板标题：缓存/搜索/消息与 SQL 助手分开，避免一律叫 SQL。 */
export function copilotAssistantTitle(context, mode = 'ssh') {
  if (String(mode || '').toLowerCase() !== 'database') return 'Shell 助手';
  if (context?.workspaceKind === 'native' || isCopilotNativeType(context?.dbType)) {
    const domain = resolveAssetDomain({
      type: 'database',
      metadata: { db_type: context?.dbType }
    });
    if (domain === 'cache') return '缓存助手';
    if (domain === 'search') return '搜索助手';
    if (domain === 'mq') return '消息助手';
    return '数据助手';
  }
  return 'SQL 助手';
}

export function copilotChatPayload(context, { sessionID, mode, message, history, terminalTail, workingDir } = {}) {
  const ctx = context || {};
  return {
    SessionID: sessionID || '',
    Mode: mode || 'ssh',
    Message: message || '',
    History: Array.isArray(history) ? history : [],
    EditorContent: ctx.editorContent || '',
    TerminalTail: terminalTail || '',
    Host: ctx.host || '',
    User: ctx.user || '',
    DBType: ctx.dbType || '',
    Database: ctx.database || '',
    Schema: ctx.schema || '',
    ObjectKind: ctx.objectKind || '',
    ObjectName: ctx.objectName || '',
    ObjectParent: ctx.objectParent || '',
    WorkingDir: workingDir || ctx.workingDir || ''
  };
}

export function buildChatHistory(messages, { maxTurns = 12, maxChars = 12000 } = {}) {
  const relevant = (Array.isArray(messages) ? messages : [])
    .filter((item) => item && (item.role === 'user' || item.role === 'assistant') && item.content)
    .slice(-Math.max(1, Number(maxTurns) || 12));
  const history = [];
  let remaining = Math.max(1, Number(maxChars) || 12000);
  for (let index = relevant.length - 1; index >= 0 && remaining > 0; index--) {
    const item = relevant[index];
    const content = String(item.content).slice(-remaining);
    history.unshift({ Role: item.role, Content: content });
    remaining -= content.length;
  }
  return history;
}
