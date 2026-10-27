/** Docker Exec 终端页签会话对象构造（纯函数，便于单测）。 */

/**
 * 前端生成 exec 会话 ID（须与后端 ^dexec-[A-Za-z0-9_-]{6,64}$ 一致），
 * 以便在 StartDockerExec 之前订阅 docker:exec:{id} / docker:exec-closed:{id}。
 */
export function newDockerExecSessionId(cryptoImpl = globalThis.crypto) {
  let hex = '';
  if (typeof cryptoImpl?.randomUUID === 'function') {
    hex = cryptoImpl.randomUUID().replace(/-/g, '');
  } else if (typeof cryptoImpl?.getRandomValues === 'function') {
    const bytes = cryptoImpl.getRandomValues(new Uint8Array(12));
    hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
  } else {
    while (hex.length < 24) hex += Math.random().toString(16).slice(2);
  }
  return `dexec-${hex.slice(0, 24)}`;
}

/** 构造 docker-exec 页签会话；sessionId 即后端 execSessionId。 */
export function buildDockerExecTabSession({
  execSessionId,
  parentDockerSessionId,
  containerId,
  containerName,
  connection,
} = {}) {
  const name = String(containerName || '').replace(/^\//, '').trim();
  return {
    sessionId: execSessionId,
    type: 'docker-exec',
    connected: true,
    parentDockerSessionId: parentDockerSessionId || '',
    containerId: containerId || '',
    containerName: name,
    connection: connection ? { ...connection } : {},
  };
}
