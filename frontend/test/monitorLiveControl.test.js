import assert from 'node:assert/strict';
import test from 'node:test';
import {
  isSessionLiveEnabled,
  setSessionLiveEnabled,
  shouldPollMonitor
} from '../src/lib/monitorLiveControl.js';

test('仅在开关开启、可用且面板可见时轮询', () => {
  assert.equal(shouldPollMonitor({ liveEnabled: true, canUseMonitor: true, panelVisible: true }), true);
  assert.equal(shouldPollMonitor({ liveEnabled: false, canUseMonitor: true, panelVisible: true }), false);
  assert.equal(shouldPollMonitor({ liveEnabled: true, canUseMonitor: false, panelVisible: true }), false);
  assert.equal(shouldPollMonitor({ liveEnabled: true, canUseMonitor: true, panelVisible: false }), false);
});

test('按会话记录实时开关，默认关闭', () => {
  assert.equal(isSessionLiveEnabled(new Map(), 's1'), false);
  const enabled = setSessionLiveEnabled(new Map(), 's1', true);
  assert.equal(isSessionLiveEnabled(enabled, 's1'), true);
  assert.equal(isSessionLiveEnabled(enabled, 's2'), false);
  const disabled = setSessionLiveEnabled(enabled, 's1', false);
  assert.equal(isSessionLiveEnabled(disabled, 's1'), false);
});
