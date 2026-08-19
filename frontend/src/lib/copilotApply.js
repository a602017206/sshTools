export const COPILOT_APPLY_SQL = 'copilot:apply-sql';
export const COPILOT_EXECUTE_SQL = 'copilot:execute-sql';
export const COPILOT_PEEK_SQL = 'copilot:peek-sql';

export function applySqlEvent(sessionId, content) {
  return new CustomEvent(COPILOT_APPLY_SQL, { detail: { sessionId, content } });
}

export function executeSqlEvent(sessionId, handled) {
  return new CustomEvent(COPILOT_EXECUTE_SQL, { detail: { sessionId, handled } });
}

export function peekSqlEvent(sessionId, out) {
  return new CustomEvent(COPILOT_PEEK_SQL, { detail: { sessionId, out } });
}

export function shellExecutePayload(content) {
  const text = String(content || '').replace(/\n+$/, '');
  return `${text}\n`;
}

// Decides whether the copilot "execute" flow should delegate to an open query/table
// panel. The panel path only applies when a panel claimed the peek AND its editor
// has a non-empty query; an empty editor must fall through to the no-panel path so
// we actually run the artifact SQL instead of silently no-oping and reporting success.
export function shouldUsePanelPath(peek) {
  return Boolean(peek?.found && String(peek?.query || '').trim());
}
