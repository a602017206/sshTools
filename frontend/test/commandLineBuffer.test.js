import assert from 'node:assert/strict';
import test from 'node:test';

import { createCommandLineBuffer } from '../src/lib/commandLineBuffer.js';

test('enter submits line and clears', () => {
  const buf = createCommandLineBuffer();
  buf.push('cd /tmp');
  const { submitted } = buf.push('\r');
  assert.deepEqual(submitted, ['cd /tmp']);
  assert.equal(buf.getLine(), '');
});

test('backspace and multiline paste', () => {
  const buf = createCommandLineBuffer();
  buf.push('ab');
  buf.push('\x7f');
  assert.equal(buf.getLine(), 'a');
  const { submitted } = buf.push('one\ntwo\r');
  assert.deepEqual(submitted, ['aone', 'two']);
});

test('\\b 退格与 \\x7f 行为一致', () => {
  const buf = createCommandLineBuffer();
  buf.push('abc');
  buf.push('\b');
  assert.equal(buf.getLine(), 'ab');
});

test('Ctrl+C 清空当前行且不提交', () => {
  const buf = createCommandLineBuffer();
  buf.push('partial');
  const { submitted } = buf.push('\x03');
  assert.deepEqual(submitted, []);
  assert.equal(buf.getLine(), '');
});

test('flushPending 提交未回车的行', () => {
  const buf = createCommandLineBuffer();
  buf.push('ls -la');
  const { submitted } = buf.flushPending();
  assert.deepEqual(submitted, ['ls -la']);
  assert.equal(buf.getLine(), '');
});

test('flushPending 空行时不提交', () => {
  const buf = createCommandLineBuffer();
  const { submitted } = buf.flushPending();
  assert.deepEqual(submitted, []);
});

test('\\r\\n 只提交一次', () => {
  const buf = createCommandLineBuffer();
  buf.push('echo hi');
  const { submitted } = buf.push('\r\n');
  assert.deepEqual(submitted, ['echo hi']);
  assert.equal(buf.getLine(), '');
});
