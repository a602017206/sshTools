import assert from 'node:assert/strict';
import test from 'node:test';

import { shouldConnectAsset } from '../src/lib/assetActivation.js';

test('连接行仅在双击或键盘 Enter 时连接', () => {
  assert.equal(shouldConnectAsset({ type: 'click' }), false);
  assert.equal(shouldConnectAsset({ type: 'dblclick' }), true);
  assert.equal(shouldConnectAsset({ type: 'keydown', key: 'Enter' }), true);
  assert.equal(shouldConnectAsset({ type: 'keydown', key: ' ' }), false);
});
