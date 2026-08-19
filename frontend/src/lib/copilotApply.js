export const COPILOT_APPLY_SQL = 'copilot:apply-sql';
export const COPILOT_EXECUTE_SQL = 'copilot:execute-sql';

export function applySqlEvent(sessionId, content) {
  return new CustomEvent(COPILOT_APPLY_SQL, { detail: { sessionId, content } });
}

export function executeSqlEvent(sessionId, handled) {
  return new CustomEvent(COPILOT_EXECUTE_SQL, { detail: { sessionId, handled } });
}

export function shellExecutePayload(content) {
  const text = String(content || '').replace(/\n+$/, '');
  return `${text}\n`;
}
