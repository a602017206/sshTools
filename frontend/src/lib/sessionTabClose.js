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

export function batchCloseConfirmCopy(idsToClose, sessions) {
  const ids = Array.isArray(idsToClose) ? idsToClose : [];
  const count = ids.length;
  if (count > 1) {
    return {
      title: '批量关闭会话',
      message: `确定要关闭这 ${count} 个会话吗？`,
    };
  }
  const target = (sessions || []).find((session) => session?.sessionId === ids[0]);
  if (target?.type === 'docker-exec') {
    return {
      title: '关闭容器 Exec',
      message: '确定要关闭此容器 Exec 会话吗？',
    };
  }
  if (target?.type === 'docker') {
    return {
      title: '关闭 Docker 连接',
      message: '确定要关闭此 Docker 连接吗？',
    };
  }
  return {
    title: '关闭 SSH 会话',
    message: '确定要关闭此 SSH 会话吗？',
  };
}

/** 关闭会话时应调用的后端绑定名（供 TerminalPanel 分派）。 */
export function closeBindingForSession(session) {
  switch (session?.type) {
    case 'docker-exec':
      return 'CloseDockerExec';
    case 'docker':
      return 'CloseDocker';
    case 'database':
      return 'CloseDatabase';
    default:
      return 'CloseSSH';
  }
}

/** docker-exec 关闭路径标志：走 CloseDockerExec，且不关闭父 Docker 连接。 */
export function dockerExecCloseFlags(session) {
  const isExec = session?.type === 'docker-exec';
  return {
    isDockerExec: isExec,
    closeBinding: isExec ? 'CloseDockerExec' : '',
    closeParent: false,
  };
}
