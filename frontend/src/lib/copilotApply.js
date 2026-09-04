export const COPILOT_APPLY_SQL = 'copilot:apply-sql';
export const COPILOT_EXECUTE_SQL = 'copilot:execute-sql';
export const COPILOT_PEEK_SQL = 'copilot:peek-sql';
export const COPILOT_APPLY_NATIVE = 'copilot:apply-native';
export const COPILOT_EXECUTE_NATIVE = 'copilot:execute-native';

export function applySqlEvent(sessionId, content) {
  return new CustomEvent(COPILOT_APPLY_SQL, { detail: { sessionId, content } });
}

export function executeSqlEvent(sessionId, handled) {
  return new CustomEvent(COPILOT_EXECUTE_SQL, { detail: { sessionId, handled } });
}

export function peekSqlEvent(sessionId, out) {
  return new CustomEvent(COPILOT_PEEK_SQL, { detail: { sessionId, out } });
}

export function applyNativeEvent(sessionId, artifact) {
  return new CustomEvent(COPILOT_APPLY_NATIVE, { detail: { sessionId, artifact } });
}

export function executeNativeEvent(sessionId, artifact) {
  return new CustomEvent(COPILOT_EXECUTE_NATIVE, { detail: { sessionId, artifact } });
}

export function shellExecutePayload(content) {
  const text = String(content || '').replace(/\n+$/, '');
  return `${text}\n`;
}

export function shouldUsePanelPath(peek) {
  return Boolean(peek?.found && String(peek?.query || '').trim());
}

export function isNativeArtifact(artifact) {
  const type = String(artifact?.type || '').toLowerCase();
  return type === 'native_mutation' || type === 'native_query';
}
