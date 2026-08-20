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
