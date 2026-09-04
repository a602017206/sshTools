/** 是否应对远端发起性能监控轮询。 */
export function shouldPollMonitor({ liveEnabled, canUseMonitor, panelVisible = true } = {}) {
  return Boolean(liveEnabled && canUseMonitor && panelVisible);
}

/** 切换某会话的实时监控开关，返回新的 Map。 */
export function setSessionLiveEnabled(enabledBySession, sessionId, enabled) {
  if (!sessionId) return new Map(enabledBySession || []);
  const next = new Map(enabledBySession || []);
  next.set(sessionId, Boolean(enabled));
  return next;
}

/** 读取某会话是否开启实时监控；未设置时默认关闭。 */
export function isSessionLiveEnabled(enabledBySession, sessionId) {
  if (!sessionId) return false;
  return Boolean(enabledBySession?.get?.(sessionId));
}
