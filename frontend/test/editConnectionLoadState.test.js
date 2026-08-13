import assert from 'node:assert/strict';
import test from 'node:test';

import { shouldApplyEditConnectionResult, shouldLoadEditConnection } from '../src/lib/editConnectionLoadState.js';

test('仅将仍处于打开状态且编辑目标未改变的异步结果写入表单', () => {
  assert.equal(shouldApplyEditConnectionResult({ isOpen: true, requestedId: 'A', activeId: 'A' }), true);
  assert.equal(shouldApplyEditConnectionResult({ isOpen: false, requestedId: 'A', activeId: 'A' }), false);
  assert.equal(shouldApplyEditConnectionResult({ isOpen: true, requestedId: 'A', activeId: 'B' }), false);
  assert.equal(shouldApplyEditConnectionResult({ isOpen: true, requestedId: 'A', activeId: null }), false);
});

test('切换到新编辑目标后可立即启动该目标的加载', () => {
  assert.equal(shouldLoadEditConnection({ isOpen: true, targetId: 'B', loaded: false }), true);
  assert.equal(shouldLoadEditConnection({ isOpen: false, targetId: 'B', loaded: false }), false);
  assert.equal(shouldLoadEditConnection({ isOpen: true, targetId: '', loaded: false }), false);
  assert.equal(shouldLoadEditConnection({ isOpen: true, targetId: 'B', loaded: true }), false);
});
