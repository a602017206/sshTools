export function databaseSchemaMenuItems() {
  return [
    { id: 'new-query', label: '新建查询' },
    { id: 'run-sql-file', label: '运行 SQL 文件…' },
    { id: 'refresh', label: '刷新' },
    { id: 'disconnect', label: '断开', danger: true }
  ];
}

export function sqlFileProgressFromEvent(payload) {
  if (!payload || typeof payload !== 'object') return null;
  const fileSize = Number(payload.fileSize ?? payload.FileSize ?? 0) || 0;
  const bytesRead = Number(payload.bytesRead ?? payload.BytesRead ?? 0) || 0;
  const done = Boolean(payload.done ?? payload.Done);
  return {
    sessionId: String(payload.sessionId ?? payload.SessionID ?? ''),
    fileName: String(payload.fileName ?? payload.FileName ?? ''),
    fileSize,
    bytesRead,
    statements: Number(payload.statements ?? payload.Statements ?? 0) || 0,
    affected: Number(payload.affected ?? payload.Affected ?? 0) || 0,
    done,
    canceled: Boolean(payload.canceled ?? payload.Canceled),
    error: String(payload.error ?? payload.Error ?? ''),
    failedSql: String(payload.failedSql ?? payload.FailedSQL ?? ''),
    percent: fileSize > 0 ? Math.min(100, Math.round((bytesRead / fileSize) * 100)) : (done ? 100 : 0)
  };
}
