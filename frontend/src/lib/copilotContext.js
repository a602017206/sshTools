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
