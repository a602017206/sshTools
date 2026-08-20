import test from 'node:test';
import assert from 'node:assert/strict';
import { appendTerminalTail, getTerminalTail } from '../src/lib/copilotContext.js';

test('terminal tails are isolated by session and capped to the latest text', () => {
  let tails = {};
  tails = appendTerminalTail(tails, 'ssh-a', 'first-', 10);
  tails = appendTerminalTail(tails, 'ssh-a', 'second', 10);
  tails = appendTerminalTail(tails, 'ssh-b', 'other', 10);

  assert.equal(getTerminalTail(tails, 'ssh-a'), 'rst-second');
  assert.equal(getTerminalTail(tails, 'ssh-b'), 'other');
});

test('terminal tails ignore missing session IDs and empty output', () => {
  const tails = { existing: 'value' };
  assert.deepEqual(appendTerminalTail(tails, '', 'ignored'), tails);
  assert.deepEqual(appendTerminalTail(tails, 'existing', ''), tails);
});
