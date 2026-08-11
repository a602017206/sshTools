export function formatConnectionError(error, fallback = '连接失败，请检查主机、端口与凭据') {
  if (error == null || error === '') {
    return fallback;
  }

  if (typeof error === 'string') {
    const trimmed = error.trim();
    return trimmed || fallback;
  }

  if (typeof error?.message === 'string' && error.message.trim()) {
    return error.message.trim();
  }

  const asString = String(error).trim();
  if (!asString || asString === '[object Object]') {
    return fallback;
  }

  return asString;
}
