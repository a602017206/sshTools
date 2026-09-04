import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import {
  batchCloseConfirmCopy,
  batchCloseNeedsConfirm,
  sessionIdsToClose,
  sessionTabCloseMenuFlags,
} from '../src/lib/sessionTabClose.js';

const ids = ['a', 'b', 'c', 'd'];

test('按当前标签计算全部关闭、关闭左侧、关闭右侧、关闭其它', () => {
  assert.deepEqual(sessionIdsToClose(ids, 'c', 'all'), ['a', 'b', 'c', 'd']);
  assert.deepEqual(sessionIdsToClose(ids, 'c', 'left'), ['a', 'b']);
  assert.deepEqual(sessionIdsToClose(ids, 'c', 'right'), ['d']);
  assert.deepEqual(sessionIdsToClose(ids, 'c', 'others'), ['a', 'b', 'd']);
  assert.deepEqual(sessionIdsToClose(ids, 'a', 'left'), []);
  assert.deepEqual(sessionIdsToClose(ids, 'd', 'right'), []);
  assert.deepEqual(sessionIdsToClose(ids, 'missing', 'all'), []);
});

test('首尾标签禁用关闭左侧或关闭右侧，仅一个标签时禁用关闭其它', () => {
  assert.deepEqual(sessionTabCloseMenuFlags(ids, 'a'), {
    canCloseLeft: false,
    canCloseRight: true,
    canCloseOthers: true,
    canCloseAll: true,
  });
  assert.deepEqual(sessionTabCloseMenuFlags(ids, 'd'), {
    canCloseLeft: true,
    canCloseRight: false,
    canCloseOthers: true,
    canCloseAll: true,
  });
  assert.deepEqual(sessionTabCloseMenuFlags(['only'], 'only'), {
    canCloseLeft: false,
    canCloseRight: false,
    canCloseOthers: false,
    canCloseAll: true,
  });
});

test('批量关闭含已连接 SSH 时需要一次确认，数据库标签不确认', () => {
  const sessions = [
    { sessionId: 'a', type: 'ssh', connected: true },
    { sessionId: 'b', type: 'database', connected: true },
    { sessionId: 'c', type: 'ssh', connected: false },
  ];
  assert.equal(batchCloseNeedsConfirm(sessions, ['a', 'b']), true);
  assert.equal(batchCloseNeedsConfirm(sessions, ['b', 'c']), false);
  assert.equal(batchCloseNeedsConfirm(sessions, ['b']), false);
  assert.deepEqual(batchCloseConfirmCopy(['a', 'b', 'c']), {
    title: '批量关闭会话',
    message: '确定要关闭这 3 个会话吗？',
  });
  assert.deepEqual(batchCloseConfirmCopy(['a']), {
    title: '关闭 SSH 会话',
    message: '确定要关闭此 SSH 会话吗？',
  });
});

test('标签栏提供右键批量关闭菜单', async () => {
  const panel = await readFile(new URL('../src/components/TerminalPanel.svelte', import.meta.url), 'utf8');
  assert.match(panel, /on:contextmenu=\{\(event\) => openTabContextMenu/);
  assert.match(panel, /全部关闭/);
  assert.match(panel, /关闭左侧/);
  assert.match(panel, /关闭右侧/);
  assert.match(panel, /关闭其它/);
  assert.match(panel, /sessionIdsToClose/);
});
