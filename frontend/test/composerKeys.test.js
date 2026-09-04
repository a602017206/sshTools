import assert from 'node:assert/strict';
import test from 'node:test';

import { isCopilotCancelError, shouldSubmitComposerOnEnter } from '../src/lib/composerKeys.js';

test('Enter 在 IME 组字中不触发发送', () => {
  assert.equal(shouldSubmitComposerOnEnter({ key: 'Enter', shiftKey: false, isComposing: true }), false);
  assert.equal(shouldSubmitComposerOnEnter({ key: 'Enter', shiftKey: false, keyCode: 229 }), false);
});

test('Enter 在非组字时触发发送，Shift+Enter 换行', () => {
  assert.equal(shouldSubmitComposerOnEnter({ key: 'Enter', shiftKey: false, isComposing: false }), true);
  assert.equal(shouldSubmitComposerOnEnter({ key: 'Enter', shiftKey: true, isComposing: false }), false);
  assert.equal(shouldSubmitComposerOnEnter({ key: 'a', shiftKey: false }), false);
});

test('识别取消生成错误', () => {
  assert.equal(isCopilotCancelError({ message: 'context canceled' }), true);
  assert.equal(isCopilotCancelError('aborted'), true);
  assert.equal(isCopilotCancelError('network error'), false);
});
