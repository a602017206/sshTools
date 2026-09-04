export function sessionIdsToClose(orderedIds, targetId, action) {
  const ids = Array.isArray(orderedIds) ? orderedIds : [];
  const index = ids.indexOf(targetId);
  if (index < 0) return [];

  switch (action) {
    case 'all':
      return [...ids];
    case 'left':
      return ids.slice(0, index);
    case 'right':
      return ids.slice(index + 1);
    case 'others':
      return ids.filter((id) => id !== targetId);
    default:
      return [];
  }
}

export function sessionTabCloseMenuFlags(orderedIds, targetId) {
  const ids = Array.isArray(orderedIds) ? orderedIds : [];
  const index = ids.indexOf(targetId);
  return {
    canCloseLeft: index > 0,
    canCloseRight: index >= 0 && index < ids.length - 1,
    canCloseOthers: ids.length > 1 && index >= 0,
    canCloseAll: ids.length > 0 && index >= 0,
  };
}

export function batchCloseNeedsConfirm(sessions, idsToClose) {
  const idSet = new Set(idsToClose || []);
  return (sessions || []).some(
    (session) => idSet.has(session?.sessionId) && session.type !== 'database' && session.connected
  );
}

export function batchCloseConfirmCopy(idsToClose) {
  const count = Array.isArray(idsToClose) ? idsToClose.length : 0;
  if (count > 1) {
    return {
      title: '批量关闭会话',
      message: `确定要关闭这 ${count} 个会话吗？`,
    };
  }
  return {
    title: '关闭 SSH 会话',
    message: '确定要关闭此 SSH 会话吗？',
  };
}
