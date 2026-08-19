import test from 'node:test';
import assert from 'node:assert/strict';
import {
  shellExecutePayload,
  applySqlEvent,
  executeSqlEvent,
  peekSqlEvent,
  shouldUsePanelPath,
  COPILOT_APPLY_SQL,
  COPILOT_EXECUTE_SQL,
  COPILOT_PEEK_SQL
} from '../src/lib/copilotApply.js';

test('shell execute appends a single newline', () => {
  assert.equal(shellExecutePayload('ls -la\n'), 'ls -la\n');
});

test('apply sql event uses session id', () => {
  const event = applySqlEvent('db-1', 'SELECT 1');
  assert.equal(event.type, COPILOT_APPLY_SQL);
  assert.equal(event.detail.sessionId, 'db-1');
});

test('execute sql event carries handled flag', () => {
  const handled = { value: false };
  const event = executeSqlEvent('db-2', handled);
  assert.equal(event.type, COPILOT_EXECUTE_SQL);
  assert.equal(event.detail.sessionId, 'db-2');
  assert.equal(event.detail.handled, handled);
});

test('peek sql event exposes out for synchronous claim', () => {
  const out = { found: false, query: '' };
  const event = peekSqlEvent('db-3', out);
  assert.equal(event.type, COPILOT_PEEK_SQL);
  assert.equal(event.detail.sessionId, 'db-3');
  assert.equal(event.detail.out, out);
  // 监听器可在 await 之前同步置位，调用方据此判断面板是否认领
  event.detail.out.found = true;
  event.detail.out.query = 'SELECT 2';
  assert.equal(out.found, true);
  assert.equal(out.query, 'SELECT 2');
});

test('shouldUsePanelPath requires both found and non-empty query', () => {
  assert.equal(shouldUsePanelPath({ found: false, query: '' }), false);
  assert.equal(shouldUsePanelPath({ found: true, query: '' }), false);
  assert.equal(shouldUsePanelPath({ found: true, query: '   ' }), false);
  assert.equal(shouldUsePanelPath({ found: false, query: 'SELECT 1' }), false);
  assert.equal(shouldUsePanelPath({ found: true, query: 'SELECT 1' }), true);
  // 空 peek / null 防御
  assert.equal(shouldUsePanelPath(null), false);
  assert.equal(shouldUsePanelPath(undefined), false);
  assert.equal(shouldUsePanelPath({}), false);
});

// 回归：peek.found 但 query 为空时不得走面板路径，否则面板 executeQuery 静默 return
// 却仍被报告为「已执行」。该用例锁定修复意图。
test('empty editor peek must not take panel path (no false success)', () => {
  const peek = { found: true, query: '' };
  assert.equal(shouldUsePanelPath(peek), false);
});
