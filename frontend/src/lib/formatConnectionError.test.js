import assert from 'node:assert/strict';
import test from 'node:test';

import { formatConnectionError } from './formatConnectionError.js';

test('prefers Error.message for UI', () => {
  assert.equal(
    formatConnectionError(new Error('authentication failed')),
    'authentication failed'
  );
});

test('keeps plain string errors', () => {
  assert.equal(formatConnectionError('主机不可达'), '主机不可达');
});

test('falls back for unknown values', () => {
  assert.equal(formatConnectionError(null), '连接失败，请检查主机、端口与凭据');
  assert.equal(formatConnectionError({}), '连接失败，请检查主机、端口与凭据');
});

test('supports custom fallback', () => {
  assert.equal(formatConnectionError(null, '无法连接数据库'), '无法连接数据库');
});
