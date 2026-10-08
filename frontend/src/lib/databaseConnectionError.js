const CONNECTION_LOST_PATTERNS = [
  'connection_lost',
  'connection has been closed',
  'connection is closed',
  'connection closed',
  'closed connection',
  'communications link failure',
  'broken pipe',
  'connection reset',
  'ora-17008',
  'ora-03113',
  '连接已关闭',
  '已关闭连接',
  '连接已断开'
];

export const CONNECTION_LOST_MESSAGE = '数据库连接已断开，可能是长时间空闲被服务器关闭';

export function isConnectionLostError(error) {
  const message = String(error?.message ?? error ?? '').toLowerCase();
  return CONNECTION_LOST_PATTERNS.some(pattern => message.includes(pattern));
}
