import test from 'node:test';
import assert from 'node:assert/strict';
import { shouldApplyTransferProgress } from '../src/lib/transferProgress.js';

test('upload and download progress both apply', () => {
  assert.equal(shouldApplyTransferProgress('upload'), true);
  assert.equal(shouldApplyTransferProgress('download'), true);
});

test('unknown kind does not apply', () => {
  assert.equal(shouldApplyTransferProgress('other'), false);
});
