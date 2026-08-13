export function canUploadDroppedFiles({ sessionId, connected, isLocal }) {
  return Boolean(sessionId && connected && !isLocal);
}

export function normalizeDroppedFilePaths(paths) {
  return [...new Set(
    (Array.isArray(paths) ? paths : [])
      .filter((path) => typeof path === 'string')
      .map((path) => path.trim())
      .filter(Boolean)
  )];
}
