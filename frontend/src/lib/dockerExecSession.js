/** Docker Exec 终端页签会话对象构造（纯函数，便于单测）。 */

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
