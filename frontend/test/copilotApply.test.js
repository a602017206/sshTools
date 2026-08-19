import test from 'node:test';
import assert from 'node:assert/strict';
import { shellExecutePayload, applySqlEvent, COPILOT_APPLY_SQL } from '../src/lib/copilotApply.js';

test('shell execute appends a single newline', () => {
  assert.equal(shellExecutePayload('ls -la\n'), 'ls -la\n');
});

test('apply sql event uses session id', () => {
  const event = applySqlEvent('db-1', 'SELECT 1');
  assert.equal(event.type, COPILOT_APPLY_SQL);
  assert.equal(event.detail.sessionId, 'db-1');
});
