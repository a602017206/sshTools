export const TERMINAL_CHARSET_OPTIONS = [
  { id: 'utf-8', label: 'UTF-8' },
  { id: 'gbk', label: 'GBK' },
  { id: 'gb2312', label: 'GB2312' },
  { id: 'gb18030', label: 'GB18030' },
  { id: 'big5', label: 'Big5' }
];

const CHARSET_ALIASES = {
  utf8: 'utf-8',
  'utf-8': 'utf-8',
  gbk: 'gbk',
  cp936: 'gbk',
  gb2312: 'gb2312',
  gb18030: 'gb18030',
  big5: 'big5',
  'big5-hkscs': 'big5'
};

const DECODER_LABEL = {
  'utf-8': 'utf-8',
  gbk: 'gbk',
  gb2312: 'gb2312',
  gb18030: 'gb18030',
  big5: 'big5'
};

export function normalizeTerminalCharset(value) {
  const key = String(value || '').trim().toLowerCase().replace(/_/g, '-');
  return CHARSET_ALIASES[key] || 'utf-8';
}

export function decodeTerminalOutput(bytes, charset) {
  const octets = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes || []);
  const name = normalizeTerminalCharset(charset);
  if (name === 'utf-8') {
    return octets;
  }
  try {
    return new TextDecoder(DECODER_LABEL[name] || name, { fatal: false }).decode(octets);
  } catch {
    return octets;
  }
}

export function terminalContextMenuItems(hasSelection) {
  return [
    { id: 'copy', label: '复制', disabled: !hasSelection },
    { id: 'paste', label: '粘贴', disabled: false }
  ];
}

export function isSshSessionForConnection(session, connectionId) {
  if (!connectionId || !session) return false;
  if (session.type === 'database') return false;
  return session.connection?.id === connectionId;
}

export function withSessionCharset(session, encoding) {
  const charset = normalizeTerminalCharset(encoding);
  const connection = session?.connection || {};
  return {
    ...session,
    connection: {
      ...connection,
      encoding: charset,
      metadata: {
        ...(connection.metadata || {}),
        encoding: charset
      }
    }
  };
}

export function applyCharsetToSessionMap(sessionMap, connectionId, encoding) {
  const charset = normalizeTerminalCharset(encoding);
  const sessions = new Map();
  const sessionIds = [];
  for (const [id, session] of sessionMap || []) {
    if (isSshSessionForConnection(session, connectionId)) {
      sessions.set(id, withSessionCharset(session, charset));
      sessionIds.push(id);
    } else {
      sessions.set(id, session);
    }
  }
  return { sessions, sessionIds, charset };
}

export function applyCharsetToSessionId(sessionMap, sessionId, encoding) {
  const charset = normalizeTerminalCharset(encoding);
  const sessions = new Map(sessionMap || []);
  const session = sessions.get(sessionId);
  if (session && session.type !== 'database') {
    sessions.set(sessionId, withSessionCharset(session, charset));
  }
  return { sessions, charset };
}
