import assert from 'node:assert/strict';
import test from 'node:test';
import { isConnectionLostError } from '../src/lib/databaseConnectionError.js';

test('识别各数据库空闲断开后的报错', () => {
  assert.equal(isConnectionLostError('[DB_CONNECT_FAILED] rpc error: code = Unknown desc = This _connection has been closed.'), true);
  assert.equal(isConnectionLostError('[CONNECTION_LOST] 数据库连接已断开，自动重连失败: Connection refused'), true);
  assert.equal(isConnectionLostError(new Error('Communications link failure')), true);
  assert.equal(isConnectionLostError('ORA-17008: 已关闭连接'), true);
});

test('普通 SQL 错误和空值不视为连接断开', () => {
  assert.equal(isConnectionLostError('[QUERY_FAILED] syntax error at or near "selec"'), false);
  assert.equal(isConnectionLostError(''), false);
  assert.equal(isConnectionLostError(null), false);
});
